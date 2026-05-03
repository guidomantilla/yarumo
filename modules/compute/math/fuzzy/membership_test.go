package fuzzy

import (
	"errors"
	"math"
	"testing"
)

func TestTriangular_peak(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(5) != 1.0 {
		t.Fatalf("expected 1.0 at peak, got %f", float64(fn(5)))
	}
}

func TestTriangular_left(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(0) != 0 {
		t.Fatalf("expected 0 at left edge, got %f", float64(fn(0)))
	}
}

func TestTriangular_right(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(10) != 0 {
		t.Fatalf("expected 0 at right edge, got %f", float64(fn(10)))
	}
}

func TestTriangular_midLeft(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := fn(2.5)

	if math.Abs(float64(result)-0.5) > 1e-9 {
		t.Fatalf("expected 0.5, got %f", float64(result))
	}
}

func TestTriangular_midRight(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := fn(7.5)

	if math.Abs(float64(result)-0.5) > 1e-9 {
		t.Fatalf("expected 0.5, got %f", float64(result))
	}
}

func TestTriangular_outside(t *testing.T) {
	t.Parallel()

	fn, err := Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(-1) != 0 {
		t.Fatalf("expected 0 outside range, got %f", float64(fn(-1)))
	}

	if fn(11) != 0 {
		t.Fatalf("expected 0 outside range, got %f", float64(fn(11)))
	}
}

func TestTriangular_invalidRange(t *testing.T) {
	t.Parallel()

	_, err := Triangular(10, 5, 0)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestTriangular_peakOutside(t *testing.T) {
	t.Parallel()

	_, err := Triangular(0, 11, 10)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestTrapezoidal_plateau(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(5) != 1.0 {
		t.Fatalf("expected 1.0 on plateau, got %f", float64(fn(5)))
	}
}

func TestTrapezoidal_edges(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(0) != 0 {
		t.Fatalf("expected 0 at left edge, got %f", float64(fn(0)))
	}

	if fn(10) != 0 {
		t.Fatalf("expected 0 at right edge, got %f", float64(fn(10)))
	}
}

func TestTrapezoidal_rise(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := fn(1.5)

	if math.Abs(float64(result)-0.5) > 1e-9 {
		t.Fatalf("expected 0.5, got %f", float64(result))
	}
}

func TestTrapezoidal_fall(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := fn(8.5)

	if math.Abs(float64(result)-0.5) > 1e-9 {
		t.Fatalf("expected 0.5, got %f", float64(result))
	}
}

func TestTrapezoidal_outside(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(-1) != 0 {
		t.Fatalf("expected 0 outside, got %f", float64(fn(-1)))
	}

	if fn(11) != 0 {
		t.Fatalf("expected 0 outside, got %f", float64(fn(11)))
	}
}

func TestTrapezoidal_plateauEdges(t *testing.T) {
	t.Parallel()

	fn, err := Trapezoidal(0, 3, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fn(3) != 1.0 {
		t.Fatalf("expected 1.0 at b, got %f", float64(fn(3)))
	}

	if fn(7) != 1.0 {
		t.Fatalf("expected 1.0 at c, got %f", float64(fn(7)))
	}
}

func TestTrapezoidal_invalidRange(t *testing.T) {
	t.Parallel()

	_, err := Trapezoidal(10, 7, 3, 0)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestTrapezoidal_invertedPlateau(t *testing.T) {
	t.Parallel()

	_, err := Trapezoidal(0, 7, 3, 10)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestGaussian_center(t *testing.T) {
	t.Parallel()

	fn, err := Gaussian(5, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(float64(fn(5))-1.0) > 1e-9 {
		t.Fatalf("expected 1.0 at center, got %f", float64(fn(5)))
	}
}

func TestGaussian_spread(t *testing.T) {
	t.Parallel()

	fn, err := Gaussian(5, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At 1 sigma, value should be ~0.6065.
	result := fn(7)
	expected := math.Exp(-0.5)

	if math.Abs(float64(result)-expected) > 1e-9 {
		t.Fatalf("expected %f at 1 sigma, got %f", expected, float64(result))
	}
}

func TestGaussian_zeroSigma(t *testing.T) {
	t.Parallel()

	_, err := Gaussian(5, 0)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestGaussian_negativeSigma(t *testing.T) {
	t.Parallel()

	_, err := Gaussian(5, -1)
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("expected ErrInvalidRange, got %v", err)
	}
}

func TestSigmoid_center(t *testing.T) {
	t.Parallel()

	fn := Sigmoid(5, 1)

	if math.Abs(float64(fn(5))-0.5) > 1e-9 {
		t.Fatalf("expected 0.5 at center, got %f", float64(fn(5)))
	}
}

func TestSigmoid_extremes(t *testing.T) {
	t.Parallel()

	fn := Sigmoid(0, 10)

	// Far right should be near 1.
	if float64(fn(10)) < 0.99 {
		t.Fatalf("expected near 1.0 far right, got %f", float64(fn(10)))
	}

	// Far left should be near 0.
	if float64(fn(-10)) > 0.01 {
		t.Fatalf("expected near 0.0 far left, got %f", float64(fn(-10)))
	}
}

func TestConstant(t *testing.T) {
	t.Parallel()

	fn := Constant(0.42)

	if fn(0) != 0.42 {
		t.Fatalf("expected 0.42, got %f", float64(fn(0)))
	}

	if fn(100) != 0.42 {
		t.Fatalf("expected 0.42, got %f", float64(fn(100)))
	}

	if fn(-50) != 0.42 {
		t.Fatalf("expected 0.42, got %f", float64(fn(-50)))
	}
}
