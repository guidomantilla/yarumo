package ahp

import (
	"math"
	"sort"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/explain"
)

// Analyze computes priority weights and consistency ratio from a pairwise comparison matrix.
// Returns error if the matrix is empty or not square.
func Analyze(matrix PairwiseMatrix) (Result, error) {
	n := len(matrix)
	if n == 0 {
		return Result{}, ErrAnalyze(mcdm.ErrEmptyMatrix)
	}

	err := validateSquare(matrix, n)
	if err != nil {
		return Result{}, ErrAnalyze(err)
	}

	weights := computeWeights(matrix, n)
	cr := consistencyRatio(matrix, weights, n)
	consistent := cr < 0.10 || n <= 2

	trace := explain.NewTrace("AHP", nil, weights)

	return Result{
		Weights:          weights,
		ConsistencyRatio: math.Abs(cr),
		Consistent:       consistent,
		Trace:            trace,
	}, nil
}

// Rank computes composite scores for alternatives given criteria weights and an evaluation matrix.
// evaluations[i][j] is the score of alternative i on criterion j.
// Returns a score per alternative (higher is better).
func Rank(weights []float64, evaluations [][]float64) ([]float64, error) {
	if len(weights) == 0 || len(evaluations) == 0 {
		return nil, ErrRank(mcdm.ErrEmptyMatrix)
	}

	nCriteria := len(weights)
	scores := make([]float64, len(evaluations))

	for i, alt := range evaluations {
		if len(alt) != nCriteria {
			return nil, ErrRank(mcdm.ErrDimensionMismatch)
		}

		for j, w := range weights {
			scores[i] += w * alt[j]
		}
	}

	return scores, nil
}

// AnalyzeAndRank performs AHP analysis and ranks alternatives, returning a Result with a full trace.
func AnalyzeAndRank(matrix PairwiseMatrix, evaluations [][]float64) (Result, error) {
	result, err := Analyze(matrix)
	if err != nil {
		return Result{}, err
	}

	scores, err := Rank(result.Weights, evaluations)
	if err != nil {
		return Result{}, err
	}

	trace := buildTrace("AHP", result.Weights, scores)

	result.Trace = trace

	return result, nil
}

func buildTrace(method string, weights, scores []float64) explain.Trace {
	trace := explain.NewTrace(method, nil, weights)

	type indexed struct {
		idx   int
		score float64
	}

	sorted := make([]indexed, len(scores))
	for i, s := range scores {
		sorted[i] = indexed{idx: i, score: s}
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	for rank, entry := range sorted {
		trace = trace.AddRanking(explain.RankEntry{
			Alternative: entry.idx,
			Score:       entry.score,
			Rank:        rank + 1,
		})
	}

	return trace
}

func validateSquare(matrix PairwiseMatrix, n int) error {
	for _, row := range matrix {
		if len(row) != n {
			return mcdm.ErrNotSquareMatrix
		}
	}

	return nil
}

func computeWeights(matrix PairwiseMatrix, n int) []float64 {
	v := make([]float64, n)

	for i := range n {
		v[i] = 1.0 / float64(n)
	}

	for range 100 {
		w := make([]float64, n)

		for i := range n {
			for j := range n {
				w[i] += matrix[i][j] * v[j]
			}
		}

		sum := 0.0

		for _, x := range w {
			sum += x
		}

		for i := range w {
			w[i] /= sum
		}

		maxDiff := 0.0

		for i := range w {
			diff := math.Abs(w[i] - v[i])

			if diff > maxDiff {
				maxDiff = diff
			}
		}

		v = w

		if maxDiff < 1e-10 {
			break
		}
	}

	return v
}

func consistencyRatio(matrix PairwiseMatrix, weights []float64, n int) float64 {
	if n <= 1 {
		return 0
	}

	lambdaMax := lambdaMax(matrix, weights, n)
	ci := (lambdaMax - float64(n)) / float64(n-1)
	ri := randomIndex(n)

	if ri == 0 {
		return 0
	}

	return ci / ri
}

func lambdaMax(matrix PairwiseMatrix, weights []float64, n int) float64 {
	sum := 0.0

	for i := range n {
		row := matrix[i]
		aw := 0.0

		for j := range n {
			aw += row[j] * weights[j]
		}

		sum += aw / weights[i]
	}

	return sum / float64(n)
}

// randomIndex returns Saaty's Random Consistency Index for matrix size n.
func randomIndex(n int) float64 {
	ri := []float64{0, 0, 0, 0.58, 0.90, 1.12, 1.24, 1.32, 1.41, 1.45, 1.49}
	if n < len(ri) {
		return ri[n]
	}

	return 1.49 // approximate for n > 10.
}
