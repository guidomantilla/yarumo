package zerolog

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("default values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.level != LevelOff {
			t.Fatalf("got level %v, want %v", opts.level, LevelOff)
		}

		if opts.writer != os.Stderr {
			t.Fatalf("got writer %v, want os.Stderr", opts.writer)
		}

		if opts.console {
			t.Fatal("got console=true, want false")
		}

		if opts.timeFormat != DefaultTimeFormat {
			t.Fatalf("got timeFormat %q, want %q", opts.timeFormat, DefaultTimeFormat)
		}

		if opts.sampling != DefaultSampling {
			t.Fatalf("got sampling %d, want %d", opts.sampling, DefaultSampling)
		}
	})

	t.Run("options applied", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		opts := NewOptions(
			WithLevel(LevelDebug),
			WithWriter(buf),
			WithConsole(true),
			WithTimeFormat(time.RFC3339),
			WithSampling(10),
		)
		if opts.level != LevelDebug {
			t.Fatalf("got level %v, want %v", opts.level, LevelDebug)
		}

		if opts.writer != buf {
			t.Fatal("writer was not set")
		}

		if !opts.console {
			t.Fatal("console was not set")
		}

		if opts.timeFormat != time.RFC3339 {
			t.Fatalf("got timeFormat %q, want %q", opts.timeFormat, time.RFC3339)
		}

		if opts.sampling != 10 {
			t.Fatalf("got sampling %d, want 10", opts.sampling)
		}
	})
}

func TestWithLevel(t *testing.T) {
	t.Parallel()

	t.Run("valid level applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithLevel(LevelTrace))
		if opts.level != LevelTrace {
			t.Fatalf("got level %v, want %v", opts.level, LevelTrace)
		}
	})

	t.Run("invalid level ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithLevel(Level(999)))
		if opts.level != LevelOff {
			t.Fatalf("got level %v, want %v (default)", opts.level, LevelOff)
		}
	})
}

func TestWithWriter(t *testing.T) {
	t.Parallel()

	t.Run("non-nil writer applied", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		opts := NewOptions(WithWriter(buf))
		if opts.writer != buf {
			t.Fatal("writer was not set")
		}
	})

	t.Run("nil writer ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithWriter(nil))
		if opts.writer != os.Stderr {
			t.Fatal("nil writer should be ignored, want os.Stderr")
		}
	})
}

func TestWithConsole(t *testing.T) {
	t.Parallel()

	t.Run("enables console mode", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConsole(true))
		if !opts.console {
			t.Fatal("console mode was not enabled")
		}
	})

	t.Run("disables console mode explicitly", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConsole(true), WithConsole(false))
		if opts.console {
			t.Fatal("console mode should have been disabled")
		}
	})
}

func TestWithTimeFormat(t *testing.T) {
	t.Parallel()

	t.Run("non-empty format applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeFormat(time.RFC822))
		if opts.timeFormat != time.RFC822 {
			t.Fatalf("got %q, want %q", opts.timeFormat, time.RFC822)
		}
	})

	t.Run("empty format ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeFormat(""))
		if opts.timeFormat != DefaultTimeFormat {
			t.Fatalf("got %q, want default %q", opts.timeFormat, DefaultTimeFormat)
		}
	})
}

func TestWithSampling(t *testing.T) {
	t.Parallel()

	t.Run("positive value applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSampling(5))
		if opts.sampling != 5 {
			t.Fatalf("got %d, want 5", opts.sampling)
		}
	})

	t.Run("zero value ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithSampling(0))
		if opts.sampling != DefaultSampling {
			t.Fatalf("got %d, want default %d", opts.sampling, DefaultSampling)
		}
	})
}

func TestOptions_effectiveWriter(t *testing.T) {
	t.Parallel()

	t.Run("json mode returns raw writer", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		opts := NewOptions(WithWriter(buf))
		if opts.effectiveWriter() != buf {
			t.Fatal("expected raw writer in JSON mode")
		}
	})

	t.Run("console mode wraps writer", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		opts := NewOptions(WithWriter(buf), WithConsole(true))
		if opts.effectiveWriter() == buf {
			t.Fatal("expected console writer wrapper, got raw writer")
		}
	})
}
