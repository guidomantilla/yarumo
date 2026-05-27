package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	cretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
)

func TestNewRetry(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil retry with defaults", func(t *testing.T) {
		t.Parallel()

		r := NewRetry()
		if r == nil {
			t.Fatal("expected non-nil retry")
		}
	})

	t.Run("returns independent instances per call", func(t *testing.T) {
		t.Parallel()

		var c1, c2 atomic.Int32
		r1 := NewRetry(WithAttempts(2), WithDelay(time.Millisecond), WithOnRetry(func(_ uint, _ error) {
			c1.Add(1)
		}))
		_ = NewRetry(WithAttempts(2), WithDelay(time.Millisecond), WithOnRetry(func(_ uint, _ error) {
			c2.Add(1)
		}))

		_ = r1.Do(context.Background(), func() error { return errors.New("boom") })

		// r2 untouched.
		if c1.Load() == 0 {
			t.Fatal("r1 should have retried at least once")
		}
		if c2.Load() != 0 {
			t.Fatalf("r2 should not have retried; got %d", c2.Load())
		}
	})
}

func TestRetry_Do(t *testing.T) {
	t.Parallel()

	t.Run("returns nil when fn succeeds on first attempt", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		r := NewRetry(WithAttempts(3), WithDelay(time.Millisecond))

		err := r.Do(context.Background(), func() error {
			calls.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		if calls.Load() != 1 {
			t.Fatalf("calls = %d, want 1", calls.Load())
		}
	})

	t.Run("retries until fn succeeds within attempt budget", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		r := NewRetry(WithAttempts(3), WithDelay(time.Millisecond), WithBackoff(cretry.BackoffFixed))

		err := r.Do(context.Background(), func() error {
			n := calls.Add(1)
			if n < 3 {
				return errors.New("transient")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		if calls.Load() != 3 {
			t.Fatalf("calls = %d, want 3", calls.Load())
		}
	})

	t.Run("returns ErrRetryFailed when attempt budget is exhausted", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		r := NewRetry(WithAttempts(3), WithDelay(time.Millisecond), WithBackoff(cretry.BackoffFixed))

		err := r.Do(context.Background(), func() error {
			calls.Add(1)
			return errors.New("permanent")
		})
		if err == nil {
			t.Fatal("expected non-nil error after budget exhaustion")
		}
		if !errors.Is(err, cretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
		if calls.Load() != 3 {
			t.Fatalf("calls = %d, want 3 (full attempt budget)", calls.Load())
		}
	})

	t.Run("stops retrying when RetryIf returns false", func(t *testing.T) {
		t.Parallel()

		var permanent = errors.New("permanent")
		var calls atomic.Int32

		r := NewRetry(
			WithAttempts(5),
			WithDelay(time.Millisecond),
			WithBackoff(cretry.BackoffFixed),
			WithRetryIf(func(err error) bool { return !errors.Is(err, permanent) }),
		)

		err := r.Do(context.Background(), func() error {
			calls.Add(1)
			return permanent
		})
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		if !errors.Is(err, cretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
		if calls.Load() != 1 {
			t.Fatalf("calls = %d, want 1 (RetryIf rejected the first error)", calls.Load())
		}
	})

	t.Run("invokes OnRetry hook before each retry", func(t *testing.T) {
		t.Parallel()

		var hookCalls atomic.Int32
		r := NewRetry(
			WithAttempts(3),
			WithDelay(time.Millisecond),
			WithBackoff(cretry.BackoffFixed),
			WithOnRetry(func(_ uint, _ error) { hookCalls.Add(1) }),
		)

		_ = r.Do(context.Background(), func() error { return errors.New("transient") })

		// avast/retry-go invokes OnRetry once per failed attempt, including
		// the final one. 3 attempts → 3 failures → 3 hook calls.
		if hookCalls.Load() != 3 {
			t.Fatalf("hookCalls = %d, want 3", hookCalls.Load())
		}
	})

	t.Run("returns ErrRetryFailed wrapping ErrContextNil when ctx is nil", func(t *testing.T) {
		t.Parallel()

		r := NewRetry()
		//nolint:staticcheck // intentionally passing nil ctx to exercise the guard
		err := r.Do(nil, func() error { return nil })
		if !errors.Is(err, cretry.ErrContextNil) {
			t.Fatalf("expected wrap of ErrContextNil, got %v", err)
		}
		if !errors.Is(err, cretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
	})

	t.Run("returns ErrRetryFailed wrapping ErrFnNil when fn is nil", func(t *testing.T) {
		t.Parallel()

		r := NewRetry()
		err := r.Do(context.Background(), nil)
		if !errors.Is(err, cretry.ErrFnNil) {
			t.Fatalf("expected wrap of ErrFnNil, got %v", err)
		}
		if !errors.Is(err, cretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
	})

	t.Run("stops retrying when ctx is canceled", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32
		r := NewRetry(WithAttempts(100), WithDelay(50*time.Millisecond), WithBackoff(cretry.BackoffFixed))

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()

		err := r.Do(ctx, func() error {
			calls.Add(1)
			return errors.New("transient")
		})
		if err == nil {
			t.Fatal("expected non-nil error when ctx times out")
		}
		if !errors.Is(err, cretry.ErrRetryFailed) {
			t.Fatalf("expected wrap of ErrRetryFailed, got %v", err)
		}
		if calls.Load() >= 100 {
			t.Fatalf("calls = %d, expected ctx cancel to stop well before budget", calls.Load())
		}
	})
}
