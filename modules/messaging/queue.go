package messaging

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/lifecycle"
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
type QueueChannel[T any] struct {
	name         string
	bufferSize   int
	workerCount  int
	drainTimeout time.Duration
	errorHandler ErrorHandler

	inbound chan Message[T]
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
func NewQueueChannel[T any](name string, opts ...Option) *QueueChannel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &QueueChannel[T]{
		name:         name,
		bufferSize:   options.bufferSize,
		workerCount:  options.workerCount,
		drainTimeout: options.drainTimeout,
		errorHandler: options.errorHandler,
		inbound:      make(chan Message[T], options.bufferSize),
		done:         make(chan struct{}),
		byID:         map[uint64]Handler[T]{},
	}
}

// Name returns the channel's identity used in lifecycle logs.
func (q *QueueChannel[T]) Name() string {
	cassert.NotNil(q, "QueueChannel is nil")

	return q.name
}

// Start spawns the worker pool. Each worker consumes from the
// inbound buffer and dispatches each message to one subscriber
// chosen by round-robin among the currently registered handlers.
// Start is idempotent.
func (q *QueueChannel[T]) Start(ctx context.Context) error {
	cassert.NotNil(q, "QueueChannel is nil")

	q.startOnce.Do(func() {
		for range q.workerCount {
			q.workerWG.Add(1)

			go q.run(ctx)
		}

		go q.awaitDrain()
	})

	return nil
}

// awaitDrain closes the done channel exactly once after every worker
// goroutine has exited.
func (q *QueueChannel[T]) awaitDrain() {
	q.workerWG.Wait()
	q.doneOnce.Do(func() { close(q.done) })
}

// Stop closes the inbound buffer and waits for every worker to
// finish processing in-flight messages, bounded by the configured
// drain timeout or by ctx's deadline (whichever is tighter). Stop is
// idempotent per the lifecycle.Component contract.
func (q *QueueChannel[T]) Stop(ctx context.Context) error {
	cassert.NotNil(q, "QueueChannel is nil")

	q.stopOnce.Do(func() {
		q.closed.Store(true)
		close(q.inbound)
	})

	timeout := q.drainTimeout

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-q.done:
		return nil
	case <-waitCtx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, waitCtx.Err())
	}
}

// Done returns the channel that is closed after every worker has
// drained the inbound buffer and exited.
func (q *QueueChannel[T]) Done() <-chan struct{} {
	cassert.NotNil(q, "QueueChannel is nil")

	return q.done
}

// Send enqueues msg for asynchronous dispatch. Send returns
// ErrSend(ErrClosed) after Stop. When the buffer is full, Send
// blocks until a slot opens or ctx expires; on ctx expiry Send
// returns ErrSend with the ctx error joined.
func (q *QueueChannel[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(q, "QueueChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if q.closed.Load() {
		return ErrSend(ErrClosed)
	}

	select {
	case q.inbound <- msg:
		return nil
	case <-ctx.Done():
		return ErrSend(ErrTimeout, ctx.Err())
	}
}

// Subscribe registers handler at the end of the rotation pool and
// returns a Cancel that detaches it. Cancel is idempotent. Subscribe
// returns ErrSubscribe(ErrHandlerNil) when handler is nil and
// ErrSubscribe(ErrClosed) when the channel has been stopped.
//
// All registered handlers are equivalent peers competing for
// messages: subsequent messages are distributed round-robin among
// the live set, skipping cancelled handlers.
func (q *QueueChannel[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(q, "QueueChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	if q.closed.Load() {
		return nil, ErrSubscribe(ErrClosed)
	}

	q.mu.Lock()
	q.nextID++
	id := q.nextID
	q.byID[id] = handler
	q.order = append(q.order, id)
	q.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			q.mu.Lock()
			defer q.mu.Unlock()

			delete(q.byID, id)

			for i, candidate := range q.order {
				if candidate != id {
					continue
				}

				q.order = append(q.order[:i], q.order[i+1:]...)

				break
			}
		})
	}

	return cancel, nil
}

// run is one worker's loop. It consumes from inbound until the
// channel is closed, then exits and decrements the worker WaitGroup.
func (q *QueueChannel[T]) run(ctx context.Context) {
	defer q.workerWG.Done()

	for msg := range q.inbound {
		q.dispatch(ctx, msg)
	}
}

// dispatch selects the next subscriber via round-robin and invokes
// its handler with panic recovery. Errors and recovered panics are
// routed through the configured ErrorHandler (default: no-op).
// Messages arriving when no subscribers are registered are dropped
// and surfaced via the hook.
func (q *QueueChannel[T]) dispatch(ctx context.Context, msg Message[T]) {
	q.mu.Lock()

	n := len(q.order)
	if n == 0 {
		q.mu.Unlock()

		if q.errorHandler != nil {
			q.errorHandler(ctx, msg, ErrNoSubscribers)
		}

		return
	}

	idx := q.rotation % uint64(n)
	q.rotation++
	handler := q.byID[q.order[idx]]
	q.mu.Unlock()

	err := invokeHandler(ctx, msg, handler)
	if err == nil {
		return
	}

	if q.errorHandler != nil {
		q.errorHandler(ctx, msg, err)
	}
}
