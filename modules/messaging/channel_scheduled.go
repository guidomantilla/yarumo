package messaging

import (
	"container/heap"
	"context"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// scheduled is the ScheduledChannel[T] implementation. Producers stage
// messages via Send (deliver immediately), SendAt (wall-clock deadline)
// or SendAfter (relative delay). A single scheduler goroutine owns an
// internal min-heap ordered by deliveryTime, sleeps until the head is
// due, then fans the message out to every subscriber synchronously
// (one dispatcher, many subscribers). Subscribers are tracked in a
// map[uint64]*scheduledSubscriber so Subscribe/Cancel are O(1) and
// Send-to-subscribers is a snapshot fan-out — exactly the shape used
// by TopicChannel for the per-Subscribe state, minus the per-subscriber
// inbox + worker (deferred delivery does not need per-sub backpressure
// because the queue lives on the producer side via the min-heap).
//
// scheduled implements both ScheduledChannel[T] and lifecycle.Component
// (worker-style). It uses the same sentinel-once + workerWG race-fix
// pattern as TopicChannel so Stop converges even when Start runs
// concurrently with Stop.
type scheduled[T any] struct {
	name         string
	drainTimeout time.Duration
	errorHandler ErrorHandler
	dlq          Channel[DeadLetter[T]]

	started      atomic.Bool
	closed       atomic.Bool
	workerCtx    context.Context
	workerCancel context.CancelFunc

	workerWG sync.WaitGroup
	done     chan struct{}

	// queueMu protects queue + wake so producers may enqueue while the
	// scheduler is parked. wake fires every time the head of the heap
	// may have changed (new earlier entry, Stop signal) so the
	// scheduler can re-arm its timer.
	queueMu sync.Mutex
	queue   scheduledHeap[T]
	wake    chan struct{}

	subsMu sync.RWMutex
	nextID uint64
	subs   map[uint64]Handler[T]

	startOnce    sync.Once
	stopOnce     sync.Once
	doneOnce     sync.Once
	sentinelOnce sync.Once
}

// scheduledItem is one entry on the scheduler's min-heap. The
// publisher's sendCtx is carried alongside the message so the
// dispatcher can merge it with the worker ctx the same way Topic
// and Queue do.
type scheduledItem[T any] struct {
	deliverAt time.Time
	sendCtx   context.Context
	msg       Message[T]
}

// scheduledHeap is a container/heap-compatible min-heap of
// scheduledItem ordered by deliverAt. Implemented as a slice so the
// heap helpers can mutate it in place without crossing the package
// boundary.
type scheduledHeap[T any] []scheduledItem[T]

// Len reports the heap size as required by heap.Interface.
func (h scheduledHeap[T]) Len() int { return len(h) }

// Less compares two heap entries by deliverAt so the head holds the
// earliest pending delivery.
func (h scheduledHeap[T]) Less(i, j int) bool { return h[i].deliverAt.Before(h[j].deliverAt) }

// Swap exchanges two heap entries in place as required by
// heap.Interface.
func (h scheduledHeap[T]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push appends x to the heap; container/heap restores the invariant
// after the call.
func (h *scheduledHeap[T]) Push(x any) {
	item, _ := x.(scheduledItem[T])
	*h = append(*h, item)
}

// Pop removes and returns the last entry; container/heap calls this
// after swapping the head with the tail.
func (h *scheduledHeap[T]) Pop() any {
	old := *h
	n := len(old)
	out := old[n-1]
	*h = old[:n-1]

	return out
}

// NewScheduledChannel constructs a ScheduledChannel[T] with the given
// name and options. name is used in lifecycle logs and must be
// non-empty. The returned channel is not running; call lifecycle.Build
// (or Start directly) to spawn the scheduler goroutine.
func NewScheduledChannel[T any](name string, opts ...Option) ScheduledChannel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &scheduled[T]{
		name:         name,
		drainTimeout: options.drainTimeout,
		errorHandler: options.errorHandler,
		dlq:          extractDLQ[T](options.dlq),
		done:         make(chan struct{}),
		wake:         make(chan struct{}, 1),
		subs:         map[uint64]Handler[T]{},
	}
}

// Name returns the channel's identity used in lifecycle logs.
func (c *scheduled[T]) Name() string {
	cassert.NotNil(c, "ScheduledChannel is nil")

	return c.name
}

// Start spawns the scheduler goroutine, which consumes from the
// internal min-heap and dispatches each message as its deliveryTime
// arrives. Start captures ctx as the worker ctx (used for Value
// propagation into handlers via mergeContexts) and is idempotent.
// Done is closed after Stop has joined the scheduler.
func (c *scheduled[T]) Start(ctx context.Context) error {
	cassert.NotNil(c, "ScheduledChannel is nil")

	c.startOnce.Do(func() {
		workerCtx, workerCancel := context.WithCancel(ctx)

		// Publishing workerCtx + flipping started under subsMu mirrors
		// the TopicChannel pattern: callers acquiring subsMu after Start
		// observe the consistent state.
		c.subsMu.Lock()
		c.workerCtx = workerCtx
		c.workerCancel = workerCancel
		c.started.Store(true)
		c.subsMu.Unlock()

		// Sentinel Add(1) keeps workerWG strictly positive between Start
		// and Stop so the awaitDrain Wait() does not race against a
		// late-spawning worker. Released via sentinelOnce by Stop on
		// the happy path or by Start itself when Stop ran first.
		c.workerWG.Add(1)

		c.workerWG.Go(func() { c.run(workerCtx) })

		go c.awaitDrain()

		if c.closed.Load() {
			c.sentinelOnce.Do(func() { c.workerWG.Done() })
		}
	})

	return nil
}

// awaitDrain closes done exactly once after every scheduler goroutine
// has exited (post-Stop).
func (c *scheduled[T]) awaitDrain() {
	c.workerWG.Wait()
	c.doneOnce.Do(func() { close(c.done) })
}

// Stop cancels the worker ctx, signals the scheduler to wake, and
// waits up to the configured drain timeout for the goroutine to
// exit. Stop is idempotent per the lifecycle.Component contract.
// Pending undelivered items are dropped — schedule semantics are
// best-effort; callers that need durability must persist intent
// outside the channel.
func (c *scheduled[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "ScheduledChannel is nil")

	c.stopOnce.Do(func() {
		c.closed.Store(true)

		c.subsMu.Lock()
		cancelFn := c.workerCancel
		startedNow := c.started.Load()
		c.subsMu.Unlock()

		if cancelFn != nil {
			cancelFn()
		}

		c.signalWake()

		if startedNow {
			c.sentinelOnce.Do(func() { c.workerWG.Done() })
		}
	})

	timeout := c.drainTimeout
	if timeout <= 0 {
		timeout = defaultDrainTimeout
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-c.done:
		return nil
	case <-waitCtx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, waitCtx.Err())
	}
}

// Done returns the channel that is closed after the scheduler
// goroutine has exited (post-Stop).
func (c *scheduled[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "ScheduledChannel is nil")

	return c.done
}

// Send delivers msg as soon as possible. Equivalent to SendAt with
// deliverAt = time.Now(). Returns ErrSend(ErrClosed) after Stop and
// ErrSend(ErrContextNil) when ctx is nil.
func (c *scheduled[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "ScheduledChannel is nil")

	return c.enqueue(ctx, time.Now(), msg)
}

// SendAt schedules msg for delivery at deliverAt (wall clock). A
// deliverAt in the past is treated as "deliver immediately". Returns
// ErrSend(ErrClosed) after Stop and ErrSend(ErrContextNil) when ctx
// is nil.
func (c *scheduled[T]) SendAt(ctx context.Context, deliverAt time.Time, msg Message[T]) error {
	cassert.NotNil(c, "ScheduledChannel is nil")

	return c.enqueue(ctx, deliverAt, msg)
}

// SendAfter schedules msg for delivery after delay has elapsed.
// Non-positive delays deliver immediately. Returns the same error set
// as SendAt.
func (c *scheduled[T]) SendAfter(ctx context.Context, delay time.Duration, msg Message[T]) error {
	cassert.NotNil(c, "ScheduledChannel is nil")

	return c.enqueue(ctx, time.Now().Add(delay), msg)
}

// Subscribe registers handler for dispatch. Subscribe is independent
// of Start: handlers registered before Start receive messages from
// the moment Start spawns the scheduler. Cancel is idempotent.
// Subscribe returns ErrSubscribe(ErrHandlerNil) when handler is nil
// and ErrSubscribe(ErrClosed) after Stop.
func (c *scheduled[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "ScheduledChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	c.subsMu.Lock()
	if c.closed.Load() {
		c.subsMu.Unlock()

		return nil, ErrSubscribe(ErrClosed)
	}

	c.nextID++
	id := c.nextID
	c.subs[id] = handler
	c.subsMu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.subsMu.Lock()
			delete(c.subs, id)
			c.subsMu.Unlock()
		})
	}

	return cancel, nil
}

// enqueue is the shared body of Send/SendAt/SendAfter. It validates
// ctx + closed state, pushes the item onto the heap and wakes the
// scheduler so it can re-arm its timer if the new head changed.
func (c *scheduled[T]) enqueue(ctx context.Context, deliverAt time.Time, msg Message[T]) error {
	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.closed.Load() {
		return ErrSend(ErrClosed)
	}

	c.queueMu.Lock()
	heap.Push(&c.queue, scheduledItem[T]{deliverAt: deliverAt, sendCtx: ctx, msg: msg})
	c.queueMu.Unlock()

	c.signalWake()

	return nil
}

// signalWake nudges the scheduler goroutine without blocking. Because
// wake is buffered with cap 1, a pending wake already covers the next
// re-arm; coalescing avoids stacking wake-ups across a burst of
// enqueues.
func (c *scheduled[T]) signalWake() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

// run is the scheduler goroutine. It loops:
//  1. Peek the head of the heap under queueMu.
//  2. If empty, wait for a wake signal or workerCtx cancellation.
//  3. If the head is due, pop it and dispatch.
//  4. Otherwise, sleep until the deliverAt, the next wake, or
//     workerCtx cancellation.
//
// Stop cancels workerCtx so the goroutine exits regardless of pending
// items; deliberately undrained items are dropped per the documented
// best-effort semantics.
func (c *scheduled[T]) run(workerCtx context.Context) {
	for {
		err := workerCtx.Err()
		if err != nil {
			return
		}

		now := time.Now()

		c.queueMu.Lock()
		ready, wait, hasItem := c.peekReady(now)
		c.queueMu.Unlock()

		if ready != nil {
			c.dispatch(workerCtx, *ready)

			continue
		}

		if !hasItem {
			select {
			case <-c.wake:
			case <-workerCtx.Done():
				return
			}

			continue
		}

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-c.wake:
			timer.Stop()
		case <-workerCtx.Done():
			timer.Stop()

			return
		}
	}
}

// peekReady returns a popped item when the head is due at now,
// otherwise the duration to wait before the next deliverAt and a
// hasItem flag distinguishing "empty heap" (sleep until wake) from
// "wait for the head". Must be called with queueMu held.
func (c *scheduled[T]) peekReady(now time.Time) (*scheduledItem[T], time.Duration, bool) {
	if c.queue.Len() == 0 {
		return nil, 0, false
	}

	head := c.queue[0]
	if !head.deliverAt.After(now) {
		popped, _ := heap.Pop(&c.queue).(scheduledItem[T])

		return &popped, 0, true
	}

	return nil, head.deliverAt.Sub(now), true
}

// dispatch fans the item out to every currently-subscribed handler.
// Per the same model as TopicChannel the worker ctx merges with the
// publisher's sendCtx so trace span / correlation id / slogctx attrs
// propagate while Done/Deadline follow the worker. Errors and
// recovered panics are routed through the configured ErrorHandler and
// (when configured) the DLQ.
func (c *scheduled[T]) dispatch(workerCtx context.Context, item scheduledItem[T]) {
	c.subsMu.RLock()
	snapshot := slices.Collect(maps.Values(c.subs))
	c.subsMu.RUnlock()

	if len(snapshot) == 0 {
		return
	}

	handlerCtx := mergeContexts(workerCtx, item.sendCtx)

	for _, handler := range snapshot {
		err := invokeHandler(handlerCtx, item.msg, handler)
		if err == nil {
			continue
		}

		if c.errorHandler != nil {
			c.errorHandler(handlerCtx, item.msg, err)
		}

		publishDeadLetter(handlerCtx, c.dlq, item.msg, err)
	}
}
