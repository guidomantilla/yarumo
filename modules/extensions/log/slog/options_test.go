package slog

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
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

		if opts.handlers != nil {
			t.Fatalf("got handlers %v, want nil", opts.handlers)
		}
	})

	t.Run("options applied", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}

		opts := NewOptions(WithLevel(LevelDebug), WithWriter(buf))
		if opts.level != LevelDebug {
			t.Fatalf("got level %v, want %v", opts.level, LevelDebug)
		}

		if opts.writer != buf {
			t.Fatal("writer was not set")
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

func TestWithHandlers(t *testing.T) {
	t.Parallel()

	t.Run("handlers appended", func(t *testing.T) {
		t.Parallel()

		h := slog.NewJSONHandler(os.Stderr, nil)

		opts := NewOptions(WithHandlers(h))
		if len(opts.handlers) != 1 {
			t.Fatalf("got %d handlers, want 1", len(opts.handlers))
		}
	})

	t.Run("multiple calls accumulate", func(t *testing.T) {
		t.Parallel()

		h1 := slog.NewJSONHandler(os.Stderr, nil)
		h2 := slog.NewJSONHandler(os.Stderr, nil)

		opts := NewOptions(WithHandlers(h1), WithHandlers(h2))
		if len(opts.handlers) != 2 {
			t.Fatalf("got %d handlers, want 2", len(opts.handlers))
		}
	})

	t.Run("empty call ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithHandlers())
		if opts.handlers != nil {
			t.Fatalf("got handlers %v, want nil", opts.handlers)
		}
	})
}
