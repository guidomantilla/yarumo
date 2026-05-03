// Package network provides Bayesian network definition and operations.
package network

import (
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
)

// Node represents a variable in a Bayesian network.
type Node struct {
	Variable stats.Var
	Parents  []stats.Var
	CPT      bayesian.CPT
	Outcomes []stats.Outcome
}

// Network defines the interface for a Bayesian network.
type Network interface {
	// AddNode adds a variable node to the network. Returns an error if the variable already exists.
	AddNode(node Node) error
	// Node returns the node for the given variable.
	Node(variable stats.Var) (Node, bool)
	// Nodes returns all nodes in the network.
	Nodes() []Node
	// Parents returns the parent variables of the given variable.
	Parents(variable stats.Var) []stats.Var
	// Children returns the child variables of the given variable.
	Children(variable stats.Var) []stats.Var
	// TopologicalOrder returns variables in topological order.
	TopologicalOrder() []stats.Var
	// Validate checks that the network is a valid DAG with valid CPTs.
	Validate() error
}

var _ Network = (*network)(nil)
