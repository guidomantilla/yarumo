package sets

import "github.com/guidomantilla/yarumo/pkg/common/pointer"

// Set is a generic set data structure that holds unique elements of type T.
type Set[T comparable] struct {
	elements map[T]struct{}
}

// New creates a new Set instance and initializes it with the provided items.
func New[T comparable](items ...T) Set[T] {
	s := Set[T]{elements: make(map[T]struct{}, len(items))}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

// Add adds an item to the set.
func (s *Set[T]) Add(item T) {
	s.elements[item] = struct{}{}
}

// Get retrieves an item from the set. If the item exists, it returns the item and true; otherwise, it returns a zero value of type T and false.
func (s *Set[T]) Get(item T) (T, bool) {
	_, ok := s.elements[item]
	if !ok {
		return pointer.Zero[T](), false
	}
	return item, true
}

// Remove deletes an item from the set. If the item does not exist, it does nothing.
func (s *Set[T]) Remove(item T) {
	delete(s.elements, item)
}

// Contains checks if the set contains the specified item. It returns true if the item is present, otherwise false.
func (s *Set[T]) Contains(item T) bool {
	_, ok := s.elements[item]
	return ok
}

// Cardinality returns the number of unique elements in the set.
func (s *Set[T]) Cardinality() int {
	return len(s.elements)
}

// Elements return a slice containing all unique elements in the set.
func (s *Set[T]) Elements() []T {
	result := make([]T, 0, len(s.elements))
	for item := range s.elements {
		result = append(result, item)
	}
	return result
}
