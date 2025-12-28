package engine

import (
	"fmt"
	"io"
)

// PrettyExplainTo writes a deterministic, human-readable representation of the
// explanation tree to the provided writer. The traversal order is stable and
// matches the construction order of the tree.
func PrettyExplainTo(w io.Writer, e *Explain) {
	if e == nil {
		_, _ = io.WriteString(w, "<nil>\n")
		return
	}

	var walk func(n *Explain, indent string)

	walk = func(n *Explain, indent string) {
		label := n.ID
		if label == "" {
			label = n.Expr
		}
		// Line: "<indent>- <label> = <value> (<why>)\n" where (why) is optional
		if n.Why != "" {
			_, _ = io.WriteString(w, fmt.Sprintf("%s- %s = %v (%s)\n", indent, label, n.Value, n.Why))
		} else {
			_, _ = io.WriteString(w, fmt.Sprintf("%s- %s = %v\n", indent, label, n.Value))
		}

		for _, k := range n.Kids { // stable order: as stored
			walk(k, indent+"  ")
		}
	}
	walk(e, "")
}
