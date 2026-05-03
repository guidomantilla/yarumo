package ahp

import (
	"errors"
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm"
)

func TestAnalyze_basic(t *testing.T) {
	t.Parallel()

	// Classic 3x3 Saaty example.
	matrix := PairwiseMatrix{
		{1, 3, 5},
		{1.0 / 3, 1, 3},
		{1.0 / 5, 1.0 / 3, 1},
	}

	result, err := Analyze(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Weights should sum to approximately 1.0.
	sum := 0.0
	for _, w := range result.Weights {
		sum += w
	}

	if math.Abs(sum-1.0) > 0.01 {
		t.Fatalf("weights sum to %f, expected ~1.0", sum)
	}

	// CR should be < 0.10 for this consistent matrix.
	if !result.Consistent {
		t.Fatalf("expected consistent matrix, CR=%f", result.ConsistencyRatio)
	}
}

func TestAnalyze_identity(t *testing.T) {
	t.Parallel()

	matrix := PairwiseMatrix{
		{1, 1, 1},
		{1, 1, 1},
		{1, 1, 1},
	}

	result, err := Analyze(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Equal weights: each should be ~1/3.
	for i, w := range result.Weights {
		if math.Abs(w-1.0/3) > 0.001 {
			t.Fatalf("weight[%d] = %f, expected ~0.333", i, w)
		}
	}

	// CR should be 0 for perfectly consistent identity.
	if result.ConsistencyRatio > 0.001 {
		t.Fatalf("expected CR~0, got %f", result.ConsistencyRatio)
	}
}

func TestAnalyze_empty(t *testing.T) {
	t.Parallel()

	_, err := Analyze(PairwiseMatrix{})
	if !errors.Is(err, mcdm.ErrEmptyMatrix) {
		t.Fatalf("expected mcdm.ErrEmptyMatrix, got %v", err)
	}
}

func TestAnalyze_notSquare(t *testing.T) {
	t.Parallel()

	matrix := PairwiseMatrix{
		{1, 2},
		{0.5, 1, 3},
	}

	_, err := Analyze(matrix)
	if !errors.Is(err, mcdm.ErrNotSquareMatrix) {
		t.Fatalf("expected mcdm.ErrNotSquareMatrix, got %v", err)
	}
}

func TestAnalyze_consistent(t *testing.T) {
	t.Parallel()

	// 2x2 matrix is always consistent.
	matrix := PairwiseMatrix{
		{1, 5},
		{1.0 / 5, 1},
	}

	result, err := Analyze(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Consistent {
		t.Fatalf("expected consistent for 2x2 matrix, CR=%f", result.ConsistencyRatio)
	}
}

func TestAnalyze_singleElement(t *testing.T) {
	t.Parallel()

	matrix := PairwiseMatrix{
		{1},
	}

	result, err := Analyze(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Weights) != 1 {
		t.Fatalf("expected 1 weight, got %d", len(result.Weights))
	}

	if math.Abs(result.Weights[0]-1.0) > 0.001 {
		t.Fatalf("expected weight 1.0, got %f", result.Weights[0])
	}
}

func TestRank_basic(t *testing.T) {
	t.Parallel()

	weights := []float64{0.6, 0.3, 0.1}
	evaluations := [][]float64{
		{0.8, 0.5, 0.9},
		{0.6, 0.9, 0.4},
		{0.9, 0.3, 0.7},
	}

	scores, err := Rank(weights, evaluations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}

	// Alt 0: 0.6*0.8 + 0.3*0.5 + 0.1*0.9 = 0.48+0.15+0.09 = 0.72.
	// Alt 1: 0.6*0.6 + 0.3*0.9 + 0.1*0.4 = 0.36+0.27+0.04 = 0.67.
	// Alt 2: 0.6*0.9 + 0.3*0.3 + 0.1*0.7 = 0.54+0.09+0.07 = 0.70.
	expected := []float64{0.72, 0.67, 0.70}
	for i, s := range scores {
		if math.Abs(s-expected[i]) > 0.001 {
			t.Fatalf("score[%d] = %f, expected %f", i, s, expected[i])
		}
	}
}

func TestRank_emptyWeights(t *testing.T) {
	t.Parallel()

	_, err := Rank([]float64{}, [][]float64{{1, 2}})
	if !errors.Is(err, mcdm.ErrEmptyMatrix) {
		t.Fatalf("expected mcdm.ErrEmptyMatrix, got %v", err)
	}
}

func TestRank_emptyEvaluations(t *testing.T) {
	t.Parallel()

	_, err := Rank([]float64{0.5, 0.5}, [][]float64{})
	if !errors.Is(err, mcdm.ErrEmptyMatrix) {
		t.Fatalf("expected mcdm.ErrEmptyMatrix, got %v", err)
	}
}

func TestRank_dimensionMismatch(t *testing.T) {
	t.Parallel()

	weights := []float64{0.5, 0.5}
	evaluations := [][]float64{
		{0.8, 0.5, 0.9}, // 3 columns, but 2 weights.
	}

	_, err := Rank(weights, evaluations)
	if !errors.Is(err, mcdm.ErrDimensionMismatch) {
		t.Fatalf("expected mcdm.ErrDimensionMismatch, got %v", err)
	}
}

func Test_randomIndex(t *testing.T) {
	t.Parallel()

	t.Run("known values", func(t *testing.T) {
		t.Parallel()

		// n=3 -> 0.58.
		ri := randomIndex(3)
		if math.Abs(ri-0.58) > 0.001 {
			t.Fatalf("expected 0.58, got %f", ri)
		}
	})

	t.Run("n=1", func(t *testing.T) {
		t.Parallel()

		ri := randomIndex(1)
		if ri != 0 {
			t.Fatalf("expected 0, got %f", ri)
		}
	})

	t.Run("large n", func(t *testing.T) {
		t.Parallel()

		ri := randomIndex(20)
		if math.Abs(ri-1.49) > 0.001 {
			t.Fatalf("expected 1.49, got %f", ri)
		}
	})
}
