package zerolog

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLevel_toZerolog(t *testing.T) {
	t.Parallel()

	t.Run("trace maps to zerolog trace", func(t *testing.T) {
		t.Parallel()

		got := LevelTrace.toZerolog()
		if got != zerolog.TraceLevel {
			t.Fatalf("got %v, want %v", got, zerolog.TraceLevel)
		}
	})

	t.Run("debug maps to zerolog debug", func(t *testing.T) {
		t.Parallel()

		got := LevelDebug.toZerolog()
		if got != zerolog.DebugLevel {
			t.Fatalf("got %v, want %v", got, zerolog.DebugLevel)
		}
	})

	t.Run("info maps to zerolog info", func(t *testing.T) {
		t.Parallel()

		got := LevelInfo.toZerolog()
		if got != zerolog.InfoLevel {
			t.Fatalf("got %v, want %v", got, zerolog.InfoLevel)
		}
	})

	t.Run("warn maps to zerolog warn", func(t *testing.T) {
		t.Parallel()

		got := LevelWarn.toZerolog()
		if got != zerolog.WarnLevel {
			t.Fatalf("got %v, want %v", got, zerolog.WarnLevel)
		}
	})

	t.Run("error maps to zerolog error", func(t *testing.T) {
		t.Parallel()

		got := LevelError.toZerolog()
		if got != zerolog.ErrorLevel {
			t.Fatalf("got %v, want %v", got, zerolog.ErrorLevel)
		}
	})

	t.Run("fatal maps to zerolog fatal", func(t *testing.T) {
		t.Parallel()

		got := LevelFatal.toZerolog()
		if got != zerolog.FatalLevel {
			t.Fatalf("got %v, want %v", got, zerolog.FatalLevel)
		}
	})

	t.Run("off maps to zerolog disabled", func(t *testing.T) {
		t.Parallel()

		got := LevelOff.toZerolog()
		if got != zerolog.Disabled {
			t.Fatalf("got %v, want %v", got, zerolog.Disabled)
		}
	})

	t.Run("unknown value maps to disabled", func(t *testing.T) {
		t.Parallel()

		got := Level(999).toZerolog()
		if got != zerolog.Disabled {
			t.Fatalf("got %v, want %v", got, zerolog.Disabled)
		}
	})
}
