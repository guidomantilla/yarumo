// Package network provides Bayesian network definition and operations.
package network

import "github.com/guidomantilla/yarumo/maths/probability"

// Node represents a variable in a Bayesian network.
type Node struct {
	Variable probability.Var
	Parents  []probability.Var
	CPT      probability.CPT
	Outcomes []probability.Outcome
}

// Network defines the interface for a Bayesian network.
type Network interface {
	// AddNode adds a variable node to the network.
	AddNode(node Node)
	// Node returns the node for the given variable.
	Node(variable probability.Var) (Node, bool)
	// Nodes returns all nodes in the network.
	Nodes() []Node
	// Parents returns the parent variables of the given variable.
	Parents(variable probability.Var) []probability.Var
	// Children returns the child variables of the given variable.
	Children(variable probability.Var) []probability.Var
	// TopologicalOrder returns variables in topological order.
	TopologicalOrder() []probability.Var
	// Validate checks that the network is a valid DAG with valid CPTs.
	Validate() error
}

var _ Network = (*network)(nil)
