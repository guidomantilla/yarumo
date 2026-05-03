package graph

import (
	"errors"
	"testing"
)

func TestNewDirected(t *testing.T) {
	t.Parallel()

	g := NewDirected()

	if g.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestDirected_AddNode(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	err := g.AddNode(Node{ID: "A"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", g.NodeCount())
	}
}

func TestDirected_AddNode_duplicate(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	err := g.AddNode(Node{ID: "A"})

	if !errors.Is(err, ErrDuplicateNode) {
		t.Fatalf("expected ErrDuplicateNode, got %v", err)
	}
}

func TestDirected_RemoveNode(t *testing.T) {
	t.Parallel()

	g := NewDirected()
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

func TestDirected_RemoveNode_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	err := g.RemoveNode("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_AddEdge(t *testing.T) {
	t.Parallel()

	g := NewDirected()
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

func TestDirected_AddEdge_custom_weight(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 2.5})

	e, _ := g.Edge("e1")

	if e.Weight != 2.5 {
		t.Fatalf("expected weight 2.5, got %f", e.Weight)
	}
}

func TestDirected_AddEdge_missing_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected ErrInvalidEdge, got %v", err)
	}
}

func TestDirected_AddEdge_duplicate_id(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := g.AddEdge(Edge{ID: "e1", From: "B", To: "A"})

	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected ErrInvalidEdge, got %v", err)
	}
}

func TestDirected_AddEdge_self_loop(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	err := g.AddEdge(Edge{ID: "e1", From: "A", To: "A"})

	if err != nil {
		t.Fatalf("self-loops should be allowed in directed graphs: %v", err)
	}
}

func TestDirected_RemoveEdge(t *testing.T) {
	t.Parallel()

	g := NewDirected()
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

func TestDirected_RemoveEdge_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	err := g.RemoveEdge("x")

	if !errors.Is(err, ErrEdgeNotFound) {
		t.Fatalf("expected ErrEdgeNotFound, got %v", err)
	}
}

func TestDirected_Node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A", Metadata: "data"})

	n, err := g.Node("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n.Metadata != "data" {
		t.Fatalf("expected metadata %q, got %v", "data", n.Metadata)
	}
}

func TestDirected_Node_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.Node("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_Nodes_sorted(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	nodes := g.Nodes()

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	if nodes[0].ID != "A" || nodes[1].ID != "B" || nodes[2].ID != "C" {
		t.Fatalf("nodes not sorted: %v", nodes)
	}
}

func TestDirected_Edge_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.Edge("X")

	if !errors.Is(err, ErrEdgeNotFound) {
		t.Fatalf("expected ErrEdgeNotFound, got %v", err)
	}
}

func TestDirected_Edges_sorted(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	edges := g.Edges()

	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}

	if edges[0].ID != "e1" || edges[1].ID != "e2" {
		t.Fatalf("edges not sorted: %v", edges)
	}
}

func TestDirected_Neighbors(t *testing.T) {
	t.Parallel()

	g := NewDirected()
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

func TestDirected_Neighbors_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.Neighbors("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_HasNode(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	if !g.HasNode("A") {
		t.Fatal("expected HasNode to return true")
	}

	if g.HasNode("X") {
		t.Fatal("expected HasNode to return false")
	}
}

func TestDirected_HasEdge(t *testing.T) {
	t.Parallel()

	g := NewDirected()
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

func TestDirected_Degree(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "A"})

	deg, err := g.Degree("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if deg != 2 {
		t.Fatalf("expected degree 2, got %d", deg)
	}
}

func TestDirected_Degree_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.Degree("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_Clone(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	clone := g.Clone()

	if clone.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", clone.NodeCount())
	}

	if clone.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", clone.EdgeCount())
	}

	_ = g.AddNode(Node{ID: "C"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected by original changes")
	}
}

func TestDirected_CloneDirected(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	clone := g.CloneDirected()

	if clone.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", clone.NodeCount())
	}
}

func TestDirected_InEdges(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "B"})

	edges, err := g.InEdges("B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(edges) != 2 {
		t.Fatalf("expected 2 in-edges, got %d", len(edges))
	}

	if edges[0] != "e1" || edges[1] != "e2" {
		t.Fatalf("expected [e1 e2], got %v", edges)
	}
}

func TestDirected_InEdges_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.InEdges("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_OutEdges(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	edges, err := g.OutEdges("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(edges) != 2 {
		t.Fatalf("expected 2 out-edges, got %d", len(edges))
	}
}

func TestDirected_OutEdges_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.OutEdges("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_Predecessors(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	preds, err := g.Predecessors("C")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(preds) != 2 {
		t.Fatalf("expected 2 predecessors, got %d", len(preds))
	}

	if preds[0] != "A" || preds[1] != "B" {
		t.Fatalf("expected [A B], got %v", preds)
	}
}

func TestDirected_Predecessors_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := g.Predecessors("X")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDirected_Successors(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	succs, err := g.Successors("A")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(succs) != 2 {
		t.Fatalf("expected 2 successors, got %d", len(succs))
	}

	if succs[0] != "B" || succs[1] != "C" {
		t.Fatalf("expected [B C], got %v", succs)
	}
}

func TestDirected_RemoveNode_cleans_incoming_edges(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	_ = g.RemoveNode("B")

	if g.HasEdge("e1") {
		t.Fatal("edge e1 should have been removed")
	}

	if g.HasEdge("e2") {
		t.Fatal("edge e2 should have been removed")
	}
}

func TestDirected_Neighbors_deduplicates(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "B"})

	neighbors, _ := g.Neighbors("A")

	if len(neighbors) != 1 {
		t.Fatalf("expected 1 unique neighbor, got %d", len(neighbors))
	}
}

func TestDirected_RemoveEdge_preserves_others(t *testing.T) {
	t.Parallel()

	// A→B, A→C: removing e1 should keep e2 in out[A].
	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	err := g.RemoveEdge("e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}
	if !g.HasEdge("e2") {
		t.Fatal("expected e2 to still exist")
	}
}

func TestDirected_interface_compliance(t *testing.T) {
	t.Parallel()

	var _ Graph = NewDirected()
}
