package graph

import (
	"errors"
	"testing"
)

func TestNewUndirected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()

	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestUndirected_AddNode(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	err := g.AddNode(Node{ID: "A"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g.NodeCount())
	}
}

func TestUndirected_AddNode_duplicate(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	err := g.AddNode(Node{ID: "A"})

	if !errors.Is(err, ErrDuplicateNode) {
		t.Fatalf("expected ErrDuplicateNode, got %v", err)
	}
}

func TestUndirected_RemoveNode(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := g.RemoveNode("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestUndirected_RemoveNode_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	err := g.RemoveNode("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestUndirected_AddEdge(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}

	e, _ := g.Edge("e1")

	if e.Weight != 0 {
		t.Fatalf("expected zero weight, got %f", e.Weight)
	}
}

func TestUndirected_AddEdge_missing_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected ErrInvalidEdge, got %v", err)
	}
}

func TestUndirected_AddEdge_duplicate_id(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected ErrInvalidEdge, got %v", err)
	}
}

func TestUndirected_AddEdge_self_loop(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "A"})

	if err != nil {
		t.Fatalf("self-loops should be allowed in undirected graphs: %v", err)
	}
}

func TestUndirected_RemoveEdge(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := g.RemoveEdge("e1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestUndirected_RemoveEdge_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	err := g.RemoveEdge("x")

	if !errors.Is(err, ErrEdgeNotFound) {
		t.Fatalf("expected ErrEdgeNotFound, got %v", err)
	}
}

func TestUndirected_Neighbors(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "B"})

	neighbors, err := g.Neighbors("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}

	if neighbors[0] != "B" || neighbors[1] != "C" {
		t.Fatalf("expected [B C], got %v", neighbors)
	}
}

func TestUndirected_Neighbors_bidirectional(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	neighborsA, _ := g.Neighbors("A")
	neighborsB, _ := g.Neighbors("B")

	if len(neighborsA) != 1 || neighborsA[0] != "B" {
		t.Fatalf("expected A's neighbor to be B, got %v", neighborsA)
	}

	if len(neighborsB) != 1 || neighborsB[0] != "A" {
		t.Fatalf("expected B's neighbor to be A, got %v", neighborsB)
	}
}

func TestUndirected_Neighbors_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_, err := g.Neighbors("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestUndirected_Degree(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	deg, err := g.Degree("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deg != 2 {
		t.Fatalf("expected degree 2, got %d", deg)
	}
}

func TestUndirected_Degree_self_loop(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "A"})

	deg, _ := g.Degree("A")

	if deg != 2 {
		t.Fatalf("expected degree 2 for self-loop, got %d", deg)
	}
}

func TestUndirected_Degree_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_, err := g.Degree("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestUndirected_Clone(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	clone := g.Clone()

	if clone.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", clone.NodeCount())
	}

	_ = g.AddNode(Node{ID: "C"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected by original changes")
	}
}

func TestUndirected_CloneUndirected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	clone := g.CloneUndirected()

	if clone.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", clone.NodeCount())
	}
}

func TestUndirected_Nodes_sorted(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	nodes := g.Nodes()

	if nodes[0].ID != "A" || nodes[1].ID != "B" || nodes[2].ID != "C" {
		t.Fatalf("nodes not sorted: %v", nodes)
	}
}

func TestUndirected_Edges_sorted(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	edges := g.Edges()

	if edges[0].ID != "e1" || edges[1].ID != "e2" {
		t.Fatalf("edges not sorted: %v", edges)
	}
}

func TestUndirected_Node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A", Metadata: "data"})

	n, err := g.Node("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n.Metadata != "data" {
		t.Fatalf("expected metadata %q, got %v", "data", n.Metadata)
	}
}

func TestUndirected_Node_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_, err := g.Node("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestUndirected_Edge_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_, err := g.Edge("X")

	if !errors.Is(err, ErrEdgeNotFound) {
		t.Fatalf("expected ErrEdgeNotFound, got %v", err)
	}
}

func TestUndirected_HasNode(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	if !g.HasNode("A") {
		t.Fatal("expected HasNode to return true")
	}

	if g.HasNode("X") {
		t.Fatal("expected HasNode to return false")
	}
}

func TestUndirected_HasEdge(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if !g.HasEdge("e1") {
		t.Fatal("expected HasEdge to return true")
	}

	if g.HasEdge("x") {
		t.Fatal("expected HasEdge to return false")
	}
}

func TestUndirected_IncidentEdges(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	edges, err := g.IncidentEdges("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(edges) != 2 {
		t.Fatalf("expected 2 incident edges, got %d", len(edges))
	}

	if edges[0] != "e1" || edges[1] != "e2" {
		t.Fatalf("expected [e1 e2], got %v", edges)
	}
}

func TestUndirected_IncidentEdges_not_found(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_, err := g.IncidentEdges("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestUndirected_interface_compliance(t *testing.T) {
	t.Parallel()

	var _ Graph = NewUndirected()
}

func TestUndirected_RemoveNode_self_loop(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "A"})

	err := g.RemoveNode("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestUndirected_Neighbors_self_loop(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "A"})

	neighbors, _ := g.Neighbors("A")

	if len(neighbors) != 1 || neighbors[0] != "A" {
		t.Fatalf("expected [A], got %v", neighbors)
	}
}
