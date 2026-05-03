// Package explain provides explanation trace types for causal inference.
package explain

// Phase identifies a stage in the causal inference process.
type Phase int

const (
	// Propagation marks the forward computation phase.
	Propagation Phase = iota
	// Intervention marks the do-operator application phase.
	Intervention
	// Counterfactual marks the hypothetical reasoning phase.
	Counterfactual
	// Attribution marks the causal attribution phase.
	Attribution
	// Complete marks the final result phase.
	Complete
)

// Step records a single computation step.
type Step struct {
	Number  int
	Phase   Phase
	Message string
	Values  map[string]float64
}

// CausalAttribution records how much each variable contributed to an outcome.
type CausalAttribution struct {
	Target       string
	Attributions map[string]float64
}

// Trace records the full causal inference computation.
type Trace struct {
	Steps        []Step
	Observations map[string]float64
	Outputs      map[string]float64
	Attributions []CausalAttribution
}

// NewTrace creates a new trace with the given observations.
func NewTrace(observations map[string]float64) Trace {
	return Trace{Observations: observations}
}

// AddStep appends a step to the trace and returns the updated trace.
func (t Trace) AddStep(step Step) Trace {
	t.Steps = append(t.Steps, step)

	return t
}

// AddOutput records a final output value and returns the updated trace.
func (t Trace) AddOutput(variable string, value float64) Trace {
	if t.Outputs == nil {
		t.Outputs = make(map[string]float64)
	}

	t.Outputs[variable] = value

	return t
}

// AddAttribution records a causal attribution and returns the updated trace.
func (t Trace) AddAttribution(attr CausalAttribution) Trace {
	t.Attributions = append(t.Attributions, attr)

	return t
}
