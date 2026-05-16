package resilience

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("allows burst tokens immediately", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		err := registry.Use("burst",
			WithRateLimiterInterval(time.Hour),
			WithRateLimiterBurst(3))
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

		registry := NewRateLimiterRegistry()

		err := registry.Use("exhaust",
			WithRateLimiterInterval(time.Hour),
			WithRateLimiterBurst(1))
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

		registry := NewRateLimiterRegistry()
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

		registry := NewRateLimiterRegistry()
		limiter := registry.Get("nil-ctx")

		err := limiter.Wait(nil)
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("returns immediately when tokens available", func(t *testing.T) {
		t.Parallel()

		registry := NewRateLimiterRegistry()

		err := registry.Use("immediate", WithRateLimiterBurst(5))
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

		registry := NewRateLimiterRegistry()

		err := registry.Use("wait",
			WithRateLimiterInterval(20*time.Millisecond),
			WithRateLimiterBurst(1))
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

		registry := NewRateLimiterRegistry()

		err := registry.Use("cancel",
			WithRateLimiterInterval(time.Hour),
			WithRateLimiterBurst(1))
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

		if !errors.Is(err, ErrRateLimiterWaitFailed) {
			t.Fatalf("expected ErrRateLimiterWaitFailed in chain, got %v", err)
		}
	})
}
