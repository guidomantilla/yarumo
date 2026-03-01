package assert

import (
	"context"
	"testing"

	clog "github.com/guidomantilla/yarumo/common/log"
	cslog "github.com/guidomantilla/yarumo/common/log/slog"
)

var _ clog.Logger = (*spyLogger)(nil)

const (
	methodError = "Error"
	methodFatal = "Fatal"
)

// spyLogger tracks which log method and message were last called.
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
	s.method = methodError
	s.msg = msg
}

func (s *spyLogger) Fatal(_ context.Context, msg string, _ ...any) {
	s.method = methodFatal
	s.msg = msg
}

// withSpy temporarily replaces the global logger with a spy and restores it after fn returns.
func withSpy(fn func(spy *spyLogger)) {
	spy := &spyLogger{}

	clog.Use(spy)

	defer clog.Use(cslog.NewLogger())

	fn(spy)
}

func TestEnable(t *testing.T) {
	t.Run("sets enabled to true", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		if !enabled.Load() {
			t.Fatal("expected enabled to be true")
		}
	})

	t.Run("sets enabled to false", func(t *testing.T) {
		Enable(true)
		Enable(false)

		if enabled.Load() {
			t.Fatal("expected enabled to be false")
		}
	})
}

func TestNotEmpty(t *testing.T) {
	t.Run("does not log when not empty", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			NotEmpty("hello", "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when empty and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			NotEmpty("", "empty-object")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when empty and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			NotEmpty("", "empty-object")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}

func TestNotNil(t *testing.T) {
	t.Run("does not log when not nil", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			NotNil("hello", "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when nil and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			NotNil(nil, "nil-object")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when nil and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			NotNil(nil, "nil-object")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}

func TestEqual(t *testing.T) {
	t.Run("does not log when equal", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			Equal(1, 1, "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when not equal and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			Equal(1, 2, "not-equal")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when not equal and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			Equal(1, 2, "not-equal")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}

func TestNotEqual(t *testing.T) {
	t.Run("does not log when not equal", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			NotEqual(1, 2, "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when equal and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			NotEqual(1, 1, "are-equal")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when equal and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			NotEqual(1, 1, "are-equal")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}

func TestTrue(t *testing.T) {
	t.Run("does not log when true", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			True(true, "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when false and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			True(false, "is-false")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when false and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			True(false, "is-false")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}

func TestFalse(t *testing.T) {
	t.Run("does not log when false", func(t *testing.T) {
		withSpy(func(spy *spyLogger) {
			False(false, "should not fail")

			if spy.method != "" {
				t.Fatalf("expected no log call, got %q", spy.method)
			}
		})
	})

	t.Run("logs error when true and disabled", func(t *testing.T) {
		Enable(false)
		withSpy(func(spy *spyLogger) {
			False(true, "is-true")

			if spy.method != methodError {
				t.Fatalf("got method %q, want %q", spy.method, methodError)
			}
		})
	})

	t.Run("logs fatal when true and enabled", func(t *testing.T) {
		Enable(true)

		defer Enable(false)

		withSpy(func(spy *spyLogger) {
			False(true, "is-true")

			if spy.method != methodFatal {
				t.Fatalf("got method %q, want %q", spy.method, methodFatal)
			}
		})
	})
}
