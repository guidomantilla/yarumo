package sets

// Add inserts items into the set.
func (s Set[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

// Remove deletes items from the set.
func (s Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

// Contains reports whether the set contains the given item.
func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

// Len returns the number of elements in the set.
func (s Set[T]) Len() int {
	return len(s)
}

// Items returns all elements of the set as a slice.
func (s Set[T]) Items() []T {
	result := make([]T, 0, len(s))

	for item := range s {
		result = append(result, item)
	}

	return result
}

// IsEmpty reports whether the set has no elements.
func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}

// Clone returns a shallow copy of the set.
func (s Set[T]) Clone() Set[T] {
	result := make(Set[T], len(s))

	for item := range s {
		result[item] = struct{}{}
	}

	return result
}

// Union returns a new set containing all elements from both sets.
func Union[T comparable](a, b Set[T]) Set[T] {
	result := a.Clone()

	for item := range b {
		result[item] = struct{}{}
	}

	return result
}

// Intersection returns a new set containing only elements present in both sets.
func Intersection[T comparable](a, b Set[T]) Set[T] {
	result := make(Set[T])

	for item := range a {
		if b.Contains(item) {
			result[item] = struct{}{}
		}
	}

	return result
}

// Difference returns a new set containing elements in a but not in b.
func Difference[T comparable](a, b Set[T]) Set[T] {
	result := make(Set[T])

	for item := range a {
		if !b.Contains(item) {
			result[item] = struct{}{}
		}
	}

	return result
}

// SymmetricDifference returns a new set with elements in either set but not both.
func SymmetricDifference[T comparable](a, b Set[T]) Set[T] {
	result := make(Set[T])

	for item := range a {
		if !b.Contains(item) {
			result[item] = struct{}{}
		}
	}

	for item := range b {
		if !a.Contains(item) {
			result[item] = struct{}{}
		}
	}

	return result
}

// IsSubset reports whether a is a subset of b.
func IsSubset[T comparable](a, b Set[T]) bool {
	for item := range a {
		if !b.Contains(item) {
			return false
		}
	}

	return true
}

// IsSuperset reports whether a is a superset of b.
func IsSuperset[T comparable](a, b Set[T]) bool {
	return IsSubset(b, a)
}

// Equal reports whether two sets contain exactly the same elements.
func Equal[T comparable](a, b Set[T]) bool {
	if len(a) != len(b) {
		return false
	}

	return IsSubset(a, b)
}
