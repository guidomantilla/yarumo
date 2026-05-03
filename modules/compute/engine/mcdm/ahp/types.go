// Package ahp implements the Analytic Hierarchy Process for multi-criteria decision making.
package ahp

import "github.com/guidomantilla/yarumo/compute/engine/mcdm/explain"

// PairwiseMatrix is a square matrix of pairwise comparisons.
// Element [i][j] represents the relative importance of criterion i over criterion j.
// Convention: [i][j] * [j][i] = 1 (reciprocal).
type PairwiseMatrix [][]float64

// Result contains the AHP analysis results.
type Result struct {
	Weights          []float64
	ConsistencyRatio float64
	Consistent       bool // CR < 0.10.
	Trace            explain.Trace
}
