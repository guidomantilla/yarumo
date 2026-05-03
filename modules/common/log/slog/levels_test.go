package slog

import (
	"log/slog"
	"testing"
)

func TestLevel_toSlog(t *testing.T) {
	t.Parallel()

	t.Run("converts to matching slog level", func(t *testing.T) {
		t.Parallel()

		if LevelTrace.toSlog() != slog.Level(-8) {
			t.Fatalf("got %v, want %v", LevelTrace.toSlog(), slog.Level(-8))
		}

		if LevelInfo.toSlog() != slog.Level(0) {
			t.Fatalf("got %v, want %v", LevelInfo.toSlog(), slog.Level(0))
		}
	})
}

func TestReplaceLevel(t *testing.T) {
	t.Parallel()

	t.Run("replaces custom levels", func(t *testing.T) {
		t.Parallel()

		trace := slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(slog.Level(LevelTrace))}

		got := ReplaceLevel(nil, trace)
		if got.Value.String() != "TRACE" {
			t.Fatalf("got %q, want %q", got.Value.String(), "TRACE")
		}

		fatal := slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(slog.Level(LevelFatal))}

		got = ReplaceLevel(nil, fatal)
		if got.Value.String() != "FATAL" {
			t.Fatalf("got %q, want %q", got.Value.String(), "FATAL")
		}
	})

	t.Run("preserves standard levels", func(t *testing.T) {
		t.Parallel()

		info := slog.Attr{Key: slog.LevelKey, Value: slog.AnyValue(slog.Level(LevelInfo))}

		got := ReplaceLevel(nil, info)
		if got.Value.Any() != slog.Level(LevelInfo) {
			t.Fatalf("standard level was modified: %v", got.Value)
		}
	})

	t.Run("returns unchanged for non-level key", func(t *testing.T) {
		t.Parallel()

		a := slog.String("msg", "hello")

		got := ReplaceLevel(nil, a)
		if got.Value.String() != "hello" {
			t.Fatalf("got %q, want %q", got.Value.String(), "hello")
		}
	})

	t.Run("returns unchanged for non-Level type", func(t *testing.T) {
		t.Parallel()

		a := slog.Attr{Key: slog.LevelKey, Value: slog.StringValue("not-a-level")}

		got := ReplaceLevel(nil, a)
		if got.Value.String() != "not-a-level" {
			t.Fatalf("got %q, want %q", got.Value.String(), "not-a-level")
		}
	})
}
