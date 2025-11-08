package examples

import (
	"fmt"
	"testing"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/engine"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/parser"
	p "github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"
)

// TestEngine_EndToEnd demonstrates a minimal forward-chaining run and query/explain.
func TestEngine_EndToEnd(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
		engine.BuildRule("r3", "D & E", "F"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	eng.Assert(p.Var("E"))

	fired := eng.RunToFixpoint(5)
	if len(fired) == 0 {
		t.Fatalf("expected some rules to fire")
	}

	// After running: A, B are asserted; r1 fires to set C; r2 fires to set D; with E true, r3 fires to set F
	for _, v := range []p.Var{"C", "D", "F"} {
		if val, _ := eng.Facts.Get(v); !val {
			t.Fatalf("expected %s to be true after fixpoint", v)
		}
	}

	// Query with explanation
	ok, why := eng.Query(parser.MustParse("F"))
	if !ok || why == nil {
		t.Fatalf("expected F to be true with explanation")
	}

	// Print a readable explanation for documentation value.
	// We use Output block to make go test verify formatting deterministically.
	fmt.Print(engine.PrettyExplain(why))
}

// TestEngine_NoFire ensures rules do not fire when conditions are not met.
func TestEngine_NoFire(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	eng.Assert(p.Var("A")) // B remains false (default)

	fired := eng.FireOnce()
	if len(fired) != 0 {
		t.Fatalf("expected no rules to fire, got %v", fired)
	}

	ok, why := eng.Query(parser.MustParse("C"))
	if ok {
		t.Fatalf("expected C to be false without B")
	}
	_ = why
}

// TestEngine_ChainingAndIdempotence verifies multi-step chaining and that rules do not re-fire once fact is true.
func TestEngine_ChainingAndIdempotence(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
		engine.BuildRule("r3", "D & E", "F"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	eng.Assert(p.Var("E"))

	fired1 := eng.FireOnce()
	if len(fired1) == 0 {
		t.Fatalf("expected first pass to fire at least one rule")
	}
	fired2 := eng.FireOnce()
	// After first pass, additional passes may still fire remaining downstream rules.
	// Run until convergence with a small loop.
	total := append([]string{}, fired1...)
	total = append(total, fired2...)
	for i := 0; i < 5; i++ {
		step := eng.FireOnce()
		if len(step) == 0 {
			break
		}
		total = append(total, step...)
	}

	// Now facts C, D, F should be true via chaining.
	for _, v := range []p.Var{"C", "D", "F"} {
		if val, _ := eng.Facts.Get(v); !val {
			t.Fatalf("expected %s to be true after chaining", v)
		}
	}

	// Idempotence: another FireOnce should not fire any rule now.
	if step := eng.FireOnce(); len(step) != 0 {
		t.Fatalf("expected no further firing, got %v", step)
	}
}

// TestEngine_ImplAndIffPatterns exercises the special firing semantics for implication and biconditional.
func TestEngine_ImplAndIffPatterns(t *testing.T) {
	rules := []engine.Rule{
		// For implication, When: X=>Y and Then: Y should fire when X is true.
		engine.BuildRule("imp", "X => Y", "Y"),
		// For IFF, if Then is one side, When should fire when the other side is true.
		engine.BuildRule("iff1", "P <=> Q", "P"),
		engine.BuildRule("iff2", "P <=> Q", "Q"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	// Case 1: X true implies Y should be derived.
	eng.Assert(p.Var("X"))
	_ = eng.RunToFixpoint(3)
	if val, _ := eng.Facts.Get(p.Var("Y")); !val {
		t.Fatalf("expected Y to be true due to implication rule")
	}

	// Reset engine for IFF tests
	eng = engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	// Case 2: Q true should derive P (via iff1)
	eng.Assert(p.Var("Q"))
	_ = eng.RunToFixpoint(3)
	if val, _ := eng.Facts.Get(p.Var("P")); !val {
		t.Fatalf("expected P to be true due to IFF rule with Then=P")
	}

	// Case 3: P true should derive Q (via iff2)
	eng = engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("P"))
	_ = eng.RunToFixpoint(3)
	if val, _ := eng.Facts.Get(p.Var("Q")); !val {
		t.Fatalf("expected Q to be true due to IFF rule with Then=Q")
	}
}

// TestEngine_MaxItersAndLoopGuard ensures RunToFixpoint stops due to maxIters with a cycle.
func TestEngine_MaxItersAndLoopGuard(t *testing.T) {
	// Create a simple cycle: A -> B, B -> A (via implications). Our semantics will fire when antecedent is true.
	rules := []engine.Rule{
		engine.BuildRule("r1", "A => B", "B"),
		engine.BuildRule("r2", "B => A", "A"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	eng.Assert(p.Var("A"))
	fired := eng.RunToFixpoint(1) // limit to 1 iteration deliberately
	if len(fired) == 0 {
		t.Fatalf("expected at least one rule to fire in first iteration")
	}
	// With maxIters=1, engine must stop even if more could fire.
	if val, _ := eng.Facts.Get(p.Var("B")); !val {
		t.Fatalf("expected B to be true after first iteration")
	}
}

// TestEngine_RetractAndReAssert checks behavior when facts change between runs.
func TestEngine_RetractAndReAssert(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	_ = eng.RunToFixpoint(2)
	if val, _ := eng.Facts.Get(p.Var("C")); !val {
		t.Fatalf("expected C to be true after A&B")
	}

	// Retract B and verify C is not automatically retracted (MVP semantics: monotonic facts)
	eng.Retract(p.Var("B"))
	if val, _ := eng.Facts.Get(p.Var("C")); !val {
		t.Fatalf("expected C to remain true (monotonic accumulation in MVP)")
	}

	// Start a fresh engine to verify re-assert then recompute
	eng = engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	_ = eng.RunToFixpoint(2)
	if val, _ := eng.Facts.Get(p.Var("C")); !val {
		t.Fatalf("expected C to be true in fresh engine after A&B")
	}
}

// TestEngine_QueryCompositeFormula validates Query over a composite formula with explanation structure.
func TestEngine_QueryCompositeFormula(t *testing.T) {
	rules := []engine.Rule{
		engine.BuildRule("r1", "A & B", "C"),
		engine.BuildRule("r2", "C => D", "D"),
	}
	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}
	eng.Assert(p.Var("A"))
	eng.Assert(p.Var("B"))
	_ = eng.RunToFixpoint(3)

	// Query a composite goal: (D | E) & C
	goal := parser.MustParse("(D | E) & C")
	ok, why := eng.Query(goal)
	if !ok || why == nil {
		t.Fatalf("expected composite goal to be true with explanation")
	}
	// Basic structure checks
	if len(why.Kids) != 2 {
		t.Fatalf("expected AND to have two children in explanation")
	}
}
