package graph

import (
	"cmp"
	"slices"
)

// Directed is a directed graph allowing self-loops.
type Directed struct {
	nodes map[string]Node
	edges map[string]Edge
	out   map[string][]string // node ID -> outgoing edge IDs
	in    map[string][]string // node ID -> incoming edge IDs
}

// NewDirected creates an empty directed graph.
func NewDirected() *Directed {
	return &Directed{
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
		out:   make(map[string][]string),
		in:    make(map[string][]string),
	}
}

// AddNode adds a node to the directed graph.
func (g *Directed) AddNode(node Node) error {
	if _, exists := g.nodes[node.ID]; exists {
		return ErrGraph(ErrDuplicateNode)
	}

	g.nodes[node.ID] = node
	g.out[node.ID] = nil
	g.in[node.ID] = nil

	return nil
}

// RemoveNode removes a node and all its incident edges.
func (g *Directed) RemoveNode(id string) error {
	if _, exists := g.nodes[id]; !exists {
		return ErrGraph(ErrNodeNotFound)
	}

	edgesToRemove := make([]string, 0)
	edgesToRemove = append(edgesToRemove, g.out[id]...)
	edgesToRemove = append(edgesToRemove, g.in[id]...)

	for _, eid := range edgesToRemove {
		g.removeEdgeInternal(eid)
	}

	delete(g.nodes, id)
	delete(g.out, id)
	delete(g.in, id)

	return nil
}

// AddEdge adds an edge to the directed graph.
func (g *Directed) AddEdge(edge Edge) error {
	if !g.HasNode(edge.From) || !g.HasNode(edge.To) {
		return ErrGraph(ErrInvalidEdge, ErrNodeNotFound)
	}

	if _, exists := g.edges[edge.ID]; exists {
		return ErrGraph(ErrInvalidEdge)
	}

	g.edges[edge.ID] = edge
	g.out[edge.From] = append(g.out[edge.From], edge.ID)
	g.in[edge.To] = append(g.in[edge.To], edge.ID)

	return nil
}

// RemoveEdge removes an edge by its ID.
func (g *Directed) RemoveEdge(id string) error {
	if _, exists := g.edges[id]; !exists {
		return ErrGraph(ErrEdgeNotFound)
	}

	g.removeEdgeInternal(id)

	return nil
}

// Node returns the node with the given ID.
func (g *Directed) Node(id string) (Node, error) {
	node, exists := g.nodes[id]
	if !exists {
		return Node{}, ErrGraph(ErrNodeNotFound)
	}

	return node, nil
}

// Nodes returns all nodes sorted by ID.
func (g *Directed) Nodes() []Node {
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
func (g *Directed) Edge(id string) (Edge, error) {
	edge, exists := g.edges[id]
	if !exists {
		return Edge{}, ErrGraph(ErrEdgeNotFound)
	}

	return edge, nil
}

// Edges returns all edges sorted by ID.
func (g *Directed) Edges() []Edge {
	result := make([]Edge, 0, len(g.edges))

	for _, e := range g.edges {
		result = append(result, e)
	}

	slices.SortFunc(result, func(a, b Edge) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Neighbors returns the IDs of nodes reachable via outgoing edges, sorted alphabetically.
func (g *Directed) Neighbors(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	seen := make(map[string]bool)

	for _, eid := range g.out[id] {
		edge := g.edges[eid]
		seen[edge.To] = true
	}

	result := make([]string, 0, len(seen))

	for nid := range seen {
		result = append(result, nid)
	}

	slices.Sort(result)

	return result, nil
}

// HasNode reports whether the graph contains a node with the given ID.
func (g *Directed) HasNode(id string) bool {
	_, exists := g.nodes[id]
	return exists
}

// HasEdge reports whether the graph contains an edge with the given ID.
func (g *Directed) HasEdge(id string) bool {
	_, exists := g.edges[id]
	return exists
}

// NodeCount returns the number of nodes.
func (g *Directed) NodeCount() int {
	return len(g.nodes)
}

// EdgeCount returns the number of edges.
func (g *Directed) EdgeCount() int {
	return len(g.edges)
}

// Degree returns the degree of a node (in-degree + out-degree).
func (g *Directed) Degree(id string) (int, error) {
	if !g.HasNode(id) {
		return 0, ErrGraph(ErrNodeNotFound)
	}

	return len(g.out[id]) + len(g.in[id]), nil
}

// Clone returns a deep copy of the directed graph.
func (g *Directed) Clone() Graph {
	return g.CloneDirected()
}

// IsDirected reports whether the graph is directed.
func (g *Directed) IsDirected() bool {
	return true
}

// CloneDirected returns a deep copy as *Directed.
func (g *Directed) CloneDirected() *Directed {
	ng := NewDirected()

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

// InEdges returns the incoming edge IDs for a node, sorted.
func (g *Directed) InEdges(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	result := make([]string, len(g.in[id]))
	copy(result, g.in[id])
	slices.Sort(result)

	return result, nil
}

// OutEdges returns the outgoing edge IDs for a node, sorted.
func (g *Directed) OutEdges(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	result := make([]string, len(g.out[id]))
	copy(result, g.out[id])
	slices.Sort(result)

	return result, nil
}

// Predecessors returns IDs of nodes with edges pointing to the given node, sorted.
func (g *Directed) Predecessors(id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	seen := make(map[string]bool)

	for _, eid := range g.in[id] {
		edge := g.edges[eid]
		seen[edge.From] = true
	}

	result := make([]string, 0, len(seen))

	for nid := range seen {
		result = append(result, nid)
	}

	slices.Sort(result)

	return result, nil
}

// Successors returns IDs of nodes reachable via outgoing edges, sorted.
func (g *Directed) Successors(id string) ([]string, error) {
	return g.Neighbors(id)
}

func (g *Directed) removeEdgeInternal(id string) {
	edge, exists := g.edges[id]
	if !exists {
		return
	}

	g.out[edge.From] = removeFromSlice(g.out[edge.From], id)
	g.in[edge.To] = removeFromSlice(g.in[edge.To], id)

	delete(g.edges, id)
}

func sortedKeys(m map[string]bool) []string {
	result := make([]string, 0, len(m))

	for k := range m {
		result = append(result, k)
	}

	slices.Sort(result)

	return result
}

func removeFromSlice(s []string, item string) []string {
	result := make([]string, 0, len(s))

	for _, v := range s {
		if v != item {
			result = append(result, v)
		}
	}

	return result
}
