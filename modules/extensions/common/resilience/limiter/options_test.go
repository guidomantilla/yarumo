package limiter

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies safe defaults when no options given", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want %v", opts.interval, DefaultInterval)
		}
		if opts.burst != DefaultBurst {
			t.Fatalf("burst = %d, want %d", opts.burst, DefaultBurst)
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithRate(5, time.Second),
			WithBurst(50),
		)
		if opts.interval != 200*time.Millisecond {
			t.Fatalf("interval = %v, want %v", opts.interval, 200*time.Millisecond)
		}
		if opts.burst != 50 {
			t.Fatalf("burst = %d, want 50", opts.burst)
		}
	})
}

func TestWithRate(t *testing.T) {
	t.Parallel()

	t.Run("computes interval as interval/perInterval", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRate(10, time.Second))
		if opts.interval != 100*time.Millisecond {
			t.Fatalf("interval = %v, want 100ms (10 rps)", opts.interval)
		}
	})

	t.Run("ignores non-positive perInterval, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRate(0, time.Second))
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want default %v", opts.interval, DefaultInterval)
		}
	})

	t.Run("ignores non-positive interval, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRate(10, 0))
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want default %v", opts.interval, DefaultInterval)
		}
	})

	t.Run("ignores negative perInterval, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRate(-1, time.Second))
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want default %v", opts.interval, DefaultInterval)
		}
	})
}

func TestWithBurst(t *testing.T) {
	t.Parallel()

	t.Run("sets burst when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBurst(20))
		if opts.burst != 20 {
			t.Fatalf("burst = %d, want 20", opts.burst)
		}
	})

	t.Run("ignores zero burst, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBurst(0))
		if opts.burst != DefaultBurst {
			t.Fatalf("burst = %d, want default %d", opts.burst, DefaultBurst)
		}
	})

	t.Run("ignores negative burst, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBurst(-3))
		if opts.burst != DefaultBurst {
			t.Fatalf("burst = %d, want default %d", opts.burst, DefaultBurst)
		}
	})
}

func TestOptions_rateLimit(t *testing.T) {
	t.Parallel()

	t.Run("returns rate.Every(interval) for positive interval", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRate(10, time.Second))
		if opts.rateLimit() != rate.Every(100*time.Millisecond) {
			t.Fatalf("rateLimit = %v, want %v", opts.rateLimit(), rate.Every(100*time.Millisecond))
		}
	})

	t.Run("falls back to DefaultInterval when interval is zero", func(t *testing.T) {
		t.Parallel()

		opts := &Options{interval: 0, burst: 1}
		if opts.rateLimit() != rate.Every(DefaultInterval) {
			t.Fatalf("rateLimit fallback failed; got %v", opts.rateLimit())
		}
	})
}
