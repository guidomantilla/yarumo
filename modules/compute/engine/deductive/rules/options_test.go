package rules

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.priority != 0 {
			t.Fatalf("expected priority 0, got %d", opts.priority)
		}
	})

	t.Run("with priority", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPriority(3))
		if opts.priority != 3 {
			t.Fatalf("expected priority 3, got %d", opts.priority)
		}
	})
}

func TestWithPriority(t *testing.T) {
	t.Parallel()

	t.Run("positive value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPriority(10))
		if opts.priority != 10 {
			t.Fatalf("expected 10, got %d", opts.priority)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPriority(0))
		if opts.priority != 0 {
			t.Fatalf("expected 0, got %d", opts.priority)
		}
	})

	t.Run("negative value ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPriority(-1))
		if opts.priority != 0 {
			t.Fatalf("expected 0 (default), got %d", opts.priority)
		}
	})
}
