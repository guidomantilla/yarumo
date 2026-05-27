package network

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/compute/math/graph"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
)

type network struct {
	g     *graph.Directed
	order []stats.Var
}

// NewNetwork creates an empty Bayesian network.
func NewNetwork() Network {
	return &network{
		g: graph.NewDirected(),
	}
}

func (n *network) AddNode(node Node) error {
	cassert.NotNil(n, "network is nil")

	err := n.g.AddNode(graph.Node{ID: string(node.Variable), Metadata: node})
	if err != nil {
		return bayesian.ErrValidation(fmt.Errorf("duplicate node: %s", string(node.Variable)))
	}

	n.order = append(n.order, node.Variable)

	for _, p := range node.Parents {
		if n.g.HasNode(string(p)) {
			_ = n.g.AddEdge(graph.Edge{
				ID:   string(p) + "->" + string(node.Variable),
				From: string(p),
				To:   string(node.Variable),
			})
		}
	}

	return nil
}

func (n *network) Node(variable stats.Var) (Node, bool) {
	cassert.NotNil(n, "network is nil")

	gn, err := n.g.Node(string(variable))
	if err != nil {
		return Node{}, false
	}

	node, ok := gn.Metadata.(Node)
	cassert.True(ok, "metadata is not a Node")

	return node, true
}

func (n *network) Nodes() []Node {
	cassert.NotNil(n, "network is nil")

	result := make([]Node, 0, len(n.order))

	for _, v := range n.order {
		gn, err := n.g.Node(string(v))
		if err != nil {
			continue
		}

		node, ok := gn.Metadata.(Node)
		cassert.True(ok, "metadata is not a Node")

		result = append(result, node)
	}

	return result
}

func (n *network) Parents(variable stats.Var) []stats.Var {
	cassert.NotNil(n, "network is nil")

	gn, err := n.g.Node(string(variable))
	if err != nil {
		return nil
	}

	node, ok := gn.Metadata.(Node)
	cassert.True(ok, "metadata is not a Node")

	return node.Parents
}

func (n *network) Children(variable stats.Var) []stats.Var {
	cassert.NotNil(n, "network is nil")

	neighbors, err := n.g.Neighbors(string(variable))
	if err != nil {
		return nil
	}

	children := make([]stats.Var, 0, len(neighbors))

	for _, id := range neighbors {
		children = append(children, stats.Var(id))
	}

	return children
}

func (n *network) TopologicalOrder() []stats.Var {
	cassert.NotNil(n, "network is nil")

	sorted, err := graph.TopologicalSort(n.g)
	if err != nil {
		return n.order
	}

	result := make([]stats.Var, 0, len(sorted))

	for _, id := range sorted {
		result = append(result, stats.Var(id))
	}

	return result
}

func (n *network) Validate() error {
	cassert.NotNil(n, "network is nil")

	// Validate all nodes and ensure all edges are present.
	for _, v := range n.order {
		gn, _ := n.g.Node(string(v))
		node, ok := gn.Metadata.(Node)
		cassert.True(ok, "metadata is not a Node")

		// Parents must exist.
		for _, p := range node.Parents {
			if !n.g.HasNode(string(p)) {
				return bayesian.ErrValidation(fmt.Errorf("parent %s of %s not in network", string(p), string(node.Variable)))
			}

			edgeID := string(p) + "->" + string(node.Variable)

			if !n.g.HasEdge(edgeID) {
				_ = n.g.AddEdge(graph.Edge{
					ID:   edgeID,
					From: string(p),
					To:   string(node.Variable),
				})
			}
		}

		// Outcomes must be non-empty.
		if len(node.Outcomes) == 0 {
			return bayesian.ErrValidation(fmt.Errorf("node %s has no outcomes", string(node.Variable)))
		}

		// CPT must be valid.
		err := node.CPT.Validate()
		if err != nil {
			return bayesian.ErrValidation(fmt.Errorf("CPT validation failed for %s: %w", string(node.Variable), err))
		}
	}

	// Check for cycles.
	if graph.HasCycle(n.g) {
		return bayesian.ErrValidation(bayesian.ErrCyclicNetwork)
	}

	return nil
}
