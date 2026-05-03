package log

import (
	"context"
	"testing"
)

// spyLogger tracks which method and message were last called.
type spyLogger struct {
	method string
	msg    string
}

func (s *spyLogger) Trace(_ context.Context, msg string, _ ...any) {
	s.method = "Trace"
	s.msg = msg
}

func (s *spyLogger) Debug(_ context.Context, msg string, _ ...any) {
	s.method = "Debug"
	s.msg = msg
}

func (s *spyLogger) Info(_ context.Context, msg string, _ ...any) {
	s.method = "Info"
	s.msg = msg
}

func (s *spyLogger) Warn(_ context.Context, msg string, _ ...any) {
	s.method = "Warn"
	s.msg = msg
}

func (s *spyLogger) Error(_ context.Context, msg string, _ ...any) {
	s.method = "Error"
	s.msg = msg
}

func (s *spyLogger) Fatal(_ context.Context, msg string, _ ...any) {
	s.method = "Fatal"
	s.msg = msg
}

// withLogger temporarily replaces the global logger and restores it after fn returns.
func withLogger(l Logger, fn func()) {
	orig := load()

	Use(l)

	defer Use(orig)

	fn()
}

func TestUse(t *testing.T) {
	t.Run("replaces the global logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			got := load()
			if got != spy {
				t.Fatal("expected custom logger to be set")
			}
		})
	})

	t.Run("nil is ignored", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Use(nil)

			got := load()
			if got != spy {
				t.Fatal("expected logger to remain unchanged after Use(nil)")
			}
		})
	})
}

func TestTrace(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Trace(context.Background(), "trace-msg")

			if spy.method != "Trace" {
				t.Fatalf("got method %q, want %q", spy.method, "Trace")
			}

			if spy.msg != "trace-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "trace-msg")
			}
		})
	})
}

func TestDebug(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Debug(context.Background(), "debug-msg")

			if spy.method != "Debug" {
				t.Fatalf("got method %q, want %q", spy.method, "Debug")
			}

			if spy.msg != "debug-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "debug-msg")
			}
		})
	})
}

func TestInfo(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Info(context.Background(), "info-msg")

			if spy.method != "Info" {
				t.Fatalf("got method %q, want %q", spy.method, "Info")
			}

			if spy.msg != "info-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "info-msg")
			}
		})
	})
}

func TestWarn(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Warn(context.Background(), "warn-msg")

			if spy.method != "Warn" {
				t.Fatalf("got method %q, want %q", spy.method, "Warn")
			}

			if spy.msg != "warn-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "warn-msg")
			}
		})
	})
}

func TestError(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Error(context.Background(), "error-msg")

			if spy.method != "Error" {
				t.Fatalf("got method %q, want %q", spy.method, "Error")
			}

			if spy.msg != "error-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "error-msg")
			}
		})
	})
}

func TestFatal(t *testing.T) {
	t.Run("delegates to current logger", func(t *testing.T) {
		spy := &spyLogger{}
		withLogger(spy, func() {
			Fatal(context.Background(), "fatal-msg")

			if spy.method != "Fatal" {
				t.Fatalf("got method %q, want %q", spy.method, "Fatal")
			}

			if spy.msg != "fatal-msg" {
				t.Fatalf("got msg %q, want %q", spy.msg, "fatal-msg")
			}
		})
	})
}
