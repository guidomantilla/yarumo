// Package sets provides a generic set data structure and common set operations.
package sets

// Set is a collection of unique elements.
type Set[T comparable] map[T]struct{}

// New creates a set containing the given items.
func New[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))

	for _, item := range items {
		s[item] = struct{}{}
	}

	return s
}
