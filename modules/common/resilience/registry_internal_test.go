package resilience

import (
	"testing"
)

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
