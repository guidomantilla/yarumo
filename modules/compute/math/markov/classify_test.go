package markov

import (
	"errors"
	"testing"
)

// --- Classify ---

func TestClassify_all_recurrent(t *testing.T) {
	t.Parallel()

	c, err := newPeriodicChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	classes := c.Classify()

	for _, s := range []string{"R", "G", "Y"} {
		if classes[s] != Recurrent {
			t.Fatalf("expected %q to be Recurrent, got %v", s, classes[s])
		}
	}
}

func TestClassify_with_absorbing(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	classes := c.Classify()

	if classes["A"] != Transient {
		t.Fatalf("expected A=Transient, got %v", classes["A"])
	}

	if classes["B"] != Transient {
		t.Fatalf("expected B=Transient, got %v", classes["B"])
	}

	if classes["C"] != Absorbing {
		t.Fatalf("expected C=Absorbing, got %v", classes["C"])
	}
}

func TestClassify_mixed(t *testing.T) {
	t.Parallel()

	// A -> B <-> C (B and C recurrent, A transient).
	c, err := NewChain(
		[]string{"A", "B", "C"},
		[][]float64{
			{0.0, 1.0, 0.0},
			{0.0, 0.5, 0.5},
			{0.0, 0.5, 0.5},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	classes := c.Classify()

	if classes["A"] != Transient {
		t.Fatalf("expected A=Transient, got %v", classes["A"])
	}

	if classes["B"] != Recurrent {
		t.Fatalf("expected B=Recurrent, got %v", classes["B"])
	}

	if classes["C"] != Recurrent {
		t.Fatalf("expected C=Recurrent, got %v", classes["C"])
	}
}

func TestClassify_single_absorbing(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"X"}, [][]float64{{1.0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	classes := c.Classify()

	if classes["X"] != Absorbing {
		t.Fatalf("expected X=Absorbing, got %v", classes["X"])
	}
}

// --- IsAbsorbing ---

func TestIsAbsorbing_true(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := c.IsAbsorbing("C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Fatal("expected C to be absorbing")
	}
}

func TestIsAbsorbing_false(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := c.IsAbsorbing("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ok {
		t.Fatal("expected A to not be absorbing")
	}
}

func TestIsAbsorbing_unknown(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.IsAbsorbing("X")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- IsIrreducible ---

func TestIsIrreducible_true(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.IsIrreducible() {
		t.Fatal("expected symmetric chain to be irreducible")
	}
}

func TestIsIrreducible_false(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.IsIrreducible() {
		t.Fatal("expected absorbing chain to not be irreducible")
	}
}

func TestIsIrreducible_single_state(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"X"}, [][]float64{{1.0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.IsIrreducible() {
		t.Fatal("expected single state chain to be irreducible")
	}
}

// --- Period ---

func TestPeriod_cyclic(t *testing.T) {
	t.Parallel()

	c, err := newPeriodicChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("R")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 3 {
		t.Fatalf("expected period 3, got %d", p)
	}
}

func TestPeriod_aperiodic(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 1 {
		t.Fatalf("expected period 1, got %d", p)
	}
}

func TestPeriod_self_loop(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"X"}, [][]float64{{1.0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 1 {
		t.Fatalf("expected period 1 for self-loop, got %d", p)
	}
}

func TestPeriod_no_cycle(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 0 {
		t.Fatalf("expected period 0 for transient state with no return, got %d", p)
	}
}

func TestPeriod_unknown(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Period("X")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestPeriod_two(t *testing.T) {
	t.Parallel()

	// A <-> B with no self-loops → period 2.
	c, err := NewChain(
		[]string{"A", "B"},
		[][]float64{
			{0, 1},
			{1, 0},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 2 {
		t.Fatalf("expected period 2, got %d", p)
	}
}

func TestPeriod_mixed_cycle_lengths(t *testing.T) {
	t.Parallel()

	// A->B->A (cycle length 2) and A->C->B->A (cycle length 3).
	// Period = gcd(2, 3) = 1 (aperiodic).
	c, err := NewChain(
		[]string{"A", "B", "C"},
		[][]float64{
			{0.0, 0.5, 0.5},
			{1.0, 0.0, 0.0},
			{0.0, 1.0, 0.0},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.Period("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p != 1 {
		t.Fatalf("expected period 1 (gcd(2,3)), got %d", p)
	}
}

// --- IsErgodic ---

func TestIsErgodic_true(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.IsErgodic() {
		t.Fatal("expected symmetric chain to be ergodic")
	}
}

func TestIsErgodic_false_not_irreducible(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.IsErgodic() {
		t.Fatal("expected absorbing chain to not be ergodic")
	}
}

func TestIsErgodic_false_periodic(t *testing.T) {
	t.Parallel()

	c, err := newPeriodicChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.IsErgodic() {
		t.Fatal("expected periodic chain to not be ergodic")
	}
}
