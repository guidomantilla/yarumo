package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewRistrettoCache(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithBackend(BackendRistretto))
	c, err := newRistrettoCache[string, []byte](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if c.getFn == nil || c.setFn == nil || c.deleteFn == nil || c.hasFn == nil || c.clearFn == nil || c.stopFn == nil {
		t.Fatal("expected all function fields populated")
	}
	stopErr := c.Stop(context.Background())
	if stopErr != nil {
		t.Fatalf("stop failed: %v", stopErr)
	}
}

func TestNewBigcacheCache(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithBackend(BackendBigcache))
	c, err := newBigcacheCache[string, []byte](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if c.getFn == nil || c.setFn == nil || c.deleteFn == nil || c.hasFn == nil || c.clearFn == nil || c.stopFn == nil {
		t.Fatal("expected all function fields populated")
	}
	stopErr := c.Stop(context.Background())
	if stopErr != nil {
		t.Fatalf("stop failed: %v", stopErr)
	}
}

func TestNewGoCacheCache(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithBackend(BackendGoCache))
	c, err := newGoCacheCache[string, []byte](opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if c.getFn == nil || c.setFn == nil || c.deleteFn == nil || c.hasFn == nil || c.clearFn == nil || c.stopFn == nil {
		t.Fatal("expected all function fields populated")
	}
	stopErr := c.Stop(context.Background())
	if stopErr != nil {
		t.Fatalf("stop failed: %v", stopErr)
	}
}

func TestNewCache_UnsupportedBackend(t *testing.T) {
	t.Parallel()

	_, err := NewCache[string, []byte](func(o *Options) { o.backend = Backend("redis") })
	if err == nil {
		t.Fatal("expected error for unsupported backend")
	}
	if !errors.Is(err, ErrUnsupportedBackend) {
		t.Fatalf("expected ErrUnsupportedBackend, got %v", err)
	}
}

func TestEffectiveTTL(t *testing.T) {
	t.Parallel()

	t.Run("positive ttl is used", func(t *testing.T) {
		t.Parallel()

		got := effectiveTTL(2*time.Second, time.Minute)
		if got != 2*time.Second {
			t.Fatalf("got %v, want 2s", got)
		}
	})

	t.Run("zero ttl falls back to default", func(t *testing.T) {
		t.Parallel()

		got := effectiveTTL(0, time.Minute)
		if got != time.Minute {
			t.Fatalf("got %v, want 1m", got)
		}
	})

	t.Run("negative ttl falls back to default", func(t *testing.T) {
		t.Parallel()

		got := effectiveTTL(-time.Second, time.Minute)
		if got != time.Minute {
			t.Fatalf("got %v, want 1m", got)
		}
	})
}

func TestNewMetricsIfEnabled(t *testing.T) {
	t.Parallel()

	t.Run("disabled returns nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if newMetricsIfEnabled(opts) != nil {
			t.Fatal("expected nil when OTel disabled")
		}
	})

	t.Run("enabled returns adapter", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOTel())
		if newMetricsIfEnabled(opts) == nil {
			t.Fatal("expected non-nil adapter when OTel enabled")
		}
	})
}

func TestBigcache_NonByteValue(t *testing.T) {
	t.Parallel()

	c, err := NewCache[string, string](WithBackend(BackendBigcache))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	setErr := c.Set(context.Background(), "k", "not-bytes", time.Minute)
	if !errors.Is(setErr, ErrSerialization) {
		t.Fatalf("expected ErrSerialization for non-[]byte bigcache value, got %v", setErr)
	}
}
