package stats

import (
	"errors"
	"math"
	"testing"
)

// --- NewHypergeometric ---

func TestNewHypergeometric(t *testing.T) {
	t.Parallel()

	h, err := NewHypergeometric(50, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if h.N != 50 || h.K != 10 || h.Draws != 5 {
		t.Fatalf("unexpected values: %+v", h)
	}
}

func TestNewHypergeometric_invalidN(t *testing.T) {
	t.Parallel()

	_, err := NewHypergeometric(0, 0, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewHypergeometric_negativeK(t *testing.T) {
	t.Parallel()

	_, err := NewHypergeometric(50, -1, 5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewHypergeometric_KGreaterThanN(t *testing.T) {
	t.Parallel()

	_, err := NewHypergeometric(50, 51, 5)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewHypergeometric_zeroDraws(t *testing.T) {
	t.Parallel()

	_, err := NewHypergeometric(50, 10, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewHypergeometric_drawsGreaterThanN(t *testing.T) {
	t.Parallel()

	_, err := NewHypergeometric(50, 10, 51)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

// --- PMF ---

func TestHypergeometric_PMF(t *testing.T) {
	t.Parallel()

	// Classic urn: N=20, K=7, n=12, k=4.
	// C(7,4)*C(13,8)/C(20,12) = 35*1287/125970 ≈ 0.3576.
	h := Hypergeometric{N: 20, K: 7, Draws: 12}
	result := h.PMF(4)

	if math.Abs(result-0.35764) > 1e-4 {
		t.Fatalf("expected ~0.3576, got %f", result)
	}
}

func TestHypergeometric_PMF_zero(t *testing.T) {
	t.Parallel()

	// N=10, K=3, n=5. Min k = max(0, 5-7)=0, Max k = min(3,5)=3.
	h := Hypergeometric{N: 10, K: 3, Draws: 5}

	if h.PMF(4) != 0 {
		t.Fatalf("expected 0 for k > max")
	}
}

func TestHypergeometric_PMF_negative(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 10, K: 3, Draws: 5}

	if h.PMF(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

func TestHypergeometric_PMF_minK(t *testing.T) {
	t.Parallel()

	// N=10, K=8, n=5. Min k = max(0, 5-2)=3.
	h := Hypergeometric{N: 10, K: 8, Draws: 5}

	if h.PMF(2) != 0 {
		t.Fatalf("expected 0 for k < minK")
	}

	if h.PMF(3) <= 0 {
		t.Fatalf("expected positive PMF at minK=3")
	}
}

// --- CDFDiscrete ---

func TestHypergeometric_CDFDiscrete(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 20, K: 7, Draws: 12}
	result := h.CDFDiscrete(7)

	if math.Abs(result-1.0) > 1e-9 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

func TestHypergeometric_CDFDiscrete_negative(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 20, K: 7, Draws: 12}

	if h.CDFDiscrete(-1) != 0 {
		t.Fatalf("expected 0 for negative k")
	}
}

func TestHypergeometric_CDFDiscrete_sums(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 10, K: 3, Draws: 5}

	// Sum of all PMF should be 1.
	sum := 0.0
	for k := range 4 {
		sum += h.PMF(k)
	}

	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("expected PMF sum 1.0, got %f", sum)
	}
}

// --- Mean ---

func TestHypergeometric_Mean(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 50, K: 10, Draws: 5}
	// E[X] = n*K/N = 5*10/50 = 1.0.
	result := h.Mean()

	if math.Abs(result-1.0) > 1e-10 {
		t.Fatalf("expected 1.0, got %f", result)
	}
}

// --- Variance ---

func TestHypergeometric_Variance(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 50, K: 10, Draws: 5}
	// Var = n*(K/N)*((N-K)/N)*((N-n)/(N-1)) = 5*(10/50)*(40/50)*(45/49).
	expected := 5.0 * (10.0 / 50.0) * (40.0 / 50.0) * (45.0 / 49.0)
	result := h.Variance()

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestHypergeometric_CDFDiscrete_partial(t *testing.T) {
	t.Parallel()

	// N=10, K=3, n=5. Valid k range: [0, 3].
	h := Hypergeometric{N: 10, K: 3, Draws: 5}

	// CDF(1) should be sum of PMF(0) + PMF(1).
	result := h.CDFDiscrete(1)
	expected := h.PMF(0) + h.PMF(1)

	if math.Abs(result-expected) > 1e-9 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestHypergeometric_CDFDiscrete_belowMin(t *testing.T) {
	t.Parallel()

	// N=10, K=8, n=5. minK = max(0, 5-2) = 3.
	h := Hypergeometric{N: 10, K: 8, Draws: 5}

	if h.CDFDiscrete(2) != 0 {
		t.Fatalf("expected 0 for k < minK")
	}
}

func TestHypergeometric_CDFDiscrete_atMax(t *testing.T) {
	t.Parallel()

	// N=10, K=3, n=5. maxK = min(3,5) = 3.
	h := Hypergeometric{N: 10, K: 3, Draws: 5}
	result := h.CDFDiscrete(3)

	if math.Abs(result-1.0) > 1e-9 {
		t.Fatalf("expected 1.0 at maxK, got %f", result)
	}
}

// --- String ---

func TestHypergeometric_String(t *testing.T) {
	t.Parallel()

	h := Hypergeometric{N: 50, K: 10, Draws: 5}
	s := h.String()

	if s != "Hypergeometric(N=50, K=10, n=5)" {
		t.Fatalf("expected %q, got %q", "Hypergeometric(N=50, K=10, n=5)", s)
	}
}
