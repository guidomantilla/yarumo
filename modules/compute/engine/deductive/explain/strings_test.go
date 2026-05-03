package explain

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/logic"
)

func TestOrigin_String(t *testing.T) {
	t.Parallel()

	t.Run("asserted", func(t *testing.T) {
		t.Parallel()

		got := Asserted.String()
		if got != "asserted" {
			t.Fatalf("expected asserted, got %s", got)
		}
	})

	t.Run("derived", func(t *testing.T) {
		t.Parallel()

		got := Derived.String()
		if got != "derived" {
			t.Fatalf("expected derived, got %s", got)
		}
	})
}

func TestStep_String(t *testing.T) {
	t.Parallel()

	t.Run("with condition and produced", func(t *testing.T) {
		t.Parallel()

		s := Step{
			Number:    1,
			RuleName:  "rule1",
			Condition: logic.Var("A"),
			Produced:  map[logic.Var]bool{"B": true},
		}
		got := s.String()

		expected := `step 1: rule "rule1" fired, condition: A, produced: B=true`
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("nil condition", func(t *testing.T) {
		t.Parallel()

		s := Step{
			Number:   2,
			RuleName: "rule2",
			Produced: map[logic.Var]bool{"C": false},
		}
		got := s.String()

		expected := `step 2: rule "rule2" fired, produced: C=false`
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("empty produced", func(t *testing.T) {
		t.Parallel()

		s := Step{
			Number:    3,
			RuleName:  "rule3",
			Condition: logic.Var("X"),
		}
		got := s.String()

		expected := `step 3: rule "rule3" fired, condition: X`
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("multiple produced sorted", func(t *testing.T) {
		t.Parallel()

		s := Step{
			Number:    1,
			RuleName:  "multi",
			Condition: logic.Var("A"),
			Produced:  map[logic.Var]bool{"C": true, "B": false},
		}
		got := s.String()

		expected := `step 1: rule "multi" fired, condition: A, produced: B=false, C=true`
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}

func TestTrace_String(t *testing.T) {
	t.Parallel()

	t.Run("forward chaining trace", func(t *testing.T) {
		t.Parallel()

		tr := Trace{
			Steps: []Step{
				{Number: 1, RuleName: "r1", Condition: logic.Var("A"), Produced: map[logic.Var]bool{"B": true}},
				{Number: 2, RuleName: "r2", Condition: logic.Var("B"), Produced: map[logic.Var]bool{"C": true}},
			},
		}
		got := tr.String()

		expected := "step 1: rule \"r1\" fired, condition: A, produced: B=true\nstep 2: rule \"r2\" fired, condition: B, produced: C=true"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("backward chaining trace with goal", func(t *testing.T) {
		t.Parallel()

		tr := Trace{
			Goal: "C",
			Steps: []Step{
				{Number: 1, RuleName: "r1", Condition: logic.Var("A"), Produced: map[logic.Var]bool{"B": true}},
			},
		}
		got := tr.String()

		expected := "goal: C\nstep 1: rule \"r1\" fired, condition: A, produced: B=true"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("empty trace", func(t *testing.T) {
		t.Parallel()

		tr := Trace{}

		got := tr.String()
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("goal only no steps", func(t *testing.T) {
		t.Parallel()

		tr := Trace{Goal: "X"}
		got := tr.String()

		expected := "goal: X\n"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}

func Test_sortedProduced(t *testing.T) {
	t.Parallel()

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()

		got := sortedProduced(nil)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		got := sortedProduced(map[logic.Var]bool{})
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("single entry", func(t *testing.T) {
		t.Parallel()

		got := sortedProduced(map[logic.Var]bool{"A": true})
		if got != "A=true" {
			t.Fatalf("expected A=true, got %q", got)
		}
	})

	t.Run("multiple entries sorted", func(t *testing.T) {
		t.Parallel()

		got := sortedProduced(map[logic.Var]bool{"C": false, "A": true, "B": true})

		expected := "A=true, B=true, C=false"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}
