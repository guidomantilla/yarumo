// Package graph provides graph data structures and algorithms.
package graph

// Node represents a vertex in a graph.
type Node struct {
	ID       string
	Metadata any
}

// Edge represents a connection between two nodes in a graph.
type Edge struct {
	ID       string
	From     string
	To       string
	Weight   float64
	Label    string
	Metadata any
}

// Path represents a sequence of nodes and edges forming a walk through a graph.
type Path struct {
	Nodes  []string
	Edges  []string
	Weight float64
}

// Matrix represents an adjacency matrix for a graph.
type Matrix struct {
	Indices map[string]int
	Data    [][]float64
}

// Graph defines the common interface for all graph types.
type Graph interface {
	// AddNode adds a node to the graph.
	AddNode(node Node) error
	// RemoveNode removes a node and all its incident edges.
	RemoveNode(id string) error
	// AddEdge adds an edge to the graph.
	AddEdge(edge Edge) error
	// RemoveEdge removes an edge by its ID.
	RemoveEdge(id string) error
	// Node returns the node with the given ID.
	Node(id string) (Node, error)
	// Nodes returns all nodes sorted by ID.
	Nodes() []Node
	// Edge returns the edge with the given ID.
	Edge(id string) (Edge, error)
	// Edges returns all edges sorted by ID.
	Edges() []Edge
	// Neighbors returns the IDs of adjacent nodes sorted alphabetically.
	Neighbors(id string) ([]string, error)
	// HasNode reports whether the graph contains a node with the given ID.
	HasNode(id string) bool
	// HasEdge reports whether the graph contains an edge with the given ID.
	HasEdge(id string) bool
	// NodeCount returns the number of nodes.
	NodeCount() int
	// EdgeCount returns the number of edges.
	EdgeCount() int
	// Degree returns the degree of a node.
	Degree(id string) (int, error)
	// Clone returns a deep copy of the graph.
	Clone() Graph
	// IsDirected reports whether the graph is directed.
	IsDirected() bool
}

// Type compliance.
var (
	_ Graph = (*Directed)(nil)
	_ Graph = (*Undirected)(nil)
	_ Graph = (*DAG)(nil)
	_ Graph = (*Tree)(nil)
	_ Graph = (*Bipartite)(nil)
	_ Graph = (*MultigraphDirected)(nil)
	_ Graph = (*MultigraphUndirected)(nil)
)
