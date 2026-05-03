package graph

import (
	"testing"
)

func TestSubgraph(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	sg := Subgraph(g, []string{"A", "B"})
	if sg.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", sg.NodeCount())
	}
	if sg.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", sg.EdgeCount())
	}
}

func TestSubgraphUndirected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	sg := SubgraphUndirected(g, []string{"A", "B"})
	if sg.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", sg.NodeCount())
	}
	if sg.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", sg.EdgeCount())
	}
}

func TestUnion(t *testing.T) {
	t.Parallel()

	a := NewDirected()
	_ = a.AddNode(Node{ID: "A"})
	_ = a.AddNode(Node{ID: "B"})
	_ = a.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	b := NewDirected()
	_ = b.AddNode(Node{ID: "B"})
	_ = b.AddNode(Node{ID: "C"})
	_ = b.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	u := Union(a, b)
	if u.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", u.NodeCount())
	}
	if u.EdgeCount() != 2 {
		t.Fatalf("expected 2 edges, got %d", u.EdgeCount())
	}
}

func TestIntersection(t *testing.T) {
	t.Parallel()

	a := NewDirected()
	_ = a.AddNode(Node{ID: "A"})
	_ = a.AddNode(Node{ID: "B"})
	_ = a.AddNode(Node{ID: "C"})
	_ = a.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	b := NewDirected()
	_ = b.AddNode(Node{ID: "B"})
	_ = b.AddNode(Node{ID: "C"})
	_ = b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	inter := Intersection(a, b)
	if inter.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", inter.NodeCount())
	}
}

func TestComplement(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	comp := Complement(g)
	if comp.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", comp.NodeCount())
	}

	// 3 nodes = 6 possible directed edges (no self-loops) - 1 existing = 5
	if comp.EdgeCount() != 5 {
		t.Fatalf("expected 5 edges, got %d", comp.EdgeCount())
	}
}

func TestReverse(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	rev := Reverse(g)
	e, _ := rev.Edge("e1")
	if e.From != "B" || e.To != "A" {
		t.Fatalf("expected reversed edge B->A, got %s->%s", e.From, e.To)
	}
}

func TestCartesianProduct(t *testing.T) {
	t.Parallel()

	a := NewDirected()
	_ = a.AddNode(Node{ID: "1"})
	_ = a.AddNode(Node{ID: "2"})
	_ = a.AddEdge(Edge{ID: "e1", From: "1", To: "2"})

	b := NewDirected()
	_ = b.AddNode(Node{ID: "X"})
	_ = b.AddNode(Node{ID: "Y"})
	_ = b.AddEdge(Edge{ID: "e2", From: "X", To: "Y"})

	cp := CartesianProduct(a, b)
	if cp.NodeCount() != 4 {
		t.Fatalf("expected 4 nodes, got %d", cp.NodeCount())
	}
	// Edge from a: 1->2 applied to X,Y = 2 edges: 1:X->2:X, 1:Y->2:Y
	// Edge from b: X->Y applied to 1,2 = 2 edges: 1:X->1:Y, 2:X->2:Y
	if cp.EdgeCount() != 4 {
		t.Fatalf("expected 4 edges, got %d", cp.EdgeCount())
	}
}

func TestSubgraph_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	sg := Subgraph(g, []string{})
	if sg.NodeCount() != 0 {
		t.Fatalf("expected 0 nodes, got %d", sg.NodeCount())
	}
}

func TestReverse_preserves_weight(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 5.0})

	rev := Reverse(g)
	e, _ := rev.Edge("e1")
	if e.Weight != 5.0 {
		t.Fatalf("expected weight 5.0, got %f", e.Weight)
	}
}

func TestIntersection_no_common_edges(t *testing.T) {
	t.Parallel()

	a := NewDirected()
	_ = a.AddNode(Node{ID: "A"})
	_ = a.AddNode(Node{ID: "B"})
	_ = a.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	b := NewDirected()
	_ = b.AddNode(Node{ID: "A"})
	_ = b.AddNode(Node{ID: "B"})
	_ = b.AddEdge(Edge{ID: "e2", From: "A", To: "B"})

	inter := Intersection(a, b)
	if inter.EdgeCount() != 0 {
		t.Fatalf("expected 0 edges, got %d", inter.EdgeCount())
	}
}

func TestIntersection_common_edges(t *testing.T) {
	t.Parallel()

	a := NewDirected()
	_ = a.AddNode(Node{ID: "A"})
	_ = a.AddNode(Node{ID: "B"})
	_ = a.AddNode(Node{ID: "C"})
	_ = a.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 3})
	_ = a.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	b := NewDirected()
	_ = b.AddNode(Node{ID: "A"})
	_ = b.AddNode(Node{ID: "B"})
	_ = b.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = b.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	inter := Intersection(a, b)
	if inter.NodeCount() != 2 {
		t.Fatalf("expected 2 nodes, got %d", inter.NodeCount())
	}
	if inter.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge (e1 only, e2's C not in intersection), got %d", inter.EdgeCount())
	}
	if !inter.HasEdge("e1") {
		t.Fatal("expected edge e1 in intersection")
	}
}
