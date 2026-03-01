package engine

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()
		if opts.maxIterations != defaultMaxIterations {
			t.Fatalf("expected %d, got %d", defaultMaxIterations, opts.maxIterations)
		}

		if opts.strategy != PriorityOrder {
			t.Fatalf("expected PriorityOrder, got %d", opts.strategy)
		}
	})

	t.Run("with options", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIterations(50), WithStrategy(FirstMatch))

		if opts.maxIterations != 50 {
			t.Fatalf("expected 50, got %d", opts.maxIterations)
		}

		if opts.strategy != FirstMatch {
			t.Fatalf("expected FirstMatch, got %d", opts.strategy)
		}
	})
}

func TestWithMaxIterations(t *testing.T) {
	t.Parallel()

	t.Run("positive value", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIterations(100))
		if opts.maxIterations != 100 {
			t.Fatalf("expected 100, got %d", opts.maxIterations)
		}
	})

	t.Run("zero ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIterations(0))
		if opts.maxIterations != defaultMaxIterations {
			t.Fatalf("expected default %d, got %d", defaultMaxIterations, opts.maxIterations)
		}
	})

	t.Run("negative ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIterations(-5))
		if opts.maxIterations != defaultMaxIterations {
			t.Fatalf("expected default %d, got %d", defaultMaxIterations, opts.maxIterations)
		}
	})
}

func TestWithStrategy(t *testing.T) {
	t.Parallel()

	t.Run("PriorityOrder", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithStrategy(PriorityOrder))
		if opts.strategy != PriorityOrder {
			t.Fatalf("expected PriorityOrder, got %d", opts.strategy)
		}
	})

	t.Run("FirstMatch", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithStrategy(FirstMatch))
		if opts.strategy != FirstMatch {
			t.Fatalf("expected FirstMatch, got %d", opts.strategy)
		}
	})

	t.Run("invalid strategy ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithStrategy(Strategy(99)))
		if opts.strategy != PriorityOrder {
			t.Fatalf("expected PriorityOrder (default), got %d", opts.strategy)
		}
	})
}
