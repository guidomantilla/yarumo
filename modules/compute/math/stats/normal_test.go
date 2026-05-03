package stats

import (
	"errors"
	"math"
	"testing"
)

func TestNormal_PDF_standard(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.PDF(0)

	expected := 1 / math.Sqrt(2*math.Pi)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestNormal_PDF_shifted(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 5, Sigma: 2}
	result := n.PDF(5)

	expected := 1 / (2 * math.Sqrt(2*math.Pi))

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestNormal_CDF_standard(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}

	if math.Abs(n.CDF(0)-0.5) > 1e-10 {
		t.Fatalf("expected 0.5, got %f", n.CDF(0))
	}
}

func TestNormal_CDF_leftTail(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.CDF(-3)

	if result > 0.01 {
		t.Fatalf("expected small value, got %f", result)
	}
}

func TestNormal_CDF_rightTail(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.CDF(3)

	if result < 0.99 {
		t.Fatalf("expected near 1, got %f", result)
	}
}

func TestNormal_Mean(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 5, Sigma: 2}

	if n.Mean() != 5 {
		t.Fatalf("expected 5, got %f", n.Mean())
	}
}

func TestNormal_Variance(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 3}

	if n.Variance() != 9 {
		t.Fatalf("expected 9, got %f", n.Variance())
	}
}

func TestNormal_Quantile_median(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.Quantile(0.5)

	if math.Abs(result) > 1e-6 {
		t.Fatalf("expected ~0, got %f", result)
	}
}

func TestNormal_Quantile_lowerTail(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.Quantile(0.01)

	// z ≈ -2.326
	if math.Abs(result-(-2.326)) > 0.01 {
		t.Fatalf("expected ~-2.326, got %f", result)
	}
}

func TestNormal_Quantile_upperTail(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}
	result := n.Quantile(0.99)

	// z ≈ 2.326
	if math.Abs(result-2.326) > 0.01 {
		t.Fatalf("expected ~2.326, got %f", result)
	}
}

func TestNewNormal(t *testing.T) {
	t.Parallel()

	n, err := NewNormal(0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n.Mu != 0 || n.Sigma != 1 {
		t.Fatalf("unexpected values: %+v", n)
	}
}

func TestNewNormal_invalidSigma(t *testing.T) {
	t.Parallel()

	_, err := NewNormal(0, -1)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewNormal_zeroSigma(t *testing.T) {
	t.Parallel()

	_, err := NewNormal(0, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}
