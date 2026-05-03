package engine

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/compute/engine/deductive/explain"
	"github.com/guidomantilla/yarumo/compute/engine/deductive/facts"
	"github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
)

// Backward attempts to prove the goal variable via backward chaining.
func (e *engine) Backward(initialFacts logic.Fact, ruleSet []rules.Rule, goal logic.Var) (bool, Result) {
	cassert.NotNil(e, "engine is nil")

	fb := facts.NewFactBaseFrom(initialFacts)
	trace := explain.NewGoalTrace(goal)
	step := 0

	proven, trace, step := prove(goal, fb, ruleSet, trace, step, 0, e.options.maxDepth)

	return proven, Result{
		Facts: fb,
		Trace: trace,
		Steps: step,
	}
}

func prove(goal logic.Var, fb facts.FactBase, ruleSet []rules.Rule, trace explain.Trace, step int, depth int, maxDepth int) (bool, explain.Trace, int) {
	val, known := fb.Get(goal)
	if known {
		return val, trace, step
	}

	if depth >= maxDepth {
		return false, trace, step
	}

	for _, r := range ruleSet {
		conclusion := r.Conclusion()

		_, targets := conclusion[goal]
		if !targets {
			continue
		}

		ok, updatedTrace, updatedStep := tryRule(r, fb, ruleSet, trace, step, depth, maxDepth)
		trace = updatedTrace
		step = updatedStep

		if !ok {
			continue
		}

		goalVal, goalKnown := fb.Get(goal)
		if goalKnown {
			return goalVal, trace, step
		}
	}

	return false, trace, step
}

func tryRule(r rules.Rule, fb facts.FactBase, ruleSet []rules.Rule, trace explain.Trace, step int, depth int, maxDepth int) (bool, explain.Trace, int) {
	// Clone fact-base to avoid contamination on failed attempts.
	clone := fb.Clone()
	condVars := r.Condition().Vars()

	for _, v := range condVars {
		_, vKnown := clone.Get(v)
		if vKnown {
			continue
		}

		_, updatedTrace, updatedStep := prove(v, clone, ruleSet, trace, step, depth+1, maxDepth)
		trace = updatedTrace
		step = updatedStep
	}

	// Evaluate the full formula against the cloned fact-base.
	snapshot := clone.Snapshot()

	if !r.Fires(snapshot) {
		return false, trace, step
	}

	// Rule fires — commit conclusions to the original fact-base.
	step++
	before := fb.Snapshot()
	conclusion := r.Conclusion()

	for v, val := range conclusion {
		fb.Derive(v, val, r.Name(), step)
	}

	// Also commit any intermediate derivations from the clone.
	cloneSnap := clone.Snapshot()

	for v, val := range cloneSnap {
		_, known := fb.Get(v)
		if !known {
			prov, hasProv := clone.Provenance(v)
			if hasProv {
				fb.Derive(v, val, prov.RuleName, prov.Step)
			}
		}
	}

	trace = trace.AddStep(explain.Step{
		Number:      step,
		RuleName:    r.Name(),
		Condition:   r.Condition(),
		FactsBefore: before,
		Produced:    conclusion,
	})

	return true, trace, step
}
