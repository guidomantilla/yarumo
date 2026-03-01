package network

import (
	"fmt"
	"slices"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian"
)

type network struct {
	nodes map[probability.Var]Node
	order []probability.Var
}

// NewNetwork creates an empty Bayesian network.
func NewNetwork() Network {
	return &network{
		nodes: make(map[probability.Var]Node),
	}
}

func (n *network) AddNode(node Node) {
	cassert.NotNil(n, "network is nil")

	n.nodes[node.Variable] = node
	n.order = append(n.order, node.Variable)
}

func (n *network) Node(variable probability.Var) (Node, bool) {
	cassert.NotNil(n, "network is nil")

	node, ok := n.nodes[variable]

	return node, ok
}

func (n *network) Nodes() []Node {
	cassert.NotNil(n, "network is nil")

	result := make([]Node, 0, len(n.nodes))

	for _, v := range n.order {
		result = append(result, n.nodes[v])
	}

	return result
}

func (n *network) Parents(variable probability.Var) []probability.Var {
	cassert.NotNil(n, "network is nil")

	node, ok := n.nodes[variable]
	if !ok {
		return nil
	}

	return node.Parents
}

func (n *network) Children(variable probability.Var) []probability.Var {
	cassert.NotNil(n, "network is nil")

	var children []probability.Var

	for _, node := range n.nodes {
		if slices.Contains(node.Parents, variable) {
			children = append(children, node.Variable)
		}
	}

	slices.Sort(children)

	return children
}

func (n *network) TopologicalOrder() []probability.Var {
	cassert.NotNil(n, "network is nil")

	visited := make(map[probability.Var]bool)

	var result []probability.Var

	var visit func(v probability.Var)

	visit = func(v probability.Var) {
		if visited[v] {
			return
		}

		visited[v] = true
		node := n.nodes[v]

		for _, parent := range node.Parents {
			visit(parent)
		}

		result = append(result, v)
	}

	for _, v := range n.order {
		visit(v)
	}

	return result
}

func (n *network) Validate() error {
	cassert.NotNil(n, "network is nil")

	// Check for cycles using DFS coloring.
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[probability.Var]int)

	var hasCycle func(v probability.Var) bool

	hasCycle = func(v probability.Var) bool {
		color[v] = gray

		for _, child := range n.Children(v) {
			if color[child] == gray {
				return true
			}

			if color[child] == white && hasCycle(child) {
				return true
			}
		}

		color[v] = black

		return false
	}

	for _, v := range n.order {
		if color[v] == white && hasCycle(v) {
			return bayesian.ErrCyclicNetwork
		}
	}

	// Validate all CPTs.
	for _, node := range n.nodes {
		err := node.CPT.Validate()
		if err != nil {
			return fmt.Errorf("CPT validation failed for %s: %w", string(node.Variable), err)
		}
	}

	return nil
}
