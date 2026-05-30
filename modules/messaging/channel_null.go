package messaging

import (
	"context"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// null implements Channel[T] as a /dev/null sink. Send accepts any
// message, immediately discards it, and fires the configured
// ErrorHandler with ErrDropped so test and observability paths see
// the drop. Subscribe accepts handlers for shape compatibility but
// they are never invoked — Send drops every message. There is no
// internal state, no goroutines, no buffering, and no lifecycle.
//
// Use it as a test double (verify "no message ever reaches the
// downstream side") or as an explicit "this event flow is disabled"
// wiring that beats nil-checks on a Channel[T] field.
type null[T any] struct {
	errorHandler ErrorHandler
}

// NewNullChannel returns a Channel[T] that drops every message sent to
// it. When an ErrorHandler is configured via WithErrorHandler, it is
// invoked once per Send with ErrDropped so tests can assert the drop.
// Subscribed handlers are accepted for interface compatibility but
// are never invoked.
func NewNullChannel[T any](opts ...Option) Channel[T] {
	return &null[T]{errorHandler: NewOptions(opts...).errorHandler}
}

// Send drops msg on the floor and notifies the ErrorHandler hook with
// ErrDropped. The returned error is always nil: a dropped message is
// not a failure of Send, only an observation. Returns ErrSend wrapping
// ErrContextNil when ctx is nil (preserves the workspace contract that
// nil ctx never reaches a handler).
func (c *null[T]) Send(ctx context.Context, msg Message[T]) error {
	cassert.NotNil(c, "nullChannel is nil")
	if ctx == nil {
		return ErrSend(ErrContextNil)
	}

	if c.errorHandler != nil {
		c.errorHandler(ctx, msg, ErrDropped)
	}

	return nil
}

// Subscribe accepts handler for interface compatibility and returns a
// no-op Cancel. The handler is never invoked — NullChannel does not
// dispatch. Returns ErrSubscribe(ErrHandlerNil) when handler is nil,
// preserving the workspace contract that nil handlers never enter the
// system.
func (c *null[T]) Subscribe(handler Handler[T]) (Cancel, error) {
	cassert.NotNil(c, "nullChannel is nil")

	if handler == nil {
		return nil, ErrSubscribe(ErrHandlerNil)
	}

	return func() {}, nil
}
