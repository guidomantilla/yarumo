package engine

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
	"github.com/guidomantilla/yarumo/inference/classical/facts"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
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

		snapshot := fb.Snapshot()
		changed := false

		for _, r := range sorted {
			if !r.Produces(snapshot) {
				continue
			}

			step++
			before := fb.Snapshot()
			conclusion := r.Conclusion()

			for v, val := range conclusion {
				_, known := fb.Get(v)
				if !known || fb.Snapshot()[v] != val {
					fb.Derive(v, val, r.Name(), step)
				}
			}

			trace = trace.AddStep(explain.Step{
				Number:      step,
				RuleName:    r.Name(),
				Condition:   r.Condition(),
				FactsBefore: before,
				Produced:    conclusion,
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
	}

	return Result{
		Facts: fb,
		Trace: trace,
		Steps: step,
	}
}
