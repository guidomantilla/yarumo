package engine

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
	"github.com/guidomantilla/yarumo/inference/classical/facts"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

// Backward attempts to prove the goal variable via backward chaining.
func (e *engine) Backward(initialFacts logic.Fact, ruleSet []rules.Rule, goal logic.Var) (bool, Result) {
	cassert.NotNil(e, "engine is nil")

	fb := facts.NewFactBaseFrom(initialFacts)
	trace := explain.NewGoalTrace(goal)
	visited := make(map[logic.Var]bool)
	step := 0

	proven, trace, step := prove(goal, fb, ruleSet, visited, trace, step)

	return proven, Result{
		Facts: fb,
		Trace: trace,
		Steps: step,
	}
}

func prove(goal logic.Var, fb facts.FactBase, ruleSet []rules.Rule, visited map[logic.Var]bool, trace explain.Trace, step int) (bool, explain.Trace, int) {
	val, known := fb.Get(goal)
	if known {
		return val, trace, step
	}

	if visited[goal] {
		return false, trace, step
	}

	visited[goal] = true

	for _, r := range ruleSet {
		conclusion := r.Conclusion()

		_, targets := conclusion[goal]
		if !targets {
			continue
		}

		ok, updatedTrace, updatedStep := tryRule(r, fb, ruleSet, visited, trace, step)
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

func tryRule(r rules.Rule, fb facts.FactBase, ruleSet []rules.Rule, visited map[logic.Var]bool, trace explain.Trace, step int) (bool, explain.Trace, int) {
	condVars := r.Condition().Vars()

	for _, v := range condVars {
		_, vKnown := fb.Get(v)
		if vKnown {
			continue
		}

		proven, updatedTrace, updatedStep := prove(v, fb, ruleSet, visited, trace, step)
		trace = updatedTrace
		step = updatedStep

		if !proven {
			return false, trace, step
		}
	}

	snapshot := fb.Snapshot()

	if !r.Fires(snapshot) {
		return false, trace, step
	}

	step++
	conclusion := r.Conclusion()

	for v, val := range conclusion {
		fb.Derive(v, val, r.Name(), step)
	}

	trace = trace.AddStep(explain.Step{
		Number:      step,
		RuleName:    r.Name(),
		Condition:   r.Condition(),
		FactsBefore: snapshot,
		Produced:    conclusion,
	})

	return true, trace, step
}
