package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

const epsilon = 1e-9

func approx(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestExample_distribution(t *testing.T) {
	t.Parallel()

	dist := stats.Distribution{
		"heads": 0.5,
		"tails": 0.5,
	}

	if !stats.IsValid(dist) {
		t.Fatal("expected valid distribution")
	}
}

func TestExample_normalize(t *testing.T) {
	t.Parallel()

	unnormalized := stats.Distribution{
		"A": 2,
		"B": 3,
		"C": 5,
	}

	normalized, err := stats.Normalize(unnormalized)
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}

	if !stats.IsValid(normalized) {
		t.Fatal("expected valid distribution after normalization")
	}

	if !approx(float64(normalized["A"]), 0.2) {
		t.Fatalf("expected A=0.2, got %v", normalized["A"])
	}
}

func TestExample_complement(t *testing.T) {
	t.Parallel()

	p := stats.Prob(0.7)
	c := stats.Complement(p)

	if !approx(float64(c), 0.3) {
		t.Fatalf("expected 0.3, got %v", c)
	}
}

func TestExample_entropy(t *testing.T) {
	t.Parallel()

	// Fair coin has maximum entropy of 1 bit.
	fair := stats.Distribution{
		"heads": 0.5,
		"tails": 0.5,
	}

	h, err := stats.Entropy(fair)
	if err != nil {
		t.Fatalf("entropy error: %v", err)
	}

	if !approx(h, 1.0) {
		t.Fatalf("expected entropy 1.0, got %v", h)
	}
}

func TestExample_bayes(t *testing.T) {
	t.Parallel()

	// Disease testing: P(Disease)=0.01, P(Positive|Disease)=0.9, P(Positive)=0.059.
	prior := stats.Prob(0.01)
	likelihood := stats.Prob(0.9)
	evidence := stats.Prob(0.059)

	posterior, err := stats.Bayes(prior, likelihood, evidence)
	if err != nil {
		t.Fatalf("bayes error: %v", err)
	}

	// P(Disease|Positive) = 0.009 / 0.059.
	expected := 0.009 / 0.059

	if !approx(float64(posterior), expected) {
		t.Fatalf("expected ~%v, got %v", expected, posterior)
	}
}

func TestExample_totalProbability(t *testing.T) {
	t.Parallel()

	// P(Positive) = P(Positive|Disease)*P(Disease) + P(Positive|Healthy)*P(Healthy).
	priors := []stats.Prob{0.01, 0.99}
	likelihoods := []stats.Prob{0.9, 0.05}

	total, err := stats.TotalProbability(priors, likelihoods)
	if err != nil {
		t.Fatalf("total probability error: %v", err)
	}

	// 0.9*0.01 + 0.05*0.99 = 0.009 + 0.0495 = 0.0585.
	if !approx(float64(total), 0.0585) {
		t.Fatalf("expected 0.0585, got %v", total)
	}
}

func TestExample_chainRule(t *testing.T) {
	t.Parallel()

	// P(A,B,C) = P(A) * P(B|A) * P(C|A,B).
	joint := stats.ChainRule(0.5, 0.8, 0.3)

	if !approx(float64(joint), 0.12) {
		t.Fatalf("expected 0.12, got %v", joint)
	}
}

func TestExample_independent(t *testing.T) {
	t.Parallel()

	// P(A n B) = P(A) * P(B) for independent events.
	joint := stats.Independent(0.6, 0.4)

	if !approx(float64(joint), 0.24) {
		t.Fatalf("expected 0.24, got %v", joint)
	}
}
