package resilience_test

import (
	"errors"
	"strings"
	"testing"

	cresilience "github.com/guidomantilla/yarumo/common/resilience"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("circuit breaker error includes type and cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("inner failure")
		err := cresilience.ErrCircuitBreakerExecute(cause)

		got := err.Error()

		if !strings.Contains(got, cresilience.CircuitBreakerType) {
			t.Fatalf("expected error to contain type %q, got %s", cresilience.CircuitBreakerType, got)
		}

		if !strings.Contains(got, "inner failure") {
			t.Fatalf("expected error to contain cause, got %s", got)
		}
	})

	t.Run("rate limiter error includes type and cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("dial timeout")
		err := cresilience.ErrRateLimiterWait(cause)

		got := err.Error()

		if !strings.Contains(got, cresilience.RateLimiterType) {
			t.Fatalf("expected error to contain type %q, got %s", cresilience.RateLimiterType, got)
		}

		if !strings.Contains(got, "rate limiter wait failed") {
			t.Fatalf("expected error to contain sentinel, got %s", got)
		}
	})

	t.Run("registry error includes type", func(t *testing.T) {
		t.Parallel()

		err := cresilience.ErrRegistryUse(cresilience.ErrRegistryNameEmpty)

		got := err.Error()

		if !strings.Contains(got, cresilience.RegistryType) {
			t.Fatalf("expected error to contain type %q, got %s", cresilience.RegistryType, got)
		}

		if !errors.Is(err, cresilience.ErrRegistryNameEmpty) {
			t.Fatalf("expected errors.Is to match ErrRegistryNameEmpty")
		}
	})
}

func TestErrCircuitBreakerExecute_IsSentinel(t *testing.T) {
	t.Parallel()

	err := cresilience.ErrCircuitBreakerExecute(cresilience.ErrCircuitBreakerOpen)

	if !errors.Is(err, cresilience.ErrCircuitBreakerOpen) {
		t.Fatalf("expected errors.Is to match ErrCircuitBreakerOpen")
	}
}

func TestErrRateLimiterWait_IsSentinel(t *testing.T) {
	t.Parallel()

	err := cresilience.ErrRateLimiterWait()

	if !errors.Is(err, cresilience.ErrRateLimiterWaitFailed) {
		t.Fatalf("expected errors.Is to match ErrRateLimiterWaitFailed")
	}
}

func TestErrRegistryUse_IsSentinel(t *testing.T) {
	t.Parallel()

	err := cresilience.ErrRegistryUse(cresilience.ErrRegistryNameEmpty)

	if !errors.Is(err, cresilience.ErrRegistryNameEmpty) {
		t.Fatalf("expected errors.Is to match ErrRegistryNameEmpty")
	}
}
