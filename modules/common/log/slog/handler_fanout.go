package slog

import (
	"context"
	"errors"
	"log/slog"
)

var _ slog.Handler = (*fanoutHandler)(nil)

type fanoutHandler struct {
	handlers []slog.Handler
}

// NewFanoutHandler returns a slog.Handler that delegates to the provided handlers,
// but only if the handler is enabled for the given log level. Nil handlers are filtered out.
func NewFanoutHandler(handlers ...slog.Handler) slog.Handler {
	filtered := make([]slog.Handler, 0, len(handlers))

	for _, h := range handlers {
		if h != nil {
			filtered = append(filtered, h)
		}
	}

	return &fanoutHandler{handlers: filtered}
}

// Enabled returns true if any of the underlying handlers are enabled for the given log level.
func (h *fanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Handle processes the provided log record by dispatching it to all enabled handlers and aggregates errors, if any.
func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	if len(h.handlers) == 1 {
		if !h.handlers[0].Enabled(ctx, r.Level) {
			return nil
		}

		return h.handlers[0].Handle(ctx, r.Clone())
	}

	var errs []error

	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := h.handlers[i].Handle(ctx, r.Clone())
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// WithAttrs returns a new fanoutHandler with the provided attributes added to all underlying handlers.
func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}

	return &fanoutHandler{handlers: handlers}
}

// WithGroup returns a new fanoutHandler with the provided group name added to all underlying handlers.
func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}

	return &fanoutHandler{handlers: handlers}
}
