package engine

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func TestNewOptions_defaults(t *testing.T) {
	t.Parallel()

	o := NewOptions()

	if o.algorithm != Enumeration {
		t.Fatalf("expected Enumeration, got %d", o.algorithm)
	}

	if o.eliminationOrder != nil {
		t.Fatalf("expected nil elimination order")
	}
}

func TestWithAlgorithm_enumeration(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithAlgorithm(Enumeration))

	if o.algorithm != Enumeration {
		t.Fatalf("expected Enumeration, got %d", o.algorithm)
	}
}

func TestWithAlgorithm_variableElimination(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithAlgorithm(VariableElimination))

	if o.algorithm != VariableElimination {
		t.Fatalf("expected VariableElimination, got %d", o.algorithm)
	}
}

func TestWithAlgorithm_invalid(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithAlgorithm(Algorithm(-1)))

	if o.algorithm != Enumeration {
		t.Fatalf("expected default Enumeration, got %d", o.algorithm)
	}
}

func TestWithEliminationOrder(t *testing.T) {
	t.Parallel()

	order := []probability.Var{"A", "B"}
	o := NewOptions(WithEliminationOrder(order))

	if len(o.eliminationOrder) != 2 {
		t.Fatalf("expected 2, got %d", len(o.eliminationOrder))
	}
}

func TestWithEliminationOrder_empty(t *testing.T) {
	t.Parallel()

	o := NewOptions(WithEliminationOrder(nil))

	if o.eliminationOrder != nil {
		t.Fatal("expected nil for empty order")
	}
}
