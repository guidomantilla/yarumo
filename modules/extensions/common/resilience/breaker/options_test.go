package breaker

import (
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies safe defaults when no options given", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.name != DefaultName {
			t.Fatalf("name = %q, want %q", opts.name, DefaultName)
		}
		if opts.maxRequests != DefaultMaxRequests {
			t.Fatalf("maxRequests = %d, want %d", opts.maxRequests, DefaultMaxRequests)
		}
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want %v", opts.interval, DefaultInterval)
		}
		if opts.timeout != DefaultTimeout {
			t.Fatalf("timeout = %v, want %v", opts.timeout, DefaultTimeout)
		}
		if opts.consecutiveFailures != DefaultConsecutiveFailures {
			t.Fatalf("consecutiveFailures = %d, want %d", opts.consecutiveFailures, DefaultConsecutiveFailures)
		}
		if opts.onStateChange == nil {
			t.Fatal("onStateChange is nil")
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithName("svc"),
			WithMaxRequests(5),
			WithInterval(30*time.Second),
			WithTimeout(5*time.Second),
			WithConsecutiveFailures(10),
		)
		if opts.name != "svc" {
			t.Fatalf("name = %q, want %q", opts.name, "svc")
		}
		if opts.maxRequests != 5 {
			t.Fatalf("maxRequests = %d, want 5", opts.maxRequests)
		}
		if opts.interval != 30*time.Second {
			t.Fatalf("interval = %v, want 30s", opts.interval)
		}
		if opts.timeout != 5*time.Second {
			t.Fatalf("timeout = %v, want 5s", opts.timeout)
		}
		if opts.consecutiveFailures != 10 {
			t.Fatalf("consecutiveFailures = %d, want 10", opts.consecutiveFailures)
		}
	})
}

func TestWithName(t *testing.T) {
	t.Parallel()

	t.Run("sets name when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithName("payments"))
		if opts.name != "payments" {
			t.Fatalf("name = %q, want %q", opts.name, "payments")
		}
	})

	t.Run("ignores empty, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithName(""))
		if opts.name != DefaultName {
			t.Fatalf("name = %q, want default %q", opts.name, DefaultName)
		}
	})
}

func TestWithMaxRequests(t *testing.T) {
	t.Parallel()

	t.Run("sets value when > 0", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxRequests(3))
		if opts.maxRequests != 3 {
			t.Fatalf("maxRequests = %d, want 3", opts.maxRequests)
		}
	})

	t.Run("ignores zero, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxRequests(0))
		if opts.maxRequests != DefaultMaxRequests {
			t.Fatalf("maxRequests = %d, want default %d", opts.maxRequests, DefaultMaxRequests)
		}
	})
}

func TestWithInterval(t *testing.T) {
	t.Parallel()

	t.Run("sets interval when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInterval(2 * time.Minute))
		if opts.interval != 2*time.Minute {
			t.Fatalf("interval = %v, want 2m", opts.interval)
		}
	})

	t.Run("ignores zero, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInterval(0))
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want default %v", opts.interval, DefaultInterval)
		}
	})

	t.Run("ignores negative, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithInterval(-1 * time.Second))
		if opts.interval != DefaultInterval {
			t.Fatalf("interval = %v, want default %v", opts.interval, DefaultInterval)
		}
	})
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("sets timeout when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(500 * time.Millisecond))
		if opts.timeout != 500*time.Millisecond {
			t.Fatalf("timeout = %v, want 500ms", opts.timeout)
		}
	})

	t.Run("ignores zero, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTimeout(0))
		if opts.timeout != DefaultTimeout {
			t.Fatalf("timeout = %v, want default %v", opts.timeout, DefaultTimeout)
		}
	})
}

func TestWithConsecutiveFailures(t *testing.T) {
	t.Parallel()

	t.Run("sets value when > 0", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConsecutiveFailures(7))
		if opts.consecutiveFailures != 7 {
			t.Fatalf("consecutiveFailures = %d, want 7", opts.consecutiveFailures)
		}
	})

	t.Run("ignores zero, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConsecutiveFailures(0))
		if opts.consecutiveFailures != DefaultConsecutiveFailures {
			t.Fatalf("consecutiveFailures = %d, want default %d", opts.consecutiveFailures, DefaultConsecutiveFailures)
		}
	})
}

func TestWithOnStateChange(t *testing.T) {
	t.Parallel()

	t.Run("sets hook when non-nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOnStateChange(func(_ string, _, _ State) {}))
		if opts.onStateChange == nil {
			t.Fatal("onStateChange not set")
		}
	})

	t.Run("ignores nil, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOnStateChange(nil))
		if opts.onStateChange == nil {
			t.Fatal("expected default onStateChange to remain")
		}
	})
}
