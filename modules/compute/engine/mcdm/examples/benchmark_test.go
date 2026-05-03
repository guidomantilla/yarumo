package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)

func buildPairwiseMatrix(n int) ahp.PairwiseMatrix {
	matrix := make(ahp.PairwiseMatrix, n)

	for i := range n {
		matrix[i] = make([]float64, n)

		for j := range n {
			switch {
			case i == j:
				matrix[i][j] = 1
			case i < j:
				matrix[i][j] = float64(j-i+1) / float64(n)
			default:
				matrix[i][j] = float64(n) / float64(i-j+1)
			}
		}
	}

	return matrix
}

func buildEvalMatrix(nAlts, nCrit int) [][]float64 {
	evals := make([][]float64, nAlts)

	for i := range nAlts {
		evals[i] = make([]float64, nCrit)

		for j := range nCrit {
			evals[i][j] = float64((i+1)*(j+1)) / float64(nAlts)
		}
	}

	return evals
}

func buildCriteria(n int) []topsis.Criterion {
	criteria := make([]topsis.Criterion, n)
	w := 1.0 / float64(n)

	for i := range n {
		criteria[i] = topsis.Criterion{Weight: w, Benefit: i%2 == 0}
	}

	return criteria
}

func BenchmarkAHPAnalyze5(b *testing.B) {
	matrix := buildPairwiseMatrix(5)

	b.ResetTimer()

	for b.Loop() {
		ahp.Analyze(matrix)
	}
}

func BenchmarkAHPAnalyze10(b *testing.B) {
	matrix := buildPairwiseMatrix(10)

	b.ResetTimer()

	for b.Loop() {
		ahp.Analyze(matrix)
	}
}

func BenchmarkAHPRank(b *testing.B) {
	result, _ := ahp.Analyze(buildPairwiseMatrix(5))
	evals := buildEvalMatrix(10, 5)

	b.ResetTimer()

	for b.Loop() {
		ahp.Rank(result.Weights, evals)
	}
}

func BenchmarkTOPSIS10x5(b *testing.B) {
	matrix := buildEvalMatrix(10, 5)
	criteria := buildCriteria(5)

	b.ResetTimer()

	for b.Loop() {
		topsis.Rank(matrix, criteria)
	}
}

func BenchmarkTOPSIS50x10(b *testing.B) {
	matrix := buildEvalMatrix(50, 10)
	criteria := buildCriteria(10)

	b.ResetTimer()

	for b.Loop() {
		topsis.Rank(matrix, criteria)
	}
}
