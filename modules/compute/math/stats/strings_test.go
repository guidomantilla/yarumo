package stats

import (
	"strings"
	"testing"
)

func TestDistribution_String_empty(t *testing.T) {
	t.Parallel()

	d := Distribution{}

	if d.String() != "{}" {
		t.Fatalf("expected {}, got %q", d.String())
	}
}

func TestDistribution_String_single(t *testing.T) {
	t.Parallel()

	d := Distribution{"heads": 1.0}
	result := d.String()

	if result != "{heads=1}" {
		t.Fatalf("expected {heads=1}, got %q", result)
	}
}

func TestDistribution_String_multiple(t *testing.T) {
	t.Parallel()

	d := Distribution{"b": 0.3, "a": 0.7}
	result := d.String()

	// Should be sorted by key.
	if !strings.Contains(result, "a=") || !strings.Contains(result, "b=") {
		t.Fatalf("expected both outcomes, got %q", result)
	}

	if !strings.HasPrefix(result, "{a=") {
		t.Fatalf("expected sorted output starting with a, got %q", result)
	}
}

func TestNormal_String(t *testing.T) {
	t.Parallel()

	n := Normal{Mu: 0, Sigma: 1}

	if n.String() != "Normal(μ=0, σ=1)" {
		t.Fatalf("unexpected: %s", n.String())
	}
}

func TestExponential_String(t *testing.T) {
	t.Parallel()

	e := Exponential{Lambda: 2}

	if e.String() != "Exponential(λ=2)" {
		t.Fatalf("unexpected: %s", e.String())
	}
}

func TestUniform_String(t *testing.T) {
	t.Parallel()

	u := Uniform{Min: 0, Max: 1}

	if u.String() != "Uniform(0, 1)" {
		t.Fatalf("unexpected: %s", u.String())
	}
}

func TestBeta_String(t *testing.T) {
	t.Parallel()

	b := Beta{Alpha: 2, Bet: 5}

	if b.String() != "Beta(α=2, β=5)" {
		t.Fatalf("unexpected: %s", b.String())
	}
}

func TestGamma_String(t *testing.T) {
	t.Parallel()

	g := Gamma{Alpha: 2, Bet: 1}

	if g.String() != "Gamma(α=2, β=1)" {
		t.Fatalf("unexpected: %s", g.String())
	}
}
