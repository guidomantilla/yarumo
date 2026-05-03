package stats

import (
	"errors"
	"math"
	"testing"
)

// --- NewPareto ---

func TestNewPareto(t *testing.T) {
	t.Parallel()

	p, err := NewPareto(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Xm != 1 || p.Alpha != 2 {
		t.Fatalf("unexpected values: %+v", p)
	}
}

func TestNewPareto_invalidXm(t *testing.T) {
	t.Parallel()

	_, err := NewPareto(0, 2)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewPareto_invalidAlpha(t *testing.T) {
	t.Parallel()

	_, err := NewPareto(1, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

// --- PDF ---

func TestPareto_PDF_atXm(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// PDF(1) = alpha * xm^alpha / xm^(alpha+1) = alpha / xm = 3.
	result := p.PDF(1)

	if math.Abs(result-3.0) > 1e-10 {
		t.Fatalf("expected 3.0, got %f", result)
	}
}

func TestPareto_PDF_belowXm(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}

	if p.PDF(0.5) != 0 {
		t.Fatalf("expected 0 below xm")
	}
}

func TestPareto_PDF_above(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// PDF(2) = 3 * 1^3 / 2^4 = 3/16 = 0.1875.
	result := p.PDF(2)

	if math.Abs(result-0.1875) > 1e-10 {
		t.Fatalf("expected 0.1875, got %f", result)
	}
}

// --- CDF ---

func TestPareto_CDF_atXm(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// CDF(1) = 1 - (1/1)^3 = 0.
	result := p.CDF(1)

	if math.Abs(result) > 1e-10 {
		t.Fatalf("expected 0, got %f", result)
	}
}

func TestPareto_CDF_above(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// CDF(2) = 1 - (1/2)^3 = 1 - 0.125 = 0.875.
	result := p.CDF(2)

	if math.Abs(result-0.875) > 1e-10 {
		t.Fatalf("expected 0.875, got %f", result)
	}
}

func TestPareto_CDF_belowXm(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}

	if p.CDF(0.5) != 0 {
		t.Fatalf("expected 0 below xm")
	}
}

// --- Mean ---

func TestPareto_Mean(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// E[X] = alpha * xm / (alpha - 1) = 3/2 = 1.5.
	result := p.Mean()

	if math.Abs(result-1.5) > 1e-10 {
		t.Fatalf("expected 1.5, got %f", result)
	}
}

func TestPareto_Mean_infinite(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 1}
	result := p.Mean()

	if !math.IsInf(result, 1) {
		t.Fatalf("expected +Inf, got %f", result)
	}
}

// --- Variance ---

func TestPareto_Variance(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}
	// Var = xm^2 * alpha / ((alpha-1)^2 * (alpha-2)) = 1*3 / (4*1) = 0.75.
	result := p.Variance()

	if math.Abs(result-0.75) > 1e-10 {
		t.Fatalf("expected 0.75, got %f", result)
	}
}

func TestPareto_Variance_infinite(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 2}
	result := p.Variance()

	if !math.IsInf(result, 1) {
		t.Fatalf("expected +Inf, got %f", result)
	}
}

// --- Quantile ---

func TestPareto_Quantile_roundtrip(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 3}

	x := p.Quantile(0.5)
	cdf := p.CDF(x)

	if math.Abs(cdf-0.5) > 1e-9 {
		t.Fatalf("expected CDF(Quantile(0.5))=0.5, got %f", cdf)
	}
}

func TestPareto_Quantile_atZero(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 2, Alpha: 3}
	result := p.Quantile(0)

	if math.Abs(result-2.0) > 1e-10 {
		t.Fatalf("expected xm=2, got %f", result)
	}
}

// --- String ---

func TestPareto_String(t *testing.T) {
	t.Parallel()

	p := Pareto{Xm: 1, Alpha: 2}
	s := p.String()

	if s != "Pareto(xm=1, α=2)" {
		t.Fatalf("expected %q, got %q", "Pareto(xm=1, α=2)", s)
	}
}
