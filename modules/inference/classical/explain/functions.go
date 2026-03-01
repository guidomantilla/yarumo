package explain

import "github.com/guidomantilla/yarumo/maths/logic"

// NewTrace creates an empty trace for forward chaining.
func NewTrace() Trace {
	return Trace{}
}

// NewGoalTrace creates an empty trace for backward chaining toward a goal.
func NewGoalTrace(goal logic.Var) Trace {
	return Trace{Goal: goal}
}

// AddStep appends a step to the trace and returns the updated trace.
func (t Trace) AddStep(step Step) Trace {
	t.Steps = append(t.Steps, step)
	return t
}

// ProvenanceOf returns provenance records for all facts in a trace.
// Initial facts are marked as asserted; derived facts record their rule and step.
func ProvenanceOf(trace Trace, initial logic.Fact) []Provenance {
	result := make([]Provenance, 0)

	for v, val := range initial {
		result = append(result, Provenance{
			Variable: v,
			Value:    val,
			Origin:   Asserted,
		})
	}

	seen := make(map[logic.Var]bool)

	for _, step := range trace.Steps {
		for v, val := range step.Produced {
			if seen[v] {
				continue
			}

			seen[v] = true

			result = append(result, Provenance{
				Variable: v,
				Value:    val,
				Origin:   Derived,
				RuleName: step.RuleName,
				Step:     step.Number,
			})
		}
	}

	return result
}
