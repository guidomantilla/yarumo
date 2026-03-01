package fuzzy

import (
	"errors"
	"math"
	"testing"
)

func TestFuzzify(t *testing.T) {
	t.Parallel()

	fn := Triangular(0, 5, 10)
	result := Fuzzify(fn, 5)

	if result != 1.0 {
		t.Fatalf("expected 1.0, got %f", float64(result))
	}
}

func TestClip_belowLevel(t *testing.T) {
	t.Parallel()

	fn := Constant(0.3)
	clipped := Clip(fn, 0.5)

	if clipped(0) != 0.3 {
		t.Fatalf("expected 0.3 (below clip level), got %f", float64(clipped(0)))
	}
}

func TestClip_aboveLevel(t *testing.T) {
	t.Parallel()

	fn := Constant(0.8)
	clipped := Clip(fn, 0.5)

	if clipped(0) != 0.5 {
		t.Fatalf("expected 0.5 (clipped), got %f", float64(clipped(0)))
	}
}

func TestClip_atLevel(t *testing.T) {
	t.Parallel()

	fn := Constant(0.5)
	clipped := Clip(fn, 0.5)

	if clipped(0) != 0.5 {
		t.Fatalf("expected 0.5, got %f", float64(clipped(0)))
	}
}

func TestScale(t *testing.T) {
	t.Parallel()

	fn := Constant(0.8)
	scaled := Scale(fn, 0.5)
	result := scaled(0)

	if math.Abs(float64(result)-0.4) > 1e-9 {
		t.Fatalf("expected 0.4, got %f", float64(result))
	}
}

func TestAggregateMax(t *testing.T) {
	t.Parallel()

	fn1 := Constant(0.3)
	fn2 := Constant(0.7)
	fn3 := Constant(0.5)

	agg := AggregateMax(fn1, fn2, fn3)

	if agg(0) != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(agg(0)))
	}
}

func TestAggregateMax_empty(t *testing.T) {
	t.Parallel()

	agg := AggregateMax()

	if agg(0) != 0 {
		t.Fatalf("expected 0, got %f", float64(agg(0)))
	}
}

func TestSample_basic(t *testing.T) {
	t.Parallel()

	xs, ys, err := Sample(Constant(0.5), 0, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(xs) != 5 {
		t.Fatalf("expected 5 points, got %d", len(xs))
	}

	if len(ys) != 5 {
		t.Fatalf("expected 5 degrees, got %d", len(ys))
	}

	if xs[0] != 0 {
		t.Fatalf("expected first x=0, got %f", xs[0])
	}

	if xs[4] != 10 {
		t.Fatalf("expected last x=10, got %f", xs[4])
	}

	for i, y := range ys {
		if y != 0.5 {
			t.Fatalf("expected 0.5 at index %d, got %f", i, float64(y))
		}
	}
}

func TestSample_singlePoint(t *testing.T) {
	t.Parallel()

	xs, ys, err := Sample(Constant(0.8), 5, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(xs) != 1 {
		t.Fatalf("expected 1 point, got %d", len(xs))
	}

	if xs[0] != 5 {
		t.Fatalf("expected x=5, got %f", xs[0])
	}

	if ys[0] != 0.8 {
		t.Fatalf("expected 0.8, got %f", float64(ys[0]))
	}
}

func TestSample_zeroPoints(t *testing.T) {
	t.Parallel()

	_, _, err := Sample(Constant(1), 0, 10, 0)
	if !errors.Is(err, ErrEmptySamples) {
		t.Fatalf("expected ErrEmptySamples, got %v", err)
	}
}

func TestSample_negativePoints(t *testing.T) {
	t.Parallel()

	_, _, err := Sample(Constant(1), 0, 10, -1)
	if !errors.Is(err, ErrEmptySamples) {
		t.Fatalf("expected ErrEmptySamples, got %v", err)
	}
}

func TestSample_invalidRange(t *testing.T) {
	t.Parallel()

	_, _, err := Sample(Constant(1), 10, 0, 5)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestSample_equalRange(t *testing.T) {
	t.Parallel()

	_, _, err := Sample(Constant(1), 5, 5, 5)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}
