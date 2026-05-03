package stats

import (
	"testing"
)

func TestVar(t *testing.T) {
	t.Parallel()

	v := Var("Rain")
	if string(v) != "Rain" {
		t.Fatalf("expected Rain, got %s", string(v))
	}
}

func TestOutcome(t *testing.T) {
	t.Parallel()

	o := Outcome("true")
	if string(o) != "true" {
		t.Fatalf("expected true, got %s", string(o))
	}
}

func TestProb(t *testing.T) {
	t.Parallel()

	p := Prob(0.75)
	if float64(p) != 0.75 {
		t.Fatalf("expected 0.75, got %f", float64(p))
	}
}

func TestDistribution(t *testing.T) {
	t.Parallel()

	d := Distribution{
		"heads": 0.5,
		"tails": 0.5,
	}

	if d["heads"] != 0.5 {
		t.Fatalf("expected 0.5 for heads, got %f", float64(d["heads"]))
	}
}

func TestAssignment(t *testing.T) {
	t.Parallel()

	a := Assignment{
		"Rain":     "true",
		"Sprinkle": "false",
	}

	if a["Rain"] != "true" {
		t.Fatalf("expected true, got %s", string(a["Rain"]))
	}
}

func TestContinuousDist_Normal(t *testing.T) {
	t.Parallel()

	var d ContinuousDist = Normal{Mu: 0, Sigma: 1}

	if d.Mean() != 0 {
		t.Fatalf("expected 0, got %f", d.Mean())
	}
}

func TestContinuousDist_Exponential(t *testing.T) {
	t.Parallel()

	var d ContinuousDist = Exponential{Lambda: 2}

	if d.Mean() != 0.5 {
		t.Fatalf("expected 0.5, got %f", d.Mean())
	}
}

func TestContinuousDist_Uniform(t *testing.T) {
	t.Parallel()

	var d ContinuousDist = Uniform{Min: 0, Max: 10}

	if d.Mean() != 5 {
		t.Fatalf("expected 5, got %f", d.Mean())
	}
}

func TestContinuousDist_Beta(t *testing.T) {
	t.Parallel()

	var d ContinuousDist = Beta{Alpha: 2, Bet: 2}

	if d.Mean() != 0.5 {
		t.Fatalf("expected 0.5, got %f", d.Mean())
	}
}

func TestContinuousDist_Gamma(t *testing.T) {
	t.Parallel()

	var d ContinuousDist = Gamma{Alpha: 3, Bet: 2}

	if d.Mean() != 1.5 {
		t.Fatalf("expected 1.5, got %f", d.Mean())
	}
}
