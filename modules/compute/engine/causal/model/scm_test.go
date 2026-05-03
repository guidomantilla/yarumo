package model

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/graph"

	"github.com/guidomantilla/yarumo/compute/engine/causal"
)

func TestNewSCM(t *testing.T) {
	t.Parallel()

	s := NewSCM()
	if s == nil {
		t.Fatal("expected non-nil SCM")
	}
}

func TestSCM_AddVariable_basic(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := s.Variable("X")
	if !ok {
		t.Fatal("expected variable X")
	}

	if v.Name != "X" {
		t.Fatalf("expected name X, got %s", v.Name)
	}
}

func TestSCM_AddVariable_withParents(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := s.Variable("Z")
	if !ok {
		t.Fatal("expected variable Z")
	}

	if len(v.Parents) != 1 || v.Parents[0] != "X" {
		t.Fatalf("expected parent X, got %v", v.Parents)
	}
}

func TestSCM_AddVariable_duplicate(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 1
	})
	if !errors.Is(err, causal.ErrDuplicateVariable) {
		t.Fatalf("expected ErrDuplicateVariable, got %v", err)
	}
}

func TestSCM_AddVariable_nilEquation(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, nil)
	if !errors.Is(err, causal.ErrNilEquation) {
		t.Fatalf("expected ErrNilEquation, got %v", err)
	}
}

func TestSCM_Variable_found(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 42
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, ok := s.Variable("X")
	if !ok {
		t.Fatal("expected variable X to be found")
	}

	if v.Name != "X" {
		t.Fatalf("expected name X, got %s", v.Name)
	}
}

func TestSCM_Variable_notFound(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	_, ok := s.Variable("X")
	if ok {
		t.Fatal("expected variable X to not be found")
	}
}

func TestSCM_Variables_topological(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Y", []string{"Z"}, func(parents map[string]float64) float64 {
		return parents["Z"] + 3
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vars := s.Variables()

	if len(vars) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(vars))
	}

	// X must come before Z, Z must come before Y.
	indexOf := func(name string) int {
		for i, v := range vars {
			if v == name {
				return i
			}
		}

		return -1
	}

	xIdx := indexOf("X")
	zIdx := indexOf("Z")
	yIdx := indexOf("Y")

	if xIdx >= zIdx {
		t.Fatalf("expected X before Z, got X=%d Z=%d", xIdx, zIdx)
	}

	if zIdx >= yIdx {
		t.Fatalf("expected Z before Y, got Z=%d Y=%d", zIdx, yIdx)
	}
}

func TestSCM_Parents(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parents := s.Parents("Z")
	if len(parents) != 1 || parents[0] != "X" {
		t.Fatalf("expected [X], got %v", parents)
	}

	parents = s.Parents("X")
	if len(parents) != 0 {
		t.Fatalf("expected empty parents for X, got %v", parents)
	}
}

func TestSCM_Parents_notFound(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	parents := s.Parents("unknown")
	if parents != nil {
		t.Fatalf("expected nil parents for unknown variable, got %v", parents)
	}
}

func TestSCM_Children(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	children := s.Children("X")
	if len(children) != 1 || children[0] != "Z" {
		t.Fatalf("expected [Z], got %v", children)
	}

	children = s.Children("Z")
	if len(children) != 0 {
		t.Fatalf("expected empty children for Z, got %v", children)
	}
}

func TestSCM_Validate_valid(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("X", nil, func(_ map[string]float64) float64 {
		return 0
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.Validate()
	if err != nil {
		t.Fatalf("expected valid model, got %v", err)
	}
}

func TestSCM_Validate_missingParent(t *testing.T) {
	t.Parallel()

	s := NewSCM()

	err := s.AddVariable("Z", []string{"X"}, func(parents map[string]float64) float64 {
		return parents["X"] * 2
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = s.Validate()
	if !errors.Is(err, causal.ErrParentNotFound) {
		t.Fatalf("expected ErrParentNotFound, got %v", err)
	}
}

func TestSCM_Validate_cycle(t *testing.T) {
	t.Parallel()

	// Create a cycle: A→B→C→A using graph.Directed.
	g := graph.NewDirected()
	eq := func(_ map[string]float64) float64 { return 0 }

	_ = g.AddNode(graph.Node{ID: "A", Metadata: Variable{Name: "A", Parents: []string{"C"}, Equation: eq}})
	_ = g.AddNode(graph.Node{ID: "B", Metadata: Variable{Name: "B", Parents: []string{"A"}, Equation: eq}})
	_ = g.AddNode(graph.Node{ID: "C", Metadata: Variable{Name: "C", Parents: []string{"B"}, Equation: eq}})
	_ = g.AddEdge(graph.Edge{ID: "C->A", From: "C", To: "A"})
	_ = g.AddEdge(graph.Edge{ID: "A->B", From: "A", To: "B"})
	_ = g.AddEdge(graph.Edge{ID: "B->C", From: "B", To: "C"})

	s := &scm{
		g:     g,
		order: []string{"A", "B", "C"},
	}

	err := s.Validate()
	if !errors.Is(err, causal.ErrCyclicModel) {
		t.Fatalf("expected ErrCyclicModel, got %v", err)
	}
}
