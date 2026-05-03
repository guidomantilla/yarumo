package graph

// MultigraphUndirected is an undirected multigraph allowing parallel edges and self-loops.
type MultigraphUndirected struct {
	g *Undirected
}

// NewMultigraphUndirected creates an empty undirected multigraph.
func NewMultigraphUndirected() *MultigraphUndirected {
	return &MultigraphUndirected{g: NewUndirected()}
}

// Undirected returns the underlying undirected graph.
func (m *MultigraphUndirected) Undirected() *Undirected {
	return m.g
}

// AddNode adds a node to the multigraph.
func (m *MultigraphUndirected) AddNode(node Node) error {
	return m.g.AddNode(node)
}

// RemoveNode removes a node and all its incident edges.
func (m *MultigraphUndirected) RemoveNode(id string) error {
	return m.g.RemoveNode(id)
}

// AddEdge adds an edge to the multigraph. Parallel edges are allowed.
func (m *MultigraphUndirected) AddEdge(edge Edge) error {
	return m.g.AddEdge(edge)
}

// RemoveEdge removes an edge by its ID.
func (m *MultigraphUndirected) RemoveEdge(id string) error {
	return m.g.RemoveEdge(id)
}

// Node returns the node with the given ID.
func (m *MultigraphUndirected) Node(id string) (Node, error) {
	return m.g.Node(id)
}

// Nodes returns all nodes sorted by ID.
func (m *MultigraphUndirected) Nodes() []Node {
	return m.g.Nodes()
}

// Edge returns the edge with the given ID.
func (m *MultigraphUndirected) Edge(id string) (Edge, error) {
	return m.g.Edge(id)
}

// Edges returns all edges sorted by ID.
func (m *MultigraphUndirected) Edges() []Edge {
	return m.g.Edges()
}

// Neighbors returns the IDs of adjacent nodes, sorted.
func (m *MultigraphUndirected) Neighbors(id string) ([]string, error) {
	return m.g.Neighbors(id)
}

// HasNode reports whether the multigraph contains a node with the given ID.
func (m *MultigraphUndirected) HasNode(id string) bool {
	return m.g.HasNode(id)
}

// HasEdge reports whether the multigraph contains an edge with the given ID.
func (m *MultigraphUndirected) HasEdge(id string) bool {
	return m.g.HasEdge(id)
}

// NodeCount returns the number of nodes.
func (m *MultigraphUndirected) NodeCount() int {
	return m.g.NodeCount()
}

// EdgeCount returns the number of edges.
func (m *MultigraphUndirected) EdgeCount() int {
	return m.g.EdgeCount()
}

// Degree returns the degree of a node.
func (m *MultigraphUndirected) Degree(id string) (int, error) {
	return m.g.Degree(id)
}

// Clone returns a deep copy of the undirected multigraph.
func (m *MultigraphUndirected) Clone() Graph {
	return m.CloneMultigraphUndirected()
}

// IsDirected reports whether the graph is directed.
func (m *MultigraphUndirected) IsDirected() bool {
	return false
}

// CloneMultigraphUndirected returns a deep copy as *MultigraphUndirected.
func (m *MultigraphUndirected) CloneMultigraphUndirected() *MultigraphUndirected {
	return &MultigraphUndirected{g: m.g.CloneUndirected()}
}

// EdgesBetween returns all edge IDs between two nodes, sorted.
func (m *MultigraphUndirected) EdgesBetween(a, b string) ([]string, error) {
	if !m.g.HasNode(a) || !m.g.HasNode(b) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	var result []string

	for _, eid := range m.g.adj[a] {
		edge := m.g.edges[eid]
		if (edge.From == a && edge.To == b) || (edge.From == b && edge.To == a) {
			result = append(result, eid)
		}
	}

	return result, nil
}
