package slog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

// spyHandler is a minimal slog.Handler that tracks calls.
type spyHandler struct {
	enabled bool
	err     error
	handled int
}

func (s *spyHandler) Enabled(_ context.Context, _ slog.Level) bool { return s.enabled }

func (s *spyHandler) Handle(_ context.Context, _ slog.Record) error {
	s.handled++

	return s.err
}

func (s *spyHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return &spyHandler{enabled: s.enabled, err: s.err}
}

func (s *spyHandler) WithGroup(_ string) slog.Handler {
	return &spyHandler{enabled: s.enabled, err: s.err}
}

func TestNewFanoutHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil handler", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler()
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("accepts multiple handlers", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{}, &spyHandler{})
		if h == nil {
			t.Fatal("expected non-nil handler")
		}
	})
}

func TestFanoutHandler_Enabled(t *testing.T) {
	t.Parallel()

	t.Run("no handlers returns false", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler()
		if h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatal("expected false with no handlers")
		}
	})

	t.Run("returns true if any handler enabled", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{enabled: false}, &spyHandler{enabled: true})
		if !h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatal("expected true when at least one handler is enabled")
		}
	})

	t.Run("returns false if none enabled", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{enabled: false}, &spyHandler{enabled: false})
		if h.Enabled(context.Background(), slog.LevelInfo) {
			t.Fatal("expected false when no handlers are enabled")
		}
	})
}

func TestFanoutHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("single handler delegates directly", func(t *testing.T) {
		t.Parallel()

		spy := &spyHandler{enabled: true}
		h := NewFanoutHandler(spy)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if spy.handled != 1 {
			t.Fatalf("got %d handled, want 1", spy.handled)
		}
	})

	t.Run("single disabled handler skips handle", func(t *testing.T) {
		t.Parallel()

		spy := &spyHandler{enabled: false}
		h := NewFanoutHandler(spy)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if spy.handled != 0 {
			t.Fatalf("got %d handled, want 0", spy.handled)
		}
	})

	t.Run("multiple handlers respect enabled", func(t *testing.T) {
		t.Parallel()

		enabled := &spyHandler{enabled: true}
		disabled := &spyHandler{enabled: false}
		h := NewFanoutHandler(enabled, disabled)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if enabled.handled != 1 {
			t.Fatalf("enabled: got %d handled, want 1", enabled.handled)
		}

		if disabled.handled != 0 {
			t.Fatalf("disabled: got %d handled, want 0", disabled.handled)
		}
	})

	t.Run("propagates handler errors", func(t *testing.T) {
		t.Parallel()

		spy := &spyHandler{enabled: true, err: errors.New("boom")}
		h := NewFanoutHandler(spy, spy)

		err := h.Handle(context.Background(), slog.Record{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("zero handlers returns nil", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler()

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("single handler error propagated", func(t *testing.T) {
		t.Parallel()

		spy := &spyHandler{enabled: true, err: errors.New("fail")}
		h := NewFanoutHandler(spy)

		err := h.Handle(context.Background(), slog.Record{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil handlers filtered out", func(t *testing.T) {
		t.Parallel()

		spy := &spyHandler{enabled: true}
		h := NewFanoutHandler(nil, spy, nil)

		err := h.Handle(context.Background(), slog.Record{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if spy.handled != 1 {
			t.Fatalf("got %d handled, want 1", spy.handled)
		}
	})
}

func TestFanoutHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	t.Run("returns new handler with attrs", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{enabled: true})

		h2 := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
		if h2 == nil {
			t.Fatal("expected non-nil handler")
		}

		if h2 == h {
			t.Fatal("expected new handler, got same reference")
		}
	})
}

func TestFanoutHandler_WithGroup(t *testing.T) {
	t.Parallel()

	t.Run("empty name returns same handler", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{enabled: true})

		h2 := h.WithGroup("")
		if h2 != h {
			t.Fatal("expected same handler for empty group name")
		}
	})

	t.Run("non-empty name returns new handler", func(t *testing.T) {
		t.Parallel()

		h := NewFanoutHandler(&spyHandler{enabled: true})

		h2 := h.WithGroup("mygroup")
		if h2 == h {
			t.Fatal("expected new handler, got same reference")
		}
	})
}
