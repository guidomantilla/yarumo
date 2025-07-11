package maths

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"

// Set is a generic set implemented as a map with empty struct values
// This gives O(1) lookup and set semantics
type Set[T comparable] map[T]struct{}

// NewSet creates a new set from a list of elements
func NewSet[T comparable](elements ...T) Set[T] {
	m := make(Set[T])
	for _, e := range elements {
		m[e] = struct{}{}
	}
	return m
}

func (s Set[T]) Union(other Set[T]) Set[T] {
	out := make(Set[T])
	for k := range s {
		out[k] = struct{}{}
	}
	for k := range other {
		out[k] = struct{}{}
	}
	return out
}

func (s Set[T]) Intersection(other Set[T]) Set[T] {
	out := make(Set[T])
	for k := range s {
		if _, ok := other[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func (s Set[T]) Difference(other Set[T]) Set[T] {
	out := make(Set[T])
	for k := range s {
		if _, ok := other[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func (s Set[T]) Complement(universe Set[T]) Set[T] {
	return universe.Difference(s)
}

func (s Set[T]) IsSubset(other Set[T]) bool {
	for k := range other {
		if _, ok := s[k]; !ok {
			return false
		}
	}
	return true
}

func (s Set[T]) Size() int {
	return len(s)
}

func (s Set[T]) Filter(pred predicates.Predicate[T]) Set[T] {
	out := make(Set[T])
	for k := range s {
		if pred(k) {
			out[k] = struct{}{}
		}
	}
	return out
}

func (s Set[T]) Equal(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

func (s Set[T]) Empty() bool {
	return len(s) == 0
}

func (s Set[T]) ToSlice() []T {
	out := make([]T, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return out
}
