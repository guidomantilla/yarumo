package sets

import (
	"cmp"
	"slices"
	"strings"
)

// String returns a sorted string representation of the set: {a, b, c}.
func (s Set[T]) String() string {
	if len(s) == 0 {
		return "{}"
	}

	items := s.Items()

	strs := make([]string, len(items))

	for i, item := range items {
		strs[i] = stringify(item)
	}

	slices.Sort(strs)

	var b strings.Builder

	b.WriteString("{")

	for i, str := range strs {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(str)
	}

	b.WriteString("}")

	return b.String()
}

func stringify[T comparable](v T) string {
	s, ok := any(v).(interface{ String() string })
	if ok {
		return s.String()
	}

	return cmp.Or(stringifyBasic(v), "?")
}
