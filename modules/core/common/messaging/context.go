package messaging

import (
	"context"
)

// envelope pairs a Message[T] with the publisher's Send ctx as it
// travels through an async channel's internal buffer. It exists so
// the dispatcher can later derive the handler ctx via mergeContexts,
// combining the worker's lifecycle ctx with the publisher's values
// (trace span, correlation id, slogctx attrs) without exposing the
// pairing in the public Message[T] envelope.
type envelope[T any] struct {
	sendCtx context.Context
	msg     Message[T]
}

// mergedContext fuses two contexts: lifecycle drives Deadline / Done /
// Err and values is consulted first on Value lookups before falling
// back to lifecycle. It is used by async channels to derive the ctx
// passed to a handler from (a) the worker's lifecycle ctx (typically
// the one passed to Start) and (b) the publisher's Send ctx that
// originated the message.
type mergedContext struct {
	context.Context
	values context.Context
}

// Value looks up key in the publisher's ctx first so trace spans,
// correlation ids and slogctx attributes propagate from Send to the
// handler. Falls back to the lifecycle ctx so values set on it remain
// visible.
func (c mergedContext) Value(key any) any {
	v := c.values.Value(key)
	if v != nil {
		return v
	}

	return c.Context.Value(key)
}

// mergeContexts returns a context whose cancellation, deadline and
// error follow lifecycle, but whose Value lookups fall through to
// values first and then to lifecycle. It preserves async fire-and-
// forget semantics — publisher cancellation does NOT abort in-flight
// handlers — while still propagating observability values like trace
// spans, correlation ids and slogctx attributes from the publisher
// down to the handler.
//
// When values is nil mergeContexts returns lifecycle unchanged. When
// lifecycle is nil it returns context.Background to keep the handler
// invariant that ctx is never nil (Send already rejects nil ctx, so
// this case only protects against future internal callers).
func mergeContexts(lifecycle, values context.Context) context.Context {
	if lifecycle == nil {
		lifecycle = context.Background()
	}

	if values == nil {
		return lifecycle
	}

	return mergedContext{Context: lifecycle, values: values}
}
