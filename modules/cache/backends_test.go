package cache

import (
	"errors"
	"testing"
	"time"
)

func TestBuildBackend(t *testing.T) {
	t.Parallel()

	t.Run("ristretto", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendRistretto))
		b, err := buildBackend(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b == nil || b.cache == nil || b.closer == nil {
			t.Fatal("expected complete backend instance")
		}
		closeErr := b.closer.Close()
		if closeErr != nil {
			t.Fatalf("close failed: %v", closeErr)
		}
	})

	t.Run("bigcache", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendBigcache))
		b, err := buildBackend(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b == nil || b.cache == nil || b.closer == nil {
			t.Fatal("expected complete backend instance")
		}
		closeErr := b.closer.Close()
		if closeErr != nil {
			t.Fatalf("close failed: %v", closeErr)
		}
	})

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendGoCache))
		b, err := buildBackend(opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b == nil || b.cache == nil || b.closer == nil {
			t.Fatal("expected complete backend instance")
		}
		closeErr := b.closer.Close()
		if closeErr != nil {
			t.Fatalf("close failed: %v", closeErr)
		}
	})

	t.Run("unsupported backend returns error", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		opts.backend = Backend("redis")
		_, err := buildBackend(opts)
		if err == nil {
			t.Fatal("expected error for unsupported backend")
		}
		if !errors.Is(err, ErrUnsupportedBackend) {
			t.Fatal("expected ErrUnsupportedBackend")
		}
	})

	t.Run("nil options returns error", func(t *testing.T) {
		t.Parallel()

		_, err := buildBackend(nil)
		if err == nil {
			t.Fatal("expected error for nil options")
		}
	})
}

func TestSetOptionsForTTL(t *testing.T) {
	t.Parallel()

	t.Run("positive ttl is used", func(t *testing.T) {
		t.Parallel()

		got := setOptionsForTTL(2*time.Second, time.Minute)
		if len(got) == 0 {
			t.Fatal("expected non-empty options")
		}
	})

	t.Run("non-positive ttl falls back to default", func(t *testing.T) {
		t.Parallel()

		got := setOptionsForTTL(0, time.Minute)
		if len(got) == 0 {
			t.Fatal("expected non-empty options")
		}
	})

	t.Run("negative ttl falls back to default", func(t *testing.T) {
		t.Parallel()

		got := setOptionsForTTL(-time.Second, time.Minute)
		if len(got) == 0 {
			t.Fatal("expected non-empty options")
		}
	})
}

func TestCloserFn(t *testing.T) {
	t.Parallel()

	called := false
	c := closerFn(func() error {
		called = true
		return nil
	})
	err := c.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected closer to be invoked")
	}
}

func TestNoopCloser(t *testing.T) {
	t.Parallel()

	c := noopCloser{}
	err := c.Close()
	if err != nil {
		t.Fatalf("unexpected error from noop close: %v", err)
	}
}
