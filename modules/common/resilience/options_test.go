package resilience

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.cbMaxRequests != DefaultCBMaxRequests {
			t.Fatalf("cbMaxRequests=%d want %d", opts.cbMaxRequests, DefaultCBMaxRequests)
		}

		if opts.cbInterval != DefaultCBInterval {
			t.Fatalf("cbInterval=%v want %v", opts.cbInterval, DefaultCBInterval)
		}

		if opts.cbTimeout != DefaultCBTimeout {
			t.Fatalf("cbTimeout=%v want %v", opts.cbTimeout, DefaultCBTimeout)
		}

		if opts.cbConsecutiveFailures != DefaultCBConsecutiveFailures {
			t.Fatalf("cbConsecutiveFailures=%d want %d", opts.cbConsecutiveFailures, DefaultCBConsecutiveFailures)
		}

		if opts.rateInterval != DefaultRateLimitInterval {
			t.Fatalf("rateInterval=%v want %v", opts.rateInterval, DefaultRateLimitInterval)
		}

		if opts.rateBurst != DefaultRateLimitBurst {
			t.Fatalf("rateBurst=%d want %d", opts.rateBurst, DefaultRateLimitBurst)
		}
	})

	t.Run("options applied in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithCircuitBreakerMaxRequests(10),
			WithCircuitBreakerInterval(2*time.Second),
			WithCircuitBreakerTimeout(3*time.Second),
			WithCircuitBreakerConsecutiveFailures(7),
			WithRateLimiterInterval(50*time.Millisecond),
			WithRateLimiterBurst(20),
		)

		if opts.cbMaxRequests != 10 {
			t.Fatalf("cbMaxRequests=%d want 10", opts.cbMaxRequests)
		}

		if opts.cbInterval != 2*time.Second {
			t.Fatalf("cbInterval=%v want 2s", opts.cbInterval)
		}

		if opts.cbTimeout != 3*time.Second {
			t.Fatalf("cbTimeout=%v want 3s", opts.cbTimeout)
		}

		if opts.cbConsecutiveFailures != 7 {
			t.Fatalf("cbConsecutiveFailures=%d want 7", opts.cbConsecutiveFailures)
		}

		if opts.rateInterval != 50*time.Millisecond {
			t.Fatalf("rateInterval=%v want 50ms", opts.rateInterval)
		}

		if opts.rateBurst != 20 {
			t.Fatalf("rateBurst=%d want 20", opts.rateBurst)
		}
	})
}

func TestWithCircuitBreakerMaxRequests(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerMaxRequests(0))

		if opts.cbMaxRequests != DefaultCBMaxRequests {
			t.Fatalf("expected default kept, got %d", opts.cbMaxRequests)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerMaxRequests(11))

		if opts.cbMaxRequests != 11 {
			t.Fatalf("expected 11, got %d", opts.cbMaxRequests)
		}
	})
}

func TestWithCircuitBreakerInterval(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerInterval(0))

		if opts.cbInterval != DefaultCBInterval {
			t.Fatalf("expected default kept, got %v", opts.cbInterval)
		}
	})

	t.Run("negative is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerInterval(-time.Second))

		if opts.cbInterval != DefaultCBInterval {
			t.Fatalf("expected default kept, got %v", opts.cbInterval)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerInterval(5 * time.Second))

		if opts.cbInterval != 5*time.Second {
			t.Fatalf("expected 5s, got %v", opts.cbInterval)
		}
	})
}

func TestWithCircuitBreakerTimeout(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerTimeout(0))

		if opts.cbTimeout != DefaultCBTimeout {
			t.Fatalf("expected default kept, got %v", opts.cbTimeout)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerTimeout(7 * time.Second))

		if opts.cbTimeout != 7*time.Second {
			t.Fatalf("expected 7s, got %v", opts.cbTimeout)
		}
	})
}

func TestWithCircuitBreakerConsecutiveFailures(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerConsecutiveFailures(0))

		if opts.cbConsecutiveFailures != DefaultCBConsecutiveFailures {
			t.Fatalf("expected default kept, got %d", opts.cbConsecutiveFailures)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCircuitBreakerConsecutiveFailures(2))

		if opts.cbConsecutiveFailures != 2 {
			t.Fatalf("expected 2, got %d", opts.cbConsecutiveFailures)
		}
	})
}

func TestWithRateLimiterInterval(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterInterval(0))

		if opts.rateInterval != DefaultRateLimitInterval {
			t.Fatalf("expected default kept, got %v", opts.rateInterval)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterInterval(40 * time.Millisecond))

		if opts.rateInterval != 40*time.Millisecond {
			t.Fatalf("expected 40ms, got %v", opts.rateInterval)
		}
	})
}

func TestWithRateLimiterBurst(t *testing.T) {
	t.Parallel()

	t.Run("zero is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterBurst(0))

		if opts.rateBurst != DefaultRateLimitBurst {
			t.Fatalf("expected default kept, got %d", opts.rateBurst)
		}
	})

	t.Run("negative is ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterBurst(-1))

		if opts.rateBurst != DefaultRateLimitBurst {
			t.Fatalf("expected default kept, got %d", opts.rateBurst)
		}
	})

	t.Run("positive applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterBurst(50))

		if opts.rateBurst != 50 {
			t.Fatalf("expected 50, got %d", opts.rateBurst)
		}
	})
}

func TestOptions_RateLimit(t *testing.T) {
	t.Parallel()

	t.Run("defaults produce non-zero rate", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		got := opts.rateLimit()
		want := rate.Every(DefaultRateLimitInterval)

		if got != want {
			t.Fatalf("rateLimit()=%v want %v", got, want)
		}
	})

	t.Run("zero interval falls back to default", func(t *testing.T) {
		t.Parallel()

		opts := &Options{rateInterval: 0}

		got := opts.rateLimit()
		want := rate.Every(DefaultRateLimitInterval)

		if got != want {
			t.Fatalf("rateLimit()=%v want %v", got, want)
		}
	})

	t.Run("custom interval applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRateLimiterInterval(10 * time.Millisecond))

		got := opts.rateLimit()
		want := rate.Every(10 * time.Millisecond)

		if got != want {
			t.Fatalf("rateLimit()=%v want %v", got, want)
		}
	})
}
