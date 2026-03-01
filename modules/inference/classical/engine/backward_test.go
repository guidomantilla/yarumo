package engine

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func TestBackward_goalAlreadyKnown(t *testing.T) {
	t.Parallel()

	t.Run("goal known true", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()

		proven, result := e.Backward(logic.Fact{"A": true}, []rules.Rule{}, "A")
		if !proven {
			t.Fatal("expected goal to be proven")
		}

		if result.Steps != 0 {
			t.Fatalf("expected 0 steps, got %d", result.Steps)
		}
	})

	t.Run("goal known false", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()

		proven, _ := e.Backward(logic.Fact{"A": false}, []rules.Rule{}, "A")
		if proven {
			t.Fatal("expected goal not proven (value is false)")
		}
	})
}

func TestBackward_goalProvableViaOneRule(t *testing.T) {
	t.Parallel()

	t.Run("proves goal via single rule", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		proven, result := e.Backward(logic.Fact{"A": true}, []rules.Rule{r}, "B")
		if !proven {
			t.Fatal("expected goal B to be proven")
		}

		snap := result.Facts.Snapshot()
		if !snap["B"] {
			t.Fatal("expected B=true in facts")
		}

		if result.Steps != 1 {
			t.Fatalf("expected 1 step, got %d", result.Steps)
		}
	})
}

func TestBackward_goalProvableViaChain(t *testing.T) {
	t.Parallel()

	t.Run("proves goal via rule chain", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
		e := NewEngine()

		proven, result := e.Backward(logic.Fact{"A": true}, []rules.Rule{r1, r2}, "C")
		if !proven {
			t.Fatal("expected goal C to be proven")
		}

		snap := result.Facts.Snapshot()
		if !snap["C"] {
			t.Fatal("expected C=true in facts")
		}
	})
}

func TestBackward_goalNotProvable(t *testing.T) {
	t.Parallel()

	t.Run("goal not provable", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("X"), map[logic.Var]bool{"Y": true})
		e := NewEngine()

		proven, _ := e.Backward(logic.Fact{"A": true}, []rules.Rule{r}, "Z")
		if proven {
			t.Fatal("expected goal Z not provable")
		}
	})
}

func TestBackward_cycleDetection(t *testing.T) {
	t.Parallel()

	t.Run("detects cycle and returns false", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("B"), map[logic.Var]bool{"A": true})
		r2 := rules.NewRule("r2", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		proven, _ := e.Backward(logic.Fact{}, []rules.Rule{r1, r2}, "A")
		if proven {
			t.Fatal("expected cycle to prevent proof")
		}
	})
}

func TestBackward_traceCorrectness(t *testing.T) {
	t.Parallel()

	t.Run("trace records goal and steps", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := NewEngine()

		_, result := e.Backward(logic.Fact{"A": true}, []rules.Rule{r}, "B")

		if result.Trace.Goal != "B" {
			t.Fatalf("expected goal B, got %s", result.Trace.Goal)
		}

		if len(result.Trace.Steps) != 1 {
			t.Fatalf("expected 1 trace step, got %d", len(result.Trace.Steps))
		}
	})
}

func TestBackward_conditionNotSatisfied(t *testing.T) {
	t.Parallel()

	t.Run("all vars known but condition false", func(t *testing.T) {
		t.Parallel()

		// Rule: (A & B) => C, but B is false so condition fails
		r := rules.NewRule("r1",
			logic.AndF{L: logic.Var("A"), R: logic.Var("B")},
			map[logic.Var]bool{"C": true},
		)
		e := NewEngine()

		proven, _ := e.Backward(logic.Fact{"A": true, "B": false}, []rules.Rule{r}, "C")
		if proven {
			t.Fatal("expected not proven when condition is false")
		}
	})
}

func TestBackward_emptyRules(t *testing.T) {
	t.Parallel()

	t.Run("unknown goal with no rules", func(t *testing.T) {
		t.Parallel()

		e := NewEngine()

		proven, _ := e.Backward(logic.Fact{}, []rules.Rule{}, "A")
		if proven {
			t.Fatal("expected not proven with no rules")
		}
	})
}
