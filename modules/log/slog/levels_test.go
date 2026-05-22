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
