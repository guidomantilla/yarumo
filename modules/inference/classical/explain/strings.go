package explain

import (
	"slices"
	"strings"

	"github.com/guidomantilla/yarumo/maths/logic"
)

// String returns the string representation of the origin.
func (o Origin) String() string {
	if o == Derived {
		return "derived"
	}

	return "asserted"
}

// String returns a human-readable description of the inference step.
func (s Step) String() string {
	var b strings.Builder

	b.WriteString("step ")
	b.WriteString(intToStr(s.Number))
	b.WriteString(": rule \"")
	b.WriteString(s.RuleName)
	b.WriteString("\" fired")

	if s.Condition != nil {
		b.WriteString(", condition: ")
		b.WriteString(logic.Format(s.Condition))
	}

	produced := sortedProduced(s.Produced)
	if len(produced) > 0 {
		b.WriteString(", produced: ")
		b.WriteString(produced)
	}

	return b.String()
}

// String returns a human-readable description of the full trace.
func (t Trace) String() string {
	var b strings.Builder

	if t.Goal != "" {
		b.WriteString("goal: ")
		b.WriteString(string(t.Goal))
		b.WriteString("\n")
	}

	for i, step := range t.Steps {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(step.String())
	}

	return b.String()
}

func sortedProduced(produced map[logic.Var]bool) string {
	if len(produced) == 0 {
		return ""
	}

	keys := make([]logic.Var, 0, len(produced))
	for k := range produced {
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

		if produced[k] {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	}

	return b.String()
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := make([]byte, 0, 4)

	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	if negative {
		digits = append(digits, '-')
	}

	slices.Reverse(digits)

	return string(digits)
}
