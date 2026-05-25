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
type topic[T any] struct {
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
func NewTopicChannel[T any](name string, opts ...Option) Channel[T] {
	cassert.NotEmpty(name, "name is empty")

	options := NewOptions(opts...)

	return &topic[T]{
		name:         name,
		bufferSize:   options.bufferSize,
		drainTimeout: options.drainTimeout,
		errorHandler: options.errorHandler,
		inbound:      make(chan Message[T], options.bufferSize),
		done:         make(chan struct{}),
		byID:         map[uint64]Handler[T]{},
	}
}

// Name returns the channel's identity used in lifecycle logs.
func (c *topic[T]) Name() string {
	cassert.NotNil(c, "TopicChannel is nil")

	return c.name
}

// Start spawns the worker goroutine that consumes from the inbound
// queue and dispatches each message to all currently registered
// handlers. It satisfies the lifecycle.Component worker-style
// contract: Start returns immediately; Done closes after Stop has
// drained the worker.
func (c *topic[T]) Start(ctx context.Context) error {
	cassert.NotNil(c, "TopicChannel is nil")

	c.startOnce.Do(func() {
		go c.run(ctx)
	})

	return nil
}

// Stop closes the inbound queue and waits for the worker to drain
// pending messages, bounded by the configured drain timeout or by
// ctx's deadline (whichever is tighter). It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when
// the drain does not complete in time. Stop is idempotent per the
// lifecycle.Component contract.
func (c *topic[T]) Stop(ctx context.Context) error {
	cassert.NotNil(c, "TopicChannel is nil")

	c.stopOnce.Do(func() {
		c.closed.Store(true)
		close(c.inbound)
	})

	timeout := c.drainTimeout

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-c.done:
		return nil
	case <-waitCtx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, waitCtx.Err())
	}
}

// Done returns the channel that is closed after the worker has
// drained the inbound queue and exited.
func (c *topic[T]) Done() <-chan struct{} {
	cassert.NotNil(c, "TopicChannel is nil")

	return c.done
}

// Send enqueues msg for asynchronous dispatch by the worker. Send
// returns ErrSend(ErrClosed) after Stop. When the buffer is full,
// Send blocks until either a slot becomes available or ctx expires;
// on ctx expiry, Send returns ErrSend with the ctx error joined.
func (c *topic[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "TopicChannel is nil")

	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.closed.Load() {
		return ErrSend(ErrClosed)
	}

	select {
	case c.inbound <- msg:
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
func (c *topic[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "TopicChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	if c.closed.Load() {
		return nil, ErrSubscribe(ErrClosed)
	}

	c.mu.Lock()
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

// run is the worker loop. It reads from inbound until the channel is
// closed, dispatching each message to every registered handler. It
// closes done exactly once when the loop exits.
func (c *topic[T]) run(ctx context.Context) {
	defer c.doneOnce.Do(func() { close(c.done) })

	for msg := range c.inbound {
		c.dispatch(ctx, msg)
	}
}

// dispatch invokes every registered handler for the given message in
// Subscribe order. Each handler runs under panic recovery so one bad
// message cannot kill the worker. Handler errors and recovered panics
// are routed through the configured ErrorHandler (default: no-op)
// since async dispatch has no return path to the publisher.
func (c *topic[T]) dispatch(ctx context.Context, msg Message[T]) {
	c.mu.RLock()
	snapshot := make([]Handler[T], 0, len(c.order))
	for _, id := range c.order {
		snapshot = append(snapshot, c.byID[id])
	}
	c.mu.RUnlock()

	for _, handler := range snapshot {
		err := invokeHandler(ctx, msg, handler)
		if err == nil {
			continue
		}

		if c.errorHandler != nil {
			c.errorHandler(ctx, msg, err)
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
