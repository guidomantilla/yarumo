package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/explain"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)

// Laptop selection: criteria are Price, Performance, Battery.
// 3 alternatives: Laptop A (cheap, slow, long battery), B (mid), C (expensive, fast, short battery).

// makeConsistentMatrix returns a 3x3 consistent pairwise comparison matrix.
// Performance strongly preferred over Price; Battery moderately preferred over Price.
func makeConsistentMatrix() ahp.PairwiseMatrix {
	return ahp.PairwiseMatrix{
		{1.0, 1.0 / 5.0, 1.0 / 3.0}, // Price vs others.
		{5.0, 1.0, 3.0},             // Performance vs others.
		{3.0, 1.0 / 3.0, 1.0},       // Battery vs others.
	}
}

// makeLaptopEvaluations returns evaluation scores (3 alternatives x 3 criteria).
// Price is raw cost (lower = cheaper); Performance and Battery are quality scores (higher = better).
func makeLaptopEvaluations() [][]float64 {
	return [][]float64{
		{2, 3, 8}, // Laptop A: cheap, slow, good battery.
		{5, 6, 5}, // Laptop B: mid-range.
		{9, 9, 3}, // Laptop C: expensive, fast, poor battery.
	}
}

// makeLaptopCriteria returns TOPSIS criteria for the laptop problem.
func makeLaptopCriteria() []topsis.Criterion {
	return []topsis.Criterion{
		{Weight: 0.1, Benefit: false}, // Price: lower is better.
		{Weight: 0.6, Benefit: true},  // Performance: higher is better.
		{Weight: 0.3, Benefit: true},  // Battery: higher is better.
	}
}

func TestAHPConsistency(t *testing.T) {
	t.Parallel()

	t.Run("consistent matrix passes check", func(t *testing.T) {
		t.Parallel()

		result, err := ahp.Analyze(makeConsistentMatrix())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Consistent {
			t.Fatalf("expected consistent, CR=%f", result.ConsistencyRatio)
		}

		if result.ConsistencyRatio >= 0.10 {
			t.Fatalf("expected CR < 0.10, got %f", result.ConsistencyRatio)
		}
	})

	t.Run("weights sum to one", func(t *testing.T) {
		t.Parallel()

		result, err := ahp.Analyze(makeConsistentMatrix())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sum := 0.0
		for _, w := range result.Weights {
			sum += w
		}

		if math.Abs(sum-1.0) > 0.01 {
			t.Fatalf("expected weights sum ≈ 1.0, got %f", sum)
		}
	})

	t.Run("performance has highest weight", func(t *testing.T) {
		t.Parallel()

		result, err := ahp.Analyze(makeConsistentMatrix())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Weights[1] <= result.Weights[0] {
			t.Fatalf("expected performance > price weight, got %f <= %f", result.Weights[1], result.Weights[0])
		}

		if result.Weights[1] <= result.Weights[2] {
			t.Fatalf("expected performance > battery weight, got %f <= %f", result.Weights[1], result.Weights[2])
		}
	})
}

func TestAHPInconsistency(t *testing.T) {
	t.Parallel()

	t.Run("inconsistent matrix detected", func(t *testing.T) {
		t.Parallel()

		inconsistent := ahp.PairwiseMatrix{
			{1, 9, 1.0 / 9.0},
			{1.0 / 9.0, 1, 9},
			{9, 1.0 / 9.0, 1},
		}

		result, err := ahp.Analyze(inconsistent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Consistent {
			t.Fatalf("expected inconsistent, CR=%f", result.ConsistencyRatio)
		}
	})
}

func TestAHPRanking(t *testing.T) {
	t.Parallel()

	t.Run("rank alternatives by weighted sum", func(t *testing.T) {
		t.Parallel()

		result, err := ahp.Analyze(makeConsistentMatrix())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		scores, err := ahp.Rank(result.Weights, makeLaptopEvaluations())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(scores) != 3 {
			t.Fatalf("expected 3 scores, got %d", len(scores))
		}

		// Performance-heavy weights → Laptop C (highest performance) should score highest.
		if scores[2] <= scores[0] {
			t.Fatalf("expected Laptop C > Laptop A, got %f <= %f", scores[2], scores[0])
		}
	})
}

func TestAHPErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty matrix", func(t *testing.T) {
		t.Parallel()

		_, err := ahp.Analyze(ahp.PairwiseMatrix{})
		if err == nil {
			t.Fatal("expected error for empty matrix")
		}
	})

	t.Run("non-square matrix", func(t *testing.T) {
		t.Parallel()

		_, err := ahp.Analyze(ahp.PairwiseMatrix{
			{1, 2},
			{0.5, 1},
			{1, 1},
		})
		if err == nil {
			t.Fatal("expected error for non-square matrix")
		}
	})

	t.Run("dimension mismatch in rank", func(t *testing.T) {
		t.Parallel()

		weights := []float64{0.5, 0.5}
		evals := [][]float64{{1, 2, 3}} // 3 columns but 2 weights.

		_, err := ahp.Rank(weights, evals)
		if err == nil {
			t.Fatal("expected error for dimension mismatch")
		}
	})
}

func TestTOPSISRanking(t *testing.T) {
	t.Parallel()

	t.Run("performance-weighted ranking favors fast laptop", func(t *testing.T) {
		t.Parallel()

		result, err := topsis.Rank(makeLaptopEvaluations(), makeLaptopCriteria())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Scores) != 3 {
			t.Fatalf("expected 3 scores, got %d", len(result.Scores))
		}

		// With 60% performance weight, Laptop C (highest performance) should rank first.
		if result.Scores[2] <= result.Scores[0] {
			t.Fatalf("expected Laptop C > Laptop A, got %f <= %f", result.Scores[2], result.Scores[0])
		}
	})

	t.Run("scores are between 0 and 1", func(t *testing.T) {
		t.Parallel()

		result, err := topsis.Rank(makeLaptopEvaluations(), makeLaptopCriteria())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, s := range result.Scores {
			if s < 0 || s > 1 {
				t.Fatalf("expected score in [0,1], alternative %d got %f", i, s)
			}
		}
	})

	t.Run("different weights produce different rankings", func(t *testing.T) {
		t.Parallel()

		evals := makeLaptopEvaluations()

		perfCriteria := []topsis.Criterion{
			{Weight: 0.1, Benefit: false},
			{Weight: 0.8, Benefit: true},
			{Weight: 0.1, Benefit: true},
		}

		priceCriteria := []topsis.Criterion{
			{Weight: 0.8, Benefit: false},
			{Weight: 0.1, Benefit: true},
			{Weight: 0.1, Benefit: true},
		}

		perfResult, err := topsis.Rank(evals, perfCriteria)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		priceResult, err := topsis.Rank(evals, priceCriteria)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Performance-weighted: C wins. Price-weighted: A wins.
		perfBest := bestAlternative(perfResult.Scores)
		priceBest := bestAlternative(priceResult.Scores)

		if perfBest == priceBest {
			t.Fatalf("expected different winners for different criteria weights, both got %d", perfBest)
		}
	})
}

func TestTOPSISErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty input", func(t *testing.T) {
		t.Parallel()

		_, err := topsis.Rank([][]float64{}, []topsis.Criterion{})
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("dimension mismatch", func(t *testing.T) {
		t.Parallel()

		matrix := [][]float64{{1, 2, 3}}
		criteria := []topsis.Criterion{{Weight: 0.5, Benefit: true}, {Weight: 0.5, Benefit: false}}

		_, err := topsis.Rank(matrix, criteria)
		if err == nil {
			t.Fatal("expected error for dimension mismatch")
		}
	})
}

func TestAHPPlusTOPSIS(t *testing.T) {
	t.Parallel()

	t.Run("AHP weights feed into TOPSIS", func(t *testing.T) {
		t.Parallel()

		ahpResult, err := ahp.Analyze(makeConsistentMatrix())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		criteria := make([]topsis.Criterion, len(ahpResult.Weights))
		criteria[0] = topsis.Criterion{Weight: ahpResult.Weights[0], Benefit: false} // Price: lower is better.
		criteria[1] = topsis.Criterion{Weight: ahpResult.Weights[1], Benefit: true}  // Performance.
		criteria[2] = topsis.Criterion{Weight: ahpResult.Weights[2], Benefit: true}  // Battery.

		topsisResult, err := topsis.Rank(makeLaptopEvaluations(), criteria)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(topsisResult.Scores) != 3 {
			t.Fatalf("expected 3 scores, got %d", len(topsisResult.Scores))
		}

		// Combined AHP+TOPSIS should still favor Laptop C (performance dominates).
		if topsisResult.Scores[2] <= topsisResult.Scores[0] {
			t.Fatalf("expected Laptop C > Laptop A in AHP+TOPSIS, got %f <= %f",
				topsisResult.Scores[2], topsisResult.Scores[0])
		}
	})
}

func TestMCDMExplainTrace(t *testing.T) {
	t.Parallel()

	t.Run("AHP trace creation", func(t *testing.T) {
		t.Parallel()

		criteria := []string{"Price", "Performance", "Battery"}
		weights := []float64{0.1, 0.6, 0.3}

		trace := explain.NewTrace("AHP", criteria, weights)
		trace = trace.AddRanking(explain.RankEntry{Alternative: 2, Score: 0.85, Rank: 1})
		trace = trace.AddRanking(explain.RankEntry{Alternative: 1, Score: 0.50, Rank: 2})
		trace = trace.AddRanking(explain.RankEntry{Alternative: 0, Score: 0.30, Rank: 3})

		if trace.Method != "AHP" {
			t.Fatalf("expected method=AHP, got %s", trace.Method)
		}

		if len(trace.Rankings) != 3 {
			t.Fatalf("expected 3 rankings, got %d", len(trace.Rankings))
		}

		if trace.Rankings[0].Rank != 1 {
			t.Fatalf("expected first entry rank=1, got %d", trace.Rankings[0].Rank)
		}
	})

	t.Run("trace string is non-empty", func(t *testing.T) {
		t.Parallel()

		trace := explain.NewTrace("TOPSIS", []string{"Cost", "Quality"}, []float64{0.4, 0.6})
		trace = trace.AddRanking(explain.RankEntry{Alternative: 0, Score: 0.75, Rank: 1})

		traceStr := trace.String()
		if traceStr == "" {
			t.Fatal("expected non-empty trace string")
		}
	})
}

func TestMCDMSentinelErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrEmptyMatrix", func(t *testing.T) {
		t.Parallel()

		if mcdm.ErrEmptyMatrix.Error() != "matrix is empty" {
			t.Fatalf("unexpected: %s", mcdm.ErrEmptyMatrix.Error())
		}
	})

	t.Run("ErrDimensionMismatch", func(t *testing.T) {
		t.Parallel()

		if mcdm.ErrDimensionMismatch.Error() != "dimensions do not match" {
			t.Fatalf("unexpected: %s", mcdm.ErrDimensionMismatch.Error())
		}
	})
}

func bestAlternative(scores []float64) int {
	best := 0

	for i := 1; i < len(scores); i++ {
		if scores[i] > scores[best] {
			best = i
		}
	}

	return best
}
