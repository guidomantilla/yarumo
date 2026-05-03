package graph

import (
	"strconv"
	"strings"
)

// String returns a human-readable representation of a Node.
func (n Node) String() string {
	return "Node(" + n.ID + ")"
}

// String returns a human-readable representation of an Edge.
func (e Edge) String() string {
	var b strings.Builder

	b.WriteString("Edge(")
	b.WriteString(e.ID)
	b.WriteString(": ")
	b.WriteString(e.From)
	b.WriteString(" -> ")
	b.WriteString(e.To)

	if e.Weight != 1.0 {
		b.WriteString(", w=")
		b.WriteString(strconv.FormatFloat(e.Weight, 'g', -1, 64))
	}

	if e.Label != "" {
		b.WriteString(", ")
		b.WriteString(e.Label)
	}

	b.WriteString(")")

	return b.String()
}

// String returns a human-readable representation of a Path.
func (p Path) String() string {
	return "Path(" + strings.Join(p.Nodes, " -> ") +
		", w=" + strconv.FormatFloat(p.Weight, 'g', -1, 64) + ")"
}
