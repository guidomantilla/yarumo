package graph

import (
	"testing"
)

func TestGraphColoring_triangle(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	coloring := GraphColoring(g)
	if len(coloring) != 3 {
		t.Fatalf("expected 3 colors, got %d", len(coloring))
	}

	for _, e := range g.Edges() {
		if coloring[e.From] == coloring[e.To] {
			t.Fatalf("adjacent nodes %s and %s have same color %d", e.From, e.To, coloring[e.From])
		}
	}
}

func TestGraphColoring_bipartite(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	coloring := GraphColoring(g)
	maxColor := 0

	for _, c := range coloring {
		if c > maxColor {
			maxColor = c
		}
	}

	if maxColor > 1 {
		t.Fatalf("bipartite graph should need at most 2 colors, used %d", maxColor+1)
	}
}

func TestGraphColoring_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	coloring := GraphColoring(g)
	if len(coloring) != 0 {
		t.Fatalf("expected empty, got %v", coloring)
	}
}

func TestGraphColoring_valid(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "D"})
	_ = g.AddEdge(Edge{ID: "e4", From: "D", To: "A"})

	coloring := GraphColoring(g)

	for _, e := range g.Edges() {
		if coloring[e.From] == coloring[e.To] {
			t.Fatalf("adjacent nodes %s and %s have same color", e.From, e.To)
		}
	}
}

func TestChromaticNumber(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	cn := ChromaticNumber(g)
	if cn != 3 {
		t.Fatalf("expected 3 colors for triangle, got %d", cn)
	}
}

func TestChromaticNumber_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	cn := ChromaticNumber(g)
	if cn != 0 {
		t.Fatalf("expected 0, got %d", cn)
	}
}

func TestChromaticNumber_single_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	cn := ChromaticNumber(g)
	if cn != 1 {
		t.Fatalf("expected 1, got %d", cn)
	}
}

func TestGraphColoring_directed(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	coloring := GraphColoring(g)
	if len(coloring) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(coloring))
	}
}
