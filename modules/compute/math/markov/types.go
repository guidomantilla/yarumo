// Package markov provides discrete-time Markov chain primitives.
package markov

import (
	"github.com/guidomantilla/yarumo/compute/math/graph"
)

// StateClass classifies a state in a Markov chain.
type StateClass int

// State classifications.
const (
	// Transient indicates a state that the chain will eventually leave and never return to.
	Transient StateClass = iota
	// Recurrent indicates a state that the chain will always return to.
	Recurrent
	// Absorbing indicates a state with P(i,i) = 1 from which the chain never leaves.
	Absorbing
)

// Chain represents a discrete-time Markov chain backed by a directed graph.
type Chain struct {
	graph  *graph.Directed
	states []string       // states in matrix order.
	index  map[string]int // state name → matrix index.
	matrix [][]float64    // transition probability matrix.
}
