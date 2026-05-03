package graph

import (
	"errors"
	"testing"
)

func TestNewTree(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "root"})
	if tr.Root() != "root" {
		t.Fatalf("expected root, got %s", tr.Root())
	}
	if tr.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", tr.NodeCount())
	}
}

func TestNewTreeFrom(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "R"})
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = d.AddEdge(Edge{ID: "e2", From: "R", To: "B"})

	tr, err := NewTreeFrom(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Root() != "R" {
		t.Fatalf("expected R, got %s", tr.Root())
	}
}

func TestNewTreeFrom_multiple_roots(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})

	_, err := NewTreeFrom(d)
	if !errors.Is(err, ErrNotTree) {
		t.Fatalf("expected ErrNotTree, got %v", err)
	}
}

func TestNewTreeFrom_multiple_parents(t *testing.T) {
	t.Parallel()

	d := NewDAG()
	_ = d.AddNode(Node{ID: "R"})
	_ = d.AddNode(Node{ID: "A"})
	_ = d.AddNode(Node{ID: "B"})
	_ = d.AddNode(Node{ID: "C"})
	_ = d.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = d.AddEdge(Edge{ID: "e2", From: "R", To: "B"})
	_ = d.AddEdge(Edge{ID: "e3", From: "A", To: "C"})
	_ = d.AddEdge(Edge{ID: "e4", From: "B", To: "C"})

	_, err := NewTreeFrom(d)
	if !errors.Is(err, ErrMultipleParents) {
		t.Fatalf("expected ErrMultipleParents, got %v", err)
	}
}

func TestTree_AddEdge_multiple_parents(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddNode(Node{ID: "B"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = tr.AddEdge(Edge{ID: "e2", From: "R", To: "B"})

	err := tr.AddEdge(Edge{ID: "e3", From: "A", To: "B"})
	if !errors.Is(err, ErrMultipleParents) {
		t.Fatalf("expected ErrMultipleParents, got %v", err)
	}
}

func TestTree_RemoveNode_root(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})

	err := tr.RemoveNode("R")
	if !errors.Is(err, ErrInvalidEdge) {
		t.Fatalf("expected error when removing root, got %v", err)
	}
}

func TestTree_RemoveNode(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})

	err := tr.RemoveNode("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.NodeCount() != 1 {
		t.Fatalf("expected 1 node, got %d", tr.NodeCount())
	}
}

func TestTree_RemoveNode_cascadesSubtree(t *testing.T) {
	t.Parallel()

	// R → A → B → C
	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddNode(Node{ID: "B"})
	_ = tr.AddNode(Node{ID: "C"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = tr.AddEdge(Edge{ID: "e2", From: "A", To: "B"})
	_ = tr.AddEdge(Edge{ID: "e3", From: "B", To: "C"})

	t.Run("removes internal node and all descendants", func(t *testing.T) {
		t.Parallel()

		clone := tr.CloneTree()
		err := clone.RemoveNode("A")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if clone.NodeCount() != 1 {
			t.Fatalf("expected 1 node (root only), got %d", clone.NodeCount())
		}

		if clone.HasNode("A") {
			t.Fatal("expected A to be removed")
		}

		if clone.HasNode("B") {
			t.Fatal("expected B to be cascade-removed")
		}

		if clone.HasNode("C") {
			t.Fatal("expected C to be cascade-removed")
		}
	})

	t.Run("removes leaf without affecting siblings", func(t *testing.T) {
		t.Parallel()

		clone := tr.CloneTree()
		err := clone.RemoveNode("C")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if clone.NodeCount() != 3 {
			t.Fatalf("expected 3 nodes, got %d", clone.NodeCount())
		}

		if clone.HasNode("C") {
			t.Fatal("expected C to be removed")
		}

		if !clone.HasNode("B") {
			t.Fatal("expected B to remain")
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		t.Parallel()

		clone := tr.CloneTree()
		err := clone.RemoveNode("X")
		if err == nil {
			t.Fatal("expected error for non-existent node")
		}
	})
}

func TestTree_Parent(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})

	parent, err := tr.Parent("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parent != "R" {
		t.Fatalf("expected R, got %s", parent)
	}
}

func TestTree_Parent_root(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})

	parent, err := tr.Parent("R")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parent != "" {
		t.Fatalf("expected empty string, got %s", parent)
	}
}

func TestTree_Parent_not_found(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})

	_, err := tr.Parent("X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestTree_Children(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddNode(Node{ID: "B"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})
	_ = tr.AddEdge(Edge{ID: "e2", From: "R", To: "B"})

	children, err := tr.Children("R")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(children) != 2 || children[0] != "A" || children[1] != "B" {
		t.Fatalf("expected [A B], got %v", children)
	}
}

func TestTree_IsLeaf(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})

	isLeaf, err := tr.IsLeaf("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isLeaf {
		t.Fatal("expected A to be a leaf")
	}

	isLeaf, _ = tr.IsLeaf("R")
	if isLeaf {
		t.Fatal("expected R to not be a leaf")
	}
}

func TestTree_IsLeaf_not_found(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})

	_, err := tr.IsLeaf("X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestTree_Clone(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})

	clone := tr.CloneTree()
	_ = tr.AddNode(Node{ID: "B"})

	if clone.NodeCount() != 2 {
		t.Fatal("clone should not be affected by original changes")
	}
	if clone.Root() != "R" {
		t.Fatalf("expected root R, got %s", clone.Root())
	}
}

func TestTree_DAG(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	if tr.DAG() == nil {
		t.Fatal("expected non-nil DAG")
	}
}

func TestTree_Directed(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	if tr.Directed() == nil {
		t.Fatal("expected non-nil Directed")
	}
}

func TestTree_delegated_methods(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})
	_ = tr.AddEdge(Edge{ID: "e1", From: "R", To: "A"})

	t.Run("Node", func(t *testing.T) {
		t.Parallel()
		n, err := tr.Node("R")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n.ID != "R" {
			t.Fatalf("expected R, got %s", n.ID)
		}
	})

	t.Run("Nodes", func(t *testing.T) {
		t.Parallel()
		if len(tr.Nodes()) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(tr.Nodes()))
		}
	})

	t.Run("Edge", func(t *testing.T) {
		t.Parallel()
		e, err := tr.Edge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.From != "R" {
			t.Fatalf("expected from R, got %s", e.From)
		}
	})

	t.Run("Edges", func(t *testing.T) {
		t.Parallel()
		if len(tr.Edges()) != 1 {
			t.Fatalf("expected 1 edge, got %d", len(tr.Edges()))
		}
	})

	t.Run("Neighbors", func(t *testing.T) {
		t.Parallel()
		nb, _ := tr.Neighbors("R")
		if len(nb) != 1 || nb[0] != "A" {
			t.Fatalf("expected [A], got %v", nb)
		}
	})

	t.Run("HasNode_HasEdge", func(t *testing.T) {
		t.Parallel()
		if !tr.HasNode("R") {
			t.Fatal("expected HasNode true")
		}
		if !tr.HasEdge("e1") {
			t.Fatal("expected HasEdge true")
		}
	})

	t.Run("Counts", func(t *testing.T) {
		t.Parallel()
		if tr.NodeCount() != 2 {
			t.Fatalf("expected 2, got %d", tr.NodeCount())
		}
		if tr.EdgeCount() != 1 {
			t.Fatalf("expected 1, got %d", tr.EdgeCount())
		}
	})

	t.Run("Degree", func(t *testing.T) {
		t.Parallel()
		deg, err := tr.Degree("R")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deg != 1 {
			t.Fatalf("expected 1, got %d", deg)
		}
	})

	t.Run("RemoveEdge", func(t *testing.T) {
		t.Parallel()
		clone := tr.CloneTree()
		err := clone.RemoveEdge("e1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestTree_Clone_interface(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.Clone()
}

func TestTree_Parent_no_incoming_edges(t *testing.T) {
	t.Parallel()

	tr := NewTree(Node{ID: "R"})
	_ = tr.AddNode(Node{ID: "A"})

	parent, err := tr.Parent("A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parent != "" {
		t.Fatalf("expected empty parent for disconnected node, got %s", parent)
	}
}
