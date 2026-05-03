package config

import (
	"context"
	"testing"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
)

// TestDefault cannot be parallel: sets environment variables and modifies
// global state (viper, assert, clog).
func TestDefault(t *testing.T) {

	t.Run("all parameters populated", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "true")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("DEBUG", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("empty parameters", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("DEBUG", "")

		ctx := Default(context.Background(), "", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")

		ctx := Default(nil, "app", "1.0", "prod") //nolint:staticcheck // testing nil context edge case
		if ctx != nil {
			t.Fatal("expected nil context returned for nil input")
		}
	})

	t.Run("asserts enabled with 1", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "1")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts enabled with yes", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "yes")
		t.Setenv("LOG_LEVEL", "warn")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "1.0", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts enabled with YES uppercase", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "YES")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts disabled", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "false")
		t.Setenv("LOG_LEVEL", "error")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "", "staging")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters name only", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "trace")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters version only", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "fatal")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "2.0", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters env only", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "off")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "", "staging")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("invalid log level defaults to info", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "invalid")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("debug mode enabled", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with 1", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("CONFIG_DUMP", "1")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with true", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("CONFIG_DUMP", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with yes", func(t *testing.T) {
		t.Setenv("ASSERTS_ENABLED", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("CONFIG_DUMP", "yes")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})
}

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
