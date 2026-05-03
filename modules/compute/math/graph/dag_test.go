package graph

import (
	"errors"
	"testing"
)

func TestNewDAG(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	if d.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", d.NodeCount())
	}
}

func TestNewDAGFrom(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	d, err := NewDAGFrom(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", d.NodeCount())
	}
}

func TestNewDAGFrom_cycle(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "A"})

	_, err := NewDAGFrom(g)
	if !errors.Is(err, ErrNotDAG) {
		t.Fatalf("expected ErrNotDAG, got %v", err)
	}
}

func TestDAG_AddEdge_self_loop(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})

	err := d.AddEdge(Edge{ID: "e1", From: "A", To: "A"})
	if !errors.Is(err, ErrSelfLoop) {
		t.Fatalf("expected ErrSelfLoop, got %v", err)
	}
}

func TestDAG_AddEdge_cycle(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddNode(Node{ID: "C"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = d.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	err := d.AddEdge(Edge{ID: "e3", From: "C", To: "A"})
	if !errors.Is(err, ErrCycleDetected) {
		t.Fatalf("expected ErrCycleDetected, got %v", err)
	}

	if d.EdgeCount() != 2 {
		t.Fatalf("cycle-causing edge should have been removed, got %d edges", d.EdgeCount())
	}
}

func TestDAG_AddEdge_invalid(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})

	err := d.AddEdge(Edge{ID: "e1", From: "A", To: "X"})
	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected ErrInvalidEdge, got %v", err)
	}
}

func TestDAG_Directed(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})

	if d.Directed() == nil {
		t.Fatal("expected non-nil Directed")
	}
}

func TestDAG_Roots(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddNode(Node{ID: "C"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = d.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	r := d.Roots()
	if len(r) != 1 || r[0] != "A" {
		t.Fatalf("expected [A], got %v", r)
	}
}

func TestDAG_Leaves(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddNode(Node{ID: "C"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = d.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	l := d.Leaves()
	if len(l) != 2 || l[0] != "B" || l[1] != "C" {
		t.Fatalf("expected [B C], got %v", l)
	}
}

func TestDAG_Clone(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	clone := d.CloneDAG()
	_ = d.AddNode(Node{ID: "C"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected by original changes")
	}
}

func TestDAG_Clone_interface(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})

	_ = d.Clone()
}

func TestDAG_delegated_methods(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	t.Run("Node", func(t *testing.T) {
		t.Parallel()
		n, err := d.Node("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.ID != "A" {
			t.Fatalf("expected A, got %s", n.ID)
		}
	})

	t.Run("Nodes", func(t *testing.T) {
		t.Parallel()
		nodes := d.Nodes()
		if len(nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(nodes))
		}
	})

	t.Run("Edge", func(t *testing.T) {
		t.Parallel()
		e, err := d.Edge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.From != "A" {
			t.Fatalf("expected from A, got %s", e.From)
		}
	})

	t.Run("Edges", func(t *testing.T) {
		t.Parallel()
		edges := d.Edges()
		if len(edges) != 1 {
			t.Fatalf("expected 1 edge, got %d", len(edges))
		}
	})

	t.Run("Neighbors", func(t *testing.T) {
		t.Parallel()
		neighbors, err := d.Neighbors("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(neighbors) != 1 || neighbors[0] != "B" {
			t.Fatalf("expected [B], got %v", neighbors)
		}
	})

	t.Run("HasNode", func(t *testing.T) {
		t.Parallel()
		if !d.HasNode("A") {
			t.Fatal("expected HasNode to return true")
		}
	})

	t.Run("HasEdge", func(t *testing.T) {
		t.Parallel()
		if !d.HasEdge("e1") {
			t.Fatal("expected HasEdge to return true")
		}
	})

	t.Run("NodeCount", func(t *testing.T) {
		t.Parallel()
		if d.NodeCount() != 2 {
			t.Fatalf("expected 2, got %d", d.NodeCount())
		}
	})

	t.Run("EdgeCount", func(t *testing.T) {
		t.Parallel()
		if d.EdgeCount() != 1 {
			t.Fatalf("expected 1, got %d", d.EdgeCount())
		}
	})

	t.Run("Degree", func(t *testing.T) {
		t.Parallel()
		deg, err := d.Degree("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deg != 1 {
			t.Fatalf("expected 1, got %d", deg)
		}
	})
}

func TestDAG_RemoveNode(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := d.RemoveNode("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", d.NodeCount())
	}
}

func TestDAG_RemoveEdge(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := d.RemoveEdge("e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", d.EdgeCount())
	}
}

func TestDAG_interface_compliance(t *testing.T) {
	t.Parallel()

	var _ Graph = NewDAG()
}
