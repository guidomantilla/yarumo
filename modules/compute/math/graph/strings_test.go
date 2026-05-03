package graph

import (
	"testing"
)

func TestNode_String(t *testing.T) {
	t.Parallel()

	n := Node{ID: "A"}
	expected := "Node(A)"

	if n.String() != expected {
		t.Fatalf("expected %q, got %q", expected, n.String())
	}
}

func TestEdge_String_default_weight(t *testing.T) {
	t.Parallel()

	e := Edge{ID: "e1", From: "A", To: "B", Weight: 1.0}
	expected := "Edge(e1: A -> B)"

	if e.String() != expected {
		t.Fatalf("expected %q, got %q", expected, e.String())
	}
}

func TestEdge_String_custom_weight(t *testing.T) {
	t.Parallel()

	e := Edge{ID: "e1", From: "A", To: "B", Weight: 2.5}
	expected := "Edge(e1: A -> B, w=2.5)"

	if e.String() != expected {
		t.Fatalf("expected %q, got %q", expected, e.String())
	}
}

func TestEdge_String_with_label(t *testing.T) {
	t.Parallel()

	e := Edge{ID: "e1", From: "A", To: "B", Weight: 1.0, Label: "connects"}
	expected := "Edge(e1: A -> B, connects)"

	if e.String() != expected {
		t.Fatalf("expected %q, got %q", expected, e.String())
	}
}

func TestEdge_String_weight_and_label(t *testing.T) {
	t.Parallel()

	e := Edge{ID: "e1", From: "A", To: "B", Weight: 3.0, Label: "road"}
	expected := "Edge(e1: A -> B, w=3, road)"

	if e.String() != expected {
		t.Fatalf("expected %q, got %q", expected, e.String())
	}
}

func TestPath_String(t *testing.T) {
	t.Parallel()

	p := Path{Nodes: []string{"A", "B", "C"}, Edges: []string{"e1", "e2"}, Weight: 5.0}
	expected := "Path(A -> B -> C, w=5)"

	if p.String() != expected {
		t.Fatalf("expected %q, got %q", expected, p.String())
	}
}

func TestPath_String_empty(t *testing.T) {
	t.Parallel()

	p := Path{}
	expected := "Path(, w=0)"

	if p.String() != expected {
		t.Fatalf("expected %q, got %q", expected, p.String())
	}
}
