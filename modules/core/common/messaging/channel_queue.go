package messaging

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

// QueueChannel is the point-to-point async Channel[T] implementation:
// every message is delivered to EXACTLY ONE subscriber, selected by
// round-robin among the registered handlers. Multiple worker
// goroutines (see WithWorkerCount) consume from the inbound buffer in
// parallel, so a slow handler on one worker does not block the next
// worker from picking up the next message.
//
// QueueChannel implements both Channel[T] and lifecycle.Component
// (worker-style): Start spawns the configured worker pool; Stop
// closes the inbound buffer and waits for every worker to drain its
// in-flight message, bounded by the configured drain timeout.
//
// Use this primitive for work distribution: N equivalent subscribers
// share the load and each message is processed once. Contrast with
// TopicChannel where each message goes to every subscriber.
//
// Context propagation: the Send ctx travels with the message in the
// internal buffer but its cancellation does NOT abort the handler —
// async dispatch must outlive the publisher's request. The handler
// receives a ctx whose Done / Deadline / Err follow the worker's
// lifecycle ctx (the one passed to Start) and whose Value lookups
// fall through to the Send ctx so trace span, correlation id and
// slogctx attributes propagate from the publisher.
type queue[T any] struct {
	name           string
	bufferSize     int
	workerCount    int
	drainTimeout   time.Duration
	errorHandler   ErrorHandler
	overflowPolicy OverflowPolicy
	dlq            Channel[DeadLetter[T]]

	inbound chan envelope[T]
	done    chan struct{}
	closed  atomic.Bool

	mu       sync.RWMutex
	nextID   uint64
	order    []uint64
	byID     map[uint64]Handler[T]
	rotation uint64

	workerWG  sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once
}

// NewQueueChannel constructs a QueueChannel[T] with the given name
// and options. name is used in lifecycle logs and must be non-empty.
//
// The returned channel is not running; call lifecycle.Build to wire
// it into the application lifecycle and spawn the worker pool.
func NewQueueChannel[T any](name string, opts ...Option) Channel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &queue[T]{
		name:           name,
		bufferSize:     options.bufferSize,
		workerCount:    options.workerCount,
		drainTimeout:   options.drainTimeout,
		errorHandler:   options.errorHandler,
		overflowPolicy: options.overflowPolicy,
		dlq:            extractDLQ[T](options.dlq),
		inbound:        make(chan envelope[T], options.bufferSize),
		done:           make(chan struct{}),
		byID:           map[uint64]Handler[T]{},
	}
}

// Name returns the channel's identity used in lifecycle logs.
func (c *queue[T]) Name() string {
	cassert.NotNil(c, "QueueChannel is nil")

	return c.name
}

// Start spawns the worker pool. Each worker consumes from the
// inbound buffer and dispatches each message to one subscriber
// chosen by round-robin among the currently registered handlers.
// Start is idempotent.
func (c *queue[T]) Start(ctx context.Context) error {
	cassert.NotNil(c, "QueueChannel is nil")

	c.startOnce.Do(func() {
		for range c.workerCount {
			c.workerWG.Add(1)

			go c.run(ctx)
		}

		go c.awaitDrain()
	})

	return nil
}

// awaitDrain closes the done channel exactly once after every worker
// goroutine has exited.
func (c *queue[T]) awaitDrain() {
	c.workerWG.Wait()
	c.doneOnce.Do(func() { close(c.done) })
}

// Stop closes the inbound buffer and waits for every worker to
// finish processing in-flight messages, bounded by the configured
// drain timeout or by ctx's deadline (whichever is tighter). Stop is
// idempotent per the lifecycle.Component contract.
func (c *queue[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "QueueChannel is nil")

	c.stopOnce.Do(func() {
		c.closed.Store(true)
		close(c.inbound)
	})

	// Defensive: WithDrainTimeout rejects non-positive values, but a
	// struct constructed outside NewQueueChannel could land here with
	// drainTimeout == 0, which would cancel waitCtx immediately and
	// skip the drain entirely. Fall back to the package default.
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

// Done returns the channel that is closed after every worker has
// drained the inbound buffer and exited.
func (c *queue[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "QueueChannel is nil")

	return c.done
}

// Send enqueues msg for asynchronous dispatch. Send returns
// ErrSend(ErrClosed) after Stop. The behavior when the internal
// buffer is full depends on the configured OverflowPolicy (see
// WithOverflowPolicy): Block waits until a slot opens or ctx
// expires; DropNewest/DropOldest drop and fire the hook; Reject
// (the default) returns ErrSend(ErrBufferFull) immediately.
func (c *queue[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "QueueChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.closed.Load() {
		return ErrSend(ErrClosed)
	}

	return sendWithPolicy(ctx, c.inbound, msg, c.overflowPolicy, c.errorHandler)
}

// Subscribe registers handler at the end of the rotation pool and
// returns a Cancel that detaches it. Cancel is idempotent. Subscribe
// returns ErrSubscribe(ErrHandlerNil) when handler is nil and
// ErrSubscribe(ErrClosed) when the channel has been stopped.
//
// All registered handlers are equivalent peers competing for
// messages: subsequent messages are distributed round-robin among
// the live set, skipping cancelled handlers.
func (c *queue[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "QueueChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	// closed check + register under the same lock as Stop's
	// close-of-inbound so a Subscribe racing with Stop either bails
	// out cleanly or completes before Stop closes the channel. No
	// handler is left registered against a closed inbound channel.
	c.mu.Lock()
	if c.closed.Load() {
		c.mu.Unlock()

		return nil, ErrSubscribe(ErrClosed)
	}

	c.nextID++
	id := c.nextID
	c.byID[id] = handler
	c.order = append(c.order, id)
	c.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.byID, id)

			for i, candidate := range c.order {
				if candidate != id {
					continue
				}

				c.order = append(c.order[:i], c.order[i+1:]...)

				break
			}
		})
	}

	return cancel, nil
}

// run is one worker's loop. It consumes from inbound until the
// channel is closed, then exits and decrements the worker WaitGroup.
func (c *queue[T]) run(workerCtx context.Context) {
	defer c.workerWG.Done()

	for env := range c.inbound {
		c.dispatch(workerCtx, env)
	}
}

// dispatch selects the next subscriber via round-robin and invokes
// its handler with panic recovery. The handler ctx is the worker's
// lifecycle ctx merged with the publisher's Send ctx values (see
// mergeContexts): cancellation, deadline and Err follow the worker
// so a publisher abandoning its request does NOT abort in-flight
// handlers; Value lookups fall through to the publisher's ctx so
// trace span, correlation id and slogctx attrs reach the handler.
// Errors and recovered panics are routed through the configured
// ErrorHandler. Messages arriving when no subscribers are registered
// are dropped and surfaced via the hook with ErrNoSubscribers.
func (c *queue[T]) dispatch(workerCtx context.Context, env envelope[T]) {
	handlerCtx := mergeContexts(workerCtx, env.sendCtx)

	c.mu.Lock()

	n := len(c.order)
	if n == 0 {
		c.mu.Unlock()

		if c.errorHandler != nil {
			c.errorHandler(handlerCtx, env.msg, ErrNoSubscribers)
		}

		return
	}

	idx := c.rotation % uint64(n)
	c.rotation++
	handler := c.byID[c.order[idx]]
	c.mu.Unlock()

	err := invokeHandler(handlerCtx, env.msg, handler)
	if err == nil {
		return
	}

	if c.errorHandler != nil {
		c.errorHandler(handlerCtx, env.msg, err)
	}

	publishDeadLetter(handlerCtx, c.dlq, env.msg, err)
}
