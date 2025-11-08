package sets

import (
	"fmt"
	"slices"
	"strings"
)

// SliceToStrings converts a slice of any comparable type to a slice of strings.
func SliceToStrings[T comparable](items []T) []string {
	out := make([]string, len(items))
	for i, v := range items {
		out[i] = fmt.Sprintf("%v", v)
	}
	return out
}

// SerializePairSet creates a unique string representation of a pair (a, b).
func SerializePairSet[A comparable, B comparable](a A, b B) string {
	return fmt.Sprintf("%v|%v", a, b)
}

// SerializeTupleSet creates a unique string representation of a tuple.
func SerializeTupleSet[T comparable](tuple []T) string {
	var sb strings.Builder
	for i, v := range tuple {
		if i > 0 {
			sb.WriteString("|")
		}
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	return sb.String()
}

// SerializeSet converts a Set[T] to a string representation.
func SerializeSet[T comparable](s Set[T]) string {
	elements := s.Elements()
	serialized := SliceToStrings(elements)
	slices.Sort(serialized)
	return fmt.Sprintf("[%s]", strings.Join(serialized, ","))
}
