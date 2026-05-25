package messaging

import (
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.bufferSize != defaultBufferSize {
			t.Fatalf("expected bufferSize %d, got %d", defaultBufferSize, opts.bufferSize)
		}
		if opts.drainTimeout != defaultDrainTimeout {
			t.Fatalf("expected drainTimeout %v, got %v", defaultDrainTimeout, opts.drainTimeout)
		}
	})

	t.Run("custom options applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBufferSize(8), WithDrainTimeout(time.Minute))
		if opts.bufferSize != 8 {
			t.Fatalf("expected bufferSize 8, got %d", opts.bufferSize)
		}
		if opts.drainTimeout != time.Minute {
			t.Fatalf("expected drainTimeout 1m, got %v", opts.drainTimeout)
		}
	})
}

func TestWithBufferSize(t *testing.T) {
	t.Parallel()

	t.Run("positive value applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBufferSize(128))
		if opts.bufferSize != 128 {
			t.Fatalf("expected 128, got %d", opts.bufferSize)
		}
	})

	t.Run("zero ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBufferSize(0))
		if opts.bufferSize != defaultBufferSize {
			t.Fatalf("expected default %d, got %d", defaultBufferSize, opts.bufferSize)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBufferSize(-5))
		if opts.bufferSize != defaultBufferSize {
			t.Fatalf("expected default %d, got %d", defaultBufferSize, opts.bufferSize)
		}
	})
}

func TestWithDrainTimeout(t *testing.T) {
	t.Parallel()

	t.Run("positive value applied", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDrainTimeout(2 * time.Second))
		if opts.drainTimeout != 2*time.Second {
			t.Fatalf("expected 2s, got %v", opts.drainTimeout)
		}
	})

	t.Run("zero ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDrainTimeout(0))
		if opts.drainTimeout != defaultDrainTimeout {
			t.Fatalf("expected default %v, got %v", defaultDrainTimeout, opts.drainTimeout)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDrainTimeout(-time.Second))
		if opts.drainTimeout != defaultDrainTimeout {
			t.Fatalf("expected default %v, got %v", defaultDrainTimeout, opts.drainTimeout)
		}
	})
}
