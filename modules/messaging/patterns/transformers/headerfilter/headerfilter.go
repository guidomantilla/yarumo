package headerfilter

import (
	"context"
	"maps"
	"sync"
	"time"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/common/lifecycle"
	"github.com/guidomantilla/yarumo/messaging"
)

// headerFilter is the Header Filter implementation. It owns a single
// subscription on the source channel (registered in Start, cancelled in
// Stop) and forwards every received message to the destination with the
// configured Headers fields cleared.
type headerFilter[T any] struct {
	name           string
	src            messaging.Channel[T]
	dst            messaging.Channel[T]
	headersToClear []string
	errorHandler   messaging.ErrorHandler

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	doneOnce  sync.Once

	mu     sync.Mutex
	cancel messaging.Cancel
}

// NewHeaderFilter constructs a HeaderFilter that subscribes to src and
// forwards every Message[T] to dst with the configured Headers fields
// cleared. The header filter is not running on return; call
// lifecycle.Build (or Start directly) to register the subscription.
//
// name is used in lifecycle logs and must be non-empty. src and dst
// are mandatory. The list of headers to clear is configured via
// WithClearHeader (one at a time) or WithHeadersToClear (variadic). An
// empty list forwards messages unchanged.
//
// Optional behaviors:
//
//   - WithErrorHandler overrides the default
//     messaging.DefaultErrorHandler (which logs via common/log) with a
//     custom hook for forward Send failures.
func NewHeaderFilter[T any](name string, src messaging.Channel[T], dst messaging.Channel[T], opts ...Option) HeaderFilter[T] {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(src, "source channel is nil")
	cassert.NotNil(dst, "destination channel is nil")

	options := NewOptions(opts...)

	return &headerFilter[T]{
		name:           name,
		src:            src,
		dst:            dst,
		headersToClear: options.headersToClear,
		errorHandler:   options.errorHandler,
		done:           make(chan struct{}),
	}
}

// Name returns the header filter's identity used in lifecycle logs.
func (f *headerFilter[T]) Name() string {
	cassert.NotNil(f, "header filter is nil")

	return f.name
}

// Start registers the filtering handler as a subscriber on the source
// channel. It satisfies the lifecycle.Component worker-style contract:
// Start returns immediately after the subscription is in place; the
// actual dispatching runs in the source channel's goroutine model.
// Start is idempotent — a second invocation returns nil without
// re-subscribing.
func (f *headerFilter[T]) Start(_ context.Context) error {
	cassert.NotNil(f, "header filter is nil")

	var startErr error

	f.startOnce.Do(func() {
		cancel, err := f.src.Subscribe(f.handle)
		if err != nil {
			startErr = lifecycle.ErrStart(err)

			return
		}

		f.mu.Lock()
		f.cancel = cancel
		f.mu.Unlock()
	})

	return startErr
}

// Stop cancels the source-channel subscription and closes Done. Stop
// is idempotent per the lifecycle.Component contract. It returns
// lifecycle.ErrShutdown wrapping lifecycle.ErrShutdownTimeout when ctx
// is already expired on entry; otherwise nil.
func (f *headerFilter[T]) Stop(ctx context.Context) error {
	cassert.NotNil(f, "header filter is nil")

	f.stopOnce.Do(func() {
		f.mu.Lock()
		cancel := f.cancel
		f.cancel = nil
		f.mu.Unlock()

		if cancel != nil {
			cancel()
		}

		f.doneOnce.Do(func() { close(f.done) })
	})

	select {
	case <-ctx.Done():
		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	default:
		return nil
	}
}

// Done returns the channel that is closed after Stop has been called.
func (f *headerFilter[T]) Done() <-chan struct{} {
	cassert.NotNil(f, "header filter is nil")

	return f.done
}

// handle is the Handler[T] subscribed on the source channel. It builds
// a forwarded copy of msg with the configured headers cleared and
// forwards it to dst. Forward failures are reported through the
// configured ErrorHandler. The function itself always returns nil so
// header filter concerns never propagate to the source channel's Send
// caller.
func (f *headerFilter[T]) handle(ctx context.Context, msg messaging.Message[T]) error {
	forwarded := messaging.Message[T]{
		Payload: msg.Payload,
		Headers: clearHeaders(msg.Headers, f.headersToClear),
	}

	err := f.dst.Send(ctx, forwarded)
	if err != nil {
		f.reportError(ctx, msg, ErrHeaderFilter(ErrForwardFailed, err))
	}

	return nil
}

// reportError forwards err to the configured ErrorHandler. ErrorHandler
// is guaranteed non-nil by NewOptions (defaults to
// messaging.DefaultErrorHandler), so the nil-guard is defensive only.
func (f *headerFilter[T]) reportError(ctx context.Context, msg messaging.Message[T], err error) {
	if f.errorHandler == nil {
		return
	}

	f.errorHandler(ctx, msg, err)
}

// clearHeaders returns a copy of h with every name in names cleared.
// Known struct fields are zeroed in place; unknown names are removed
// from the Custom map (a defensive copy is taken before deletion so the
// source map is never mutated). When names is empty h is returned
// unchanged (no Custom-map clone is made).
func clearHeaders(h messaging.Headers, names []string) messaging.Headers {
	if len(names) == 0 {
		return h
	}

	out := h

	cloneNeeded := false

	for _, name := range names {
		if !isKnownField(name) && out.Custom != nil {
			if _, ok := out.Custom[name]; ok {
				cloneNeeded = true

				break
			}
		}
	}

	if cloneNeeded {
		out.Custom = maps.Clone(out.Custom)
	}

	for _, name := range names {
		clearOne(&out, name)
	}

	return out
}

// fieldClearers maps a recognised messaging.Headers struct field name
// to the function that zeroes that field on a Headers value. Names
// absent from this table are treated as Custom map keys.
var fieldClearers = map[string]func(*messaging.Headers){
	"MessageID":      func(h *messaging.Headers) { h.MessageID = "" },
	"CorrelationID":  func(h *messaging.Headers) { h.CorrelationID = "" },
	"CausationID":    func(h *messaging.Headers) { h.CausationID = "" },
	"ReplyTo":        func(h *messaging.Headers) { h.ReplyTo = "" },
	"Type":           func(h *messaging.Headers) { h.Type = "" },
	"Source":         func(h *messaging.Headers) { h.Source = "" },
	"ContentType":    func(h *messaging.Headers) { h.ContentType = "" },
	"Priority":       func(h *messaging.Headers) { h.Priority = 0 },
	"ExpirationTime": func(h *messaging.Headers) { h.ExpirationTime = time.Time{} },
	"SequenceNumber": func(h *messaging.Headers) { h.SequenceNumber = 0 },
	"SequenceSize":   func(h *messaging.Headers) { h.SequenceSize = 0 },
	"Timestamp":      func(h *messaging.Headers) { h.Timestamp = time.Time{} },
}

// clearOne zeroes the named struct field on out, or deletes the entry
// from out.Custom when the name is not a recognised struct field. The
// caller is responsible for cloning out.Custom before invoking
// clearOne with Custom keys.
func clearOne(out *messaging.Headers, name string) {
	clearer, ok := fieldClearers[name]
	if ok {
		clearer(out)

		return
	}

	if out.Custom != nil {
		delete(out.Custom, name)
	}
}

// isKnownField reports whether name matches one of the recognised
// messaging.Headers struct fields cleared in place.
func isKnownField(name string) bool {
	_, ok := fieldClearers[name]

	return ok
}
