package probability

import (
	"math"
	"testing"
)

func TestNewFactor(t *testing.T) {
	t.Parallel()

	table := map[string]Prob{"A=t": 0.7, "A=f": 0.3}
	f := NewFactor([]Var{"A"}, table)

	if len(f.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(f.Variables))
	}

	if f.Table["A=t"] != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(f.Table["A=t"]))
	}
}

func TestNewFactor_copiesInputs(t *testing.T) {
	t.Parallel()

	vars := []Var{"A"}
	table := map[string]Prob{"A=t": 0.7}

	f := NewFactor(vars, table)

	// Modify originals.
	vars[0] = "Z"
	table["A=t"] = 0.1

	if f.Variables[0] != "A" {
		t.Fatalf("expected A, got %s", string(f.Variables[0]))
	}

	if f.Table["A=t"] != 0.7 {
		t.Fatalf("expected 0.7, got %f", float64(f.Table["A=t"]))
	}
}

func TestMultiply_basic(t *testing.T) {
	t.Parallel()

	a := NewFactor([]Var{"A"}, map[string]Prob{
		"A=t": 0.6,
		"A=f": 0.4,
	})
	b := NewFactor([]Var{"A", "B"}, map[string]Prob{
		"A=t,B=t": 0.2,
		"A=t,B=f": 0.8,
		"A=f,B=t": 0.5,
		"A=f,B=f": 0.5,
	})

	result := Multiply(a, b)

	if len(result.Variables) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(result.Variables))
	}

	// P(A=t)*P(A=t,B=t) = 0.6*0.2 = 0.12
	expected := 0.6 * 0.2
	if math.Abs(float64(result.Table["A=t,B=t"])-expected) > epsilon {
		t.Fatalf("expected %f for A=t,B=t, got %f", expected, float64(result.Table["A=t,B=t"]))
	}
}

func TestMultiply_conflictingAssignments(t *testing.T) {
	t.Parallel()

	a := NewFactor([]Var{"A"}, map[string]Prob{
		"A=t": 0.6,
	})
	b := NewFactor([]Var{"A"}, map[string]Prob{
		"A=f": 0.4,
	})

	result := Multiply(a, b)

	// A=t and A=f conflict, so no entry should be produced.
	if len(result.Table) != 0 {
		t.Fatalf("expected empty table for conflicting assignments, got %d entries", len(result.Table))
	}
}

func TestSumOut_basic(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A", "B"}, map[string]Prob{
		"A=t,B=t": 0.12,
		"A=t,B=f": 0.48,
		"A=f,B=t": 0.20,
		"A=f,B=f": 0.20,
	})

	result := SumOut(f, "B")

	if len(result.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(result.Variables))
	}

	// Sum over B: A=t -> 0.12 + 0.48 = 0.60
	expected := 0.12 + 0.48
	if math.Abs(float64(result.Table["A=t"])-expected) > epsilon {
		t.Fatalf("expected %f for A=t, got %f", expected, float64(result.Table["A=t"]))
	}
}

func TestRestrict_basic(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A", "B"}, map[string]Prob{
		"A=t,B=t": 0.12,
		"A=t,B=f": 0.48,
		"A=f,B=t": 0.20,
		"A=f,B=f": 0.20,
	})

	result := Restrict(f, "A", "t")

	if len(result.Variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(result.Variables))
	}

	if result.Table["B=t"] != 0.12 {
		t.Fatalf("expected 0.12 for B=t, got %f", float64(result.Table["B=t"]))
	}

	if result.Table["B=f"] != 0.48 {
		t.Fatalf("expected 0.48 for B=f, got %f", float64(result.Table["B=f"]))
	}
}

func TestRestrict_noMatch(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A"}, map[string]Prob{
		"A=t": 0.6,
	})

	result := Restrict(f, "A", "f")

	if len(result.Table) != 0 {
		t.Fatalf("expected empty table, got %d entries", len(result.Table))
	}
}

func TestNormalizeFactor_basic(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A"}, map[string]Prob{
		"A=t": 2,
		"A=f": 3,
	})

	result := NormalizeFactor(f)

	if math.Abs(float64(result.Table["A=t"])-0.4) > epsilon {
		t.Fatalf("expected 0.4, got %f", float64(result.Table["A=t"]))
	}

	if math.Abs(float64(result.Table["A=f"])-0.6) > epsilon {
		t.Fatalf("expected 0.6, got %f", float64(result.Table["A=f"]))
	}
}

func TestNormalizeFactor_zeroSum(t *testing.T) {
	t.Parallel()

	f := NewFactor([]Var{"A"}, map[string]Prob{
		"A=t": 0,
		"A=f": 0,
	})

	result := NormalizeFactor(f)

	if result.Table["A=t"] != 0 {
		t.Fatalf("expected 0 for zero sum, got %f", float64(result.Table["A=t"]))
	}
}

func TestDeserializeAssignment_empty(t *testing.T) {
	t.Parallel()

	result := DeserializeAssignment("")

	if len(result) != 0 {
		t.Fatalf("expected empty assignment, got %d", len(result))
	}
}

func TestDeserializeAssignment_single(t *testing.T) {
	t.Parallel()

	result := DeserializeAssignment("A=t")

	if result["A"] != "t" {
		t.Fatalf("expected t, got %s", string(result["A"]))
	}
}

func TestDeserializeAssignment_multiple(t *testing.T) {
	t.Parallel()

	result := DeserializeAssignment("A=t,B=f")

	if result["A"] != "t" {
		t.Fatalf("expected t for A, got %s", string(result["A"]))
	}

	if result["B"] != "f" {
		t.Fatalf("expected f for B, got %s", string(result["B"]))
	}
}

func TestMergeVars(t *testing.T) {
	t.Parallel()

	result := mergeVars([]Var{"B", "A"}, []Var{"C", "A"})

	if len(result) != 3 {
		t.Fatalf("expected 3 vars, got %d", len(result))
	}

	// Should be sorted.
	if result[0] != "A" || result[1] != "B" || result[2] != "C" {
		t.Fatalf("expected [A B C], got %v", result)
	}
}

func TestRemoveVar(t *testing.T) {
	t.Parallel()

	result := removeVar([]Var{"A", "B", "C"}, "B")

	if len(result) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(result))
	}

	if result[0] != "A" || result[1] != "C" {
		t.Fatalf("expected [A C], got %v", result)
	}
}

func TestMergeAssignments_noConflict(t *testing.T) {
	t.Parallel()

	a := Assignment{"A": "t"}
	b := Assignment{"B": "f"}

	result, ok := mergeAssignments(a, b)
	if !ok {
		t.Fatal("expected successful merge")
	}

	if result["A"] != "t" || result["B"] != "f" {
		t.Fatalf("unexpected merge result: %v", result)
	}
}

func TestMergeAssignments_conflict(t *testing.T) {
	t.Parallel()

	a := Assignment{"A": "t"}
	b := Assignment{"A": "f"}

	_, ok := mergeAssignments(a, b)
	if ok {
		t.Fatal("expected conflict")
	}
}

func TestMergeAssignments_sameValue(t *testing.T) {
	t.Parallel()

	a := Assignment{"A": "t"}
	b := Assignment{"A": "t"}

	result, ok := mergeAssignments(a, b)
	if !ok {
		t.Fatal("expected successful merge")
	}

	if result["A"] != "t" {
		t.Fatalf("expected t, got %s", string(result["A"]))
	}
}
