package graph

import (
	"cmp"
	"slices"
)

// Undirected is an undirected graph allowing self-loops.
type Undirected struct {
	nodes map[string]Node
	edges map[string]Edge
	adj   map[string][]string // node ID -> incident edge IDs
}

// NewUndirected creates an empty undirected graph.
func NewUndirected() *Undirected {
	return &Undirected{
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
		adj:   make(map[string][]string),
	}
}

// AddNode adds a node to the undirected graph.
func (g *Undirected) AddNode(node Node) error {
	if _, exists := g.nodes[node.ID]; exists {
		return ErrGraph(ErrDuplicateNode)
	}

	g.nodes[node.ID] = node
	g.adj[node.ID] = nil

	return nil
}

// RemoveNode removes a node and all its incident edges.
func (g *Undirected) RemoveNode(id string) error {
	if _, exists := g.nodes[id]; !exists {
		return ErrGraph(ErrNodeNotFound)
	}

	edgesToRemove := make([]string, len(g.adj[id]))
	copy(edgesToRemove, g.adj[id])

	for _, eid := range edgesToRemove {
		g.removeEdgeInternal(eid)
	}

	delete(g.nodes, id)
	delete(g.adj, id)

	return nil
}

// AddEdge adds an edge to the undirected graph.
func (g *Undirected) AddEdge(edge Edge) error {
	if !g.HasNode(edge.From) || !g.HasNode(edge.To) {
		return ErrGraph(ErrInvalidEdge, ErrNodeNotFound)
	}

	if _, exists := g.edges[edge.ID]; exists {
		return ErrGraph(ErrInvalidEdge)
	}

	g.edges[edge.ID] = edge
	g.adj[edge.From] = append(g.adj[edge.From], edge.ID)

	if edge.From != edge.To {
		g.adj[edge.To] = append(g.adj[edge.To], edge.ID)
	}

	return nil
}

// RemoveEdge removes an edge by its ID.
func (g *Undirected) RemoveEdge(id string) error {
	if _, exists := g.edges[id]; !exists {
		return ErrGraph(ErrEdgeNotFound)
	}

	g.removeEdgeInternal(id)

	return nil
}

// Node returns the node with the given ID.
func (g *Undirected) Node(id string) (Node, error) {
	node, exists := g.nodes[id]
	if !exists {
		return Node{}, ErrGraph(ErrNodeNotFound)
	}

	return node, nil
}

// Nodes returns all nodes sorted by ID.
func (g *Undirected) Nodes() []Node {
	result := make([]Node, 0, len(g.nodes))

	for _, n := range g.nodes {
		result = append(result, n)
	}

	slices.SortFunc(result, func(a, b Node) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Edge returns the edge with the given ID.
func (g *Undirected) Edge(id string) (Edge, error) {
	edge, exists := g.edges[id]
	if !exists {
		return Edge{}, ErrGraph(ErrEdgeNotFound)
	}

	return edge, nil
}

// Edges returns all edges sorted by ID.
func (g *Undirected) Edges() []Edge {
	result := make([]Edge, 0, len(g.edges))

	for _, e := range g.edges {
		result = append(result, e)
	}

	slices.SortFunc(result, func(a, b Edge) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Neighbors returns the IDs of adjacent nodes, sorted alphabetically.
func (g *Undirected) Neighbors(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	seen := make(map[string]bool)

	for _, eid := range g.adj[id] {
		edge := g.edges[eid]

		if edge.From == id {
			seen[edge.To] = true
		} else {
			seen[edge.From] = true
		}
	}

	result := make([]string, 0, len(seen))

	for nid := range seen {
		result = append(result, nid)
	}

	slices.Sort(result)

	return result, nil
}

// HasNode reports whether the graph contains a node with the given ID.
func (g *Undirected) HasNode(id string) bool {
	_, exists := g.nodes[id]
	return exists
}

// HasEdge reports whether the graph contains an edge with the given ID.
func (g *Undirected) HasEdge(id string) bool {
	_, exists := g.edges[id]
	return exists
}

// NodeCount returns the number of nodes.
func (g *Undirected) NodeCount() int {
	return len(g.nodes)
}

// EdgeCount returns the number of edges.
func (g *Undirected) EdgeCount() int {
	return len(g.edges)
}

// Degree returns the degree of a node.
func (g *Undirected) Degree(id string) (int, error) {
	if !g.HasNode(id) {
		return 0, ErrGraph(ErrNodeNotFound)
	}

	degree := 0

	for _, eid := range g.adj[id] {
		edge := g.edges[eid]

		if edge.From == edge.To {
			degree += 2
		} else {
			degree++
		}
	}

	return degree, nil
}

// Clone returns a deep copy of the undirected graph.
func (g *Undirected) Clone() Graph {
	return g.CloneUndirected()
}

// IsDirected reports whether the graph is directed.
func (g *Undirected) IsDirected() bool {
	return false
}

// CloneUndirected returns a deep copy as *Undirected.
func (g *Undirected) CloneUndirected() *Undirected {
	ng := NewUndirected()

	for _, n := range g.Nodes() {
		_ = ng.AddNode(Node{ID: n.ID, Metadata: n.Metadata})
	}

	for _, e := range g.Edges() {
		_ = ng.AddEdge(Edge{
			ID: e.ID, From: e.From, To: e.To,
			Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
		})
	}

	return ng
}

// IncidentEdges returns the incident edge IDs for a node, sorted.
func (g *Undirected) IncidentEdges(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	result := make([]string, len(g.adj[id]))
	copy(result, g.adj[id])
	slices.Sort(result)

	return result, nil
}

func (g *Undirected) removeEdgeInternal(id string) {
	edge, exists := g.edges[id]
	if !exists {
		return
	}

	g.adj[edge.From] = removeFromSlice(g.adj[edge.From], id)

	if edge.From != edge.To {
		g.adj[edge.To] = removeFromSlice(g.adj[edge.To], id)
	}

	delete(g.edges, id)
}
