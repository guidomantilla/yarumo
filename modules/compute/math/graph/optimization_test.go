package graph

import (
	"errors"
	"testing"
)

func TestMinimumSpanningTree(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 3})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D", Weight: 4})

	mst, err := MinimumSpanningTree(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mst) != 3 {
		t.Fatalf("expected 3 MST edges, got %d", len(mst))
	}
}

func TestMinimumSpanningTree_disconnected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	_, err := MinimumSpanningTree(g)
	if !errors.Is(err, ErrDisconnected) {
		t.Fatalf("expected ErrDisconnected, got %v", err)
	}
}

func TestMinimumSpanningTree_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()

	mst, err := MinimumSpanningTree(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mst != nil {
		t.Fatalf("expected nil, got %v", mst)
	}
}

func TestMinimumSpanningTree_single_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	mst, err := MinimumSpanningTree(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mst) != 0 {
		t.Fatalf("expected 0 edges, got %d", len(mst))
	}
}

func TestMaxFlow(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "S"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "T"})
	_ = g.AddEdge(Edge{ID: "e1", From: "S", To: "A", Weight: 10})
	_ = g.AddEdge(Edge{ID: "e2", From: "S", To: "B", Weight: 5})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "B", Weight: 15})
	_ = g.AddEdge(Edge{ID: "e4", From: "A", To: "T", Weight: 10})
	_ = g.AddEdge(Edge{ID: "e5", From: "B", To: "T", Weight: 10})

	flow, err := MaxFlow(g, "S", "T")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flow != 15 {
		t.Fatalf("expected max flow 15, got %f", flow)
	}
}

func TestMaxFlow_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := MaxFlow(g, "X", "Y")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestMaxFlow_no_path(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "S"})
	_ = g.AddNode(Node{ID: "T"})

	flow, err := MaxFlow(g, "S", "T")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flow != 0 {
		t.Fatalf("expected 0, got %f", flow)
	}
}

func TestMinCut(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "S"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "T"})
	_ = g.AddEdge(Edge{ID: "e1", From: "S", To: "A", Weight: 3})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "T", Weight: 2})

	cut, err := MinCut(g, "S", "T")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cut) != 1 || cut[0] != "e2" {
		t.Fatalf("expected [e2], got %v", cut)
	}
}

func TestMinCut_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := MinCut(g, "X", "Y")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestBipartiteMatching(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "L1"})
	_ = b.AddNodeLeft(Node{ID: "L2"})
	_ = b.AddNodeLeft(Node{ID: "L3"})
	_ = b.AddNodeRight(Node{ID: "R1"})
	_ = b.AddNodeRight(Node{ID: "R2"})
	_ = b.AddNodeRight(Node{ID: "R3"})
	_ = b.AddEdge(Edge{ID: "e1", From: "L1", To: "R1"})
	_ = b.AddEdge(Edge{ID: "e2", From: "L1", To: "R2"})
	_ = b.AddEdge(Edge{ID: "e3", From: "L2", To: "R1"})
	_ = b.AddEdge(Edge{ID: "e4", From: "L2", To: "R3"})
	_ = b.AddEdge(Edge{ID: "e5", From: "L3", To: "R2"})

	matching := BipartiteMatching(b)
	if len(matching) != 3 {
		t.Fatalf("expected maximum matching of size 3, got %d: %v", len(matching), matching)
	}
}

func TestBipartiteMatching_empty(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	matching := BipartiteMatching(b)
	if len(matching) != 0 {
		t.Fatalf("expected 0, got %d", len(matching))
	}
}

func TestBipartiteMatching_no_edges(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "L1"})
	_ = b.AddNodeRight(Node{ID: "R1"})

	matching := BipartiteMatching(b)
	if len(matching) != 0 {
		t.Fatalf("expected 0, got %d", len(matching))
	}
}

func TestBipartiteMatching_augmenting_paths(t *testing.T) {
	t.Parallel()

	// K_{3,3}-like graph forces multi-layer BFS and DFS failure paths in Hopcroft-Karp.
	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "L1"})
	_ = b.AddNodeLeft(Node{ID: "L2"})
	_ = b.AddNodeLeft(Node{ID: "L3"})
	_ = b.AddNodeLeft(Node{ID: "L4"})
	_ = b.AddNodeRight(Node{ID: "R1"})
	_ = b.AddNodeRight(Node{ID: "R2"})
	_ = b.AddNodeRight(Node{ID: "R3"})
	// L1→R1, L2→R1, L3→R2, L4→R3 (L1 and L2 compete for R1)
	_ = b.AddEdge(Edge{ID: "e1", From: "L1", To: "R1"})
	_ = b.AddEdge(Edge{ID: "e2", From: "L2", To: "R1"})
	_ = b.AddEdge(Edge{ID: "e3", From: "L2", To: "R2"})
	_ = b.AddEdge(Edge{ID: "e4", From: "L3", To: "R2"})
	_ = b.AddEdge(Edge{ID: "e5", From: "L3", To: "R3"})
	_ = b.AddEdge(Edge{ID: "e6", From: "L4", To: "R3"})

	matching := BipartiteMatching(b)
	if len(matching) != 3 {
		t.Fatalf("expected maximum matching of size 3, got %d: %v", len(matching), matching)
	}
}

func TestBipartiteMatching_partial(t *testing.T) {
	t.Parallel()

	// Two left nodes compete for same right node, only one can match.
	// Forces hopcroftDFS to return false for one node.
	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "L1"})
	_ = b.AddNodeLeft(Node{ID: "L2"})
	_ = b.AddNodeRight(Node{ID: "R1"})
	_ = b.AddEdge(Edge{ID: "e1", From: "L1", To: "R1"})
	_ = b.AddEdge(Edge{ID: "e2", From: "L2", To: "R1"})

	matching := BipartiteMatching(b)
	if len(matching) != 1 {
		t.Fatalf("expected maximum matching of size 1, got %d: %v", len(matching), matching)
	}
}

func TestMinimumSpanningTree_selects_cheapest(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 5})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "B", Weight: 1})

	mst, _ := MinimumSpanningTree(g)
	if len(mst) != 1 || mst[0] != "e2" {
		t.Fatalf("expected cheapest edge [e2], got %v", mst)
	}
}
