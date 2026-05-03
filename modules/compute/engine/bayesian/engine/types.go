// Package engine provides Bayesian network inference algorithms.
package engine

import (
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/explain"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

// Algorithm identifies the inference algorithm.
type Algorithm int

const (
	// Enumeration is exact inference by full enumeration.
	Enumeration Algorithm = iota
	// VariableElimination is exact inference using factor operations.
	VariableElimination
)

// Result holds the outcome of a Bayesian inference query.
type Result struct {
	Posterior stats.Distribution
	Trace     explain.Trace
}

// Engine defines the interface for a Bayesian inference engine.
type Engine interface {
	// Query computes P(query | evidence) in the given network.
	Query(net network.Network, ev evidence.EvidenceBase, query stats.Var) Result
}

var _ Engine = (*engine)(nil)
