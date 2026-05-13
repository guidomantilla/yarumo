package resilience_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	cresilience "github.com/guidomantilla/yarumo/common/resilience"
)

// errSentinelCB is a sentinel error used inside test closures to satisfy
// lint rules while keeping the test focus on the behavior under inspection.
var errSentinelCB = errors.New("test sentinel circuit breaker")

// wantValue is the canonical success return string used across success-path tests.
const wantValue = "value"

func TestNewCircuitBreakerRegistry(t *testing.T) {
	t.Parallel()

	registry := cresilience.NewCircuitBreakerRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestCircuitBreakerRegistry_Get(t *testing.T) {
	t.Parallel()

	t.Run("lazy-creates on first call", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		breaker := registry.Get("svc-a")
		if breaker == nil {
			t.Fatal("expected non-nil breaker")
		}

		if breaker.State() != cresilience.StateClosed {
			t.Fatalf("expected StateClosed, got %v", breaker.State())
		}
	})

	t.Run("returns same instance for same name", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		first := registry.Get("svc-b")
		second := registry.Get("svc-b")

		if first != second {
			t.Fatal("expected same instance for same name")
		}
	})

	t.Run("returns distinct instances per name", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		a := registry.Get("svc-x")
		b := registry.Get("svc-y")

		if a == b {
			t.Fatal("expected distinct instances per name")
		}
	})

	t.Run("concurrent Get is safe and yields a single instance", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		const workers = 32

		instances := make([]cresilience.CircuitBreaker, workers)

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

		breaker := cresilience.DefaultCircuitBreakerRegistry.Get("default-cb-test")
		if breaker == nil {
			t.Fatal("expected non-nil breaker")
		}
	})

	t.Run("many concurrent Get + Use stress double-check path", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

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
					_ = registry.Use(name, cresilience.WithCircuitBreakerConsecutiveFailures(2))
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

		registry := cresilience.NewCircuitBreakerRegistry()

		err := registry.Use("")
		if !errors.Is(err, cresilience.ErrRegistryNameEmpty) {
			t.Fatalf("expected ErrRegistryNameEmpty, got %v", err)
		}
	})

	t.Run("creates new entry", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		err := registry.Use("svc-new", cresilience.WithCircuitBreakerConsecutiveFailures(2))
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

		registry := cresilience.NewCircuitBreakerRegistry()

		original := registry.Get("svc-replace")

		err := registry.Use("svc-replace", cresilience.WithCircuitBreakerConsecutiveFailures(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		replacement := registry.Get("svc-replace")
		if replacement == original {
			t.Fatal("expected replacement to be a different instance")
		}
	})
}

func TestCircuitBreaker_Execute(t *testing.T) {
	t.Parallel()

	t.Run("success returns value", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()
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

		registry := cresilience.NewCircuitBreakerRegistry()
		breaker := registry.Get("nil-ctx")

		fn := func() (any, error) { return "noop", errSentinelCB }

		_, err := breaker.Execute(nil, fn)
		if !errors.Is(err, cresilience.ErrContextNil) {
			t.Fatalf("expected ErrContextNil, got %v", err)
		}
	})

	t.Run("nil fn returns sentinel", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()
		breaker := registry.Get("nil-fn")

		_, err := breaker.Execute(t.Context(), nil)
		if !errors.Is(err, cresilience.ErrCircuitBreakerExecuteFnNil) {
			t.Fatalf("expected ErrCircuitBreakerExecuteFnNil, got %v", err)
		}
	})

	t.Run("function error is wrapped", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()
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

		registry := cresilience.NewCircuitBreakerRegistry()

		err := registry.Use("trip",
			cresilience.WithCircuitBreakerConsecutiveFailures(3),
			cresilience.WithCircuitBreakerTimeout(time.Hour))
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

		if breaker.State() != cresilience.StateOpen {
			t.Fatalf("expected StateOpen after 3 failures, got %v", breaker.State())
		}

		_, runErr := breaker.Execute(t.Context(), func() (any, error) {
			return "ignored", nil
		})
		if !errors.Is(runErr, cresilience.ErrCircuitBreakerOpen) {
			t.Fatalf("expected ErrCircuitBreakerOpen, got %v", runErr)
		}
	})

	t.Run("transitions to half-open after short timeout", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		err := registry.Use("half",
			cresilience.WithCircuitBreakerConsecutiveFailures(2),
			cresilience.WithCircuitBreakerTimeout(50*time.Millisecond),
			cresilience.WithCircuitBreakerMaxRequests(1))
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

		if breaker.State() != cresilience.StateOpen {
			t.Fatalf("expected StateOpen, got %v", breaker.State())
		}

		time.Sleep(80 * time.Millisecond)

		_, runErr := breaker.Execute(t.Context(), func() (any, error) {
			return "ok", nil
		})
		if runErr != nil {
			t.Fatalf("expected probe to succeed, got %v", runErr)
		}

		if breaker.State() != cresilience.StateClosed {
			t.Fatalf("expected StateClosed after successful probe, got %v", breaker.State())
		}
	})

	t.Run("cancelled context is recorded as failure without invoking fn", func(t *testing.T) {
		t.Parallel()

		registry := cresilience.NewCircuitBreakerRegistry()

		err := registry.Use("cancel", cresilience.WithCircuitBreakerConsecutiveFailures(5))
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

		registry := cresilience.NewCircuitBreakerRegistry()
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
