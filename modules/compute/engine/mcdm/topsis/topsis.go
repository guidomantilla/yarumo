package topsis

import (
	"math"
	"sort"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/explain"
)

// Rank evaluates alternatives using the TOPSIS method.
// matrix[i][j] is the value of alternative i on criterion j.
// criteria must have the same length as columns in the matrix.
func Rank(matrix [][]float64, criteria []Criterion) (Result, error) {
	if len(matrix) == 0 || len(criteria) == 0 {
		return Result{}, ErrRank(mcdm.ErrEmptyInput)
	}

	nCrit := len(criteria)

	err := validateDimensions(matrix, nCrit)
	if err != nil {
		return Result{}, ErrRank(err)
	}

	weights := make([]float64, nCrit)
	for i, c := range criteria {
		weights[i] = c.Weight
	}

	weighted := weightedNormalized(matrix, criteria, nCrit)
	idealPos, idealNeg := idealSolutions(weighted, criteria, nCrit)
	scores := closeness(weighted, idealPos, idealNeg, nCrit)
	trace := buildTrace(weights, scores)

	return Result{Scores: scores, Trace: trace}, nil
}

func validateDimensions(matrix [][]float64, nCrit int) error {
	for _, row := range matrix {
		if len(row) != nCrit {
			return mcdm.ErrDimensionMismatch
		}
	}

	return nil
}

func weightedNormalized(matrix [][]float64, criteria []Criterion, nCrit int) [][]float64 {
	nAlts := len(matrix)
	norms := columnNorms(matrix, nAlts, nCrit)

	weighted := make([][]float64, nAlts)

	for i := range nAlts {
		weighted[i] = make([]float64, nCrit)
		row := matrix[i]

		for j := range nCrit {
			normalized := 0.0

			if norms[j] > 0 {
				normalized = row[j] / norms[j]
			}

			weighted[i][j] = normalized * criteria[j].Weight
		}
	}

	return weighted
}

func columnNorms(matrix [][]float64, nAlts, nCrit int) []float64 {
	norms := make([]float64, nCrit)

	for j := range nCrit {
		sumSq := 0.0

		for i := range nAlts {
			val := matrix[i][j]
			sumSq += val * val
		}

		norms[j] = math.Sqrt(sumSq)
	}

	return norms
}

func idealSolutions(weighted [][]float64, criteria []Criterion, nCrit int) ([]float64, []float64) {
	idealPos := make([]float64, nCrit)
	idealNeg := make([]float64, nCrit)

	for j := range nCrit {
		maxVal, minVal := columnExtremes(weighted, j)

		if criteria[j].Benefit {
			idealPos[j] = maxVal
			idealNeg[j] = minVal
		} else {
			idealPos[j] = minVal
			idealNeg[j] = maxVal
		}
	}

	return idealPos, idealNeg
}

func columnExtremes(weighted [][]float64, col int) (float64, float64) {
	maxVal := weighted[0][col]
	minVal := weighted[0][col]

	for i := 1; i < len(weighted); i++ {
		val := weighted[i][col]

		if val > maxVal {
			maxVal = val
		}

		if val < minVal {
			minVal = val
		}
	}

	return maxVal, minVal
}

func buildTrace(weights, scores []float64) explain.Trace {
	trace := explain.NewTrace("TOPSIS", nil, weights)

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

func closeness(weighted [][]float64, idealPos, idealNeg []float64, nCrit int) []float64 {
	scores := make([]float64, len(weighted))

	for i, row := range weighted {
		distPos := 0.0
		distNeg := 0.0

		for j := range nCrit {
			dp := row[j] - idealPos[j]
			dn := row[j] - idealNeg[j]
			distPos += dp * dp
			distNeg += dn * dn
		}

		distPos = math.Sqrt(distPos)
		distNeg = math.Sqrt(distNeg)

		denom := distPos + distNeg
		if denom == 0 {
			scores[i] = 0
		} else {
			scores[i] = distNeg / denom
		}
	}

	return scores
}
