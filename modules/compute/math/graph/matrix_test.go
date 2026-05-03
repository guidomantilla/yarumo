package graph

import (
	"math"
	"testing"
)

func TestToMatrix(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})

	m := ToMatrix(g)
	if len(m.Indices) != 3 {
		t.Fatalf("expected 3 indices, got %d", len(m.Indices))
	}

	ai := m.Indices["A"]
	bi := m.Indices["B"]
	ci := m.Indices["C"]

	if m.Data[ai][bi] != 1 {
		t.Fatalf("expected A->B weight 1, got %f", m.Data[ai][bi])
	}
	if m.Data[bi][ci] != 2 {
		t.Fatalf("expected B->C weight 2, got %f", m.Data[bi][ci])
	}
	if !math.IsInf(m.Data[ai][ci], 1) {
		t.Fatalf("expected A->C weight Inf, got %f", m.Data[ai][ci])
	}
	if m.Data[ai][ai] != 0 {
		t.Fatalf("expected A->A weight 0, got %f", m.Data[ai][ai])
	}
}

func TestToMatrix_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	m := ToMatrix(g)
	if len(m.Data) != 0 {
		t.Fatalf("expected empty matrix, got %d rows", len(m.Data))
	}
}

func TestMatrixMultiply(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})

	m := ToMatrix(g)
	m2 := MatrixMultiply(m, m)

	ai := m2.Indices["A"]
	ci := m2.Indices["C"]

	if m2.Data[ai][ci] != 3 {
		t.Fatalf("expected A->C in M^2 to be 3, got %f", m2.Data[ai][ci])
	}
}

func TestMatrixPower(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})

	m := ToMatrix(g)
	m2 := MatrixPower(m, 2)

	ai := m2.Indices["A"]
	ci := m2.Indices["C"]

	if m2.Data[ai][ci] != 3 {
		t.Fatalf("expected A->C in M^2 to be 3, got %f", m2.Data[ai][ci])
	}
}

func TestMatrixPower_zero(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})

	m := ToMatrix(g)
	m0 := MatrixPower(m, 0)

	ai := m0.Indices["A"]

	if m0.Data[ai][ai] != 0 {
		t.Fatalf("expected identity diagonal 0, got %f", m0.Data[ai][ai])
	}
}

func TestMatrixPower_one(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 3})

	m := ToMatrix(g)
	m1 := MatrixPower(m, 1)

	ai := m1.Indices["A"]
	bi := m1.Indices["B"]

	if m1.Data[ai][bi] != 3 {
		t.Fatalf("expected A->B weight 3, got %f", m1.Data[ai][bi])
	}
}

func TestTransitiveClosure(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	tc := TransitiveClosure(g)

	if !tc["A"]["B"] {
		t.Fatal("expected A can reach B")
	}
	if !tc["A"]["C"] {
		t.Fatal("expected A can reach C")
	}
	if tc["A"]["D"] {
		t.Fatal("expected A cannot reach D")
	}
	if tc["C"]["A"] {
		t.Fatal("expected C cannot reach A")
	}
	if !tc["A"]["A"] {
		t.Fatal("expected self-reachability")
	}
}

func TestTransitiveClosure_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	tc := TransitiveClosure(g)
	if len(tc) != 0 {
		t.Fatalf("expected empty, got %v", tc)
	}
}
