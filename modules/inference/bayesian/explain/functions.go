package explain

import "github.com/guidomantilla/yarumo/maths/probability"

// NewTrace creates an empty trace for the given query and evidence.
func NewTrace(query probability.Var, evidence probability.Assignment) Trace {
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
