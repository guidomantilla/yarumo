package explain

import (
	"slices"
	"strconv"
	"strings"
)

// String returns the string representation of a phase.
func (p Phase) String() string {
	names := [...]string{
		Propagation:    "propagation",
		Intervention:   "intervention",
		Counterfactual: "counterfactual",
		Attribution:    "attribution",
		Complete:       "complete",
	}

	if int(p) >= 0 && int(p) < len(names) {
		return names[p]
	}

	return "unknown"
}

// String returns a human-readable description of a computation step.
func (s Step) String() string {
	var b strings.Builder

	b.WriteString("step ")
	b.WriteString(strconv.Itoa(s.Number))
	b.WriteString(": [")
	b.WriteString(s.Phase.String())
	b.WriteString("] ")
	b.WriteString(s.Message)

	if len(s.Values) > 0 {
		b.WriteString(" {")
		b.WriteString(formatValues(s.Values))
		b.WriteString("}")
	}

	return b.String()
}

// String returns a human-readable description of a causal attribution.
func (a CausalAttribution) String() string {
	var b strings.Builder

	b.WriteString("attribution(")
	b.WriteString(a.Target)
	b.WriteString("): ")
	b.WriteString(formatValues(a.Attributions))

	return b.String()
}

// String returns a human-readable description of the full trace.
func (t Trace) String() string {
	var b strings.Builder

	b.WriteString("observations: ")
	b.WriteString(formatValues(t.Observations))

	for _, step := range t.Steps {
		b.WriteString("\n")
		b.WriteString(step.String())
	}

	if len(t.Outputs) > 0 {
		b.WriteString("\noutputs: ")
		b.WriteString(formatValues(t.Outputs))
	}

	for _, attr := range t.Attributions {
		b.WriteString("\n")
		b.WriteString(attr.String())
	}

	return b.String()
}

func formatValues(vals map[string]float64) string {
	keys := make([]string, 0, len(vals))

	for k := range vals {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	var b strings.Builder

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(strconv.FormatFloat(vals[k], 'f', 4, 64))
	}

	return b.String()
}
