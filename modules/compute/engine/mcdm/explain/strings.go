package explain

import (
	"strconv"
	"strings"
)

// String returns a human-readable description of a rank entry.
func (r RankEntry) String() string {
	var b strings.Builder

	b.WriteString("#")
	b.WriteString(strconv.Itoa(r.Rank))
	b.WriteString(" alternative ")
	b.WriteString(strconv.Itoa(r.Alternative))
	b.WriteString(" (score: ")
	b.WriteString(formatFloat(r.Score))
	b.WriteString(")")

	return b.String()
}

// String returns a human-readable description of the full MCDM trace.
func (t Trace) String() string {
	var b strings.Builder

	b.WriteString("method: ")
	b.WriteString(t.Method)

	if len(t.Criteria) > 0 {
		b.WriteString("\ncriteria: ")
		b.WriteString(formatCriteria(t.Criteria, t.Weights))
	}

	if len(t.Rankings) > 0 {
		b.WriteString("\nrankings:")

		for _, r := range t.Rankings {
			b.WriteString("\n  ")
			b.WriteString(r.String())
		}
	}

	return b.String()
}

func formatCriteria(names []string, weights []float64) string {
	var b strings.Builder

	for i, name := range names {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(name)

		if i < len(weights) {
			b.WriteString(" (")
			b.WriteString(formatFloat(weights[i]))
			b.WriteString(")")
		}
	}

	return b.String()
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}
