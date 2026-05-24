package ristretto

import (
	"context"
	"errors"
	"testing"
	"time"

	ccache "github.com/guidomantilla/yarumo/common/cache"
	lctests "github.com/guidomantilla/yarumo/common/lifecycle/tests"
)

func newTestRistrettoCache(t *testing.T) ccache.Cache[string, string] {
	t.Helper()

	c := NewRistrettoCache[string]("test")

	err := c.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	t.Cleanup(func() { _ = c.Stop(context.Background()) })

	return c
}

func TestNewRistrettoCache(t *testing.T) {
	t.Parallel()

	t.Run("returns a usable cache after Start", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)
		if c == nil {
			t.Fatal("expected non-nil cache")
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

	t.Run("constructor itself does no I/O and cannot fail", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("ctor-only")
		if c == nil {
			t.Fatal("expected non-nil cache from constructor")
		}

		t.Cleanup(func() { _ = c.Stop(context.Background()) })
	})
}

func TestRistrettoCache_Start(t *testing.T) {
	t.Parallel()

	t.Run("succeeds with default configuration", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("start-ok")
		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		err := c.Start(context.Background())
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
	})
}

func TestRistrettoCache_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the configured name", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("alpha")
		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		if c.Name() != "alpha" {
			t.Fatalf("Name() = %q, want %q", c.Name(), "alpha")
		}
	})
}

func TestRistrettoCache_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns stored value on hit", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

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

	t.Run("returns ErrCacheMiss when key absent", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

		got, err := c.Get(ctx, "missing")
		if err == nil {
			t.Fatal("expected error for missing key")
		}

		if !errors.Is(err, ccache.ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}

		if got != "" {
			t.Fatalf("expected zero value on miss, got %q", got)
		}
	})
}

func TestRistrettoCache_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("stores value with per-call ttl", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

		err := c.Set(ctx, "k", "v", 10*time.Second)
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

	t.Run("falls back to default ttl when per-call ttl is non-positive", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("test-default-ttl", WithTTL(10*time.Second))

		err := c.Start(ctx)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("expected key to be present after Set with default ttl")
		}
	})
}

func TestRistrettoCache_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes stored value", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
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

		c := newTestRistrettoCache(t)

		err := c.Delete(ctx, "missing")
		if err != nil {
			t.Fatalf("Delete on missing key: %v", err)
		}
	})
}

func TestRistrettoCache_Has(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns true for present key", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

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

		c := newTestRistrettoCache(t)

		has, hasErr := c.Has(ctx, "missing")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("Has = true, want false")
		}
	})
}

func TestRistrettoCache_Clear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes every entry", func(t *testing.T) {
		t.Parallel()

		c := newTestRistrettoCache(t)

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

func TestRistrettoCache_StopIsIdempotent(t *testing.T) {
	t.Parallel()

	c := NewRistrettoCache[string]("idempotent-stop")

	err := c.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	lctests.AssertIdempotentStop(t, c)
}

func TestRistrettoCache_KeyPrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("default prefix uses name", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("alpha")

		err := c.Start(ctx)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		impl, ok := c.(*ristrettoCache[string])
		if !ok {
			t.Fatalf("expected *ristrettoCache, got %T", c)
		}

		if impl.keyPrefix != "alpha:" {
			t.Fatalf("keyPrefix = %q, want %q", impl.keyPrefix, "alpha:")
		}
	})

	t.Run("WithKeyPrefix overrides default", func(t *testing.T) {
		t.Parallel()

		c := NewRistrettoCache[string]("alpha", WithKeyPrefix("custom::"))

		err := c.Start(ctx)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		impl, ok := c.(*ristrettoCache[string])
		if !ok {
			t.Fatalf("expected *ristrettoCache, got %T", c)
		}

		if impl.keyPrefix != "custom::" {
			t.Fatalf("keyPrefix = %q, want %q", impl.keyPrefix, "custom::")
		}
	})
}
