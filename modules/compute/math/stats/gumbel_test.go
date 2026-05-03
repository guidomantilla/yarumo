package stats

import (
	"errors"
	"math"
	"testing"
)

// --- NewGumbel ---

func TestNewGumbel(t *testing.T) {
	t.Parallel()

	g, err := NewGumbel(0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Mu != 0 || g.Beta != 1 {
		t.Fatalf("unexpected values: %+v", g)
	}
}

func TestNewGumbel_negativeMu(t *testing.T) {
	t.Parallel()

	g, err := NewGumbel(-5, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Mu != -5 {
		t.Fatalf("expected Mu=-5, got %f", g.Mu)
	}
}

func TestNewGumbel_invalidBeta(t *testing.T) {
	t.Parallel()

	_, err := NewGumbel(0, 0)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

func TestNewGumbel_negativeBeta(t *testing.T) {
	t.Parallel()

	_, err := NewGumbel(0, -1)
	if !errors.Is(err, ErrInvalidParameter) {
		t.Fatalf("expected ErrInvalidParameter, got %v", err)
	}
}

// --- PDF ---

func TestGumbel_PDF_mode(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	// PDF at mode (x=mu) = exp(-exp(0)) / beta = exp(-1) / 1 ≈ 0.3679.
	result := g.PDF(0)
	expected := math.Exp(-1)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGumbel_PDF_positive(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}

	if g.PDF(5) <= 0 {
		t.Fatalf("expected positive PDF")
	}
}

// --- CDF ---

func TestGumbel_CDF_atMu(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	// CDF(mu) = exp(-exp(0)) = exp(-1) ≈ 0.3679.
	result := g.CDF(0)
	expected := math.Exp(-1)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGumbel_CDF_large(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	result := g.CDF(10)

	if result < 0.999 {
		t.Fatalf("expected CDF near 1.0, got %f", result)
	}
}

// --- Mean ---

func TestGumbel_Mean(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	// E[X] = mu + beta * γ ≈ 0.5772.
	result := g.Mean()

	if math.Abs(result-eulerMascheroni) > 1e-10 {
		t.Fatalf("expected %f, got %f", eulerMascheroni, result)
	}
}

// --- Variance ---

func TestGumbel_Variance(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	// Var = π² * β² / 6 ≈ 1.6449.
	expected := math.Pi * math.Pi / 6
	result := g.Variance()

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

// --- Quantile ---

func TestGumbel_Quantile_median(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 0, Beta: 1}
	// Median = mu - beta * ln(ln(2)) = -ln(ln(2)).
	expected := -math.Log(math.Log(2))
	result := g.Quantile(0.5)

	if math.Abs(result-expected) > 1e-10 {
		t.Fatalf("expected %f, got %f", expected, result)
	}
}

func TestGumbel_Quantile_roundtrip(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 2, Beta: 3}

	x := g.Quantile(0.75)
	p := g.CDF(x)

	if math.Abs(p-0.75) > 1e-9 {
		t.Fatalf("expected CDF(Quantile(0.75))=0.75, got %f", p)
	}
}

// --- String ---

func TestGumbel_String(t *testing.T) {
	t.Parallel()

	g := Gumbel{Mu: 1, Beta: 2}
	s := g.String()

	if s != "Gumbel(μ=1, β=2)" {
		t.Fatalf("expected %q, got %q", "Gumbel(μ=1, β=2)", s)
	}
}
