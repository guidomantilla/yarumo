package history

import (
	"context"
	"maps"
	"sync"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// history is the Message History implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled
// in Stop) and forwards each received message to the destination after
// appending its own configured name to the per-message history trail
// stored under Headers.Custom[historyKey].
type history[T any] struct {
	name         string
	src          messaging.Channel[T]
	dst          messaging.Channel[T]
	historyKey   string
	errorHandler messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewHistory constructs a Message History endpoint that subscribes to
// src and forwards every Message[T] to dst after appending its own
// name to the per-message history trail. The endpoint is not running
// on return; call lifecycle.Build (or Start directly) to register the
// subscription.
//
// name is used both in lifecycle logs and as the entry appended to the
// history trail. It must be non-empty. src and dst are mandatory.
// Optional behaviors:
//
//   - WithHistoryKey overrides the Headers.Custom map key used to store
//     the trail (default "History"). Use when the default would
//     collide with another caller-defined entry in Custom.
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for forward Send failures.
func NewHistory[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], opts ...Option) History[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")

	options := NewOptions(opts...)

	return &history[T]{
		name:         name,
		src:          src,
		dst:          dst,
		historyKey:   options.historyKey,
		errorHandler: options.errorHandler,
		done:         make(chan struct{}),
	}
}

// Name returns the history endpoint's identity used in lifecycle logs.
// It is also the value appended to the history trail on each pass-
// through.
func (h *history[T]) Name() string {
	cassert.NotNil(h, "history is nil")

	return h.name
}

// Start registers the history handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (h *history[T]) Start(_ context.Context) error {
	cassert.NotNil(h, "history is nil")

	var startErr error

	h.startOnce.Do(func() {
		cancel, err := h.src.Subscribe(h.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		h.mu.Lock()
		h.cancel = cancel
		h.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (h *history[T]) Stop(ctx context.Context) error {
	cassert.NotNil(h, "history is nil")

	h.stopOnce.Do(func() {
		h.mu.Lock()
		cancel := h.cancel
		h.cancel = nil
		h.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		h.doneOnce.Do(func() { close(h.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (h *history[T]) Done() <-chan struct{} {
	cassert.NotNil(h, "history is nil")

	return h.done
}

// handle is the Handler[T] subscribed on the source channel. It
// appends the endpoint's name to the message's history trail (working
// on a defensive copy so source-side state is not mutated) and
// forwards the message to the destination. Forward failures are
// reported through the configured ErrorHandler; the function itself
// always returns nil so history concerns never propagate to the
// source channel's Send caller.
func (h *history[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	stamped := h.stamp(msg)

	err := h.dst.Send(ctx, stamped)
	if err != nil && h.errorHandler != nil {
		h.errorHandler(ctx, stamped, ErrHistory(ErrForwardFailed, err))
	}

	return nil
}

// stamp returns a copy of msg with the endpoint's name appended to
// Headers.Custom[historyKey]. The function preserves an existing
// []string trail (append + store back); otherwise it seeds a fresh
// []string{name}. Headers.Custom is copied so downstream mutations do
// not leak back into the source message.
func (h *history[T]) stamp(msg messaging.Message[T]) messaging.Message[T] {
	out := msg
	out.Headers = msg.Headers
	out.Headers.Custom = cloneCustom(msg.Headers.Custom)

	if out.Headers.Custom == nil {
		out.Headers.Custom = map[string]any{}
	}

	existing, ok := out.Headers.Custom[h.historyKey].([]string)
	if !ok {
		out.Headers.Custom[h.historyKey] = []string{h.name}

		return out
	}

	trail := make([]string, len(existing), len(existing)+1)
	copy(trail, existing)
	trail = append(trail, h.name)
	out.Headers.Custom[h.historyKey] = trail

	return out
}

// cloneCustom returns a shallow copy of in or nil when in is nil.
// Cloning is required so appending to the trail does not mutate the
// caller's Headers.Custom map.
func cloneCustom(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}

	return maps.Clone(in)
}
