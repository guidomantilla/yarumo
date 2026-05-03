// Package topsis implements the TOPSIS method for multi-criteria decision making.
package topsis

import "github.com/guidomantilla/yarumo/compute/engine/mcdm/explain"

// Criterion defines a decision criterion with weight and optimization direction.
type Criterion struct {
	Weight  float64
	Benefit bool // true = maximize (higher is better), false = minimize (lower is better).
}

// Result contains the TOPSIS analysis results.
type Result struct {
	Scores []float64 // Relative closeness to ideal solution, one per alternative.
	Trace  explain.Trace
}
