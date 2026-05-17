package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	t.Parallel()

	t.Run("returns a usable Cache", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")
		if c == nil {
			t.Fatal("expected non-nil Cache")
		}

		ctx := context.Background()

		err := c.Set(ctx, "k", "v", 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		got, getErr := c.Get(ctx, "k")
		if getErr != nil {
			t.Fatalf("Get: %v", getErr)
		}

		if got != "v" {
			t.Fatalf("Get = %q, want %q", got, "v")
		}
	})

	t.Run("exposes the name passed at construction", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("my-name")
		if c.Name() != "my-name" {
			t.Fatalf("Name() = %q, want %q", c.Name(), "my-name")
		}
	})
}

func TestMemoryCache_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the configured name", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("alpha")
		if c.Name() != "alpha" {
			t.Fatalf("Name() = %q, want %q", c.Name(), "alpha")
		}
	})
}

func TestMemoryCache_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns stored value", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, int]("test")

		err := c.Set(ctx, "k", 42, 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		got, getErr := c.Get(ctx, "k")
		if getErr != nil {
			t.Fatalf("Get: %v", getErr)
		}

		if got != 42 {
			t.Fatalf("Get = %d, want 42", got)
		}
	})

	t.Run("returns ErrCacheMiss when key absent", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, int]("test")

		got, err := c.Get(ctx, "missing")
		if err == nil {
			t.Fatal("expected error for missing key")
		}

		if !errors.Is(err, ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}

		if got != 0 {
			t.Fatalf("expected zero value on miss, got %d", got)
		}
	})

	t.Run("returns ErrCacheTypeAssertion when stored type does not match V", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		impl, ok := c.(*memoryCache[string, string])
		if !ok {
			t.Fatalf("expected *memoryCache, got %T", c)
		}

		impl.data.Store("k", 123)

		got, err := c.Get(ctx, "k")
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}

		if got != "" {
			t.Fatalf("expected zero value on cast fail, got %q", got)
		}
	})
}

func TestMemoryCache_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("stores value", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Set(ctx, "k", "v", 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("expected key to be present after Set")
		}
	})

	t.Run("ignores ttl", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Set(ctx, "k", "v", time.Nanosecond)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("expected key to persist past ttl (memoryCache ignores ttl)")
		}
	})
}

func TestMemoryCache_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes stored value", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Set(ctx, "k", "v", 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		delErr := c.Delete(ctx, "k")
		if delErr != nil {
			t.Fatalf("Delete: %v", delErr)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("expected key to be absent after Delete")
		}
	})

	t.Run("is a no-op when key absent", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Delete(ctx, "missing")
		if err != nil {
			t.Fatalf("Delete on missing key: %v", err)
		}
	})
}

func TestMemoryCache_Has(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns true for present key", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Set(ctx, "k", "v", 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("Has = false, want true")
		}
	})

	t.Run("returns false for absent key", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		has, hasErr := c.Has(ctx, "missing")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("Has = true, want false")
		}
	})
}

func TestMemoryCache_Clear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes every entry", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Set(ctx, "k1", "v1", 0)
		if err != nil {
			t.Fatalf("Set k1: %v", err)
		}

		err = c.Set(ctx, "k2", "v2", 0)
		if err != nil {
			t.Fatalf("Set k2: %v", err)
		}

		clearErr := c.Clear(ctx)
		if clearErr != nil {
			t.Fatalf("Clear: %v", clearErr)
		}

		has1, _ := c.Has(ctx, "k1")
		has2, _ := c.Has(ctx, "k2")

		if has1 || has2 {
			t.Fatal("expected cache to be empty after Clear")
		}
	})
}

func TestMemoryCache_Stop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("is a no-op and safe to call more than once", func(t *testing.T) {
		t.Parallel()

		c := NewMemoryCache[string, string]("test")

		err := c.Stop(ctx)
		if err != nil {
			t.Fatalf("first Stop: %v", err)
		}

		err = c.Stop(ctx)
		if err != nil {
			t.Fatalf("second Stop: %v", err)
		}
	})
}
