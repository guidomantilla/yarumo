package stats

import (
	"errors"
	"math"
	"testing"
)

// --- NewGeometric ---

func TestNewGeometric(t *testing.T) {
	t.Parallel()

	g, err := NewGeometric(0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.P != 0.5 {
		t.Fatalf("expected P=0.5, got %f", g.P)
	}
}

func TestNewGeometric_one(t *testing.T) {
	t.Parallel()

	g, err := NewGeometric(1.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.P != 1.0 {
		t.Fatalf("expected P=1.0, got %f", g.P)
	}
}

func TestNewGeometric_zeroP(t *testing.T) {
	t.Parallel()

	_, err := NewGeometric(0)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

func TestNewGeometric_negativeP(t *testing.T) {
	t.Parallel()

	_, err := NewGeometric(-0.1)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

func TestNewGeometric_overOne(t *testing.T) {
	t.Parallel()

	_, err := NewGeometric(1.1)
	if !errors.Is(err, ErrInvalidProb) {
		t.Fatalf("expected ErrInvalidProb, got %v", err)
	}
}

// --- PMF ---

func TestGeometric_PMF_k0(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.3}
	// P(X=0) = p = 0.3.
	result := g.PMF(0)

	if math.Abs(result-0.3) > 1e-10 {
		t.Fatalf("expected 0.3, got %f", result)
	}
}

func TestGeometric_PMF_k2(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.3}
	// P(X=2) = (0.7)^2 * 0.3 = 0.147.
	result := g.PMF(2)

	if math.Abs(result-0.147) > 1e-10 {
		t.Fatalf("expected 0.147, got %f", result)
	}
}

func TestGeometric_PMF_negative(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.5}

	if g.PMF(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

// --- CDFDiscrete ---

func TestGeometric_CDFDiscrete_k0(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.5}
	// P(X<=0) = 1 - (1-0.5)^1 = 0.5.
	result := g.CDFDiscrete(0)

	if math.Abs(result-0.5) > 1e-10 {
		t.Fatalf("expected 0.5, got %f", result)
	}
}

func TestGeometric_CDFDiscrete_k3(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.5}
	// P(X<=3) = 1 - (0.5)^4 = 0.9375.
	result := g.CDFDiscrete(3)

	if math.Abs(result-0.9375) > 1e-10 {
		t.Fatalf("expected 0.9375, got %f", result)
	}
}

func TestGeometric_CDFDiscrete_negative(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.5}

	if g.CDFDiscrete(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

// --- Mean ---

func TestGeometric_Mean(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.25}
	// E[X] = (1-p)/p = 0.75/0.25 = 3.
	result := g.Mean()

	if math.Abs(result-3.0) > 1e-10 {
		t.Fatalf("expected 3.0, got %f", result)
	}
}

// --- Variance ---

func TestGeometric_Variance(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.25}
	// Var = (1-p)/p^2 = 0.75/0.0625 = 12.
	result := g.Variance()

	if math.Abs(result-12.0) > 1e-10 {
		t.Fatalf("expected 12.0, got %f", result)
	}
}

// --- String ---

func TestGeometric_String(t *testing.T) {
	t.Parallel()

	g := Geometric{P: 0.3}
	s := g.String()

	if s != "Geometric(p=0.3)" {
		t.Fatalf("expected %q, got %q", "Geometric(p=0.3)", s)
	}
}
