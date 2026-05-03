package markov

import (
	"errors"
	rand "math/rand/v2"
	"testing"
)

func newSymmetricChain() (*Chain, error) {
	return NewChain(
		[]string{"A", "B"},
		[][]float64{
			{0.5, 0.5},
			{0.5, 0.5},
		},
	)
}

func newAbsorbingChain() (*Chain, error) {
	return NewChain(
		[]string{"A", "B", "C"},
		[][]float64{
			{0.0, 1.0, 0.0},
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
		},
	)
}

func newPeriodicChain() (*Chain, error) {
	return NewChain(
		[]string{"R", "G", "Y"},
		[][]float64{
			{0, 1, 0},
			{0, 0, 1},
			{1, 0, 0},
		},
	)
}

// --- NewChain ---

func TestNewChain(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	states := c.States()

	if len(states) != 2 {
		t.Fatalf("expected 2 states, got %d", len(states))
	}
}

func TestNewChain_single_absorbing(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"X"}, [][]float64{{1.0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.States()) != 1 {
		t.Fatalf("expected 1 state, got %d", len(c.States()))
	}
}

func TestNewChain_empty(t *testing.T) {
	t.Parallel()

	_, err := NewChain(nil, nil)

	if !errors.Is(err, ErrEmptyChain) {
		t.Fatalf("expected ErrEmptyChain, got %v", err)
	}
}

func TestNewChain_duplicate_state(t *testing.T) {
	t.Parallel()

	_, err := NewChain([]string{"A", "A"}, [][]float64{{0.5, 0.5}, {0.5, 0.5}})

	if !errors.Is(err, ErrDuplicateState) {
		t.Fatalf("expected ErrDuplicateState, got %v", err)
	}
}

func TestNewChain_invalid_matrix_rows(t *testing.T) {
	t.Parallel()

	_, err := NewChain([]string{"A", "B"}, [][]float64{{1.0}})

	if !errors.Is(err, ErrInvalidMatrix) {
		t.Fatalf("expected ErrInvalidMatrix, got %v", err)
	}
}

func TestNewChain_invalid_matrix_cols(t *testing.T) {
	t.Parallel()

	_, err := NewChain([]string{"A", "B"}, [][]float64{{1.0}, {1.0}})

	if !errors.Is(err, ErrInvalidMatrix) {
		t.Fatalf("expected ErrInvalidMatrix, got %v", err)
	}
}

func TestNewChain_negative_probability(t *testing.T) {
	t.Parallel()

	_, err := NewChain([]string{"A", "B"}, [][]float64{{1.5, -0.5}, {0.5, 0.5}})

	if !errors.Is(err, ErrInvalidProbability) {
		t.Fatalf("expected ErrInvalidProbability, got %v", err)
	}
}

func TestNewChain_row_not_sum_to_one(t *testing.T) {
	t.Parallel()

	_, err := NewChain([]string{"A", "B"}, [][]float64{{0.5, 0.3}, {0.5, 0.5}})

	if !errors.Is(err, ErrInvalidRow) {
		t.Fatalf("expected ErrInvalidRow, got %v", err)
	}
}

func TestNewChain_deep_copies_matrix(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{{0.5, 0.5}, {0.5, 0.5}}

	c, err := NewChain([]string{"A", "B"}, matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	matrix[0][0] = 0.9

	p, _ := c.P("A", "A")

	if !approxEqual(p, 0.5) {
		t.Fatalf("expected P(A,A)=0.5 after mutating input, got %f", p)
	}
}

// --- P ---

func TestP(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.P("A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(p, 0.5) {
		t.Fatalf("expected 0.5, got %f", p)
	}
}

func TestP_zero(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, err := c.P("A", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(p, 0.0) {
		t.Fatalf("expected 0.0, got %f", p)
	}
}

func TestP_unknown_from(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.P("X", "A")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestP_unknown_to(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.P("A", "X")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- States ---

func TestStates_sorted(t *testing.T) {
	t.Parallel()

	c, err := NewChain([]string{"C", "A", "B"}, [][]float64{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	states := c.States()

	if states[0] != "A" || states[1] != "B" || states[2] != "C" {
		t.Fatalf("expected sorted states [A B C], got %v", states)
	}
}

// --- Graph ---

func TestGraph_returns_clone(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := c.Graph()

	if g.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", g.NodeCount())
	}

	_ = g.RemoveNode("A")
	g2 := c.Graph()

	if g2.NodeCount() != 2 {
		t.Fatalf("original graph should be unaffected, got %d nodes", g2.NodeCount())
	}
}

func TestGraph_edges_match_matrix(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := c.Graph()

	// A->B, B->C, C->C = 3 edges (only non-zero probabilities).
	if g.EdgeCount() != 3 {
		t.Fatalf("expected 3 edges, got %d", g.EdgeCount())
	}
}

// --- StepN ---

func TestStepN_zero_steps(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.StepN("A", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["A"]), 1.0) {
		t.Fatalf("expected P(A)=1.0 at step 0, got %f", dist["A"])
	}
}

func TestStepN_one_step(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.StepN("A", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["A"]), 0.5) {
		t.Fatalf("expected P(A)=0.5 at step 1, got %f", dist["A"])
	}

	if !approxEqual(float64(dist["B"]), 0.5) {
		t.Fatalf("expected P(B)=0.5 at step 1, got %f", dist["B"])
	}
}

func TestStepN_deterministic(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dist, err := c.StepN("A", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(float64(dist["C"]), 1.0) {
		t.Fatalf("expected P(C)=1.0 after 2 steps from A, got %f", dist["C"])
	}
}

func TestStepN_unknown_state(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.StepN("X", 1)

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- Simulate ---

func TestSimulate_deterministic(t *testing.T) {
	t.Parallel()

	c, err := newAbsorbingChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng := rand.New(rand.NewPCG(42, 42)) //nolint:gosec // deterministic seed for tests

	path, err := c.Simulate("A", 3, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(path) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(path))
	}

	if path[0] != "A" {
		t.Fatalf("expected path[0]=%q, got %q", "A", path[0])
	}

	if path[1] != "B" {
		t.Fatalf("expected path[1]=%q, got %q", "B", path[1])
	}

	if path[2] != "C" {
		t.Fatalf("expected path[2]=%q, got %q", "C", path[2])
	}

	if path[3] != "C" {
		t.Fatalf("expected path[3]=%q, got %q", "C", path[3])
	}
}

func TestSimulate_zero_steps(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng := rand.New(rand.NewPCG(1, 1)) //nolint:gosec // deterministic seed for tests

	path, err := c.Simulate("A", 0, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(path) != 1 {
		t.Fatalf("expected 1 element, got %d", len(path))
	}

	if path[0] != "A" {
		t.Fatalf("expected %q, got %q", "A", path[0])
	}
}

func TestSimulate_unknown_state(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng := rand.New(rand.NewPCG(1, 1)) //nolint:gosec // deterministic seed for tests

	_, err = c.Simulate("X", 1, rng)

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestSimulate_reproducible(t *testing.T) {
	t.Parallel()

	c, err := newSymmetricChain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng1 := rand.New(rand.NewPCG(99, 99)) //nolint:gosec // deterministic seed for tests
	path1, err := c.Simulate("A", 10, rng1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rng2 := rand.New(rand.NewPCG(99, 99)) //nolint:gosec // deterministic seed for tests
	path2, err := c.Simulate("A", 10, rng2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range path1 {
		if path1[i] != path2[i] {
			t.Fatalf("paths differ at index %d: %q vs %q", i, path1[i], path2[i])
		}
	}
}
