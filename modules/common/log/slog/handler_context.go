package slog

import (
	"context"
	"log/slog"
)

// contextHandler is a slog.Handler middleware that runs the configured
// AttrExtractors on each record's context and merges the resulting attrs
// into the record before delegating to the wrapped handler. Extractors that
// return nil are skipped without allocation.
type contextHandler struct {
	inner      slog.Handler
	extractors []AttrExtractor
}

// NewContextHandler returns a slog.Handler that wraps inner and enriches
// every record with attrs returned by the supplied extractors. Nil
// extractors are filtered. If no extractors remain, inner is returned
// unchanged so callers pay no overhead.
func NewContextHandler(inner slog.Handler, extractors ...AttrExtractor) slog.Handler {
	var filtered []AttrExtractor

	for _, extractor := range extractors {
		if extractor != nil {
			filtered = append(filtered, extractor)
		}
	}

	if len(filtered) == 0 {
		return inner
	}

	return &contextHandler{inner: inner, extractors: filtered}
}

// Enabled delegates to the inner handler.
func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle enriches the record with attrs collected from every extractor and
// delegates to the inner handler. Allocation is amortised: only when at
// least one extractor returns attrs does the record get cloned.
func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	var extracted []slog.Attr

	for _, extractor := range h.extractors {
		attrs := extractor(ctx)
		if len(attrs) > 0 {
			extracted = append(extracted, attrs...)
		}
	}

	if len(extracted) > 0 {
		record.AddAttrs(extracted...)
	}

	return h.inner.Handle(ctx, record)
}

// WithAttrs re-wraps the inner handler with the same extractors so any
// child handler produced by upstream `slog.Logger.With(...)` keeps
// extraction active.
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{
		inner:      h.inner.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

// WithGroup re-wraps the inner handler with the same extractors so any
// grouped handler keeps extraction active.
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{
		inner:      h.inner.WithGroup(name),
		extractors: h.extractors,
	}
}
