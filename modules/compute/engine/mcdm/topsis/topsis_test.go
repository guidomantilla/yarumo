package topsis

import (
	"errors"
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm"
)

func TestRank_basic(t *testing.T) {
	t.Parallel()

	// 3 alternatives, 3 criteria (2 benefit, 1 cost).
	matrix := [][]float64{
		{250, 16, 12},
		{200, 16, 8},
		{300, 32, 16},
	}

	criteria := []Criterion{
		{Weight: 0.4, Benefit: true},
		{Weight: 0.3, Benefit: true},
		{Weight: 0.3, Benefit: false},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(result.Scores))
	}

	// All scores should be in [0, 1].
	for i, s := range result.Scores {
		if s < 0 || s > 1 {
			t.Fatalf("score[%d] = %f, expected in [0,1]", i, s)
		}
	}
}

func TestRank_identical(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{
		{10, 20, 30},
		{10, 20, 30},
		{10, 20, 30},
	}

	criteria := []Criterion{
		{Weight: 0.5, Benefit: true},
		{Weight: 0.3, Benefit: true},
		{Weight: 0.2, Benefit: false},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All identical alternatives should have equal scores.
	for i := 1; i < len(result.Scores); i++ {
		if math.Abs(result.Scores[i]-result.Scores[0]) > 0.001 {
			t.Fatalf("scores should be equal: %v", result.Scores)
		}
	}
}

func TestRank_empty(t *testing.T) {
	t.Parallel()

	t.Run("empty matrix", func(t *testing.T) {
		t.Parallel()

		_, err := Rank([][]float64{}, []Criterion{{Weight: 1, Benefit: true}})
		if !errors.Is(err, mcdm.ErrEmptyInput) {
			t.Fatalf("expected mcdm.ErrEmptyInput, got %v", err)
		}
	})

	t.Run("empty criteria", func(t *testing.T) {
		t.Parallel()

		_, err := Rank([][]float64{{1, 2}}, []Criterion{})
		if !errors.Is(err, mcdm.ErrEmptyInput) {
			t.Fatalf("expected mcdm.ErrEmptyInput, got %v", err)
		}
	})
}

func TestRank_dimensionMismatch(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{
		{1, 2, 3},
		{4, 5}, // only 2 columns.
	}

	criteria := []Criterion{
		{Weight: 0.3, Benefit: true},
		{Weight: 0.3, Benefit: true},
		{Weight: 0.4, Benefit: false},
	}

	_, err := Rank(matrix, criteria)
	if !errors.Is(err, mcdm.ErrDimensionMismatch) {
		t.Fatalf("expected mcdm.ErrDimensionMismatch, got %v", err)
	}
}

func TestRank_singleAlternative(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{
		{10, 20},
	}

	criteria := []Criterion{
		{Weight: 0.6, Benefit: true},
		{Weight: 0.4, Benefit: false},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(result.Scores))
	}

	// Single alternative: dist to positive and negative are both 0 -> score = 0.
	if result.Scores[0] != 0 {
		t.Fatalf("expected score 0 for single alternative, got %f", result.Scores[0])
	}
}

func TestRank_allBenefit(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{
		{10, 20},
		{30, 40},
		{20, 10},
	}

	criteria := []Criterion{
		{Weight: 0.5, Benefit: true},
		{Weight: 0.5, Benefit: true},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Alternative 1 (30, 40) should have the highest score.
	best := 0
	for i := 1; i < len(result.Scores); i++ {
		if result.Scores[i] > result.Scores[best] {
			best = i
		}
	}

	if best != 1 {
		t.Fatalf("expected alternative 1 to be best, got %d (scores: %v)", best, result.Scores)
	}
}

func TestRank_allCost(t *testing.T) {
	t.Parallel()

	matrix := [][]float64{
		{10, 20},
		{30, 40},
		{5, 8},
	}

	criteria := []Criterion{
		{Weight: 0.5, Benefit: false},
		{Weight: 0.5, Benefit: false},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Alternative 2 (5, 8) should have the highest score (lowest cost wins).
	best := 0
	for i := 1; i < len(result.Scores); i++ {
		if result.Scores[i] > result.Scores[best] {
			best = i
		}
	}

	if best != 2 {
		t.Fatalf("expected alternative 2 to be best, got %d (scores: %v)", best, result.Scores)
	}
}

func TestRank_zeroColumn(t *testing.T) {
	t.Parallel()

	// All zeros in one column should not panic.
	matrix := [][]float64{
		{0, 20},
		{0, 40},
	}

	criteria := []Criterion{
		{Weight: 0.5, Benefit: true},
		{Weight: 0.5, Benefit: true},
	}

	result, err := Rank(matrix, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(result.Scores))
	}
}
