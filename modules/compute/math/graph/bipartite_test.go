package graph

import (
	"errors"
	"testing"
)

func TestNewBipartite(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	if b.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", b.NodeCount())
	}
}

func TestNewBipartiteFrom(t *testing.T) {
	t.Parallel()

	u := NewUndirected()
	_ = u.AddNode(Node{ID: "A"})
	_ = u.AddNode(Node{ID: "B"})
	_ = u.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	b, err := NewBipartiteFrom(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", b.NodeCount())
	}
}

func TestNewBipartiteFrom_not_bipartite(t *testing.T) {
	t.Parallel()

	u := NewUndirected()
	_ = u.AddNode(Node{ID: "A"})
	_ = u.AddNode(Node{ID: "B"})
	_ = u.AddNode(Node{ID: "C"})
	_ = u.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = u.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = u.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	_, err := NewBipartiteFrom(u)
	if !errors.Is(err, ErrNotBipartite) {
		t.Fatalf("expected ErrNotBipartite, got %v", err)
	}
}

func TestBipartite_AddNodeLeft(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	err := b.AddNodeLeft(Node{ID: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	left := b.Left()
	if len(left) != 1 || left[0] != "A" {
		t.Fatalf("expected [A], got %v", left)
	}
}

func TestBipartite_AddNodeRight(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	err := b.AddNodeRight(Node{ID: "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	right := b.Right()
	if len(right) != 1 || right[0] != "B" {
		t.Fatalf("expected [B], got %v", right)
	}
}

func TestNewBipartiteFrom_path_graph(t *testing.T) {
	t.Parallel()

	// A-B-C path: BFS starts at A(left), colors B(right), then B colors C(left).
	// This exercises the else branch in checkBipartite where color[curr]==2.
	u := NewUndirected()
	_ = u.AddNode(Node{ID: "A"})
	_ = u.AddNode(Node{ID: "B"})
	_ = u.AddNode(Node{ID: "C"})
	_ = u.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = u.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	b, err := NewBipartiteFrom(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", b.NodeCount())
	}

	left := b.Left()
	right := b.Right()
	if len(left) != 2 {
		t.Fatalf("expected 2 left nodes, got %d: %v", len(left), left)
	}
	if len(right) != 1 {
		t.Fatalf("expected 1 right node, got %d: %v", len(right), right)
	}
}

func TestBipartite_AddNodeLeft_already_right(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeRight(Node{ID: "A"})

	err := b.AddNodeLeft(Node{ID: "A"})
	if err == nil {
		t.Fatal("expected error adding right-partition node to left")
	}
}

func TestBipartite_AddNodeRight_already_left(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})

	err := b.AddNodeRight(Node{ID: "A"})
	if err == nil {
		t.Fatal("expected error adding left-partition node to right")
	}
}

func TestBipartite_AddEdge_same_partition(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	_ = b.AddNodeLeft(Node{ID: "B"})

	err := b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	if !errors.Is(err, ErrNotBipartite) {
		t.Fatalf("expected ErrNotBipartite, got %v", err)
	}
}

func TestBipartite_AddEdge_different_partitions(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	_ = b.AddNodeRight(Node{ID: "B"})

	err := b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBipartite_RemoveNode(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	_ = b.AddNodeRight(Node{ID: "B"})
	_ = b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	err := b.RemoveNode("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	left := b.Left()
	if len(left) != 0 {
		t.Fatalf("expected empty left partition, got %v", left)
	}
}

func TestBipartite_RemoveNode_not_found(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	err := b.RemoveNode("X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestBipartite_Undirected(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	if b.Undirected() == nil {
		t.Fatal("expected non-nil Undirected")
	}
}

func TestBipartite_Clone(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	_ = b.AddNodeRight(Node{ID: "B"})
	_ = b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	clone := b.CloneBipartite()
	_ = b.AddNodeLeft(Node{ID: "C"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected by original changes")
	}

	if len(clone.Left()) != 1 || clone.Left()[0] != "A" {
		t.Fatalf("expected left [A], got %v", clone.Left())
	}
}

func TestBipartite_Clone_interface(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.Clone()
}

func TestBipartite_AddNode_defaults_left(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNode(Node{ID: "A"})

	left := b.Left()
	if len(left) != 1 || left[0] != "A" {
		t.Fatalf("expected AddNode to default to left, got left=%v", left)
	}
}

func TestBipartite_delegated_methods(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	_ = b.AddNodeRight(Node{ID: "B"})
	_ = b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	t.Run("Node", func(t *testing.T) {
		t.Parallel()
		n, err := b.Node("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.ID != "A" {
			t.Fatalf("expected A, got %s", n.ID)
		}
	})

	t.Run("Nodes", func(t *testing.T) {
		t.Parallel()
		if len(b.Nodes()) != 2 {
			t.Fatalf("expected 2, got %d", len(b.Nodes()))
		}
	})

	t.Run("Edge", func(t *testing.T) {
		t.Parallel()
		e, err := b.Edge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.ID != "e1" {
			t.Fatalf("expected e1, got %s", e.ID)
		}
	})

	t.Run("Edges", func(t *testing.T) {
		t.Parallel()
		if len(b.Edges()) != 1 {
			t.Fatalf("expected 1, got %d", len(b.Edges()))
		}
	})

	t.Run("Neighbors", func(t *testing.T) {
		t.Parallel()
		nb, _ := b.Neighbors("A")
		if len(nb) != 1 || nb[0] != "B" {
			t.Fatalf("expected [B], got %v", nb)
		}
	})

	t.Run("HasNode_HasEdge", func(t *testing.T) {
		t.Parallel()
		if !b.HasNode("A") {
			t.Fatal("expected true")
		}
		if !b.HasEdge("e1") {
			t.Fatal("expected true")
		}
	})

	t.Run("Counts", func(t *testing.T) {
		t.Parallel()
		if b.NodeCount() != 2 {
			t.Fatalf("expected 2, got %d", b.NodeCount())
		}
		if b.EdgeCount() != 1 {
			t.Fatalf("expected 1, got %d", b.EdgeCount())
		}
	})

	t.Run("Degree", func(t *testing.T) {
		t.Parallel()
		deg, _ := b.Degree("A")
		if deg != 1 {
			t.Fatalf("expected 1, got %d", deg)
		}
	})

	t.Run("RemoveEdge", func(t *testing.T) {
		t.Parallel()
		clone := b.CloneBipartite()
		err := clone.RemoveEdge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestBipartite_AddNodeLeft_duplicate(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeLeft(Node{ID: "A"})
	err := b.AddNodeLeft(Node{ID: "A"})
	if !errors.Is(err, ErrDuplicateNode) {
		t.Fatalf("expected ErrDuplicateNode, got %v", err)
	}
}

func TestBipartite_AddNodeRight_duplicate(t *testing.T) {
	t.Parallel()

	b := NewBipartite()
	_ = b.AddNodeRight(Node{ID: "A"})
	err := b.AddNodeRight(Node{ID: "A"})
	if !errors.Is(err, ErrDuplicateNode) {
		t.Fatalf("expected ErrDuplicateNode, got %v", err)
	}
}

func TestNewBipartiteFrom_disconnected(t *testing.T) {
	t.Parallel()

	u := NewUndirected()
	_ = u.AddNode(Node{ID: "A"})
	_ = u.AddNode(Node{ID: "B"})
	_ = u.AddNode(Node{ID: "C"})
	_ = u.AddNode(Node{ID: "D"})
	_ = u.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = u.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	b, err := NewBipartiteFrom(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.NodeCount() != 4 {
		t.Fatalf("expected 4 nodes, got %d", b.NodeCount())
	}
}
