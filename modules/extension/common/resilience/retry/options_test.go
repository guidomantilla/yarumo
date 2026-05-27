package retry

import (
	"errors"
	"testing"
	"time"

	cretry "github.com/guidomantilla/yarumo/core/common/resilience/retry"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies safe defaults when no options given", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.attempts != DefaultAttempts {
			t.Fatalf("attempts = %d, want %d", opts.attempts, DefaultAttempts)
		}
		if opts.delay != DefaultDelay {
			t.Fatalf("delay = %v, want %v", opts.delay, DefaultDelay)
		}
		if opts.maxDelay != DefaultMaxDelay {
			t.Fatalf("maxDelay = %v, want %v", opts.maxDelay, DefaultMaxDelay)
		}
		if opts.backoff != DefaultBackoff {
			t.Fatalf("backoff = %d, want %d", opts.backoff, DefaultBackoff)
		}
		if opts.retryIf == nil {
			t.Fatal("retryIf is nil")
		}
		if opts.onRetry == nil {
			t.Fatal("onRetry is nil")
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithAttempts(5),
			WithDelay(50*time.Millisecond),
			WithMaxDelay(time.Second),
			WithBackoff(cretry.BackoffFixed),
		)
		if opts.attempts != 5 {
			t.Fatalf("attempts = %d, want 5", opts.attempts)
		}
		if opts.delay != 50*time.Millisecond {
			t.Fatalf("delay = %v, want 50ms", opts.delay)
		}
		if opts.maxDelay != time.Second {
			t.Fatalf("maxDelay = %v, want 1s", opts.maxDelay)
		}
		if opts.backoff != cretry.BackoffFixed {
			t.Fatalf("backoff = %d, want %d", opts.backoff, cretry.BackoffFixed)
		}
	})
}

func TestWithAttempts(t *testing.T) {
	t.Parallel()

	t.Run("sets attempts when > 1", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAttempts(10))
		if opts.attempts != 10 {
			t.Fatalf("attempts = %d, want 10", opts.attempts)
		}
	})

	t.Run("ignores 0 attempts, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAttempts(0))
		if opts.attempts != DefaultAttempts {
			t.Fatalf("attempts = %d, want default %d", opts.attempts, DefaultAttempts)
		}
	})

	t.Run("ignores 1 attempt, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAttempts(1))
		if opts.attempts != DefaultAttempts {
			t.Fatalf("attempts = %d, want default %d", opts.attempts, DefaultAttempts)
		}
	})
}

func TestWithDelay(t *testing.T) {
	t.Parallel()

	t.Run("sets delay when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDelay(250 * time.Millisecond))
		if opts.delay != 250*time.Millisecond {
			t.Fatalf("delay = %v, want 250ms", opts.delay)
		}
	})

	t.Run("ignores zero delay, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDelay(0))
		if opts.delay != DefaultDelay {
			t.Fatalf("delay = %v, want default %v", opts.delay, DefaultDelay)
		}
	})

	t.Run("ignores negative delay, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDelay(-1 * time.Second))
		if opts.delay != DefaultDelay {
			t.Fatalf("delay = %v, want default %v", opts.delay, DefaultDelay)
		}
	})
}

func TestWithMaxDelay(t *testing.T) {
	t.Parallel()

	t.Run("sets maxDelay when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxDelay(30 * time.Second))
		if opts.maxDelay != 30*time.Second {
			t.Fatalf("maxDelay = %v, want 30s", opts.maxDelay)
		}
	})

	t.Run("ignores zero, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxDelay(0))
		if opts.maxDelay != DefaultMaxDelay {
			t.Fatalf("maxDelay = %v, want default %v", opts.maxDelay, DefaultMaxDelay)
		}
	})
}

func TestWithBackoff(t *testing.T) {
	t.Parallel()

	t.Run("sets known backoff types", func(t *testing.T) {
		t.Parallel()

		for _, b := range []cretry.Backoff{cretry.BackoffFixed, cretry.BackoffExponential, cretry.BackoffRandom} {
			opts := NewOptions(WithBackoff(b))
			if opts.backoff != b {
				t.Fatalf("backoff = %d, want %d", opts.backoff, b)
			}
		}
	})

	t.Run("ignores invalid backoff, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBackoff(cretry.Backoff(99)))
		if opts.backoff != DefaultBackoff {
			t.Fatalf("backoff = %d, want default %d", opts.backoff, DefaultBackoff)
		}
	})
}

func TestWithRetryIf(t *testing.T) {
	t.Parallel()

	t.Run("sets predicate when non-nil", func(t *testing.T) {
		t.Parallel()

		fn := func(err error) bool { return errors.Is(err, errors.New("specific")) }
		opts := NewOptions(WithRetryIf(fn))
		if opts.retryIf == nil {
			t.Fatal("retryIf not set")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRetryIf(nil))
		if opts.retryIf == nil {
			t.Fatal("expected default retryIf to remain")
		}
	})
}

func TestWithOnRetry(t *testing.T) {
	t.Parallel()

	t.Run("sets hook when non-nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOnRetry(func(_ uint, _ error) {}))
		if opts.onRetry == nil {
			t.Fatal("onRetry not set")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOnRetry(nil))
		if opts.onRetry == nil {
			t.Fatal("expected default onRetry to remain")
		}
	})
}
