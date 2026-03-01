package engine

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func TestForward_emptyRules(t *testing.T) {
	t.Parallel()

	t.Run("returns initial facts unchanged", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{})
		if result.Steps != 0 {
			t.Fatalf("expected 0 steps, got %d", result.Steps)
		}

		snap := result.Facts.Snapshot()
		if !snap["A"] {
			t.Fatal("expected A=true in result")
		}
	})
}

func TestForward_singleRule(t *testing.T) {
	t.Parallel()

	t.Run("fires single rule", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r})
		if result.Steps != 1 {
			t.Fatalf("expected 1 step, got %d", result.Steps)
		}

		snap := result.Facts.Snapshot()
		if !snap["B"] {
			t.Fatal("expected B=true")
		}
	})
}

func TestForward_chain(t *testing.T) {
	t.Parallel()

	t.Run("multi-rule chain A to B to C", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2})

		snap := result.Facts.Snapshot()
		if !snap["B"] {
			t.Fatal("expected B=true")
		}

		if !snap["C"] {
			t.Fatal("expected C=true")
		}

		if result.Steps < 2 {
			t.Fatalf("expected at least 2 steps, got %d", result.Steps)
		}
	})
}

func TestForward_fixpoint(t *testing.T) {
	t.Parallel()

	t.Run("stops at fixpoint", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r})

		if result.Steps != 1 {
			t.Fatalf("expected exactly 1 step (fixpoint), got %d", result.Steps)
		}
	})
}

func TestForward_maxIterations(t *testing.T) {
	t.Parallel()

	t.Run("respects max iterations", func(t *testing.T) {
		t.Parallel()

		e := NewEngine(WithMaxIterations(2))
		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
		r3 := rules.NewRule("r3", logic.Var("C"), map[logic.Var]bool{"D": true})

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2, r3})

		snap := result.Facts.Snapshot()
		if !snap["B"] {
			t.Fatal("expected B=true")
		}

		if !snap["C"] {
			t.Fatal("expected C=true")
		}
	})
}

func TestForward_priority(t *testing.T) {
	t.Parallel()

	t.Run("fires rules in priority order", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("low", logic.Var("A"), map[logic.Var]bool{"B": true}, rules.WithPriority(10))
		r2 := rules.NewRule("high", logic.Var("A"), map[logic.Var]bool{"C": true}, rules.WithPriority(1))
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2})
		if len(result.Trace.Steps) < 2 {
			t.Fatalf("expected at least 2 steps, got %d", len(result.Trace.Steps))
		}

		if result.Trace.Steps[0].RuleName != "high" {
			t.Fatalf("expected high priority first, got %s", result.Trace.Steps[0].RuleName)
		}
	})
}

func TestForward_firstMatch(t *testing.T) {
	t.Parallel()

	t.Run("fires only first match per pass", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("A"), map[logic.Var]bool{"C": true})
		e := NewEngine(WithStrategy(FirstMatch))

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2})

		snap := result.Facts.Snapshot()
		if !snap["B"] {
			t.Fatal("expected B=true")
		}

		if !snap["C"] {
			t.Fatal("expected C=true (from second pass)")
		}

		if result.Steps != 2 {
			t.Fatalf("expected 2 steps (one per pass), got %d", result.Steps)
		}
	})
}

func TestForward_traceCorrectness(t *testing.T) {
	t.Parallel()

	t.Run("trace records all steps", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r})

		if len(result.Trace.Steps) != 1 {
			t.Fatalf("expected 1 trace step, got %d", len(result.Trace.Steps))
		}

		step := result.Trace.Steps[0]
		if step.Number != 1 {
			t.Fatalf("expected step number 1, got %d", step.Number)
		}

		if step.RuleName != "r1" {
			t.Fatalf("expected r1, got %s", step.RuleName)
		}

		if step.Condition == nil {
			t.Fatal("expected non-nil condition")
		}

		if step.FactsBefore == nil {
			t.Fatal("expected non-nil facts before")
		}
	})
}

func TestForward_conditionNotMet(t *testing.T) {
	t.Parallel()

	t.Run("rule does not fire when condition is false", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		result := e.Forward(logic.Fact{"A": false}, []rules.Rule{r})
		if result.Steps != 0 {
			t.Fatalf("expected 0 steps, got %d", result.Steps)
		}
	})
}
