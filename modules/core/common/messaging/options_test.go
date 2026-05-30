package messaging

import (
	"context"
	"reflect"
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

func TestNewOptions_DefaultErrorHandlerInstalled(t *testing.T) {
	t.Parallel()

	opts := NewOptions()
	if opts.errorHandler == nil {
		t.Fatal("expected DefaultErrorHandler installed, got nil")
	}

	want := reflect.ValueOf(DefaultErrorHandler).Pointer()
	got := reflect.ValueOf(opts.errorHandler).Pointer()
	if got != want {
		t.Fatalf("expected DefaultErrorHandler pointer, got different function")
	}
}

func TestWithErrorHandler_OverridesDefault(t *testing.T) {
	t.Parallel()

	custom := func(_ context.Context, _ any, _ error) {}

	opts := NewOptions(WithErrorHandler(custom))
	if opts.errorHandler == nil {
		t.Fatal("expected custom hook installed, got nil")
	}

	want := reflect.ValueOf(custom).Pointer()
	got := reflect.ValueOf(opts.errorHandler).Pointer()
	if got != want {
		t.Fatalf("expected custom hook, got DefaultErrorHandler")
	}
}

func TestWithErrorHandler_NilPreservesDefault(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithErrorHandler(nil))

	want := reflect.ValueOf(DefaultErrorHandler).Pointer()
	got := reflect.ValueOf(opts.errorHandler).Pointer()
	if got != want {
		t.Fatalf("expected DefaultErrorHandler when nil passed, got different function")
	}
}

func TestNewOptions_DefaultOverflowPolicyIsReject(t *testing.T) {
	t.Parallel()

	opts := NewOptions()
	if opts.overflowPolicy != OverflowReject {
		t.Fatalf("expected default OverflowReject, got %v", opts.overflowPolicy)
	}
}

func TestWithOverflowPolicy_AppliesValid(t *testing.T) {
	t.Parallel()

	opts := NewOptions(WithOverflowPolicy(OverflowDropOldest))
	if opts.overflowPolicy != OverflowDropOldest {
		t.Fatalf("expected OverflowDropOldest, got %v", opts.overflowPolicy)
	}
}

func TestWithOverflowPolicy_OutOfRangeIgnored(t *testing.T) {
	t.Parallel()

	t.Run("positive out of range", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOverflowPolicy(OverflowPolicy(99)))
		if opts.overflowPolicy != OverflowReject {
			t.Fatalf("expected default preserved, got %v", opts.overflowPolicy)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithOverflowPolicy(OverflowPolicy(-1)))
		if opts.overflowPolicy != OverflowReject {
			t.Fatalf("expected default preserved, got %v", opts.overflowPolicy)
		}
	})
}
