package graph

// MultigraphDirected is a directed multigraph allowing parallel edges and self-loops.
type MultigraphDirected struct {
	g *Directed
}

// NewMultigraphDirected creates an empty directed multigraph.
func NewMultigraphDirected() *MultigraphDirected {
	return &MultigraphDirected{g: NewDirected()}
}

// Directed returns the underlying directed graph.
func (m *MultigraphDirected) Directed() *Directed {
	return m.g
}

// AddNode adds a node to the multigraph.
func (m *MultigraphDirected) AddNode(node Node) error {
	return m.g.AddNode(node)
}

// RemoveNode removes a node and all its incident edges.
func (m *MultigraphDirected) RemoveNode(id string) error {
	return m.g.RemoveNode(id)
}

// AddEdge adds an edge to the multigraph. Parallel edges are allowed.
func (m *MultigraphDirected) AddEdge(edge Edge) error {
	return m.g.AddEdge(edge)
}

// RemoveEdge removes an edge by its ID.
func (m *MultigraphDirected) RemoveEdge(id string) error {
	return m.g.RemoveEdge(id)
}

// Node returns the node with the given ID.
func (m *MultigraphDirected) Node(id string) (Node, error) {
	return m.g.Node(id)
}

// Nodes returns all nodes sorted by ID.
func (m *MultigraphDirected) Nodes() []Node {
	return m.g.Nodes()
}

// Edge returns the edge with the given ID.
func (m *MultigraphDirected) Edge(id string) (Edge, error) {
	return m.g.Edge(id)
}

// Edges returns all edges sorted by ID.
func (m *MultigraphDirected) Edges() []Edge {
	return m.g.Edges()
}

// Neighbors returns the IDs of successor nodes, sorted.
func (m *MultigraphDirected) Neighbors(id string) ([]string, error) {
	return m.g.Neighbors(id)
}

// HasNode reports whether the multigraph contains a node with the given ID.
func (m *MultigraphDirected) HasNode(id string) bool {
	return m.g.HasNode(id)
}

// HasEdge reports whether the multigraph contains an edge with the given ID.
func (m *MultigraphDirected) HasEdge(id string) bool {
	return m.g.HasEdge(id)
}

// NodeCount returns the number of nodes.
func (m *MultigraphDirected) NodeCount() int {
	return m.g.NodeCount()
}

// EdgeCount returns the number of edges.
func (m *MultigraphDirected) EdgeCount() int {
	return m.g.EdgeCount()
}

// Degree returns the degree of a node.
func (m *MultigraphDirected) Degree(id string) (int, error) {
	return m.g.Degree(id)
}

// Clone returns a deep copy of the directed multigraph.
func (m *MultigraphDirected) Clone() Graph {
	return m.CloneMultigraphDirected()
}

// IsDirected reports whether the graph is directed.
func (m *MultigraphDirected) IsDirected() bool {
	return true
}

// CloneMultigraphDirected returns a deep copy as *MultigraphDirected.
func (m *MultigraphDirected) CloneMultigraphDirected() *MultigraphDirected {
	return &MultigraphDirected{g: m.g.CloneDirected()}
}

// EdgesBetween returns all edge IDs from src to dst, sorted.
func (m *MultigraphDirected) EdgesBetween(src, dst string) ([]string, error) {
	if !m.g.HasNode(src) || !m.g.HasNode(dst) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	var result []string

	for _, eid := range m.g.out[src] {
		edge := m.g.edges[eid]
		if edge.To == dst {
			result = append(result, eid)
		}
	}

	return result, nil
}
