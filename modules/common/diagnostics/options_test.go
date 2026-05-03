package diagnostics

import (
	"testing"
	"time"
)

func TestWithMinAge(t *testing.T) {
	t.Parallel()

	t.Run("overrides default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinAge(30 * time.Second))
		if opts.minAge != 30*time.Second {
			t.Fatalf("got minAge %v, want %v", opts.minAge, 30*time.Second)
		}
	})

	t.Run("zero keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinAge(0))
		if opts.minAge != 10*time.Second {
			t.Fatalf("got minAge %v, want default %v", opts.minAge, 10*time.Second)
		}
	})

	t.Run("negative keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinAge(-5 * time.Second))
		if opts.minAge != 10*time.Second {
			t.Fatalf("got minAge %v, want default %v", opts.minAge, 10*time.Second)
		}
	})

	t.Run("does not affect maxBytes", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinAge(30 * time.Second))
		if opts.maxBytes != 10<<20 {
			t.Fatalf("got maxBytes %d, want %d", opts.maxBytes, uint64(10<<20))
		}
	})
}

func TestWithMaxBytes(t *testing.T) {
	t.Parallel()

	t.Run("overrides default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxBytes(20 << 20))
		if opts.maxBytes != 20<<20 {
			t.Fatalf("got maxBytes %d, want %d", opts.maxBytes, uint64(20<<20))
		}
	})

	t.Run("zero keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxBytes(0))
		if opts.maxBytes != 10<<20 {
			t.Fatalf("got maxBytes %d, want default %d", opts.maxBytes, uint64(10<<20))
		}
	})

	t.Run("does not affect minAge", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxBytes(20 << 20))
		if opts.minAge != 10*time.Second {
			t.Fatalf("got minAge %v, want %v", opts.minAge, 10*time.Second)
		}
	})
}

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("default values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.minAge != 10*time.Second {
			t.Fatalf("got minAge %v, want %v", opts.minAge, 10*time.Second)
		}

		if opts.maxBytes != 10<<20 {
			t.Fatalf("got maxBytes %d, want %d", opts.maxBytes, uint64(10<<20))
		}
	})

	t.Run("both options applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMinAge(5*time.Second), WithMaxBytes(20<<20))
		if opts.minAge != 5*time.Second {
			t.Fatalf("got minAge %v, want %v", opts.minAge, 5*time.Second)
		}

		if opts.maxBytes != 20<<20 {
			t.Fatalf("got maxBytes %d, want %d", opts.maxBytes, uint64(20<<20))
		}
	})
}
