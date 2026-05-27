package engine

import (
	"maps"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/compute/engine/deductive/explain"
	"github.com/guidomantilla/yarumo/compute/engine/deductive/facts"
	"github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
)

// Forward runs forward chaining from the given initial facts and rule set.
func (e *engine) Forward(initialFacts logic.Fact, ruleSet []rules.Rule) Result {
	cassert.NotNil(e, "engine is nil")

	fb := facts.NewFactBaseFrom(initialFacts)
	sorted := rules.SortByPriority(ruleSet)
	trace := explain.NewTrace()
	step := 0

	for i := range e.options.maxIterations {
		_ = i

		iterStart := fb.Snapshot()
		snapshot := fb.Snapshot()
		changed := false

		for _, r := range sorted {
			if !r.Produces(snapshot) {
				continue
			}

			step++
			before := fb.Snapshot()
			conclusion := r.Conclusion()
			produced := make(map[logic.Var]bool)

			for v, val := range conclusion {
				current, known := before[v]
				if !known || current != val {
					fb.Derive(v, val, r.Name(), step)
					produced[v] = val
				}
			}

			trace = trace.AddStep(explain.Step{
				Number:      step,
				RuleName:    r.Name(),
				Condition:   r.Condition(),
				FactsBefore: before,
				Produced:    produced,
			})

			changed = true
			snapshot = fb.Snapshot()

			if e.options.strategy == FirstMatch {
				break
			}
		}

		if !changed {
			break
		}

		// Detect oscillation: if net state unchanged after full pass, stop.
		if maps.Equal(iterStart, fb.Snapshot()) {
			break
		}
	}

	return Result{
		Facts: fb,
		Trace: trace,
		Steps: step,
	}
}
