package resilience

import (
	"errors"
	"sync"
	"testing"
)

func TestNewRateLimiterRegistry(t *testing.T) {
	t.Parallel()

	registry := NewRateLimiterRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRateLimiterRegistry_Get(t *testing.T) {
	t.Parallel()

	t.Run("lazy-creates on first call", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		limiter := registry.Get("svc-a")
		if limiter == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("returns same instance for same name", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		first := registry.Get("svc-b")
		second := registry.Get("svc-b")

		if first != second {
			t.Fatal("expected same instance for same name")
		}
	})

	t.Run("returns distinct instances per name", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		a := registry.Get("svc-x")
		b := registry.Get("svc-y")

		if a == b {
			t.Fatal("expected distinct instances per name")
		}
	})

	t.Run("concurrent Get is safe and yields a single instance", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		const workers = 32

		instances := make([]RateLimiter, workers)

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := range workers {
			go func(idx int) {
				defer wg.Done()
				instances[idx] = registry.Get("shared")
			}(i)
		}

		wg.Wait()

		first := instances[0]
		for i, ins := range instances {
			if ins != first {
				t.Fatalf("instance %d differs from instance 0", i)
			}
		}
	})

	t.Run("DefaultRateLimiterRegistry is usable", func(t *testing.T) {
		t.Parallel()

		limiter := DefaultRateLimiterRegistry.Get("default-rl-test")
		if limiter == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("many concurrent Get + Use stress double-check path", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		const (
			names   = 4
			workers = 200
		)

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := range workers {
			go func(idx int) {
				defer wg.Done()

				name := []string{"dl-a", "dl-b", "dl-c", "dl-d"}[idx%names]

				switch idx % 3 {
				case 0:
					_ = registry.Use(name, WithRateLimiterBurst(2))
				default:
					_ = registry.Get(name)
				}
			}(i)
		}

		wg.Wait()

		for _, name := range []string{"dl-a", "dl-b", "dl-c", "dl-d"} {
			limiter := registry.Get(name)
			if limiter == nil {
				t.Fatalf("expected non-nil limiter for %q", name)
			}
		}
	})
}

func TestRateLimiterRegistry_Use(t *testing.T) {
	t.Parallel()

	t.Run("empty name returns error", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		err := registry.Use("")
		if !errors.Is(err, ErrRegistryNameEmpty) {
			t.Fatalf("expected ErrRegistryNameEmpty, got %v", err)
		}
	})

	t.Run("creates new entry", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		err := registry.Use("svc-new", WithRateLimiterBurst(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("svc-new")
		if limiter == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("replaces existing entry", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		original := registry.Get("svc-replace")

		err := registry.Use("svc-replace", WithRateLimiterBurst(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		replacement := registry.Get("svc-replace")
		if replacement == original {
			t.Fatal("expected replacement to be a different instance")
		}
	})
}

// TestRateLimiterRegistry_LoadOrCreate_RecheckBranch mirrors the
// circuit-breaker test for the limiter registry.
func TestRateLimiterRegistry_LoadOrCreate_RecheckBranch(t *testing.T) {
	t.Parallel()

	registry, ok := NewRateLimiterRegistry().(*rateLimiterRegistry)
	if !ok {
		t.Fatal("expected *rateLimiterRegistry")
	}

	name := "seeded"
	seed := newRateLimiter(NewOptions())

	registry.lock.Lock()
	registry.limiters[name] = seed
	registry.lock.Unlock()

	got := registry.loadOrCreate(name)
	if got != seed {
		t.Fatal("expected loadOrCreate to return the seeded limiter")
	}
}

// TestRateLimiterRegistry_LoadOrCreate_FreshBranch covers the path where no
// entry exists yet.
func TestRateLimiterRegistry_LoadOrCreate_FreshBranch(t *testing.T) {
	t.Parallel()

	registry, ok := NewRateLimiterRegistry().(*rateLimiterRegistry)
	if !ok {
		t.Fatal("expected *rateLimiterRegistry")
	}

	got := registry.loadOrCreate("fresh")
	if got == nil {
		t.Fatal("expected non-nil limiter from loadOrCreate")
	}
}
