package markov

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// --- Steady ---

func TestSteady_symmetric(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.Steady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["A"]), 0.5) {
		t.Fatalf("expected π(A)=0.5, got %f", dist["A"])
	}

	if !approxEqual(float64(dist["B"]), 0.5) {
		t.Fatalf("expected π(B)=0.5, got %f", dist["B"])
	}
}

func TestSteady_asymmetric(t *testing.T) {
	t.Parallel()

	// P = [[0.7, 0.3], [0.4, 0.6]] → π = (4/7, 3/7).
	c, err := NewChain(
		[]string{"A", "B"},
		[][]float64{
			{0.7, 0.3},
			{0.4, 0.6},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.Steady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["A"]), 4.0/7.0) {
		t.Fatalf("expected π(A)=4/7, got %f", dist["A"])
	}

	if !approxEqual(float64(dist["B"]), 3.0/7.0) {
		t.Fatalf("expected π(B)=3/7, got %f", dist["B"])
	}
}

func TestSteady_not_irreducible(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Steady()

	if !errors.Is(err, ErrNotIrreducible) {
		t.Fatalf("expected ErrNotIrreducible, got %v", err)
	}
}

func TestSteady_single_state(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"X"}, [][]float64{{1.0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.Steady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["X"]), 1.0) {
		t.Fatalf("expected π(X)=1.0, got %f", dist["X"])
	}
}

func TestSteady_periodic(t *testing.T) {
	t.Parallel()

	// Periodic but irreducible — steady state still exists.
	c, err := newPeriodicChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.Steady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Uniform distribution for symmetric periodic chain.
	for _, s := range []string{"R", "G", "Y"} {
		if !approxEqual(float64(dist[stats.Outcome(s)]), 1.0/3.0) {
			t.Fatalf("expected π(%s)=1/3, got %f", s, dist[stats.Outcome(s)])
		}
	}
}

// --- MeanFirstPassage ---

func TestMeanFirstPassage_deterministic(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfp, err := c.MeanFirstPassage("A", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mfp, 2.0) {
		t.Fatalf("expected MFP(A→C)=2, got %f", mfp)
	}
}

func TestMeanFirstPassage_one_step(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfp, err := c.MeanFirstPassage("B", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mfp, 1.0) {
		t.Fatalf("expected MFP(B→C)=1, got %f", mfp)
	}
}

func TestMeanFirstPassage_same_state(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfp, err := c.MeanFirstPassage("A", "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mfp, 0.0) {
		t.Fatalf("expected MFP(A→A)=0, got %f", mfp)
	}
}

func TestMeanFirstPassage_stochastic(t *testing.T) {
	t.Parallel()

	// P = [[0, 1], [0.5, 0.5]] → MFP(B→A) = 2.
	c, err := NewChain(
		[]string{"A", "B"},
		[][]float64{
			{0, 1},
			{0.5, 0.5},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mfp, err := c.MeanFirstPassage("B", "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mfp, 2.0) {
		t.Fatalf("expected MFP(B→A)=2, got %f", mfp)
	}
}

func TestMeanFirstPassage_unknown_from(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.MeanFirstPassage("X", "A")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestMeanFirstPassage_unknown_to(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.MeanFirstPassage("A", "X")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- AbsorptionProbabilities ---

func TestAbsorptionProbabilities(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs, err := c.AbsorptionProbabilities()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(probs["A"]["C"], 1.0) {
		t.Fatalf("expected P(A→C)=1.0, got %f", probs["A"]["C"])
	}

	if !approxEqual(probs["B"]["C"], 1.0) {
		t.Fatalf("expected P(B→C)=1.0, got %f", probs["B"]["C"])
	}
}

func TestAbsorptionProbabilities_multiple_absorbing(t *testing.T) {
	t.Parallel()

	// A with 50% chance to B or C, both absorbing.
	c, err := NewChain(
		[]string{"A", "B", "C"},
		[][]float64{
			{0.0, 0.5, 0.5},
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs, err := c.AbsorptionProbabilities()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(probs["A"]["B"], 0.5) {
		t.Fatalf("expected P(A→B)=0.5, got %f", probs["A"]["B"])
	}

	if !approxEqual(probs["A"]["C"], 0.5) {
		t.Fatalf("expected P(A→C)=0.5, got %f", probs["A"]["C"])
	}
}

func TestAbsorptionProbabilities_no_absorbing(t *testing.T) {
	t.Parallel()

	c, err := newPeriodicChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.AbsorptionProbabilities()

	if !errors.Is(err, ErrNoAbsorbingStates) {
		t.Fatalf("expected ErrNoAbsorbingStates, got %v", err)
	}
}

func TestAbsorptionProbabilities_no_transient(t *testing.T) {
	t.Parallel()

	// All states absorbing.
	c, err := NewChain(
		[]string{"A", "B"},
		[][]float64{
			{1.0, 0.0},
			{0.0, 1.0},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs, err := c.AbsorptionProbabilities()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(probs) != 0 {
		t.Fatalf("expected empty map, got %v", probs)
	}
}

// --- MeanAbsorptionTime ---

func TestMeanAbsorptionTime(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mat, err := c.MeanAbsorptionTime("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mat, 2.0) {
		t.Fatalf("expected MAT(A)=2, got %f", mat)
	}
}

func TestMeanAbsorptionTime_one_step(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mat, err := c.MeanAbsorptionTime("B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mat, 1.0) {
		t.Fatalf("expected MAT(B)=1, got %f", mat)
	}
}

func TestMeanAbsorptionTime_unknown(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.MeanAbsorptionTime("X")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestMeanAbsorptionTime_not_transient(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.MeanAbsorptionTime("C")

	if !errors.Is(err, ErrNotTransient) {
		t.Fatalf("expected ErrNotTransient, got %v", err)
	}
}

func TestMeanAbsorptionTime_stochastic(t *testing.T) {
	t.Parallel()

	// A → B (p=0.5) or stays (p=0.5), B absorbing → MAT(A) = 2.
	c, err := NewChain(
		[]string{"A", "B"},
		[][]float64{
			{0.5, 0.5},
			{0.0, 1.0},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mat, err := c.MeanAbsorptionTime("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(mat, 2.0) {
		t.Fatalf("expected MAT(A)=2, got %f", mat)
	}
}
