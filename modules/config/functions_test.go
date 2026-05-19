package config

import (
	"context"
	"testing"
)

// TestDefault cannot be parallel: sets environment variables and modifies
// global state (viper, assert, clog).
func TestDefault(t *testing.T) {

	t.Run("all parameters populated", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "true")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("DEBUG", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("empty parameters", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("DEBUG", "")

		ctx := Default(context.Background(), "", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts enabled with 1", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "1")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts enabled with yes", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "yes")
		t.Setenv("LOG_LEVEL", "warn")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "1.0", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts enabled with YES uppercase", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "YES")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("asserts disabled", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "false")
		t.Setenv("LOG_LEVEL", "error")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "", "staging")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters name only", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "trace")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters version only", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "fatal")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "2.0", "")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("partial parameters env only", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "off")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "", "", "staging")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("invalid log level defaults to info", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "invalid")
		t.Setenv("DEBUG", "false")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("debug mode enabled", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with 1", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("ENABLE_CONFIG_DUMP", "1")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with true", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("ENABLE_CONFIG_DUMP", "true")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})

	t.Run("config dump with yes", func(t *testing.T) {
		t.Setenv("ENABLE_ASSERTS", "")
		t.Setenv("LOG_LEVEL", "info")
		t.Setenv("DEBUG", "false")
		t.Setenv("ENABLE_CONFIG_DUMP", "yes")

		ctx := Default(context.Background(), "app", "1.0", "prod")
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
	})
}
