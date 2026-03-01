package slog

import (
	"bytes"
	"context"
	"log/slog"
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

	t.Run("default creates json handler", func(t *testing.T) {
		t.Parallel()

		l := NewLogger()
		if l == nil {
			t.Fatal("expected non-nil logger")
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

	t.Run("with custom handlers", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		h := &testHandler{buf: buf, enabled: true}
		l := NewLogger(WithHandlers(h))
		l.Info(context.Background(), "custom")

		if !strings.Contains(buf.String(), "custom") {
			t.Fatalf("output %q does not contain %q", buf.String(), "custom")
		}
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

// testHandler is a minimal slog.Handler for testing.
type testHandler struct {
	buf     *bytes.Buffer
	enabled bool
}

func (h *testHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *testHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := h.buf.WriteString(r.Message)

	return err
}

func (h *testHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(_ string) slog.Handler {
	return h
}
