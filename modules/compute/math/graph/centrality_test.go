package graph

import (
	"math"
	"testing"
)

func TestDegreeCentrality(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	dc := DegreeCentrality(g)
	if dc["A"] != 1.0 {
		t.Fatalf("expected A centrality 1.0, got %f", dc["A"])
	}
	if dc["B"] != 0.5 {
		t.Fatalf("expected B centrality 0.5, got %f", dc["B"])
	}
}

func TestDegreeCentrality_single_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	dc := DegreeCentrality(g)
	if dc["A"] != 0 {
		t.Fatalf("expected 0, got %f", dc["A"])
	}
}

func TestDegreeCentrality_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	dc := DegreeCentrality(g)
	if len(dc) != 0 {
		t.Fatalf("expected empty, got %v", dc)
	}
}

func TestBetweennessCentrality(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	bc := BetweennessCentrality(g)
	if bc["B"] < bc["A"] {
		t.Fatalf("expected B to have higher betweenness than A, got A=%f B=%f", bc["A"], bc["B"])
	}
}

func TestBetweennessCentrality_star(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "Center"})
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "Center", To: "A"})
	_ = g.AddEdge(Edge{ID: "e2", From: "Center", To: "B"})
	_ = g.AddEdge(Edge{ID: "e3", From: "Center", To: "C"})

	bc := BetweennessCentrality(g)
	if bc["Center"] == 0 {
		t.Fatal("expected center to have positive betweenness")
	}
	if bc["A"] != 0 {
		t.Fatalf("expected leaf betweenness 0, got %f", bc["A"])
	}
}

func TestClosenessCentrality(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	cc := ClosenessCentrality(g)
	if cc["B"] < cc["A"] {
		t.Fatalf("expected B closeness >= A, got A=%f B=%f", cc["A"], cc["B"])
	}
}

func TestClosenessCentrality_single_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	cc := ClosenessCentrality(g)
	if cc["A"] != 0 {
		t.Fatalf("expected 0, got %f", cc["A"])
	}
}

func TestClosenessCentrality_disconnected(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	cc := ClosenessCentrality(g)
	if cc["A"] != 0 {
		t.Fatalf("expected 0 for disconnected node, got %f", cc["A"])
	}
}

func TestPageRank(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	pr := PageRank(g, 0.85, 100)

	total := 0.0
	for _, v := range pr {
		total += v
	}
	if math.Abs(total-1.0) > 0.001 {
		t.Fatalf("expected total rank ~1.0, got %f", total)
	}
}

func TestPageRank_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	pr := PageRank(g, 0.85, 10)
	if len(pr) != 0 {
		t.Fatalf("expected empty, got %v", pr)
	}
}

func TestPageRank_dangling_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	pr := PageRank(g, 0.85, 100)
	if pr["B"] < pr["A"] {
		t.Fatalf("expected B to have higher PageRank, A=%f B=%f", pr["A"], pr["B"])
	}
}
