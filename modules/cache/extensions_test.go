package cache

import (
	"context"
	"testing"
	"time"
)

func TestBuildCache(t *testing.T) {
	t.Parallel()

	t.Run("returns cache and stop fn", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c, stopFn, err := BuildCache[string, []byte](ctx, "test-cache", WithBackend(BackendGoCache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected non-nil cache")
		}
		if stopFn == nil {
			t.Fatal("expected non-nil stop fn")
		}

		err = c.Set(ctx, "k", []byte("v"), time.Minute)
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		stopFn(ctx, time.Second)
	})

	t.Run("stop is timeout-bounded", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		_, stopFn, err := BuildCache[string, []byte](ctx, "timeout-cache", WithBackend(BackendBigcache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stopFn(ctx, 100*time.Millisecond)
	})

	t.Run("invalid backend bubbles up", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		opts := []Option{
			func(o *Options) { o.backend = Backend("invalid") },
		}
		_, _, err := BuildCache[string, []byte](ctx, "bad-cache", opts...)
		if err == nil {
			t.Fatal("expected error for invalid backend")
		}
	})
}
