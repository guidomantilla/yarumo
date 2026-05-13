package resilience_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	cresilience "github.com/guidomantilla/yarumo/common/resilience"
)

func TestNewRateLimiterRegistry(t *testing.T) {
	t.Parallel()

	registry := cresilience.NewRateLimiterRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRateLimiterRegistry_Get(t *testing.T) {
	t.Parallel()

	t.Run("lazy-creates on first call", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		limiter := registry.Get("svc-a")
		if limiter == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("returns same instance for same name", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		first := registry.Get("svc-b")
		second := registry.Get("svc-b")

		if first != second {
			t.Fatal("expected same instance for same name")
		}
	})

	t.Run("returns distinct instances per name", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		a := registry.Get("svc-x")
		b := registry.Get("svc-y")

		if a == b {
			t.Fatal("expected distinct instances per name")
		}
	})

	t.Run("concurrent Get is safe and yields a single instance", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		const workers = 32

		instances := make([]cresilience.RateLimiter, workers)

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

		limiter := cresilience.DefaultRateLimiterRegistry.Get("default-rl-test")
		if limiter == nil {
			t.Fatal("expected non-nil limiter")
		}
	})

	t.Run("many concurrent Get + Use stress double-check path", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

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
					_ = registry.Use(name, cresilience.WithRateLimiterBurst(2))
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

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("")
		if !errors.Is(err, cresilience.ErrRegistryNameEmpty) {
			t.Fatalf("expected ErrRegistryNameEmpty, got %v", err)
		}
	})

	t.Run("creates new entry", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("svc-new", cresilience.WithRateLimiterBurst(2))
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

		registry := cresilience.NewRateLimiterRegistry()

		original := registry.Get("svc-replace")

		err := registry.Use("svc-replace", cresilience.WithRateLimiterBurst(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		replacement := registry.Get("svc-replace")
		if replacement == original {
			t.Fatal("expected replacement to be a different instance")
		}
	})
}

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("allows burst tokens immediately", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("burst",
			cresilience.WithRateLimiterInterval(time.Hour),
			cresilience.WithRateLimiterBurst(3))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("burst")

		for i := range 3 {
			if !limiter.Allow() {
				t.Fatalf("expected Allow to return true for token %d", i)
			}
		}
	})

	t.Run("returns false when bucket exhausted", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("exhaust",
			cresilience.WithRateLimiterInterval(time.Hour),
			cresilience.WithRateLimiterBurst(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("exhaust")

		if !limiter.Allow() {
			t.Fatal("expected first Allow to succeed")
		}

		if limiter.Allow() {
			t.Fatal("expected second Allow to fail (bucket exhausted)")
		}
	})

	t.Run("concurrent Allow is safe", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()
		limiter := registry.Get("race")

		const workers = 50

		var wg sync.WaitGroup
		wg.Add(workers)

		for range workers {
			go func() {
				defer wg.Done()
				_ = limiter.Allow()
			}()
		}

		wg.Wait()
	})
}

func TestRateLimiter_Wait(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns sentinel", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()
		limiter := registry.Get("nil-ctx")

		err := limiter.Wait(nil)
		if !errors.Is(err, cresilience.ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns immediately when tokens available", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("immediate", cresilience.WithRateLimiterBurst(5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("immediate")

		err = limiter.Wait(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("blocks until token available", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("wait",
			cresilience.WithRateLimiterInterval(20*time.Millisecond),
			cresilience.WithRateLimiterBurst(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("wait")

		// Consume the initial token.
		if !limiter.Allow() {
			t.Fatal("expected initial Allow to succeed")
		}

		// Now a fresh Wait must block ~20ms before returning.
		start := time.Now()

		err = limiter.Wait(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		elapsed := time.Since(start)
		if elapsed < 5*time.Millisecond {
			t.Fatalf("expected Wait to block, only %v elapsed", elapsed)
		}
	})

	t.Run("cancelled context returns wrapped error", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewRateLimiterRegistry()

		err := registry.Use("cancel",
			cresilience.WithRateLimiterInterval(time.Hour),
			cresilience.WithRateLimiterBurst(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		limiter := registry.Get("cancel")

		// Consume the only token.
		if !limiter.Allow() {
			t.Fatal("expected initial Allow to succeed")
		}

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
		defer cancel()

		err = limiter.Wait(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, cresilience.ErrRateLimiterWaitFailed) {
			t.Fatalf("expected ErrRateLimiterWaitFailed in chain, got %v", err)
		}
	})
}
