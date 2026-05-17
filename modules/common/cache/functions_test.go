package cache

import (
	"context"
	"errors"
	"testing"
)

func TestGet(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the default registered cache", func(t *testing.T) {
		setErr := INMEMORY_CACHE.Set(ctx, "test-get-hit", "v", 0)
		if setErr != nil {
			t.Fatalf("INMEMORY_CACHE.Set: %v", setErr)
		}
		defer func() { _ = INMEMORY_CACHE.Delete(ctx, "test-get-hit") }()

		got, getErr := Get[string, any](ctx, "test-get-hit")
		if getErr != nil {
			t.Fatalf("Get: %v", getErr)
		}

		if got != "v" {
			t.Fatalf("Get = %v, want %q", got, "v")
		}
	})

	t.Run("propagates ErrCacheMiss", func(t *testing.T) {
		_, err := Get[string, any](ctx, "test-get-missing")
		if err == nil {
			t.Fatal("expected error for missing key")
		}

		if !errors.Is(err, ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		_, err := Get[int, int](ctx, 1)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestSet(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the default registered cache", func(t *testing.T) {
		defer func() { _ = INMEMORY_CACHE.Delete(ctx, "test-set") }()

		setErr := Set[string, any](ctx, "test-set", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		has, hasErr := INMEMORY_CACHE.Has(ctx, "test-set")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("expected INMEMORY_CACHE to hold the value")
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		err := Set[int, int](ctx, 1, 2, 0)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the default registered cache", func(t *testing.T) {
		setErr := INMEMORY_CACHE.Set(ctx, "test-delete", "v", 0)
		if setErr != nil {
			t.Fatalf("INMEMORY_CACHE.Set: %v", setErr)
		}

		delErr := Delete[string, any](ctx, "test-delete")
		if delErr != nil {
			t.Fatalf("Delete: %v", delErr)
		}

		has, hasErr := INMEMORY_CACHE.Has(ctx, "test-delete")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("expected INMEMORY_CACHE to no longer hold the value")
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		err := Delete[int, int](ctx, 1)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestHas(t *testing.T) {
	ctx := context.Background()

	t.Run("returns true for present key", func(t *testing.T) {
		setErr := INMEMORY_CACHE.Set(ctx, "test-has-hit", "v", 0)
		if setErr != nil {
			t.Fatalf("INMEMORY_CACHE.Set: %v", setErr)
		}
		defer func() { _ = INMEMORY_CACHE.Delete(ctx, "test-has-hit") }()

		has, hasErr := Has[string, any](ctx, "test-has-hit")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("Has = false, want true")
		}
	})

	t.Run("returns false for absent key", func(t *testing.T) {
		has, hasErr := Has[string, any](ctx, "test-has-missing")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("Has = true, want false")
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		_, err := Has[int, int](ctx, 1)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestClear(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the default registered cache", func(t *testing.T) {
		setErr := INMEMORY_CACHE.Set(ctx, "test-clear", "v", 0)
		if setErr != nil {
			t.Fatalf("INMEMORY_CACHE.Set: %v", setErr)
		}

		clearErr := Clear[string, any](ctx)
		if clearErr != nil {
			t.Fatalf("Clear: %v", clearErr)
		}

		has, hasErr := INMEMORY_CACHE.Has(ctx, "test-clear")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("expected INMEMORY_CACHE to be empty after Clear")
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		err := Clear[int, int](ctx)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestStop(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to the default registered cache", func(t *testing.T) {
		stopErr := Stop[string, any](ctx)
		if stopErr != nil {
			t.Fatalf("Stop: %v", stopErr)
		}
	})

	t.Run("returns ErrCacheTypeAssertion when default does not match K/V", func(t *testing.T) {
		err := Stop[int, int](ctx)
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}
	})
}

func TestResolveKeyPrefix(t *testing.T) {
	t.Parallel()

	t.Run("returns configured prefix when non-empty", func(t *testing.T) {
		t.Parallel()

		got := ResolveKeyPrefix("ignored-name", "explicit::")
		if got != "explicit::" {
			t.Fatalf("got %q, want %q", got, "explicit::")
		}
	})

	t.Run("falls back to name+: when configured prefix is empty", func(t *testing.T) {
		t.Parallel()

		got := ResolveKeyPrefix("alpha", "")
		if got != "alpha:" {
			t.Fatalf("got %q, want %q", got, "alpha:")
		}
	})
}
