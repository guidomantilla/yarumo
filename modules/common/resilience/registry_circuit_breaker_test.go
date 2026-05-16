package resilience

import (
	"errors"
	"sync"
	"testing"
)

func TestNewCircuitBreakerRegistry(t *testing.T) {
	t.Parallel()

	registry := NewCircuitBreakerRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestCircuitBreakerRegistry_Get(t *testing.T) {
	t.Parallel()

	t.Run("lazy-creates on first call", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		breaker := registry.Get("svc-a")
		if breaker == nil {
			t.Fatal("expected non-nil breaker")
		}

		if breaker.State() != StateClosed {
			t.Fatalf("expected StateClosed, got %v", breaker.State())
		}
	})

	t.Run("returns same instance for same name", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		first := registry.Get("svc-b")
		second := registry.Get("svc-b")

		if first != second {
			t.Fatal("expected same instance for same name")
		}
	})

	t.Run("returns distinct instances per name", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		a := registry.Get("svc-x")
		b := registry.Get("svc-y")

		if a == b {
			t.Fatal("expected distinct instances per name")
		}
	})

	t.Run("concurrent Get is safe and yields a single instance", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		const workers = 32

		instances := make([]CircuitBreaker, workers)

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

	t.Run("DefaultCircuitBreakerRegistry is usable", func(t *testing.T) {
		t.Parallel()

		breaker := DefaultCircuitBreakerRegistry.Get("default-cb-test")
		if breaker == nil {
			t.Fatal("expected non-nil breaker")
		}
	})

	t.Run("many concurrent Get + Use stress double-check path", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		const (
			names   = 4
			workers = 200
		)

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := range workers {
			go func(idx int) {
				defer wg.Done()

				name := []string{"dc-a", "dc-b", "dc-c", "dc-d"}[idx%names]

				switch idx % 3 {
				case 0:
					_ = registry.Use(name, WithCircuitBreakerConsecutiveFailures(2))
				default:
					_ = registry.Get(name)
				}
			}(i)
		}

		wg.Wait()

		// Sanity: every name has a breaker we can resolve.
		for _, name := range []string{"dc-a", "dc-b", "dc-c", "dc-d"} {
			breaker := registry.Get(name)
			if breaker == nil {
				t.Fatalf("expected non-nil breaker for %q", name)
			}
		}
	})
}

func TestCircuitBreakerRegistry_Use(t *testing.T) {
	t.Parallel()

	t.Run("empty name returns error", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		err := registry.Use("")
		if !errors.Is(err, ErrRegistryNameEmpty) {
			t.Fatalf("expected ErrRegistryNameEmpty, got %v", err)
		}
	})

	t.Run("creates new entry", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		err := registry.Use("svc-new", WithCircuitBreakerConsecutiveFailures(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		breaker := registry.Get("svc-new")
		if breaker == nil {
			t.Fatal("expected non-nil breaker")
		}
	})

	t.Run("replaces existing entry", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		original := registry.Get("svc-replace")

		err := registry.Use("svc-replace", WithCircuitBreakerConsecutiveFailures(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		replacement := registry.Get("svc-replace")
		if replacement == original {
			t.Fatal("expected replacement to be a different instance")
		}
	})
}

// TestCircuitBreakerRegistry_LoadOrCreate_RecheckBranch deterministically
// drives the recheck branch of loadOrCreate by pre-seeding the registry,
// emulating a concurrent writer that inserted between Get's RUnlock and
// loadOrCreate's Lock.
func TestCircuitBreakerRegistry_LoadOrCreate_RecheckBranch(t *testing.T) {
	t.Parallel()

	registry, ok := NewCircuitBreakerRegistry().(*circuitBreakerRegistry)
	if !ok {
		t.Fatal("expected *circuitBreakerRegistry")
	}

	name := "seeded"
	seed := newCircuitBreaker(name, NewOptions())

	registry.lock.Lock()
	registry.breakers[name] = seed
	registry.lock.Unlock()

	got := registry.loadOrCreate(name)
	if got != seed {
		t.Fatal("expected loadOrCreate to return the seeded breaker")
	}
}

// TestCircuitBreakerRegistry_LoadOrCreate_FreshBranch covers the path where
// no entry exists yet (the common lazy-create case).
func TestCircuitBreakerRegistry_LoadOrCreate_FreshBranch(t *testing.T) {
	t.Parallel()

	registry, ok := NewCircuitBreakerRegistry().(*circuitBreakerRegistry)
	if !ok {
		t.Fatal("expected *circuitBreakerRegistry")
	}

	got := registry.loadOrCreate("fresh")
	if got == nil {
		t.Fatal("expected non-nil breaker from loadOrCreate")
	}
}
