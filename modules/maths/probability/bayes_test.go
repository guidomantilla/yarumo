package probability

import (
	"math"
	"testing"
)

func TestBayes_basic(t *testing.T) {
	t.Parallel()

	// P(Disease|Positive) = P(Positive|Disease)*P(Disease) / P(Positive)
	// = 0.99 * 0.01 / 0.0594 ≈ 0.1667
	prior := Prob(0.01)
	likelihood := Prob(0.99)
	evidence := Prob(0.0594)

	result := Bayes(prior, likelihood, evidence)

	expected := 0.99 * 0.01 / 0.0594
	if math.Abs(float64(result)-expected) > 0.001 {
		t.Fatalf("expected ~%f, got %f", expected, float64(result))
	}
}

func TestBayes_zeroEvidence(t *testing.T) {
	t.Parallel()

	result := Bayes(0.5, 0.8, 0)

	if float64(result) != 0 {
		t.Fatalf("expected 0 for zero evidence, got %f", float64(result))
	}
}

func TestTotalProbability_basic(t *testing.T) {
	t.Parallel()

	// P(Positive) = P(Positive|Disease)*P(Disease) + P(Positive|Healthy)*P(Healthy)
	// = 0.99*0.01 + 0.05*0.99 = 0.0099 + 0.0495 = 0.0594
	priors := []Prob{0.01, 0.99}
	likelihoods := []Prob{0.99, 0.05}

	result := TotalProbability(priors, likelihoods)

	expected := 0.99*0.01 + 0.05*0.99
	if math.Abs(float64(result)-expected) > epsilon {
		t.Fatalf("expected %f, got %f", expected, float64(result))
	}
}

func TestTotalProbability_empty(t *testing.T) {
	t.Parallel()

	result := TotalProbability(nil, nil)

	if float64(result) != 0 {
		t.Fatalf("expected 0 for empty, got %f", float64(result))
	}
}

func TestChainRule_basic(t *testing.T) {
	t.Parallel()

	// P(A,B,C) = P(A)*P(B|A)*P(C|A,B) = 0.5 * 0.4 * 0.3 = 0.06
	result := ChainRule(0.5, 0.4, 0.3)

	expected := 0.5 * 0.4 * 0.3
	if math.Abs(float64(result)-expected) > epsilon {
		t.Fatalf("expected %f, got %f", expected, float64(result))
	}
}

func TestChainRule_single(t *testing.T) {
	t.Parallel()

	result := ChainRule(0.7)

	if math.Abs(float64(result)-0.7) > epsilon {
		t.Fatalf("expected 0.7, got %f", float64(result))
	}
}

func TestChainRule_empty(t *testing.T) {
	t.Parallel()

	result := ChainRule()

	if float64(result) != 1.0 {
		t.Fatalf("expected 1.0 for empty chain, got %f", float64(result))
	}
}

func TestIndependent(t *testing.T) {
	t.Parallel()

	result := Independent(0.5, 0.3)

	expected := 0.5 * 0.3
	if math.Abs(float64(result)-expected) > epsilon {
		t.Fatalf("expected %f, got %f", expected, float64(result))
	}
}
