package graph

import (
	"errors"
	"testing"
)

func TestIsDAG_true(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if !IsDAG(g) {
		t.Fatal("expected IsDAG to return true")
	}
}

func TestIsDAG_false(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "A"})

	if IsDAG(g) {
		t.Fatal("expected IsDAG to return false")
	}
}

func TestIsTree_true(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "R"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = g.AddEdge(Edge{ID: "e2", From: "R", To: "B"})

	if !IsTree(g) {
		t.Fatal("expected IsTree to return true")
	}
}

func TestIsTree_multiple_roots(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	if IsTree(g) {
		t.Fatal("expected IsTree to return false for multiple roots")
	}
}

func TestIsTree_multiple_parents(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "R"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = g.AddEdge(Edge{ID: "e2", From: "R", To: "B"})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e4", From: "B", To: "C"})

	if IsTree(g) {
		t.Fatal("expected IsTree to return false for multiple parents")
	}
}

func TestIsBipartite_true(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	if !IsBipartite(g) {
		t.Fatal("expected IsBipartite to return true")
	}
}

func TestIsBipartite_false(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	if IsBipartite(g) {
		t.Fatal("expected IsBipartite to return false")
	}
}

func TestInDegree(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	deg, err := InDegree(g, "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deg != 2 {
		t.Fatalf("expected in-degree 2, got %d", deg)
	}
}

func TestInDegree_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := InDegree(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestOutDegree(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	deg, err := OutDegree(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deg != 2 {
		t.Fatalf("expected out-degree 2, got %d", deg)
	}
}

func TestOutDegree_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := OutDegree(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestIsDAG_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	if !IsDAG(g) {
		t.Fatal("empty graph should be a DAG")
	}
}

func TestIsTree_single_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "R"})

	if !IsTree(g) {
		t.Fatal("single node should be a tree")
	}
}

func TestIsBipartite_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	if !IsBipartite(g) {
		t.Fatal("empty graph should be bipartite")
	}
}
