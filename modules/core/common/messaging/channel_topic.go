package messaging

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
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
//
// Context propagation: the Send ctx travels with the message in the
// internal buffer but its cancellation does NOT abort the handler —
// async dispatch must outlive the publisher's request. The handler
// receives a ctx whose Done / Deadline / Err follow the worker's
// lifecycle ctx (the one passed to Start) and whose Value lookups
// fall through to the Send ctx so trace span, correlation id and
// slogctx attributes propagate from the publisher.
type topic[T any] struct {
	name         string
	bufferSize   int
	drainTimeout time.Duration
	errorHandler ErrorHandler

	inbound chan envelope[T]
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
		inbound:      make(chan envelope[T], options.bufferSize),
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
	case c.inbound <- envelope[T]{sendCtx: ctx, msg: msg}:
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
func (c *topic[T]) run(workerCtx context.Context) {
	defer c.doneOnce.Do(func() { close(c.done) })

	for env := range c.inbound {
		c.dispatch(workerCtx, env)
	}
}

// dispatch invokes every registered handler for the given message in
// Subscribe order. The handler ctx is the worker's lifecycle ctx
// merged with the publisher's Send ctx values (see mergeContexts):
// cancellation, deadline and Err follow the worker so a publisher
// abandoning its request does NOT abort in-flight handlers; Value
// lookups fall through to the publisher's ctx so trace span,
// correlation id and slogctx attrs reach the handler. Each handler
// runs under panic recovery so one bad message cannot kill the
// worker. Handler errors and recovered panics are routed through
// the configured ErrorHandler since async dispatch has no return
// path to the publisher.
func (c *topic[T]) dispatch(workerCtx context.Context, env envelope[T]) {
	snapshot := snapshotHandlers(&c.mu, &c.order, c.byID)

	handlerCtx := mergeContexts(workerCtx, env.sendCtx)

	for _, handler := range snapshot {
		err := invokeHandler(handlerCtx, env.msg, handler)
		if err == nil {
			continue
		}

		if c.errorHandler != nil {
			c.errorHandler(handlerCtx, env.msg, err)
		}
	}
}
