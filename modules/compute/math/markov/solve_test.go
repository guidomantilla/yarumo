package markov

import (
	"errors"
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestSolveLinearSystem_2x2(t *testing.T) {
	t.Parallel()

	// 2x + 3y = 8, x + y = 3 → x=1, y=2
	a := [][]float64{
		{2, 3},
		{1, 1},
	}
	b := []float64{8, 3}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(x[0], 1.0) {
		t.Fatalf("expected x[0]=1, got %f", x[0])
	}

	if !approxEqual(x[1], 2.0) {
		t.Fatalf("expected x[1]=2, got %f", x[1])
	}
}

func TestSolveLinearSystem_3x3(t *testing.T) {
	t.Parallel()

	// x + y + z = 6, 2y + 5z = -4, 2x + 5y - z = 27 → x=5, y=3, z=-2
	a := [][]float64{
		{1, 1, 1},
		{0, 2, 5},
		{2, 5, -1},
	}
	b := []float64{6, -4, 27}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(x[0], 5.0) {
		t.Fatalf("expected x[0]=5, got %f", x[0])
	}

	if !approxEqual(x[1], 3.0) {
		t.Fatalf("expected x[1]=3, got %f", x[1])
	}

	if !approxEqual(x[2], -2.0) {
		t.Fatalf("expected x[2]=-2, got %f", x[2])
	}
}

func TestSolveLinearSystem_identity(t *testing.T) {
	t.Parallel()

	a := [][]float64{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}
	b := []float64{7, 8, 9}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, expected := range b {
		if !approxEqual(x[i], expected) {
			t.Fatalf("expected x[%d]=%f, got %f", i, expected, x[i])
		}
	}
}

func TestSolveLinearSystem_singular(t *testing.T) {
	t.Parallel()

	a := [][]float64{
		{1, 2},
		{2, 4},
	}
	b := []float64{3, 6}

	_, err := solveLinearSystem(a, b)

	if !errors.Is(err, ErrSingularMatrix) {
		t.Fatalf("expected ErrSingularMatrix, got %v", err)
	}
}

func TestSolveLinearSystem_needs_pivoting(t *testing.T) {
	t.Parallel()

	// First pivot element is 0, needs row swap.
	a := [][]float64{
		{0, 1},
		{1, 0},
	}
	b := []float64{5, 3}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(x[0], 3.0) {
		t.Fatalf("expected x[0]=3, got %f", x[0])
	}

	if !approxEqual(x[1], 5.0) {
		t.Fatalf("expected x[1]=5, got %f", x[1])
	}
}
