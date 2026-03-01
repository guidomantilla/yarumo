package examples

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/logic"
	"github.com/guidomantilla/yarumo/maths/logic/parser"

	"github.com/guidomantilla/yarumo/inference/classical/engine"
	"github.com/guidomantilla/yarumo/inference/classical/explain"
	"github.com/guidomantilla/yarumo/inference/classical/facts"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func TestModusPonens(t *testing.T) {
	t.Parallel()

	t.Run("rain causes wet ground causes slippery", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("rain-wet",
			logic.Var("rain"),
			map[logic.Var]bool{"wet_ground": true},
		)
		r2 := rules.NewRule("wet-slippery",
			logic.Var("wet_ground"),
			map[logic.Var]bool{"slippery": true},
		)

		e := engine.NewEngine()
		result := e.Forward(logic.Fact{"rain": true}, []rules.Rule{r1, r2})

		snap := result.Facts.Snapshot()
		if !snap["rain"] {
			t.Fatal("expected rain=true")
		}

		if !snap["wet_ground"] {
			t.Fatal("expected wet_ground=true")
		}

		if !snap["slippery"] {
			t.Fatal("expected slippery=true")
		}

		if result.Steps != 2 {
			t.Fatalf("expected 2 steps, got %d", result.Steps)
		}
	})
}

func TestBusinessScenario(t *testing.T) {
	t.Parallel()

	t.Run("multi-rule business logic", func(t *testing.T) {
		t.Parallel()

		premium := rules.NewRule("premium-customer",
			logic.AndF{L: logic.Var("high_spend"), R: logic.Var("loyal")},
			map[logic.Var]bool{"premium": true},
		)
		discount := rules.NewRule("discount-eligible",
			logic.Var("premium"),
			map[logic.Var]bool{"discount": true},
		)
		notification := rules.NewRule("notify",
			logic.Var("discount"),
			map[logic.Var]bool{"notify": true},
		)

		e := engine.NewEngine()
		initial := logic.Fact{"high_spend": true, "loyal": true}
		result := e.Forward(initial, []rules.Rule{premium, discount, notification})

		snap := result.Facts.Snapshot()
		if !snap["premium"] {
			t.Fatal("expected premium=true")
		}

		if !snap["discount"] {
			t.Fatal("expected discount=true")
		}

		if !snap["notify"] {
			t.Fatal("expected notify=true")
		}
	})
}

func TestBackwardChaining(t *testing.T) {
	t.Parallel()

	t.Run("prove reachable goal", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
		r3 := rules.NewRule("r3", logic.Var("C"), map[logic.Var]bool{"D": true})

		e := engine.NewEngine()
		proven, result := e.Backward(logic.Fact{"A": true}, []rules.Rule{r1, r2, r3}, "D")

		if !proven {
			t.Fatal("expected D to be provable")
		}

		snap := result.Facts.Snapshot()
		if !snap["D"] {
			t.Fatal("expected D=true in facts")
		}
	})

	t.Run("unreachable goal", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("X"), map[logic.Var]bool{"Y": true})
		e := engine.NewEngine()

		proven, _ := e.Backward(logic.Fact{"A": true}, []rules.Rule{r}, "Z")
		if proven {
			t.Fatal("expected Z not provable")
		}
	})
}

func TestExplanationTrace(t *testing.T) {
	t.Parallel()

	t.Run("trace inspection", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})

		e := engine.NewEngine()
		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2})
		traceStr := result.Trace.String()

		if traceStr == "" {
			t.Fatal("expected non-empty trace")
		}

		if len(result.Trace.Steps) != 2 {
			t.Fatalf("expected 2 trace steps, got %d", len(result.Trace.Steps))
		}

		if result.Trace.Steps[0].RuleName != "r1" {
			t.Fatalf("expected r1 first, got %s", result.Trace.Steps[0].RuleName)
		}

		if result.Trace.Steps[1].RuleName != "r2" {
			t.Fatalf("expected r2 second, got %s", result.Trace.Steps[1].RuleName)
		}
	})
}

func TestPriorityConflict(t *testing.T) {
	t.Parallel()

	t.Run("higher priority rule fires first", func(t *testing.T) {
		t.Parallel()

		low := rules.NewRule("low",
			logic.Var("A"),
			map[logic.Var]bool{"B": true},
			rules.WithPriority(10),
		)
		high := rules.NewRule("high",
			logic.Var("A"),
			map[logic.Var]bool{"C": true},
			rules.WithPriority(1),
		)

		e := engine.NewEngine()
		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{low, high})

		if result.Trace.Steps[0].RuleName != "high" {
			t.Fatalf("expected high priority first, got %s", result.Trace.Steps[0].RuleName)
		}
	})
}

func TestProvenanceTracking(t *testing.T) {
	t.Parallel()

	t.Run("asserted vs derived", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := engine.NewEngine()
		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r})

		provA, okA := result.Facts.Provenance("A")
		if !okA {
			t.Fatal("expected provenance for A")
		}

		if provA.Origin != explain.Asserted {
			t.Fatal("expected A to be asserted")
		}

		provB, okB := result.Facts.Provenance("B")
		if !okB {
			t.Fatal("expected provenance for B")
		}

		if provB.Origin != explain.Derived {
			t.Fatal("expected B to be derived")
		}

		if provB.RuleName != "r1" {
			t.Fatalf("expected rule r1, got %s", provB.RuleName)
		}
	})

	t.Run("ProvenanceOf utility", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		e := engine.NewEngine()
		initial := logic.Fact{"A": true}
		result := e.Forward(initial, []rules.Rule{r})

		provs := explain.ProvenanceOf(result.Trace, initial)
		if len(provs) < 2 {
			t.Fatalf("expected at least 2 provenances, got %d", len(provs))
		}
	})
}

func TestParserIntegration(t *testing.T) {
	t.Parallel()

	t.Run("parser formulas used as conditions", func(t *testing.T) {
		t.Parallel()

		condition := parser.MustParse("A & B")
		r := rules.NewRule("parsed-rule", condition, map[logic.Var]bool{"C": true})

		e := engine.NewEngine()
		result := e.Forward(logic.Fact{"A": true, "B": true}, []rules.Rule{r})

		snap := result.Facts.Snapshot()
		if !snap["C"] {
			t.Fatal("expected C=true from parsed condition")
		}
	})
}

func TestFactBaseOperations(t *testing.T) {
	t.Parallel()

	t.Run("clone independence", func(t *testing.T) {
		t.Parallel()

		fb := facts.NewFactBase()
		fb.Assert("A", true)

		clone := fb.Clone()
		clone.Assert("B", false)

		if fb.Len() != 1 {
			t.Fatal("expected original unchanged")
		}

		if clone.Len() != 2 {
			t.Fatal("expected clone to have 2 facts")
		}
	})

	t.Run("retract and re-assert", func(t *testing.T) {
		t.Parallel()

		fb := facts.NewFactBase()
		fb.Assert("A", true)
		fb.Retract("A")

		_, known := fb.Get("A")
		if known {
			t.Fatal("expected A retracted")
		}

		fb.Assert("A", false)

		val, known := fb.Get("A")
		if !known || val {
			t.Fatal("expected A=false after re-assert")
		}
	})
}

func TestRuleVariables(t *testing.T) {
	t.Parallel()

	t.Run("returns all unique variables", func(t *testing.T) {
		t.Parallel()

		r := rules.NewRule("r1",
			logic.AndF{L: logic.Var("A"), R: logic.Var("B")},
			map[logic.Var]bool{"C": true, "A": false},
		)

		vars := rules.Variables(r)
		if len(vars) != 3 {
			t.Fatalf("expected 3, got %d", len(vars))
		}
	})
}

func TestFirstMatchStrategy(t *testing.T) {
	t.Parallel()

	t.Run("only one rule per pass", func(t *testing.T) {
		t.Parallel()

		r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
		r2 := rules.NewRule("r2", logic.Var("A"), map[logic.Var]bool{"C": true})

		e := engine.NewEngine(engine.WithStrategy(engine.FirstMatch))
		result := e.Forward(logic.Fact{"A": true}, []rules.Rule{r1, r2})

		if result.Steps != 2 {
			t.Fatalf("expected 2 steps (one per pass), got %d", result.Steps)
		}
	})
}
