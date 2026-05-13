package health

import (
	"runtime"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies defaults", func(t *testing.T) {
		t.Parallel()

		o := NewOptions()
		if o == nil {
			t.Fatalf("NewOptions returned nil")
		}

		want := max(runtime.NumCPU(), 1)

		if o.concurrency != want {
			t.Fatalf("default concurrency = %d, want %d", o.concurrency, want)
		}
	})

	t.Run("applies provided options", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithConcurrency(7))

		if o.concurrency != 7 {
			t.Fatalf("concurrency = %d, want 7", o.concurrency)
		}
	})
}

func TestWithConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("positive value is applied", func(t *testing.T) {
		t.Parallel()

		o := NewOptions(WithConcurrency(12))
		if o.concurrency != 12 {
			t.Fatalf("concurrency = %d, want 12", o.concurrency)
		}
	})

	t.Run("zero is ignored — default preserved", func(t *testing.T) {
		t.Parallel()

		base := NewOptions()
		o := NewOptions(WithConcurrency(0))

		if o.concurrency != base.concurrency {
			t.Fatalf("concurrency = %d, want default %d (zero must be ignored)", o.concurrency, base.concurrency)
		}
	})

	t.Run("negative is ignored — default preserved", func(t *testing.T) {
		t.Parallel()

		base := NewOptions()
		o := NewOptions(WithConcurrency(-3))

		if o.concurrency != base.concurrency {
			t.Fatalf("concurrency = %d, want default %d (negative must be ignored)", o.concurrency, base.concurrency)
		}
	})
}

func TestDefaultConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("matches runtime.NumCPU", func(t *testing.T) {
		t.Parallel()

		got := defaultConcurrency()
		want := runtime.NumCPU()

		if got != want {
			t.Fatalf("defaultConcurrency() = %d, want %d", got, want)
		}

		if got < 1 {
			t.Fatalf("defaultConcurrency() = %d, want >= 1", got)
		}
	})
}
