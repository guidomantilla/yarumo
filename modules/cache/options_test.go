package cache

import (
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.backend != BackendRistretto {
			t.Fatalf("got backend %q, want ristretto", opts.backend)
		}
		if opts.ttl <= 0 {
			t.Fatalf("expected positive default ttl, got %v", opts.ttl)
		}
		if opts.otelEnabled {
			t.Fatal("OTel should be disabled by default")
		}
		if opts.slogEnabled {
			t.Fatal("slog should be disabled by default")
		}
	})

	t.Run("applies options in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendBigcache), WithTTL(30*time.Second), WithOTel(), WithSlog())
		if opts.backend != BackendBigcache {
			t.Fatalf("got backend %q, want bigcache", opts.backend)
		}
		if opts.ttl != 30*time.Second {
			t.Fatalf("got ttl %v, want 30s", opts.ttl)
		}
		if !opts.otelEnabled {
			t.Fatal("expected OTel enabled")
		}
		if !opts.slogEnabled {
			t.Fatal("expected slog enabled")
		}
	})
}

func TestWithBackend(t *testing.T) {
	t.Parallel()

	t.Run("ristretto", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendRistretto))
		if opts.backend != BackendRistretto {
			t.Fatalf("got %q, want ristretto", opts.backend)
		}
	})

	t.Run("bigcache", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendBigcache))
		if opts.backend != BackendBigcache {
			t.Fatalf("got %q, want bigcache", opts.backend)
		}
	})

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(BackendGoCache))
		if opts.backend != BackendGoCache {
			t.Fatalf("got %q, want go-cache", opts.backend)
		}
	})

	t.Run("unknown backend is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackend(Backend("redis")))
		if opts.backend != BackendRistretto {
			t.Fatalf("got %q, want ristretto fallback", opts.backend)
		}
	})
}

func TestWithTTL(t *testing.T) {
	t.Parallel()

	t.Run("positive is applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(time.Second))
		if opts.ttl != time.Second {
			t.Fatalf("got %v, want 1s", opts.ttl)
		}
	})

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(0))
		if opts.ttl <= 0 {
			t.Fatal("expected positive ttl preserved")
		}
	})

	t.Run("negative is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(-time.Second))
		if opts.ttl <= 0 {
			t.Fatal("expected positive ttl preserved")
		}
	})
}

func TestWithOTel(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithOTel())
	if !opts.otelEnabled {
		t.Fatal("expected OTel enabled")
	}
}

func TestWithOTelMeterName(t *testing.T) {
	t.Parallel()

	t.Run("custom name is applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOTelMeterName("my-cache"))
		if opts.otelMeterName != "my-cache" {
			t.Fatalf("got %q, want my-cache", opts.otelMeterName)
		}
	})

	t.Run("empty name is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOTelMeterName(""))
		if opts.otelMeterName == "" {
			t.Fatal("expected default meter name preserved")
		}
	})
}

func TestWithSlog(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithSlog())
	if !opts.slogEnabled {
		t.Fatal("expected slog enabled")
	}
}

func TestWithRistrettoCapacity(t *testing.T) {
	t.Parallel()

	t.Run("positive values are applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(10, 1000, 4))
		if opts.ristrettoNumCtrs != 10 {
			t.Fatalf("got num counters %d, want 10", opts.ristrettoNumCtrs)
		}
		if opts.ristrettoMaxCost != 1000 {
			t.Fatalf("got max cost %d, want 1000", opts.ristrettoMaxCost)
		}
		if opts.ristrettoBufItems != 4 {
			t.Fatalf("got buffer items %d, want 4", opts.ristrettoBufItems)
		}
	})

	t.Run("non-positive values are ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(0, -1, 0))
		if opts.ristrettoNumCtrs <= 0 || opts.ristrettoMaxCost <= 0 || opts.ristrettoBufItems <= 0 {
			t.Fatal("expected defaults preserved")
		}
	})
}

func TestWithBigcacheCapacity(t *testing.T) {
	t.Parallel()

	t.Run("positive values are applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBigcacheCapacity(8, time.Second, time.Second, 16, 1024))
		if opts.bigcacheShards != 8 {
			t.Fatalf("got shards %d, want 8", opts.bigcacheShards)
		}
		if opts.bigcacheLifeWin != time.Second {
			t.Fatalf("got life window %v, want 1s", opts.bigcacheLifeWin)
		}
		if opts.bigcacheCleanWin != time.Second {
			t.Fatalf("got clean window %v, want 1s", opts.bigcacheCleanWin)
		}
		if opts.bigcacheMaxSize != 16 {
			t.Fatalf("got max size %d, want 16", opts.bigcacheMaxSize)
		}
		if opts.bigcacheMaxEntry != 1024 {
			t.Fatalf("got max entry %d, want 1024", opts.bigcacheMaxEntry)
		}
	})

	t.Run("non-positive values are ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBigcacheCapacity(0, 0, 0, 0, 0))
		if opts.bigcacheShards <= 0 {
			t.Fatal("expected default shards preserved")
		}
	})
}

func TestWithGoCacheCapacity(t *testing.T) {
	t.Parallel()

	t.Run("positive values are applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGoCacheCapacity(2*time.Second, 3*time.Second))
		if opts.gocacheDefault != 2*time.Second {
			t.Fatalf("got default %v, want 2s", opts.gocacheDefault)
		}
		if opts.gocacheCleanup != 3*time.Second {
			t.Fatalf("got cleanup %v, want 3s", opts.gocacheCleanup)
		}
	})

	t.Run("non-positive values are ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGoCacheCapacity(0, 0))
		if opts.gocacheDefault <= 0 || opts.gocacheCleanup <= 0 {
			t.Fatal("expected defaults preserved")
		}
	})
}
