package stats

import (
	"errors"
	"math"
	"testing"
)

func TestUniform_PDF(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 5}

	if math.Abs(u.PDF(2.5)-0.2) > 1e-10 {
		t.Fatalf("expected 0.2, got %f", u.PDF(2.5))
	}
}

func TestUniform_PDF_outside(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 1}

	if u.PDF(-0.1) != 0 {
		t.Fatalf("expected 0 below range")
	}

	if u.PDF(1.1) != 0 {
		t.Fatalf("expected 0 above range")
	}
}

func TestUniform_CDF(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 1}

	if math.Abs(u.CDF(0.5)-0.5) > 1e-10 {
		t.Fatalf("expected 0.5, got %f", u.CDF(0.5))
	}
}

func TestUniform_CDF_bounds(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 1}

	if u.CDF(-1) != 0 {
		t.Fatalf("expected 0 below range")
	}

	if u.CDF(2) != 1 {
		t.Fatalf("expected 1 above range")
	}
}

func TestUniform_Mean(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 2, Max: 8}

	if u.Mean() != 5 {
		t.Fatalf("expected 5, got %f", u.Mean())
	}
}

func TestUniform_Variance(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 12}

	if math.Abs(u.Variance()-12) > 1e-10 {
		t.Fatalf("expected 12, got %f", u.Variance())
	}
}

func TestUniform_Quantile(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 10}

	if math.Abs(u.Quantile(0.5)-5) > 1e-10 {
		t.Fatalf("expected 5, got %f", u.Quantile(0.5))
	}
}

func TestNewUniform(t *testing.T) {
	t.Parallel()

	u, err := NewUniform(0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Min != 0 || u.Max != 10 {
		t.Fatalf("unexpected values: %+v", u)
	}
}

func TestNewUniform_invertedRange(t *testing.T) {
	t.Parallel()

	_, err := NewUniform(10, 5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewUniform_equalBounds(t *testing.T) {
	t.Parallel()

	_, err := NewUniform(5, 5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}
