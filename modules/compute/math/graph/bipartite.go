package graph

// Bipartite is an undirected bipartite graph with two explicit partitions.
type Bipartite struct {
	g     *Undirected
	left  map[string]bool
	right map[string]bool
}

// NewBipartite creates an empty bipartite graph.
func NewBipartite() *Bipartite {
	return &Bipartite{
		g:     NewUndirected(),
		left:  make(map[string]bool),
		right: make(map[string]bool),
	}
}

// NewBipartiteFrom creates a bipartite graph from an undirected graph after verifying bipartiteness.
func NewBipartiteFrom(u *Undirected) (*Bipartite, error) {
	left, right, ok := checkBipartite(u)
	if !ok {
		return nil, ErrGraph(ErrNotBipartite)
	}

	return &Bipartite{
		g:     u.CloneUndirected(),
		left:  left,
		right: right,
	}, nil
}

// Undirected returns the underlying undirected graph.
func (b *Bipartite) Undirected() *Undirected {
	return b.g
}

// Left returns the left partition node IDs, sorted.
func (b *Bipartite) Left() []string {
	return sortedKeys(b.left)
}

// Right returns the right partition node IDs, sorted.
func (b *Bipartite) Right() []string {
	return sortedKeys(b.right)
}

// AddNodeLeft adds a node to the left partition.
func (b *Bipartite) AddNodeLeft(node Node) error {
	if b.right[node.ID] {
		return ErrGraph(ErrInvalidEdge)
	}

	err := b.g.AddNode(node)
	if err != nil {
		return err
	}

	b.left[node.ID] = true

	return nil
}

// AddNodeRight adds a node to the right partition.
func (b *Bipartite) AddNodeRight(node Node) error {
	if b.left[node.ID] {
		return ErrGraph(ErrInvalidEdge)
	}

	err := b.g.AddNode(node)
	if err != nil {
		return err
	}

	b.right[node.ID] = true

	return nil
}

// AddNode adds a node (defaults to left partition).
func (b *Bipartite) AddNode(node Node) error {
	return b.AddNodeLeft(node)
}

// RemoveNode removes a node and all its incident edges.
func (b *Bipartite) RemoveNode(id string) error {
	err := b.g.RemoveNode(id)
	if err != nil {
		return err
	}

	delete(b.left, id)
	delete(b.right, id)

	return nil
}

// AddEdge adds an edge between nodes in different partitions.
func (b *Bipartite) AddEdge(edge Edge) error {
	samePartition := (b.left[edge.From] && b.left[edge.To]) ||
		(b.right[edge.From] && b.right[edge.To])

	if samePartition {
		return ErrGraph(ErrNotBipartite)
	}

	return b.g.AddEdge(edge)
}

// RemoveEdge removes an edge by its ID.
func (b *Bipartite) RemoveEdge(id string) error {
	return b.g.RemoveEdge(id)
}

// Node returns the node with the given ID.
func (b *Bipartite) Node(id string) (Node, error) {
	return b.g.Node(id)
}

// Nodes returns all nodes sorted by ID.
func (b *Bipartite) Nodes() []Node {
	return b.g.Nodes()
}

// Edge returns the edge with the given ID.
func (b *Bipartite) Edge(id string) (Edge, error) {
	return b.g.Edge(id)
}

// Edges returns all edges sorted by ID.
func (b *Bipartite) Edges() []Edge {
	return b.g.Edges()
}

// Neighbors returns the IDs of adjacent nodes, sorted.
func (b *Bipartite) Neighbors(id string) ([]string, error) {
	return b.g.Neighbors(id)
}

// HasNode reports whether the graph contains a node with the given ID.
func (b *Bipartite) HasNode(id string) bool {
	return b.g.HasNode(id)
}

// HasEdge reports whether the graph contains an edge with the given ID.
func (b *Bipartite) HasEdge(id string) bool {
	return b.g.HasEdge(id)
}

// NodeCount returns the number of nodes.
func (b *Bipartite) NodeCount() int {
	return b.g.NodeCount()
}

// EdgeCount returns the number of edges.
func (b *Bipartite) EdgeCount() int {
	return b.g.EdgeCount()
}

// Degree returns the degree of a node.
func (b *Bipartite) Degree(id string) (int, error) {
	return b.g.Degree(id)
}

// Clone returns a deep copy of the bipartite graph.
func (b *Bipartite) Clone() Graph {
	return b.CloneBipartite()
}

// IsDirected reports whether the graph is directed.
func (b *Bipartite) IsDirected() bool {
	return false
}

// CloneBipartite returns a deep copy as *Bipartite.
func (b *Bipartite) CloneBipartite() *Bipartite {
	nb := &Bipartite{
		g:     b.g.CloneUndirected(),
		left:  make(map[string]bool, len(b.left)),
		right: make(map[string]bool, len(b.right)),
	}

	for k := range b.left {
		nb.left[k] = true
	}

	for k := range b.right {
		nb.right[k] = true
	}

	return nb
}

// checkBipartite checks if an undirected graph is bipartite using BFS coloring.
func checkBipartite(g *Undirected) (map[string]bool, map[string]bool, bool) {
	leftSet := make(map[string]bool)
	rightSet := make(map[string]bool)
	color := make(map[string]int) // 0=unvisited, 1=left, 2=right

	for _, n := range g.Nodes() {
		if color[n.ID] != 0 {
			continue
		}

		queue := []string{n.ID}
		color[n.ID] = 1
		leftSet[n.ID] = true

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			neighbors, _ := g.Neighbors(curr)

			for _, nb := range neighbors {
				if color[nb] == color[curr] && color[nb] != 0 {
					return nil, nil, false
				}

				if color[nb] != 0 {
					continue
				}

				if color[curr] == 1 {
					color[nb] = 2
					rightSet[nb] = true
				} else {
					color[nb] = 1
					leftSet[nb] = true
				}

				queue = append(queue, nb)
			}
		}
	}

	return leftSet, rightSet, true
}
