package slog

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// captureHandler is a slog.Handler that records every record it receives.
// It is used to assert that the context handler merges extractor attrs
// before delegating.
type captureHandler struct {
	records []slog.Record
}

func (c *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (c *captureHandler) Handle(_ context.Context, r slog.Record) error {
	c.records = append(c.records, r)
	return nil
}

func (c *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return c }

func (c *captureHandler) WithGroup(_ string) slog.Handler { return c }

func TestNewContextHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns inner when no extractors", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}

		got := NewContextHandler(inner)
		if got != slog.Handler(inner) {
			t.Fatalf("expected inner returned unchanged when no extractors")
		}
	})

	t.Run("filters nil extractors", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}

		got := NewContextHandler(inner, nil, nil)
		if got != slog.Handler(inner) {
			t.Fatalf("expected inner returned when all extractors are nil")
		}
	})
}

func TestContextHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("merges extractor attrs into record", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("ext", "v")}
		}

		handler := NewContextHandler(inner, extractor)

		record := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
		record.AddAttrs(slog.String("orig", "1"))

		err := handler.Handle(context.Background(), record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(inner.records) != 1 {
			t.Fatalf("got %d records, want 1", len(inner.records))
		}

		var keys []string
		inner.records[0].Attrs(func(a slog.Attr) bool {
			keys = append(keys, a.Key)
			return true
		})

		if len(keys) != 2 || keys[0] != "orig" || keys[1] != "ext" {
			t.Fatalf("got attrs %v, want [orig ext]", keys)
		}
	})

	t.Run("no allocation when extractor returns nil", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}
		extractor := func(_ context.Context) []slog.Attr { return nil }

		handler := NewContextHandler(inner, extractor)

		record := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
		record.AddAttrs(slog.String("orig", "1"))

		err := handler.Handle(context.Background(), record)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var keys []string
		inner.records[0].Attrs(func(a slog.Attr) bool {
			keys = append(keys, a.Key)
			return true
		})

		if len(keys) != 1 || keys[0] != "orig" {
			t.Fatalf("got attrs %v, want [orig]", keys)
		}
	})
}

func TestContextHandler_Enabled(t *testing.T) {
	t.Parallel()

	t.Run("delegates to inner", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}
		extractor := func(_ context.Context) []slog.Attr { return nil }

		handler := NewContextHandler(inner, extractor)

		if !handler.Enabled(context.Background(), slog.LevelDebug) {
			t.Fatalf("expected Enabled to return true")
		}
	})
}

func TestContextHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	t.Run("re-wraps inner keeping extractors active", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("ext", "v")}
		}

		handler := NewContextHandler(inner, extractor)

		child := handler.WithAttrs([]slog.Attr{slog.String("a", "1")})

		record := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
		_ = child.Handle(context.Background(), record)

		if len(inner.records) != 1 {
			t.Fatalf("expected child to delegate to same inner")
		}

		var keys []string
		inner.records[0].Attrs(func(a slog.Attr) bool {
			keys = append(keys, a.Key)
			return true
		})

		if len(keys) != 1 || keys[0] != "ext" {
			t.Fatalf("got attrs %v, want [ext]", keys)
		}
	})
}

func TestContextHandler_WithGroup(t *testing.T) {
	t.Parallel()

	t.Run("re-wraps inner keeping extractors active", func(t *testing.T) {
		t.Parallel()

		inner := &captureHandler{}
		extractor := func(_ context.Context) []slog.Attr {
			return []slog.Attr{slog.String("ext", "v")}
		}

		handler := NewContextHandler(inner, extractor)

		child := handler.WithGroup("g")

		record := slog.NewRecord(time.Time{}, slog.LevelInfo, "msg", 0)
		_ = child.Handle(context.Background(), record)

		if len(inner.records) != 1 {
			t.Fatalf("expected child to delegate to same inner")
		}
	})
}
