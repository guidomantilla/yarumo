package graph

// Tree is a rooted tree (a DAG where each node has at most one parent).
type Tree struct {
	dag  *DAG
	root string
}

// NewTree creates an empty tree with the given root node.
func NewTree(root Node) *Tree {
	d := NewDAG()
	_ = d.AddNode(root)

	return &Tree{dag: d, root: root.ID}
}

// NewTreeFrom creates a tree from an existing DAG after verifying single-parent constraint.
func NewTreeFrom(d *DAG) (*Tree, error) {
	r := d.Roots()
	if len(r) != 1 {
		return nil, ErrGraph(ErrNotTree)
	}

	for _, n := range d.Nodes() {
		if n.ID == r[0] {
			continue
		}

		inEdges := d.g.in[n.ID]
		if len(inEdges) != 1 {
			return nil, ErrGraph(ErrMultipleParents)
		}
	}

	return &Tree{dag: d.CloneDAG(), root: r[0]}, nil
}

// DAG returns the underlying DAG.
func (t *Tree) DAG() *DAG {
	return t.dag
}

// Directed returns the underlying directed graph.
func (t *Tree) Directed() *Directed {
	return t.dag.Directed()
}

// Root returns the root node ID.
func (t *Tree) Root() string {
	return t.root
}

// AddNode adds a node to the tree.
func (t *Tree) AddNode(node Node) error {
	return t.dag.AddNode(node)
}

// RemoveNode removes a node, its incident edges, and cascade-removes the disconnected subtree.
func (t *Tree) RemoveNode(id string) error {
	if id == t.root {
		return ErrGraph(ErrInvalidEdge)
	}

	if !t.HasNode(id) {
		return ErrGraph(ErrNodeNotFound)
	}

	descendants, _ := Descendants(t.dag.g, id)

	err := t.dag.RemoveNode(id)
	if err != nil {
		return err
	}

	for _, d := range descendants {
		_ = t.dag.RemoveNode(d)
	}

	return nil
}

// AddEdge adds an edge to the tree, enforcing single-parent constraint.
func (t *Tree) AddEdge(edge Edge) error {
	if len(t.dag.g.in[edge.To]) > 0 {
		return ErrGraph(ErrMultipleParents)
	}

	return t.dag.AddEdge(edge)
}

// RemoveEdge removes an edge and cascade-removes the disconnected subtree.
// In a tree, removing any edge disconnects the child subtree from the root.
func (t *Tree) RemoveEdge(id string) error {
	edge, err := t.dag.Edge(id)
	if err != nil {
		return err
	}

	orphanRoot := edge.To
	descendants, _ := Descendants(t.dag.g, orphanRoot)

	err = t.dag.RemoveEdge(id)
	if err != nil {
		return err
	}

	for _, d := range descendants {
		_ = t.dag.RemoveNode(d)
	}

	_ = t.dag.RemoveNode(orphanRoot)

	return nil
}

// Node returns the node with the given ID.
func (t *Tree) Node(id string) (Node, error) {
	return t.dag.Node(id)
}

// Nodes returns all nodes sorted by ID.
func (t *Tree) Nodes() []Node {
	return t.dag.Nodes()
}

// Edge returns the edge with the given ID.
func (t *Tree) Edge(id string) (Edge, error) {
	return t.dag.Edge(id)
}

// Edges returns all edges sorted by ID.
func (t *Tree) Edges() []Edge {
	return t.dag.Edges()
}

// Neighbors returns the IDs of child nodes, sorted.
func (t *Tree) Neighbors(id string) ([]string, error) {
	return t.dag.Neighbors(id)
}

// HasNode reports whether the tree contains a node with the given ID.
func (t *Tree) HasNode(id string) bool {
	return t.dag.HasNode(id)
}

// HasEdge reports whether the tree contains an edge with the given ID.
func (t *Tree) HasEdge(id string) bool {
	return t.dag.HasEdge(id)
}

// NodeCount returns the number of nodes.
func (t *Tree) NodeCount() int {
	return t.dag.NodeCount()
}

// EdgeCount returns the number of edges.
func (t *Tree) EdgeCount() int {
	return t.dag.EdgeCount()
}

// Degree returns the degree of a node.
func (t *Tree) Degree(id string) (int, error) {
	return t.dag.Degree(id)
}

// Clone returns a deep copy of the tree.
func (t *Tree) Clone() Graph {
	return t.CloneTree()
}

// IsDirected reports whether the graph is directed.
func (t *Tree) IsDirected() bool {
	return true
}

// CloneTree returns a deep copy as *Tree.
func (t *Tree) CloneTree() *Tree {
	return &Tree{dag: t.dag.CloneDAG(), root: t.root}
}

// Parent returns the parent node ID, or empty string for root.
func (t *Tree) Parent(id string) (string, error) {
	if !t.HasNode(id) {
		return "", ErrGraph(ErrNodeNotFound)
	}

	if id == t.root {
		return "", nil
	}

	inEdges := t.dag.g.in[id]
	if len(inEdges) == 0 {
		return "", nil
	}

	edge := t.dag.g.edges[inEdges[0]]

	return edge.From, nil
}

// Children returns the child node IDs, sorted.
func (t *Tree) Children(id string) ([]string, error) {
	return t.dag.Neighbors(id)
}

// IsLeaf reports whether the node has no children.
func (t *Tree) IsLeaf(id string) (bool, error) {
	if !t.HasNode(id) {
		return false, ErrGraph(ErrNodeNotFound)
	}

	return len(t.dag.g.out[id]) == 0, nil
}
