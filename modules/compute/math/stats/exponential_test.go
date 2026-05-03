package stats

import (
	"errors"
	"math"
	"testing"
)

func TestExponential_PDF(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 2}
	result := e.PDF(0)

	if math.Abs(result-2) > 1e-10 {
		t.Fatalf("expected 2, got %f", result)
	}
}

func TestExponential_PDF_negative(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 1}

	if e.PDF(-1) != 0 {
		t.Fatalf("expected 0 for negative x, got %f", e.PDF(-1))
	}
}

func TestExponential_CDF(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 1}
	result := e.CDF(1)

	expected := 1 - math.Exp(-1)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestExponential_CDF_negative(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 1}

	if e.CDF(-1) != 0 {
		t.Fatalf("expected 0 for negative x")
	}
}

func TestExponential_Mean(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 2}

	if e.Mean() != 0.5 {
		t.Fatalf("expected 0.5, got %f", e.Mean())
	}
}

func TestExponential_Variance(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 2}

	if e.Variance() != 0.25 {
		t.Fatalf("expected 0.25, got %f", e.Variance())
	}
}

func TestExponential_Quantile(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 1}
	result := e.Quantile(0.5)

	expected := math.Ln2

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestNewExponential(t *testing.T) {
	t.Parallel()

	e, err := NewExponential(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Lambda != 2 {
		t.Fatalf("expected lambda 2, got %f", e.Lambda)
	}
}

func TestNewExponential_invalidLambda(t *testing.T) {
	t.Parallel()

	_, err := NewExponential(0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewExponential_negativeLambda(t *testing.T) {
	t.Parallel()

	_, err := NewExponential(-1)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}
