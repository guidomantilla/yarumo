package stats

import (
	"errors"
	"math"
	"testing"
)

func TestGamma_PDF(t *testing.T) {
	t.Parallel()

	// Gamma(1, 1) = Exponential(1).
	g := Gamma{Alpha: 1, Bet: 1}
	result := g.PDF(1)

	expected := math.Exp(-1)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGamma_PDF_negative(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 2, Bet: 1}

	if g.PDF(-1) != 0 {
		t.Fatalf("expected 0 for negative x")
	}
}

func TestGamma_PDF_zero(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 2, Bet: 1}

	if g.PDF(0) != 0 {
		t.Fatalf("expected 0 at x=0 for alpha>1")
	}
}

func TestGamma_CDF(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 1, Bet: 1}
	result := g.CDF(1)

	// Gamma(1,1) = Exp(1): CDF(1) = 1 - e^{-1}
	expected := 1 - math.Exp(-1)

	if math.Abs(result-expected) > 0.01 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGamma_CDF_negative(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 2, Bet: 1}

	if g.CDF(-1) != 0 {
		t.Fatalf("expected 0 for negative x")
	}
}

func TestGamma_Mean(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 3, Bet: 2}

	if g.Mean() != 1.5 {
		t.Fatalf("expected 1.5, got %f", g.Mean())
	}
}

func TestGamma_Variance(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 3, Bet: 2}

	if g.Variance() != 0.75 {
		t.Fatalf("expected 0.75, got %f", g.Variance())
	}
}

func TestGamma_Quantile(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 1, Bet: 1}
	result := g.Quantile(0.5)

	// Gamma(1,1) = Exp(1): quantile(0.5) = ln(2)
	expected := math.Ln2

	if math.Abs(result-expected) > 0.01 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGamma_Quantile_expandBound(t *testing.T) {
	t.Parallel()

	// Small alpha, small beta: mean = 0.001, mean*10 = 0.01.
	// CDF(0.01) is very small for Gamma(0.001, 1), so hi needs expanding.
	g := Gamma{Alpha: 0.001, Bet: 1}
	result := g.Quantile(0.999)

	// Must be significantly larger than mean*10.
	if result <= g.Mean()*10 {
		t.Fatalf("expected quantile > mean*10, got %f", result)
	}
}

func TestNewGamma(t *testing.T) {
	t.Parallel()

	g, err := NewGamma(3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Alpha != 3 || g.Bet != 2 {
		t.Fatalf("unexpected values: %+v", g)
	}
}

func TestNewGamma_invalidAlpha(t *testing.T) {
	t.Parallel()

	_, err := NewGamma(0, 2)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewGamma_invalidBeta(t *testing.T) {
	t.Parallel()

	_, err := NewGamma(3, -1)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}
