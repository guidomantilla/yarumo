package config

import (
	"context"
	"testing"

	cslog "github.com/guidomantilla/yarumo/log/slog"
)

// Test_dump cannot be parallel: calls clog.Info which uses global logger state.
func Test_dump(t *testing.T) {

	t.Run("logs environment variables", func(t *testing.T) {
		t.Setenv("TEST_VAR", "hello")
		t.Setenv("TEST_SECRET", "mysecret")

		dump(context.Background())
	})
}

func Test_shouldMask(t *testing.T) {
	t.Parallel()

	t.Run("key containing PASSWORD", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("DB_PASSWORD") {
			t.Fatal("expected true for key containing PASSWORD")
		}
	})

	t.Run("key containing SECRET", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("APP_SECRET") {
			t.Fatal("expected true for key containing SECRET")
		}
	})

	t.Run("key containing TOKEN", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("AUTH_TOKEN") {
			t.Fatal("expected true for key containing TOKEN")
		}
	})

	t.Run("key containing KEY", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("API_KEY") {
			t.Fatal("expected true for key containing KEY")
		}
	})

	t.Run("key containing CREDENTIAL", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("AWS_CREDENTIAL") {
			t.Fatal("expected true for key containing CREDENTIAL")
		}
	})

	t.Run("key containing PRIVATE", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("PRIVATE_DATA") {
			t.Fatal("expected true for key containing PRIVATE")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		t.Parallel()

		if !shouldMask("db_password") {
			t.Fatal("expected true for lowercase key containing password")
		}
	})

	t.Run("non-sensitive key", func(t *testing.T) {
		t.Parallel()

		if shouldMask("LOG_LEVEL") {
			t.Fatal("expected false for non-sensitive key")
		}
	})
}

func Test_maskValue(t *testing.T) {
	t.Parallel()

	t.Run("masks non-empty value", func(t *testing.T) {
		t.Parallel()

		got := maskValue("supersecret")
		if got != maskedValue {
			t.Fatalf("expected \"********\", got %q", got)
		}
	})

	t.Run("masks empty value", func(t *testing.T) {
		t.Parallel()

		got := maskValue("")
		if got != maskedValue {
			t.Fatalf("expected \"********\", got %q", got)
		}
	})
}

func Test_parseLevel(t *testing.T) {
	t.Parallel()

	t.Run("trace", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("trace")
		if got != cslog.LevelTrace {
			t.Fatalf("expected LevelTrace (%d), got %d", cslog.LevelTrace, got)
		}
	})

	t.Run("debug", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("debug")
		if got != cslog.LevelDebug {
			t.Fatalf("expected LevelDebug (%d), got %d", cslog.LevelDebug, got)
		}
	})

	t.Run("info", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("info")
		if got != cslog.LevelInfo {
			t.Fatalf("expected LevelInfo (%d), got %d", cslog.LevelInfo, got)
		}
	})

	t.Run("warn", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("warn")
		if got != cslog.LevelWarn {
			t.Fatalf("expected LevelWarn (%d), got %d", cslog.LevelWarn, got)
		}
	})

	t.Run("warning", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("warning")
		if got != cslog.LevelWarn {
			t.Fatalf("expected LevelWarn (%d), got %d", cslog.LevelWarn, got)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("error")
		if got != cslog.LevelError {
			t.Fatalf("expected LevelError (%d), got %d", cslog.LevelError, got)
		}
	})

	t.Run("fatal", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("fatal")
		if got != cslog.LevelFatal {
			t.Fatalf("expected LevelFatal (%d), got %d", cslog.LevelFatal, got)
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("off")
		if got != cslog.LevelOff {
			t.Fatalf("expected LevelOff (%d), got %d", cslog.LevelOff, got)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("disabled")
		if got != cslog.LevelOff {
			t.Fatalf("expected LevelOff (%d), got %d", cslog.LevelOff, got)
		}
	})

	t.Run("unknown defaults to info", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("unknown")
		if got != cslog.LevelInfo {
			t.Fatalf("expected LevelInfo (%d), got %d", cslog.LevelInfo, got)
		}
	})

	t.Run("empty defaults to info", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("")
		if got != cslog.LevelInfo {
			t.Fatalf("expected LevelInfo (%d), got %d", cslog.LevelInfo, got)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		t.Parallel()

		got := parseLevel("DEBUG")
		if got != cslog.LevelDebug {
			t.Fatalf("expected LevelDebug (%d), got %d", cslog.LevelDebug, got)
		}
	})
}
