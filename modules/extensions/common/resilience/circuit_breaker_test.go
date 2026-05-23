package resilience

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// errSentinelCB is a sentinel error used inside test closures to satisfy
// lint rules while keeping the test focus on the behavior under inspection.
var errSentinelCB = errors.New("test sentinel circuit breaker")

// wantValue is the canonical success return string used across success-path tests.
const wantValue = "value"

func TestCircuitBreaker_Execute(t *testing.T) {
	t.Parallel()

	t.Run("success returns value", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()
		breaker := registry.Get("ok")

		raw, err := breaker.Execute(t.Context(), func() (any, error) {
			return wantValue, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := raw.(string)
		if !ok {
			t.Fatalf("expected string result, got %T", raw)
		}

		if got != wantValue {
			t.Fatalf("expected value, got %v", got)
		}
	})

	t.Run("nil context returns ErrContextNil", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()
		breaker := registry.Get("nil-ctx")

		fn := func() (any, error) { return "noop", errSentinelCB }

		_, err := breaker.Execute(nil, fn)
		if !errors.Is(err, ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("nil fn returns sentinel", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()
		breaker := registry.Get("nil-fn")

		_, err := breaker.Execute(t.Context(), nil)
		if !errors.Is(err, ErrCircuitBreakerExecuteFnNil) {
			t.Fatalf("expected ErrCircuitBreakerExecuteFnNil, got %v", err)
		}
	})

	t.Run("function error is wrapped", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()
		breaker := registry.Get("err")

		cause := errors.New("call failed")

		_, err := breaker.Execute(t.Context(), func() (any, error) {
			return nil, cause
		})
		if !errors.Is(err, cause) {
			t.Fatalf("expected errors.Is to match cause, got %v", err)
		}
	})

	t.Run("trips after configured consecutive failures", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		err := registry.Use("trip",
			WithCircuitBreakerConsecutiveFailures(3),
			WithCircuitBreakerTimeout(time.Hour))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		breaker := registry.Get("trip")
		cause := errors.New("boom")

		for i := range 3 {
			_, runErr := breaker.Execute(t.Context(), func() (any, error) {
				return nil, cause
			})
			if !errors.Is(runErr, cause) {
				t.Fatalf("call %d: expected wrapped cause, got %v", i, runErr)
			}
		}

		if breaker.State() != StateOpen {
			t.Fatalf("expected StateOpen after 3 failures, got %v", breaker.State())
		}

		_, runErr := breaker.Execute(t.Context(), func() (any, error) {
			return "ignored", nil
		})
		if !errors.Is(runErr, ErrCircuitBreakerOpen) {
			t.Fatalf("expected ErrCircuitBreakerOpen, got %v", runErr)
		}
	})

	t.Run("transitions to half-open after short timeout", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		err := registry.Use("half",
			WithCircuitBreakerConsecutiveFailures(2),
			WithCircuitBreakerTimeout(50*time.Millisecond),
			WithCircuitBreakerMaxRequests(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		breaker := registry.Get("half")
		cause := errors.New("boom")

		for i := range 2 {
			_, runErr := breaker.Execute(t.Context(), func() (any, error) {
				return nil, cause
			})
			if runErr == nil {
				t.Fatalf("call %d: expected error", i)
			}
		}

		if breaker.State() != StateOpen {
			t.Fatalf("expected StateOpen, got %v", breaker.State())
		}

		time.Sleep(80 * time.Millisecond)

		_, runErr := breaker.Execute(t.Context(), func() (any, error) {
			return "ok", nil
		})
		if runErr != nil {
			t.Fatalf("expected probe to succeed, got %v", runErr)
		}

		if breaker.State() != StateClosed {
			t.Fatalf("expected StateClosed after successful probe, got %v", breaker.State())
		}
	})

	t.Run("cancelled context is recorded as failure without invoking fn", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()

		err := registry.Use("cancel", WithCircuitBreakerConsecutiveFailures(5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		breaker := registry.Get("cancel")
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		called := atomic.Bool{}

		_, runErr := breaker.Execute(ctx, func() (any, error) {
			called.Store(true)
			return wantValue, nil
		})
		if runErr == nil {
			t.Fatal("expected error, got nil")
		}

		if called.Load() {
			t.Fatal("expected fn not to be called when ctx is canceled")
		}

		if !errors.Is(runErr, context.Canceled) {
			t.Fatalf("expected context.Canceled in error chain, got %v", runErr)
		}
	})

	t.Run("concurrent Execute is safe", func(t *testing.T) {
		t.Parallel()

		registry := NewCircuitBreakerRegistry()
		breaker := registry.Get("race")

		const workers = 50

		var wg sync.WaitGroup
		wg.Add(workers)

		for range workers {
			go func() {
				defer wg.Done()

				_, _ = breaker.Execute(t.Context(), func() (any, error) {
					return 1, nil
				})
			}()
		}

		wg.Wait()
	})
}
