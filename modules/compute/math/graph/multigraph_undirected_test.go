package graph

import (
	"errors"
	"testing"
)

func TestNewMultigraphUndirected(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	if m.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", m.NodeCount())
	}
}

func TestMultigraphUndirected_parallel_edges(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})
	_ = m.AddNode(Node{ID: "B"})
	_ = m.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = m.AddEdge(Edge{ID: "e2", From: "A", To: "B", Weight: 2.0})

	if m.EdgeCount() != 2 {
		t.Fatalf("expected 2 edges, got %d", m.EdgeCount())
	}
}

func TestMultigraphUndirected_EdgesBetween(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})
	_ = m.AddNode(Node{ID: "B"})
	_ = m.AddNode(Node{ID: "C"})
	_ = m.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = m.AddEdge(Edge{ID: "e2", From: "A", To: "B"})
	_ = m.AddEdge(Edge{ID: "e3", From: "A", To: "C"})

	edges, err := m.EdgesBetween("A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges between A and B, got %d", len(edges))
	}
}

func TestMultigraphUndirected_EdgesBetween_reverse(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})
	_ = m.AddNode(Node{ID: "B"})
	_ = m.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	edges, _ := m.EdgesBetween("B", "A")
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge between B and A, got %d", len(edges))
	}
}

func TestMultigraphUndirected_EdgesBetween_not_found(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_, err := m.EdgesBetween("A", "B")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestMultigraphUndirected_Undirected(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	if m.Undirected() == nil {
		t.Fatal("expected non-nil Undirected")
	}
}

func TestMultigraphUndirected_Clone(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})
	_ = m.AddNode(Node{ID: "B"})
	_ = m.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	clone := m.CloneMultigraphUndirected()
	_ = m.AddNode(Node{ID: "C"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected")
	}
}

func TestMultigraphUndirected_Clone_interface(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.Clone()
}

func TestMultigraphUndirected_delegated_methods(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})
	_ = m.AddNode(Node{ID: "B"})
	_ = m.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	t.Run("Node", func(t *testing.T) {
		t.Parallel()
		n, _ := m.Node("A")
		if n.ID != "A" {
			t.Fatalf("expected A, got %s", n.ID)
		}
	})

	t.Run("Nodes", func(t *testing.T) {
		t.Parallel()
		if len(m.Nodes()) != 2 {
			t.Fatalf("expected 2, got %d", len(m.Nodes()))
		}
	})

	t.Run("Edge", func(t *testing.T) {
		t.Parallel()
		e, _ := m.Edge("e1")
		if e.ID != "e1" {
			t.Fatalf("expected e1, got %s", e.ID)
		}
	})

	t.Run("Edges", func(t *testing.T) {
		t.Parallel()
		if len(m.Edges()) != 1 {
			t.Fatalf("expected 1, got %d", len(m.Edges()))
		}
	})

	t.Run("Neighbors", func(t *testing.T) {
		t.Parallel()
		nb, _ := m.Neighbors("A")
		if len(nb) != 1 {
			t.Fatalf("expected 1, got %d", len(nb))
		}
	})

	t.Run("HasNode_HasEdge", func(t *testing.T) {
		t.Parallel()
		if !m.HasNode("A") {
			t.Fatal("expected true")
		}
		if !m.HasEdge("e1") {
			t.Fatal("expected true")
		}
	})

	t.Run("Counts", func(t *testing.T) {
		t.Parallel()
		if m.NodeCount() != 2 {
			t.Fatalf("expected 2, got %d", m.NodeCount())
		}
		if m.EdgeCount() != 1 {
			t.Fatalf("expected 1, got %d", m.EdgeCount())
		}
	})

	t.Run("Degree", func(t *testing.T) {
		t.Parallel()
		deg, _ := m.Degree("A")
		if deg != 1 {
			t.Fatalf("expected 1, got %d", deg)
		}
	})

	t.Run("RemoveEdge", func(t *testing.T) {
		t.Parallel()
		clone := m.CloneMultigraphUndirected()
		err := clone.RemoveEdge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("RemoveNode", func(t *testing.T) {
		t.Parallel()
		clone := m.CloneMultigraphUndirected()
		err := clone.RemoveNode("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMultigraphUndirected_self_loop(t *testing.T) {
	t.Parallel()

	m := NewMultigraphUndirected()
	_ = m.AddNode(Node{ID: "A"})

	err := m.AddEdge(Edge{ID: "e1", From: "A", To: "A"})
	if err != nil {
		t.Fatalf("self-loops should be allowed: %v", err)
	}
}
