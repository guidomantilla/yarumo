package config

import (
	"context"
	"testing"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// fakeLogger is a no-op clog.Logger used to verify WithLogger replaces the
// default. It records whether any method was called.
type fakeLogger struct {
	called bool
}

func (f *fakeLogger) Trace(_ context.Context, _ string, _ ...any) { f.called = true }
func (f *fakeLogger) Debug(_ context.Context, _ string, _ ...any) { f.called = true }
func (f *fakeLogger) Info(_ context.Context, _ string, _ ...any)  { f.called = true }
func (f *fakeLogger) Warn(_ context.Context, _ string, _ ...any)  { f.called = true }
func (f *fakeLogger) Error(_ context.Context, _ string, _ ...any) { f.called = true }
func (f *fakeLogger) Fatal(_ context.Context, _ string, _ ...any) { f.called = true }

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("default eagerly builds the slog logger", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions("app", "1.0", "prod")
		if opts.logger == nil {
			t.Fatal("expected non-nil default logger")
		}
	})

	t.Run("applies each option in order, overriding the default", func(t *testing.T) {
		t.Parallel()

		fake := &fakeLogger{}
		opts := NewOptions("app", "1.0", "prod", WithLogger(fake))

		if opts.logger != clog.Logger(fake) {
			t.Fatalf("expected logger to be the fake, got %T", opts.logger)
		}
	})
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("replaces the default when non-nil", func(t *testing.T) {
		t.Parallel()

		fake := &fakeLogger{}
		opts := NewOptions("app", "1.0", "prod", WithLogger(fake))

		if opts.logger != clog.Logger(fake) {
			t.Fatalf("expected fake logger, got %T", opts.logger)
		}
	})

	t.Run("ignores nil, preserves the default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions("app", "1.0", "prod", WithLogger(nil))
		if opts.logger == nil {
			t.Fatal("expected default logger to remain in place when WithLogger(nil)")
		}
	})
}

// TestDefault_WithLoggerOverride verifies that WithLogger replaces the
// slog default at the Default() entry point. Cannot be parallel: mutates
// the global logger via clog.Use.
func TestDefault_WithLoggerOverride(t *testing.T) {
	t.Setenv("ENABLE_ASSERTS", "")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("DEBUG", "false")
	t.Setenv("ENABLE_CONFIG_DUMP", "")

	fake := &fakeLogger{}

	ctx := Default(context.Background(), "app", "1.0", "prod", WithLogger(fake))
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	clog.Info(ctx, "ping")

	if !fake.called {
		t.Fatal("expected the injected logger to receive the call; default was used instead")
	}
}
