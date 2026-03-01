package probability

import (
	"slices"
	"strconv"
	"strings"
)

// String returns a human-readable representation of a distribution.
func (d Distribution) String() string {
	if len(d) == 0 {
		return "{}"
	}

	keys := make([]Outcome, 0, len(d))

	for k := range d {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	var b strings.Builder

	b.WriteString("{")

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(k))
		b.WriteString("=")
		b.WriteString(strconv.FormatFloat(float64(d[k]), 'f', -1, 64))
	}

	b.WriteString("}")

	return b.String()
}

// String returns a human-readable representation of a CPT.
func (c CPT) String() string {
	var b strings.Builder

	b.WriteString("CPT(")
	b.WriteString(string(c.Variable))

	if len(c.Parents) > 0 {
		b.WriteString(" | ")

		for i, p := range c.Parents {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(string(p))
		}
	}

	b.WriteString(")")

	if len(c.Entries) > 0 {
		keys := make([]string, 0, len(c.Entries))

		for k := range c.Entries {
			keys = append(keys, k)
		}

		slices.Sort(keys)

		for _, k := range keys {
			b.WriteString("\n  ")

			if k != "" {
				b.WriteString(k)
				b.WriteString(": ")
			}

			b.WriteString(c.Entries[k].String())
		}
	}

	return b.String()
}

// String returns a human-readable representation of a factor.
func (f Factor) String() string {
	var b strings.Builder

	b.WriteString("Factor(")

	for i, v := range f.Variables {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(v))
	}

	b.WriteString(")")

	if len(f.Table) > 0 {
		keys := make([]string, 0, len(f.Table))

		for k := range f.Table {
			keys = append(keys, k)
		}

		slices.Sort(keys)

		for _, k := range keys {
			b.WriteString("\n  ")

			if k != "" {
				b.WriteString(k)
				b.WriteString(": ")
			}

			b.WriteString(strconv.FormatFloat(float64(f.Table[k]), 'f', -1, 64))
		}
	}

	return b.String()
}
