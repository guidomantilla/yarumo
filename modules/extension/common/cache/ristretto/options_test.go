package ristretto

import (
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies safe defaults when no options given", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 5*time.Minute)
		}

		if opts.keyPrefix != "" {
			t.Fatalf("keyPrefix = %q, want empty", opts.keyPrefix)
		}

		if opts.numCtrs != 1_000_000 {
			t.Fatalf("numCtrs = %d, want %d", opts.numCtrs, 1_000_000)
		}

		if opts.maxCost != 100<<20 {
			t.Fatalf("maxCost = %d, want %d", opts.maxCost, 100<<20)
		}

		if opts.bufItems != 64 {
			t.Fatalf("bufItems = %d, want %d", opts.bufItems, 64)
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithTTL(10*time.Minute),
			WithCapacity(2_000_000, 200<<20, 128),
		)

		if opts.ttl != 10*time.Minute {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 10*time.Minute)
		}

		if opts.numCtrs != 2_000_000 {
			t.Fatalf("numCtrs = %d, want %d", opts.numCtrs, 2_000_000)
		}
	})
}

func TestWithTTL(t *testing.T) {
	t.Parallel()

	t.Run("sets the ttl when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(30 * time.Second))
		if opts.ttl != 30*time.Second {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 30*time.Second)
		}
	})

	t.Run("ignores zero ttl, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(0))
		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want default %v", opts.ttl, 5*time.Minute)
		}
	})

	t.Run("ignores negative ttl, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(-1 * time.Second))
		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want default %v", opts.ttl, 5*time.Minute)
		}
	})
}

func TestWithCapacity(t *testing.T) {
	t.Parallel()

	t.Run("sets every parameter when all positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCapacity(500_000, 50<<20, 32))

		if opts.numCtrs != 500_000 {
			t.Fatalf("numCounters = %d, want %d", opts.numCtrs, 500_000)
		}

		if opts.maxCost != 50<<20 {
			t.Fatalf("maxCost = %d, want %d", opts.maxCost, 50<<20)
		}

		if opts.bufItems != 32 {
			t.Fatalf("bufferItems = %d, want %d", opts.bufItems, 32)
		}
	})

	t.Run("ignores non-positive numCounters", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCapacity(0, 50<<20, 32))
		if opts.numCtrs != 1_000_000 {
			t.Fatalf("numCounters = %d, want default %d", opts.numCtrs, 1_000_000)
		}
	})

	t.Run("ignores non-positive maxCost", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCapacity(500_000, 0, 32))
		if opts.maxCost != 100<<20 {
			t.Fatalf("maxCost = %d, want default %d", opts.maxCost, 100<<20)
		}
	})

	t.Run("ignores non-positive bufferItems", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCapacity(500_000, 50<<20, 0))
		if opts.bufItems != 64 {
			t.Fatalf("bufferItems = %d, want default %d", opts.bufItems, 64)
		}
	})
}

func TestWithKeyPrefix(t *testing.T) {
	t.Parallel()

	t.Run("sets prefix when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyPrefix("svc::"))
		if opts.keyPrefix != "svc::" {
			t.Fatalf("keyPrefix = %q, want %q", opts.keyPrefix, "svc::")
		}
	})

	t.Run("ignores empty prefix, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyPrefix(""))
		if opts.keyPrefix != "" {
			t.Fatalf("keyPrefix = %q, want empty", opts.keyPrefix)
		}
	})
}
