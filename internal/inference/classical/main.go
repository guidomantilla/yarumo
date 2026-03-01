package main

import (
	"fmt"

	"github.com/guidomantilla/yarumo/maths/logic"
	"github.com/guidomantilla/yarumo/maths/logic/parser"

	"github.com/guidomantilla/yarumo/inference/classical/engine"
	"github.com/guidomantilla/yarumo/inference/classical/explain"
	"github.com/guidomantilla/yarumo/inference/classical/facts"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

func main() {
	forwardChaining()
	backwardChaining()
	priorityAndConflict()
	provenanceTracking()
	parserIntegration()
	factBaseOperations()
}

// forwardChaining shows data-driven reasoning: given facts, what can we conclude?
func forwardChaining() {
	fmt.Println("=== Forward Chaining ===")

	// Define rules:
	// Rule 1: IF rain THEN wet_ground
	// Rule 2: IF wet_ground THEN slippery
	// Rule 3: IF slippery AND windy THEN dangerous
	r1 := rules.NewRule("rain-causes-wet",
		logic.Var("rain"),
		map[logic.Var]bool{"wet_ground": true},
	)
	r2 := rules.NewRule("wet-causes-slippery",
		logic.Var("wet_ground"),
		map[logic.Var]bool{"slippery": true},
	)
	r3 := rules.NewRule("slippery-and-windy-dangerous",
		logic.AndF{L: logic.Var("slippery"), R: logic.Var("windy")},
		map[logic.Var]bool{"dangerous": true},
	)

	// Start with initial facts
	initial := logic.Fact{"rain": true, "windy": true}

	// Forward chaining: apply all applicable rules until no more fire
	eng := engine.NewEngine()
	result := eng.Forward(initial, []rules.Rule{r1, r2, r3})

	// Show all derived facts
	snap := result.Facts.Snapshot()
	fmt.Println("Initial facts: rain=true, windy=true")
	fmt.Println("Derived facts:")
	for k, v := range snap {
		fmt.Printf("  %s = %v\n", k, v)
	}
	fmt.Printf("Steps taken: %d\n", result.Steps)
	fmt.Println()
}

// backwardChaining shows goal-driven reasoning: can we prove a specific goal?
func backwardChaining() {
	fmt.Println("=== Backward Chaining ===")

	// Same rules as before
	r1 := rules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
	r2 := rules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
	r3 := rules.NewRule("r3", logic.Var("C"), map[logic.Var]bool{"D": true})

	eng := engine.NewEngine()

	// Can we prove D starting from A?
	proven, result := eng.Backward(logic.Fact{"A": true}, []rules.Rule{r1, r2, r3}, "D")
	fmt.Println("Can we prove D from {A=true}?", proven)
	fmt.Printf("  Steps: %d\n", result.Steps)

	// Can we prove Z? (no rule derives Z)
	proven, _ = eng.Backward(logic.Fact{"A": true}, []rules.Rule{r1, r2, r3}, "Z")
	fmt.Println("Can we prove Z from {A=true}?", proven)
	fmt.Println()
}

// priorityAndConflict shows how rule priority resolves conflicts.
func priorityAndConflict() {
	fmt.Println("=== Priority and Conflict Resolution ===")

	// Two rules match the same condition, but with different priorities
	// Lower number = higher priority (like urgency levels)
	lowPriority := rules.NewRule("default-action",
		logic.Var("alert"),
		map[logic.Var]bool{"action_log": true},
		rules.WithPriority(10),
	)
	highPriority := rules.NewRule("emergency-action",
		logic.Var("alert"),
		map[logic.Var]bool{"action_evacuate": true},
		rules.WithPriority(1),
	)

	eng := engine.NewEngine()
	result := eng.Forward(logic.Fact{"alert": true}, []rules.Rule{lowPriority, highPriority})

	// The high priority rule fires first
	fmt.Println("Rule firing order:")
	for i, step := range result.Trace.Steps {
		fmt.Printf("  %d. %s (produced: %v)\n", i+1, step.RuleName, step.Produced)
	}

	// FirstMatch strategy: only one rule fires per pass
	fmt.Println()
	fmt.Println("With FirstMatch strategy (one rule per pass):")
	fmEngine := engine.NewEngine(engine.WithStrategy(engine.FirstMatch))
	fmResult := fmEngine.Forward(logic.Fact{"alert": true}, []rules.Rule{lowPriority, highPriority})
	fmt.Printf("  Steps: %d (one rule per iteration)\n", fmResult.Steps)
	fmt.Println()
}

// provenanceTracking shows how to trace where each fact came from.
func provenanceTracking() {
	fmt.Println("=== Provenance Tracking ===")

	r1 := rules.NewRule("r1", logic.Var("symptom_fever"), map[logic.Var]bool{"possible_infection": true})
	r2 := rules.NewRule("r2",
		logic.AndF{L: logic.Var("possible_infection"), R: logic.Var("symptom_cough")},
		map[logic.Var]bool{"likely_flu": true},
	)

	eng := engine.NewEngine()
	initial := logic.Fact{"symptom_fever": true, "symptom_cough": true}
	result := eng.Forward(initial, []rules.Rule{r1, r2})

	// Check provenance: where did each fact come from?
	fmt.Println("Fact origins:")
	for _, v := range []logic.Var{"symptom_fever", "possible_infection", "likely_flu"} {
		prov, ok := result.Facts.Provenance(v)
		if !ok {
			continue
		}

		if prov.Origin == explain.Asserted {
			fmt.Printf("  %s: ASSERTED (initial fact)\n", v)
		} else {
			fmt.Printf("  %s: DERIVED by rule '%s' at step %d\n", v, prov.RuleName, prov.Step)
		}
	}

	// Full trace
	fmt.Println()
	fmt.Println("Execution trace:")
	fmt.Println(result.Trace.String())
}

// parserIntegration shows how to use the logic parser for rule conditions.
func parserIntegration() {
	fmt.Println("=== Parser Integration ===")

	// Parse complex conditions from strings instead of building them manually
	condition := parser.MustParse("eligible & verified & !suspended")
	fmt.Println("Condition:", logic.Format(condition))

	r := rules.NewRule("approve",
		condition,
		map[logic.Var]bool{"approved": true},
	)

	eng := engine.NewEngine()

	// Scenario 1: all conditions met
	result1 := eng.Forward(
		logic.Fact{"eligible": true, "verified": true, "suspended": false},
		[]rules.Rule{r},
	)
	fmt.Println("Eligible+Verified+NotSuspended => approved:", result1.Facts.Snapshot()["approved"])

	// Scenario 2: suspended blocks approval
	result2 := eng.Forward(
		logic.Fact{"eligible": true, "verified": true, "suspended": true},
		[]rules.Rule{r},
	)
	_, hasApproved := result2.Facts.Snapshot()["approved"]
	fmt.Println("Eligible+Verified+Suspended => approved:", hasApproved)
	fmt.Println()
}

// factBaseOperations shows direct manipulation of the fact base.
func factBaseOperations() {
	fmt.Println("=== FactBase Operations ===")

	fb := facts.NewFactBase()

	// Assert facts
	fb.Assert("temperature", true)
	fb.Assert("humidity", true)
	fb.Assert("pressure", false)
	fmt.Println("Facts count:", fb.Len())

	// Get a specific fact
	val, known := fb.Get("temperature")
	fmt.Printf("temperature: value=%v, known=%v\n", val, known)

	// Retract (remove) a fact
	fb.Retract("pressure")
	fmt.Println("After retracting pressure:", fb.Len(), "facts")

	// Clone: independent copy
	clone := fb.Clone()
	clone.Assert("wind", true)
	fmt.Printf("Original: %d facts, Clone: %d facts\n", fb.Len(), clone.Len())

	// Snapshot: get all facts as a map
	snap := fb.Snapshot()
	fmt.Println("Snapshot:", snap)
}
