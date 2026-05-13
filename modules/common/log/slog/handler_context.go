package slog

import (
	"context"
	"log/slog"
)

// AttrExtractor pulls slog attributes from a context. It must be safe for
// concurrent use and must return nil (not panic) when ctx carries no attrs.
// A nil context, when received, must be treated as an empty context.
type AttrExtractor func(ctx context.Context) []slog.Attr

var _ slog.Handler = (*contextHandler)(nil)

// contextHandler is a slog.Handler middleware that, for every log record,
// invokes the configured extractors against the record's context and adds the
// returned attributes to the record before delegating to the wrapped handler.
//
// Multiple extractors can be composed: their attributes are appended in the
// order the extractors were registered. Extractors that return nil are skipped
// for free — they incur no allocation.
type contextHandler struct {
	inner      slog.Handler
	extractors []AttrExtractor
}

// NewContextHandler wraps inner so that, on every record, the supplied
// extractors are invoked against the record's context and their attributes
// are added to the record before delegation.
//
// If inner is nil or no extractors are supplied, NewContextHandler returns
// inner unchanged (a noop wrapper would add cost without value).
func NewContextHandler(inner slog.Handler, extractors ...AttrExtractor) slog.Handler {
	if inner == nil {
		return nil
	}

	filtered := make([]AttrExtractor, 0, len(extractors))

	for _, fn := range extractors {
		if fn != nil {
			filtered = append(filtered, fn)
		}
	}

	if len(filtered) == 0 {
		return inner
	}

	return &contextHandler{inner: inner, extractors: filtered}
}

// Enabled reports whether the wrapped handler is enabled for the given level.
func (h *contextHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

// Handle invokes each extractor against ctx, appends the resulting attributes
// to a clone of r, then delegates to the wrapped handler.
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	clone := r.Clone()

	for _, fn := range h.extractors {
		attrs := fn(ctx)
		if len(attrs) == 0 {
			continue
		}

		clone.AddAttrs(attrs...)
	}

	return h.inner.Handle(ctx, clone)
}

// WithAttrs delegates to the wrapped handler and re-wraps the result so that
// context extraction remains active for child handlers.
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{
		inner:      h.inner.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

// WithGroup delegates to the wrapped handler and re-wraps the result so that
// context extraction remains active for child handlers.
func (h *contextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &contextHandler{
		inner:      h.inner.WithGroup(name),
		extractors: h.extractors,
	}
}
