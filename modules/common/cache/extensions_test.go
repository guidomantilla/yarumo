package cache

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"
)

func snapshotState() map[string]any {
	lock.RLock()
	defer lock.RUnlock()

	cp := make(map[string]any, len(caches))
	maps.Copy(cp, caches)

	return cp
}

func restoreState(snap map[string]any) {
	lock.Lock()
	defer lock.Unlock()

	caches = make(map[string]any, len(snap))
	maps.Copy(caches, snap)
}

func TestRegister(t *testing.T) {
	snap := snapshotState()
	defer restoreState(snap)

	t.Run("registers a new Cache[string, any]", func(t *testing.T) {
		Register("custom", NewMemoryCache[string, any]("custom"))

		got, err := Lookup[string, any]("custom")
		if err != nil {
			t.Fatalf("Lookup after Register: %v", err)
		}

		if got == nil {
			t.Fatal("Lookup returned nil cache")
		}
	})

	t.Run("registers a typed Cache[int, int]", func(t *testing.T) {
		Register("typed", NewMemoryCache[int, int]("typed"))

		got, err := Lookup[int, int]("typed")
		if err != nil {
			t.Fatalf("Lookup[int, int] after Register: %v", err)
		}

		if got == nil {
			t.Fatal("Lookup returned nil cache")
		}
	})

	t.Run("overwrites existing registration", func(t *testing.T) {
		first := NewMemoryCache[string, any]("first")
		second := NewMemoryCache[string, any]("second")

		Register("over", first)
		Register("over", second)

		got, err := Lookup[string, any]("over")
		if err != nil {
			t.Fatalf("Lookup after overwrite: %v", err)
		}

		ctx := context.Background()

		setErr := got.Set(ctx, "marker", "v", 0)
		if setErr != nil {
			t.Fatalf("got.Set: %v", setErr)
		}

		has, hasErr := second.Has(ctx, "marker")
		if hasErr != nil {
			t.Fatalf("second.Has: %v", hasErr)
		}

		if !has {
			t.Fatal("expected Lookup to return the second registration")
		}
	})
}

func TestLookup(t *testing.T) {
	snap := snapshotState()
	defer restoreState(snap)

	t.Run("returns default cache", func(t *testing.T) {
		got, err := Lookup[string, any]("default")
		if err != nil {
			t.Fatalf("Lookup(default): %v", err)
		}

		if got == nil {
			t.Fatal("expected non-nil default cache")
		}
	})

	t.Run("returns ErrCacheNotRegistered for unknown name", func(t *testing.T) {
		got, err := Lookup[string, any]("does-not-exist")
		if err == nil {
			t.Fatal("expected error for unknown name")
		}

		if !errors.Is(err, ErrCacheNotRegistered) {
			t.Fatalf("expected wrap of ErrCacheNotRegistered, got %v", err)
		}

		if got != nil {
			t.Fatalf("expected nil cache on error, got %T", got)
		}
	})

	t.Run("returns ErrCacheTypeAssertion when K/V do not match", func(t *testing.T) {
		Register("mismatch", NewMemoryCache[string, any]("mismatch"))

		got, err := Lookup[int, int]("mismatch")
		if err == nil {
			t.Fatal("expected type-assertion error")
		}

		if !errors.Is(err, ErrCacheTypeAssertion) {
			t.Fatalf("expected wrap of ErrCacheTypeAssertion, got %v", err)
		}

		if got != nil {
			t.Fatalf("expected nil cache on type error, got %T", got)
		}
	})
}

func TestSupported(t *testing.T) {
	snap := snapshotState()
	defer restoreState(snap)

	t.Run("includes the default cache", func(t *testing.T) {
		names := Supported()

		if !slices.Contains(names, "default") {
			t.Fatalf("Supported() = %v, want to include %q", names, "default")
		}
	})

	t.Run("includes the INMEMORY_CACHE name entry", func(t *testing.T) {
		names := Supported()

		if !slices.Contains(names, INMEMORY_CACHE.Name()) {
			t.Fatalf("Supported() = %v, want to include %q", names, INMEMORY_CACHE.Name())
		}
	})

	t.Run("includes newly registered cache", func(t *testing.T) {
		Register("extra", NewMemoryCache[string, any]("extra"))

		names := Supported()

		if !slices.Contains(names, "extra") {
			t.Fatalf("Supported() = %v, want to include %q", names, "extra")
		}
	})
}
