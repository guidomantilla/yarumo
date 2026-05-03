package graph

import (
	"errors"
	"testing"
)

func TestBFS_directed(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "B", To: "D"})

	result, err := BFSAll(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(result))
	}
	if result[0] != "A" {
		t.Fatalf("expected first node to be A, got %s", result[0])
	}
}

func TestBFS_undirected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	result, err := BFSAll(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result))
	}
}

func TestBFS_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := BFSAll(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestBFS_single_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	result, _ := BFSAll(g, "A")
	if len(result) != 1 || result[0] != "A" {
		t.Fatalf("expected [A], got %v", result)
	}
}

func TestDFS_directed(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "B", To: "D"})

	result, err := DFSAll(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(result))
	}
	if result[0] != "A" {
		t.Fatalf("expected first node to be A, got %s", result[0])
	}
}

func TestDFS_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := DFSAll(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDFS_single_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	result, _ := DFSAll(g, "A")
	if len(result) != 1 || result[0] != "A" {
		t.Fatalf("expected [A], got %v", result)
	}
}

func TestBFS_callback(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	count := 0
	_ = BFS(g, "A", func(_ string) {
		count++
	})
	if count != 2 {
		t.Fatalf("expected 2 visits, got %d", count)
	}
}

func TestDFS_callback(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	count := 0
	_ = DFS(g, "A", func(_ string) {
		count++
	})
	if count != 2 {
		t.Fatalf("expected 2 visits, got %d", count)
	}
}

func TestBFS_undirected_cycle(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	result, _ := BFSAll(g, "A")
	if len(result) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result))
	}
}

func TestDFS_undirected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	result, _ := DFSAll(g, "A")
	if len(result) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result))
	}
}
