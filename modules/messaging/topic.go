package messaging

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/lifecycle"
)

// TopicChannel is the buffered async Channel[T] implementation.
//
// TopicChannel implements both Channel[T] and lifecycle.Component
// (worker-style): Start spawns a single worker goroutine that consumes
// from an internal buffered channel and invokes every registered
// handler for each message. Stop closes the inbound channel and waits
// for the worker to drain the in-flight messages, bounded by the
// configured drain timeout (see WithDrainTimeout) or by the ctx
// deadline passed to Stop, whichever is tighter.
//
// TopicChannel is safe for concurrent use. Send is non-blocking
// (returns immediately after enqueue) until the buffer fills; once
// full, Send blocks until either the worker drains a slot or ctx
// expires.
type TopicChannel[T any] struct {
	name         string
	bufferSize   int
	drainTimeout time.Duration
	errorHandler ErrorHandler

	inbound chan Message[T]
	done    chan struct{}
	closed  atomic.Bool

	mu     sync.RWMutex
	nextID uint64
	order  []uint64
	byID   map[uint64]Handler[T]

	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once
}

// NewTopicChannel constructs a TopicChannel[T] with the given name and
// options. name is used in lifecycle logs and must be non-empty.
//
// The returned channel is not running; call lifecycle.Build (or the
// BuildTopicChannel convenience) to spawn the worker goroutine.
func NewTopicChannel[T any](name string, opts ...Option) *TopicChannel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &TopicChannel[T]{
		name:         name,
		bufferSize:   options.bufferSize,
		drainTimeout: options.drainTimeout,
		errorHandler: options.errorHandler,
		inbound:      make(chan Message[T], options.bufferSize),
		done:         make(chan struct{}),
		byID:         map[uint64]Handler[T]{},
	}
}

// BuildTopicChannel wires a TopicChannel[T] into the application
// lifecycle via lifecycle.Build. It returns the CloseFn for graceful
// shutdown.
func BuildTopicChannel[T any](ctx context.Context, qc *TopicChannel[T], errChan lifecycle.ErrChan) (lifecycle.CloseFn, error) {
	cassert.NotNil(qc, "TopicChannel is nil")

	closeFn, err := lifecycle.Build(ctx, qc, errChan)
	if err != nil {
		return nil, err
	}

	return closeFn, nil
}

// Name returns the channel's identity used in lifecycle logs.
func (q *TopicChannel[T]) Name() string {
	cassert.NotNil(q, "TopicChannel is nil")

	return q.name
}

// Start spawns the worker goroutine that consumes from the inbound
// queue and dispatches each message to all currently registered
// handlers. It satisfies the lifecycle.Component worker-style
// contract: Start returns immediately; Done closes after Stop has
// drained the worker.
func (q *TopicChannel[T]) Start(ctx context.Context) error {
	cassert.NotNil(q, "TopicChannel is nil")

	q.startOnce.Do(func() {
		go q.run(ctx)
	})

	return nil
}

// Stop closes the inbound queue and waits for the worker to drain
// pending messages, bounded by the configured drain timeout or by
// ctx's deadline (whichever is tighter). It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when
// the drain does not complete in time. Stop is idempotent per the
// lifecycle.Component contract.
func (q *TopicChannel[T]) Stop(ctx context.Context) error {
	cassert.NotNil(q, "TopicChannel is nil")

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

// Done returns the channel that is closed after the worker has
// drained the inbound queue and exited.
func (q *TopicChannel[T]) Done() <-chan struct{} {
	cassert.NotNil(q, "TopicChannel is nil")

	return q.done
}

// Send enqueues msg for asynchronous dispatch by the worker. Send
// returns ErrSend(ErrClosed) after Stop. When the buffer is full,
// Send blocks until either a slot becomes available or ctx expires;
// on ctx expiry, Send returns ErrSend with the ctx error joined.
func (q *TopicChannel[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(q, "TopicChannel is nil")

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

// Subscribe registers handler at the end of the subscriber list and
// returns a Cancel that detaches it. Subscribers receive messages in
// Subscribe order. Cancel is idempotent. Subscribe returns
// ErrSubscribe(ErrHandlerNil) when handler is nil and
// ErrSubscribe(ErrClosed) when the channel has been stopped.
func (q *TopicChannel[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(q, "TopicChannel is nil")

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

// run is the worker loop. It reads from inbound until the channel is
// closed, dispatching each message to every registered handler. It
// closes done exactly once when the loop exits.
func (q *TopicChannel[T]) run(ctx context.Context) {
	defer q.doneOnce.Do(func() { close(q.done) })

	for msg := range q.inbound {
		q.dispatch(ctx, msg)
	}
}

// dispatch invokes every registered handler for the given message in
// Subscribe order. Each handler runs under panic recovery so one bad
// message cannot kill the worker. Handler errors and recovered panics
// are routed through the configured ErrorHandler (default: no-op)
// since async dispatch has no return path to the publisher.
func (q *TopicChannel[T]) dispatch(ctx context.Context, msg Message[T]) {
	q.mu.RLock()
	snapshot := make([]Handler[T], 0, len(q.order))
	for _, id := range q.order {
		snapshot = append(snapshot, q.byID[id])
	}
	q.mu.RUnlock()

	for _, handler := range snapshot {
		err := invokeHandler(ctx, msg, handler)
		if err == nil {
			continue
		}

		if q.errorHandler != nil {
			q.errorHandler(ctx, msg, err)
		}
	}
}

// invokeHandler runs one handler with panic recovery. Returned error
// is nil on success, the handler's error on a normal failure, or an
// ErrHandlerPanic-wrapping error on panic.
func invokeHandler[T any](ctx context.Context, msg Message[T], handler Handler[T]) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		err = fmt.Errorf("%w: %v", ErrHandlerPanic, r)
	}()

	return handler(ctx, msg)
}
