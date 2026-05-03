package graph

import (
	"strconv"
)

// Subgraph returns a new directed graph containing only the specified nodes and their connecting edges.
func Subgraph(g *Directed, nodeIDs []string) *Directed {
	ng := NewDirected()
	keep := make(map[string]bool, len(nodeIDs))

	for _, id := range nodeIDs {
		keep[id] = true
	}

	for _, n := range g.Nodes() {
		if keep[n.ID] {
			_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
		}
	}

	for _, e := range g.Edges() {
		if keep[e.From] && keep[e.To] {
			_ = ng.AddEdge(Edge{
				ID: e.ID, From: e.From, To: e.To,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return ng
}

// SubgraphUndirected returns a new undirected graph containing only the specified nodes.
func SubgraphUndirected(g *Undirected, nodeIDs []string) *Undirected {
	ng := NewUndirected()
	keep := make(map[string]bool, len(nodeIDs))

	for _, id := range nodeIDs {
		keep[id] = true
	}

	for _, n := range g.Nodes() {
		if keep[n.ID] {
			_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
		}
	}

	for _, e := range g.Edges() {
		if keep[e.From] && keep[e.To] {
			_ = ng.AddEdge(Edge{
				ID: e.ID, From: e.From, To: e.To,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return ng
}

// Union returns a new directed graph containing all nodes and edges from both graphs.
func Union(a, b *Directed) *Directed {
	ng := a.CloneDirected()

	for _, n := range b.Nodes() {
		if !ng.HasNode(n.ID) {
			_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
		}
	}

	for _, e := range b.Edges() {
		if !ng.HasEdge(e.ID) {
			_ = ng.AddEdge(Edge{
				ID: e.ID, From: e.From, To: e.To,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return ng
}

// Intersection returns a new directed graph containing only nodes and edges present in both graphs.
func Intersection(a, b *Directed) *Directed {
	ng := NewDirected()

	for _, n := range a.Nodes() {
		if b.HasNode(n.ID) {
			_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
		}
	}

	for _, e := range a.Edges() {
		if b.HasEdge(e.ID) && ng.HasNode(e.From) && ng.HasNode(e.To) {
			_ = ng.AddEdge(Edge{
				ID: e.ID, From: e.From, To: e.To,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return ng
}

// Complement returns the complement of a directed graph.
// Contains edges for every pair not connected in the original.
func Complement(g *Directed) *Directed {
	ng := NewDirected()

	for _, n := range g.Nodes() {
		_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
	}

	edgeID := 0
	nodes := g.Nodes()

	for _, u := range nodes {
		for _, v := range nodes {
			if u.ID == v.ID {
				continue
			}

			hasEdge := false

			for _, eid := range g.out[u.ID] {
				edge := g.edges[eid]
				if edge.To == v.ID {
					hasEdge = true

					break
				}
			}

			if !hasEdge {
				_ = ng.AddEdge(Edge{
					ID:   "c" + strconv.Itoa(edgeID),
					From: u.ID, To: v.ID,
				})
				edgeID++
			}
		}
	}

	return ng
}

// Reverse returns a new directed graph with all edges reversed.
func Reverse(g *Directed) *Directed {
	ng := NewDirected()

	for _, n := range g.Nodes() {
		_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
	}

	for _, e := range g.Edges() {
		_ = ng.AddEdge(Edge{
			ID: e.ID, From: e.To, To: e.From,
			Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
		})
	}

	return ng
}

// CartesianProduct returns the Cartesian product of two directed graphs.
func CartesianProduct(a, b *Directed) *Directed {
	ng := NewDirected()

	aNodes := a.Nodes()
	bNodes := b.Nodes()

	for _, u := range aNodes {
		for _, v := range bNodes {
			_ = ng.AddNode(Node{ID: u.ID + ":" + v.ID})
		}
	}

	edgeID := 0

	for _, u := range aNodes {
		for _, e := range a.Edges() {
			if e.From != u.ID {
				continue
			}

			for _, v := range bNodes {
				_ = ng.AddEdge(Edge{
					ID:     "cp" + strconv.Itoa(edgeID),
					From:   u.ID + ":" + v.ID,
					To:     e.To + ":" + v.ID,
					Weight: e.Weight,
				})
				edgeID++
			}
		}
	}

	for _, v := range bNodes {
		for _, e := range b.Edges() {
			if e.From != v.ID {
				continue
			}

			for _, u := range aNodes {
				_ = ng.AddEdge(Edge{
					ID:     "cp" + strconv.Itoa(edgeID),
					From:   u.ID + ":" + v.ID,
					To:     u.ID + ":" + e.To,
					Weight: e.Weight,
				})
				edgeID++
			}
		}
	}

	return ng
}
