package stats

import (
	"errors"
	"math"
	"testing"
)

func TestBeta_PDF_symmetric(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 2}
	result := b.PDF(0.5)

	// Beta(2,2) peak at x=0.5, PDF = 6*0.5*0.5 = 1.5
	if math.Abs(result-1.5) > 1e-6 {
		t.Fatalf("expected 1.5, got %f", result)
	}
}

func TestBeta_PDF_boundary(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 2}

	if b.PDF(0) != 0 {
		t.Fatalf("expected 0 at x=0")
	}

	if b.PDF(1) != 0 {
		t.Fatalf("expected 0 at x=1")
	}
}

func TestBeta_CDF_midpoint(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 2}
	result := b.CDF(0.5)

	// Symmetric: CDF at 0.5 should be ~0.5.
	if math.Abs(result-0.5) > 0.01 {
		t.Fatalf("expected ~0.5, got %f", result)
	}
}

func TestBeta_CDF_bounds(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 2}

	if b.CDF(0) != 0 {
		t.Fatalf("expected 0 at x=0")
	}

	if b.CDF(1) != 1 {
		t.Fatalf("expected 1 at x=1")
	}
}

func TestBeta_Mean(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 5}
	expected := 2.0 / 7.0

	if math.Abs(b.Mean()-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, b.Mean())
	}
}

func TestBeta_Variance(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 5}
	expected := (2.0 * 5.0) / (49.0 * 8.0)

	if math.Abs(b.Variance()-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, b.Variance())
	}
}

func TestBeta_Quantile_median(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 2}
	result := b.Quantile(0.5)

	if math.Abs(result-0.5) > 0.01 {
		t.Fatalf("expected ~0.5, got %f", result)
	}
}

func TestNewBeta(t *testing.T) {
	t.Parallel()

	b, err := NewBeta(2, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.Alpha != 2 || b.Bet != 5 {
		t.Fatalf("unexpected values: %+v", b)
	}
}

func TestNewBeta_invalidAlpha(t *testing.T) {
	t.Parallel()

	_, err := NewBeta(0, 5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewBeta_invalidBeta(t *testing.T) {
	t.Parallel()

	_, err := NewBeta(2, -1)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}
