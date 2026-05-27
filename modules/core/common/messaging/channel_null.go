package messaging

import (
	"context"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// nullChannel implements Channel[T] as a /dev/null sink. Send accepts
// any message, immediately discards it, and fires the configured
// ErrorHandler with ErrDropped so test and observability paths see the
// drop. Subscribe registers handlers for shape compatibility but they
// are never invoked. There are no goroutines, no buffering, and no
// lifecycle to manage.
//
// Use it as a test double (verify "no message ever reaches the
// downstream side") or as an explicit "this event flow is disabled"
// wiring that beats nil-checks on a Channel[T] field.
type null[T any] struct {
	errorHandler ErrorHandler

	mu     sync.Mutex
	nextID uint64
	byID   map[uint64]Handler[T]
}

// NewNullChannel returns a Channel[T] that drops every message sent to
// it. When an ErrorHandler is configured via WithErrorHandler, it is
// invoked once per Send with ErrDropped so tests can assert the drop.
// Subscribed handlers are tracked for Cancel idempotency but are
// never invoked.
func NewNullChannel[T any](opts ...Option) Channel[T] {
	options := NewOptions(opts...)

	return &null[T]{
		errorHandler: options.errorHandler,
		byID:         map[uint64]Handler[T]{},
	}
}

// Send drops msg on the floor and notifies the ErrorHandler hook with
// ErrDropped. The returned error is always nil: a dropped message is
// not a failure of Send, only an observation. Returns ErrSend wrapping
// ErrContextNil when ctx is nil (preserves the workspace contract that
// nil ctx never reaches a handler).
func (c *null[T]) Send(ctx context.Context, msg Message[T]) error {
	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	cassert.NotNil(c, "nullChannel is nil")

	if c.errorHandler != nil {
		c.errorHandler(ctx, msg, ErrDropped)
	}

	return nil
}

// Subscribe registers handler and returns a Cancel that detaches it.
// The handler is never invoked — Send drops every message — but the
// registration is tracked so the Cancel/idempotency contract matches
// the other Channel implementations. Returns ErrSubscribe(ErrHandlerNil)
// when handler is nil.
func (c *null[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "nullChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	c.byID[id] = handler
	c.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()

			delete(c.byID, id)
		})
	}

	return cancel, nil
}
