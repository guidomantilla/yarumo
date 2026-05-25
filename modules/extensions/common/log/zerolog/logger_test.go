package zerolog

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// withOsExit temporarily replaces osExit and restores it after fn returns.
func withOsExit(temp func(code int), fn func()) {
	orig := osExit
	osExit = temp

	defer func() { osExit = orig }()

	fn()
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("default returns non-nil logger", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
	})

	t.Run("default level is silent", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf))
		l.Info(context.Background(), "should-not-appear")

		if buf.Len() != 0 {
			t.Fatalf("expected silent default, got %q", buf.String())
		}
	})

	t.Run("with custom writer and level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "test-msg")

		if !strings.Contains(buf.String(), "test-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "test-msg")
		}
	})

	t.Run("console mode produces output", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo), WithConsole(true))
		l.Info(context.Background(), "console-msg")

		if !strings.Contains(buf.String(), "console-msg") {
			t.Fatalf("console output %q does not contain %q", buf.String(), "console-msg")
		}
	})

	t.Run("sampling configured does not break logger", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo), WithSampling(1000))
		// With a 1-in-1000 sampler the chance of any single log appearing
		// is low but non-zero; we only assert the logger does not panic.
		l.Info(context.Background(), "sampled")
	})
}

func TestLogger_Trace(t *testing.T) {
	t.Parallel()

	t.Run("logs at trace level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelTrace))
		l.Trace(context.Background(), "trace-msg")

		if !strings.Contains(buf.String(), "trace-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "trace-msg")
		}
	})

	t.Run("filtered out below level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Trace(context.Background(), "trace-msg")

		if buf.Len() != 0 {
			t.Fatalf("expected empty buffer, got %q", buf.String())
		}
	})
}

func TestLogger_Debug(t *testing.T) {
	t.Parallel()

	t.Run("logs at debug level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelDebug))
		l.Debug(context.Background(), "debug-msg")

		if !strings.Contains(buf.String(), "debug-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "debug-msg")
		}
	})
}

func TestLogger_Info(t *testing.T) {
	t.Parallel()

	t.Run("logs at info level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "info-msg")

		if !strings.Contains(buf.String(), "info-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "info-msg")
		}
	})

	t.Run("args become structured fields", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "with-args", "user_id", 42, "request_id", "abc")

		out := buf.String()
		if !strings.Contains(out, `"user_id":42`) {
			t.Fatalf("output %q does not contain user_id field", out)
		}

		if !strings.Contains(out, `"request_id":"abc"`) {
			t.Fatalf("output %q does not contain request_id field", out)
		}
	})

	t.Run("empty args are a no-op", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		l.Info(context.Background(), "no-args")

		if !strings.Contains(buf.String(), "no-args") {
			t.Fatalf("output %q does not contain %q", buf.String(), "no-args")
		}
	})

	t.Run("nil context is tolerated", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))
		//nolint:staticcheck // explicitly testing nil-context tolerance.
		l.Info(nil, "nil-ctx")

		if !strings.Contains(buf.String(), "nil-ctx") {
			t.Fatalf("output %q does not contain %q", buf.String(), "nil-ctx")
		}
	})

	t.Run("cancelled context surfaces ctx_err", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelInfo))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		l.Info(ctx, "cancelled")

		if !strings.Contains(buf.String(), "ctx_err") {
			t.Fatalf("output %q does not contain ctx_err", buf.String())
		}
	})
}

func TestLogger_Warn(t *testing.T) {
	t.Parallel()

	t.Run("logs at warn level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelWarn))
		l.Warn(context.Background(), "warn-msg")

		if !strings.Contains(buf.String(), "warn-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "warn-msg")
		}
	})
}

func TestLogger_Error(t *testing.T) {
	t.Parallel()

	t.Run("logs at error level", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		l := NewLogger(WithWriter(buf), WithLevel(LevelError))
		l.Error(context.Background(), "error-msg")

		if !strings.Contains(buf.String(), "error-msg") {
			t.Fatalf("output %q does not contain %q", buf.String(), "error-msg")
		}
	})
}

func TestLogger_Fatal(t *testing.T) {
	t.Run("logs and calls os exit", func(t *testing.T) {
		var exitCode int

		withOsExit(func(code int) {
			exitCode = code
		}, func() {
			buf := &bytes.Buffer{}
			l := NewLogger(WithWriter(buf), WithLevel(LevelFatal))
			l.Fatal(context.Background(), "fatal-msg")

			if !strings.Contains(buf.String(), "fatal-msg") {
				t.Fatalf("output %q does not contain %q", buf.String(), "fatal-msg")
			}
		})

		if exitCode != 1 {
			t.Fatalf("got exit code %d, want 1", exitCode)
		}
	})
}
