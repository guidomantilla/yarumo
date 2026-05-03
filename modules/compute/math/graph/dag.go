package graph

// DAG is a directed acyclic graph. Self-loops are not allowed.
type DAG struct {
	g *Directed
}

// NewDAG creates an empty DAG.
func NewDAG() *DAG {
	return &DAG{g: NewDirected()}
}

// NewDAGFrom creates a DAG wrapping an existing directed graph after verifying acyclicity.
func NewDAGFrom(d *Directed) (*DAG, error) {
	if hasCycleDirected(d) {
		return nil, ErrGraph(ErrNotDAG)
	}

	return &DAG{g: d.CloneDirected()}, nil
}

// Directed returns the underlying directed graph.
func (d *DAG) Directed() *Directed {
	return d.g
}

// AddNode adds a node to the DAG.
func (d *DAG) AddNode(node Node) error {
	return d.g.AddNode(node)
}

// RemoveNode removes a node and all its incident edges.
func (d *DAG) RemoveNode(id string) error {
	return d.g.RemoveNode(id)
}

// AddEdge adds an edge to the DAG, rejecting self-loops and cycles.
func (d *DAG) AddEdge(edge Edge) error {
	if edge.From == edge.To {
		return ErrGraph(ErrSelfLoop)
	}

	err := d.g.AddEdge(edge)
	if err != nil {
		return err
	}

	if hasCycleDirected(d.g) {
		d.g.removeEdgeInternal(edge.ID)

		return ErrGraph(ErrCycleDetected)
	}

	return nil
}

// RemoveEdge removes an edge by its ID.
func (d *DAG) RemoveEdge(id string) error {
	return d.g.RemoveEdge(id)
}

// Node returns the node with the given ID.
func (d *DAG) Node(id string) (Node, error) {
	return d.g.Node(id)
}

// Nodes returns all nodes sorted by ID.
func (d *DAG) Nodes() []Node {
	return d.g.Nodes()
}

// Edge returns the edge with the given ID.
func (d *DAG) Edge(id string) (Edge, error) {
	return d.g.Edge(id)
}

// Edges returns all edges sorted by ID.
func (d *DAG) Edges() []Edge {
	return d.g.Edges()
}

// Neighbors returns the IDs of successor nodes, sorted.
func (d *DAG) Neighbors(id string) ([]string, error) {
	return d.g.Neighbors(id)
}

// HasNode reports whether the DAG contains a node with the given ID.
func (d *DAG) HasNode(id string) bool {
	return d.g.HasNode(id)
}

// HasEdge reports whether the DAG contains an edge with the given ID.
func (d *DAG) HasEdge(id string) bool {
	return d.g.HasEdge(id)
}

// NodeCount returns the number of nodes.
func (d *DAG) NodeCount() int {
	return d.g.NodeCount()
}

// EdgeCount returns the number of edges.
func (d *DAG) EdgeCount() int {
	return d.g.EdgeCount()
}

// Degree returns the degree of a node.
func (d *DAG) Degree(id string) (int, error) {
	return d.g.Degree(id)
}

// Clone returns a deep copy of the DAG.
func (d *DAG) Clone() Graph {
	return d.CloneDAG()
}

// IsDirected reports whether the graph is directed.
func (d *DAG) IsDirected() bool {
	return true
}

// CloneDAG returns a deep copy as *DAG.
func (d *DAG) CloneDAG() *DAG {
	return &DAG{g: d.g.CloneDirected()}
}

// Roots returns all nodes with no incoming edges, sorted by ID.
func (d *DAG) Roots() []string {
	return roots(d.g)
}

// Leaves returns all nodes with no outgoing edges, sorted by ID.
func (d *DAG) Leaves() []string {
	return leaves(d.g)
}

// hasCycleDirected checks if a directed graph has a cycle using DFS coloring.
func hasCycleDirected(g *Directed) bool {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)

	for _, n := range g.Nodes() {
		color[n.ID] = white
	}

	var visit func(string) bool
	visit = func(id string) bool {
		color[id] = gray

		for _, eid := range g.out[id] {
			edge := g.edges[eid]
			target := edge.To

			switch color[target] {
			case gray:
				return true
			case white:
				if visit(target) {
					return true
				}
			}
		}

		color[id] = black

		return false
	}

	for _, n := range g.Nodes() {
		if color[n.ID] == white {
			if visit(n.ID) {
				return true
			}
		}
	}

	return false
}

func roots(g *Directed) []string {
	var result []string

	for _, n := range g.Nodes() {
		if len(g.in[n.ID]) == 0 {
			result = append(result, n.ID)
		}
	}

	return result
}

func leaves(g *Directed) []string {
	var result []string

	for _, n := range g.Nodes() {
		if len(g.out[n.ID]) == 0 {
			result = append(result, n.ID)
		}
	}

	return result
}
