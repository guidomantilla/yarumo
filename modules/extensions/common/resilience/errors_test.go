package resilience

import (
	"errors"
	"strings"
	"testing"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	t.Run("circuit breaker error includes type and cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("inner failure")
		err := ErrCircuitBreakerExecute(cause)

		got := err.Error()

		if !strings.Contains(got, CircuitBreakerType) {
			t.Fatalf("expected error to contain type %q, got %s", CircuitBreakerType, got)
		}

		if !strings.Contains(got, "inner failure") {
			t.Fatalf("expected error to contain cause, got %s", got)
		}
	})

	t.Run("rate limiter error includes type and cause", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("dial timeout")
		err := ErrRateLimiterWait(cause)

		got := err.Error()

		if !strings.Contains(got, RateLimiterType) {
			t.Fatalf("expected error to contain type %q, got %s", RateLimiterType, got)
		}

		if !strings.Contains(got, "rate limiter wait failed") {
			t.Fatalf("expected error to contain sentinel, got %s", got)
		}
	})

	t.Run("registry error includes type", func(t *testing.T) {
		t.Parallel()

		err := ErrRegistryUse(ErrRegistryNameEmpty)

		got := err.Error()

		if !strings.Contains(got, RegistryType) {
			t.Fatalf("expected error to contain type %q, got %s", RegistryType, got)
		}

		if !errors.Is(err, ErrRegistryNameEmpty) {
			t.Fatalf("expected errors.Is to match ErrRegistryNameEmpty")
		}
	})
}

func TestErrCircuitBreakerExecute_IsSentinel(t *testing.T) {
	t.Parallel()

	err := ErrCircuitBreakerExecute(ErrCircuitBreakerOpen)

	if !errors.Is(err, ErrCircuitBreakerOpen) {
		t.Fatalf("expected errors.Is to match ErrCircuitBreakerOpen")
	}
}

func TestErrRateLimiterWait_IsSentinel(t *testing.T) {
	t.Parallel()

	err := ErrRateLimiterWait()

	if !errors.Is(err, ErrRateLimiterWaitFailed) {
		t.Fatalf("expected errors.Is to match ErrRateLimiterWaitFailed")
	}
}

func TestErrRegistryUse_IsSentinel(t *testing.T) {
	t.Parallel()

	err := ErrRegistryUse(ErrRegistryNameEmpty)

	if !errors.Is(err, ErrRegistryNameEmpty) {
		t.Fatalf("expected errors.Is to match ErrRegistryNameEmpty")
	}
}
