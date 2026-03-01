package explain

import (
	"slices"
	"strconv"
	"strings"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

// String returns the string representation of a phase.
func (p Phase) String() string {
	names := [...]string{
		Fuzzification:   "fuzzification",
		RuleEvaluation:  "rule-evaluation",
		Aggregation:     "aggregation",
		Defuzzification: "defuzzification",
		Complete:        "complete",
	}

	if int(p) < len(names) {
		return names[p]
	}

	return names[Fuzzification]
}

// String returns a human-readable description of a membership result.
func (m Membership) String() string {
	return m.Variable + "/" + m.Term + "=" + fuzzym.Degree(m.Degree).String()
}

// String returns a human-readable description of a rule activation.
func (a Activation) String() string {
	return a.RuleName + " -> " + a.Output + "/" + a.Term + " [" + a.Strength.String() + "]"
}

// String returns a human-readable description of an inference step.
func (s Step) String() string {
	var b strings.Builder

	b.WriteString("step ")
	b.WriteString(strconv.Itoa(s.Number))
	b.WriteString(": [")
	b.WriteString(s.Phase.String())
	b.WriteString("] ")
	b.WriteString(s.Message)

	for _, m := range s.Memberships {
		b.WriteString("\n  ")
		b.WriteString(m.String())
	}

	for _, a := range s.Activations {
		b.WriteString("\n  ")
		b.WriteString(a.String())
	}

	return b.String()
}

// String returns a human-readable description of an output.
func (o Output) String() string {
	return o.Variable + " = " + strconv.FormatFloat(o.CrispValue, 'f', 4, 64)
}

// String returns a human-readable description of the full trace.
func (t Trace) String() string {
	var b strings.Builder

	b.WriteString("inputs: ")
	b.WriteString(formatInputs(t.Inputs))

	for _, step := range t.Steps {
		b.WriteString("\n")
		b.WriteString(step.String())
	}

	for _, out := range t.Outputs {
		b.WriteString("\n")
		b.WriteString(out.String())
	}

	return b.String()
}

func formatInputs(inputs map[string]float64) string {
	keys := make([]string, 0, len(inputs))

	for k := range inputs {
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
		b.WriteString(strconv.FormatFloat(inputs[k], 'f', 2, 64))
	}

	return b.String()
}
