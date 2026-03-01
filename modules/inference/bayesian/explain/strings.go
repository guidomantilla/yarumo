package explain

import (
	"slices"
	"strconv"
	"strings"

	"github.com/guidomantilla/yarumo/maths/probability"
)

// String returns the string representation of a phase.
func (p Phase) String() string {
	names := [...]string{
		Initialize:  "initialize",
		Propagate:   "propagate",
		Marginalize: "marginalize",
		Complete:    "complete",
	}

	if int(p) < len(names) {
		return names[p]
	}

	return names[Initialize]
}

// String returns a human-readable description of a factor.
func (f Factor) String() string {
	var b strings.Builder

	b.WriteString("Factor(")

	for i, v := range f.Variables {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(v))
	}

	b.WriteString(")[")
	b.WriteString(strconv.Itoa(f.Size))
	b.WriteString("]")

	return b.String()
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

	if len(s.Factor.Variables) > 0 {
		b.WriteString(" — ")
		b.WriteString(s.Factor.String())
	}

	return b.String()
}

// String returns a human-readable description of a posterior.
func (p Posterior) String() string {
	return "P(" + string(p.Variable) + ") = " + p.Distribution.String()
}

// String returns a human-readable description of the full trace.
func (t Trace) String() string {
	var b strings.Builder

	b.WriteString("query: ")
	b.WriteString(string(t.Query))

	if len(t.Evidence) > 0 {
		b.WriteString(" | evidence: ")
		b.WriteString(formatEvidence(t.Evidence))
	}

	for _, step := range t.Steps {
		b.WriteString("\n")
		b.WriteString(step.String())
	}

	for _, post := range t.Posteriors {
		b.WriteString("\n")
		b.WriteString(post.String())
	}

	return b.String()
}

func formatEvidence(ev probability.Assignment) string {
	keys := make([]probability.Var, 0, len(ev))

	for k := range ev {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	var b strings.Builder

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(k))
		b.WriteString("=")
		b.WriteString(string(ev[k]))
	}

	return b.String()
}
