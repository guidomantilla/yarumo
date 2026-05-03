package explain

import "github.com/guidomantilla/yarumo/compute/math/stats"

// NewTrace creates an empty trace for the given query and evidence.
func NewTrace(query stats.Var, evidence stats.Assignment) Trace {
	return Trace{
		Query:    query,
		Evidence: evidence,
	}
}

// AddStep appends a step to the trace and returns the updated trace.
func (t Trace) AddStep(step Step) Trace {
	t.Steps = append(t.Steps, step)

	return t
}

// AddPosterior appends a posterior to the trace and returns the updated trace.
func (t Trace) AddPosterior(post Posterior) Trace {
	t.Posteriors = append(t.Posteriors, post)

	return t
}
