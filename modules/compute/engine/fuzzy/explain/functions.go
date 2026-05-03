package explain

import "maps"

// NewTrace creates an empty trace with the given input values.
func NewTrace(inputs map[string]float64) Trace {
	copied := make(map[string]float64, len(inputs))
	maps.Copy(copied, inputs)

	return Trace{
		Inputs: copied,
	}
}

// AddStep appends a step to the trace and returns the updated trace.
func (t Trace) AddStep(step Step) Trace {
	t.Steps = append(t.Steps, step)

	return t
}

// AddOutput appends an output to the trace and returns the updated trace.
func (t Trace) AddOutput(out Output) Trace {
	t.Outputs = append(t.Outputs, out)

	return t
}
