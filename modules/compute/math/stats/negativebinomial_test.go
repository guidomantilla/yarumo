package stats

import (
	"errors"
	"math"
	"testing"
)

// --- NewNegativeBinomial ---

func TestNewNegativeBinomial(t *testing.T) {
	t.Parallel()

	nb, err := NewNegativeBinomial(3, 0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nb.R != 3 || nb.P != 0.5 {
		t.Fatalf("unexpected values: %+v", nb)
	}
}

func TestNewNegativeBinomial_invalidR(t *testing.T) {
	t.Parallel()

	_, err := NewNegativeBinomial(0, 0.5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewNegativeBinomial_zeroP(t *testing.T) {
	t.Parallel()

	_, err := NewNegativeBinomial(3, 0)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

func TestNewNegativeBinomial_overOne(t *testing.T) {
	t.Parallel()

	_, err := NewNegativeBinomial(3, 1.1)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

// --- PMF ---

func TestNegativeBinomial_PMF_k0(t *testing.T) {
	t.Parallel()

	// r=1, p=0.5: P(X=0) = C(0,0) * 0.5^1 * 0.5^0 = 0.5.
	nb := NegativeBinomial{R: 1, P: 0.5}
	result := nb.PMF(0)

	if math.Abs(result-0.5) > 1e-10 {
		t.Fatalf("expected 0.5, got %f", result)
	}
}

func TestNegativeBinomial_PMF_k3(t *testing.T) {
	t.Parallel()

	// r=2, p=0.3: P(X=3) = C(4,3) * 0.3^2 * 0.7^3 = 4 * 0.09 * 0.343 = 0.12348.
	nb := NegativeBinomial{R: 2, P: 0.3}
	result := nb.PMF(3)

	if math.Abs(result-0.12348) > 1e-5 {
		t.Fatalf("expected ~0.12348, got %f", result)
	}
}

func TestNegativeBinomial_PMF_negative(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 1, P: 0.5}

	if nb.PMF(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

// --- CDFDiscrete ---

func TestNegativeBinomial_CDFDiscrete(t *testing.T) {
	t.Parallel()

	// r=1, p=1.0: P(X=0) = 1, so CDF(0) = 1.
	nb := NegativeBinomial{R: 1, P: 1.0}
	result := nb.CDFDiscrete(0)

	if math.Abs(result-1.0) > 1e-10 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

func TestNegativeBinomial_CDFDiscrete_negative(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 1, P: 0.5}

	if nb.CDFDiscrete(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

func TestNegativeBinomial_CDFDiscrete_converges(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 1, P: 0.5}
	result := nb.CDFDiscrete(20)

	if result < 0.999 {
		t.Fatalf("expected CDF near 1.0 for large k, got %f", result)
	}
}

// --- Mean ---

func TestNegativeBinomial_Mean(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 5, P: 0.5}
	// E[X] = r*(1-p)/p = 5*0.5/0.5 = 5.
	result := nb.Mean()

	if math.Abs(result-5.0) > 1e-10 {
		t.Fatalf("expected 5.0, got %f", result)
	}
}

// --- Variance ---

func TestNegativeBinomial_Variance(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 5, P: 0.5}
	// Var = r*(1-p)/p^2 = 5*0.5/0.25 = 10.
	result := nb.Variance()

	if math.Abs(result-10.0) > 1e-10 {
		t.Fatalf("expected 10.0, got %f", result)
	}
}

// --- String ---

func TestNegativeBinomial_String(t *testing.T) {
	t.Parallel()

	nb := NegativeBinomial{R: 3, P: 0.5}
	s := nb.String()

	if s != "NegativeBinomial(r=3, p=0.5)" {
		t.Fatalf("expected %q, got %q", "NegativeBinomial(r=3, p=0.5)", s)
	}
}
