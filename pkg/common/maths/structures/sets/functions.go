package sets

// Set is a generic set implemented as a map with empty struct values
// This gives O(1) lookup and set semantics
type Set[T comparable] map[T]struct{}

// Union returns the union of two sets: A ∪ B
func Union[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T])
	for k := range a {
		out[k] = struct{}{}
	}
	for k := range b {
		out[k] = struct{}{}
	}
	return out
}

// Intersection returns the intersection of two sets: A ∩ B
func Intersection[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T])
	for k := range a {
		if _, ok := b[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// Difference returns the difference of two sets: A - B
func Difference[T comparable](a, b Set[T]) Set[T] {
	out := make(Set[T])
	for k := range a {
		if _, ok := b[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

// Complement returns the complement of a set A relative to the universe U: \overline{A} = U - A
func Complement[T comparable](a, universe Set[T]) Set[T] {
	return Difference(universe, a)
}

// DeMorganUnion returns !(A ∪B) = !A ∩ !B
func DeMorganUnion[T comparable](a, b, universe Set[T]) Set[T] {
	return Intersection(Complement(a, universe), Complement(b, universe))
}

// DeMorganIntersection returns !(A ∩ B) = !A ∪ !B
func DeMorganIntersection[T comparable](a, b, universe Set[T]) Set[T] {
	return Union(Complement(a, universe), Complement(b, universe))
}

// AbsorptionUnion returns A ∪ (A ∩ B) = A
func AbsorptionUnion[T comparable](a, b Set[T]) Set[T] {
	return a
}

// AbsorptionIntersection returns A ∩ (A ∪ B) = A
func AbsorptionIntersection[T comparable](a, b Set[T]) Set[T] {
	return a
}

// IdempotentUnion returns A ∪ A = A
func IdempotentUnion[T comparable](a Set[T]) Set[T] {
	return a
}

// IdempotentIntersection returns A ∩ A = A
func IdempotentIntersection[T comparable](a Set[T]) Set[T] {
	return a
}

// DominationUnion returns A ∪ U = U
func DominationUnion[T comparable](a, universe Set[T]) Set[T] {
	return universe
}

// DominationIntersection returns A ∩ ∅ = ∅
func DominationIntersection[T comparable](a Set[T]) Set[T] {
	return make(Set[T])
}

// IdentityUnion returns A ∪ ∅ = A
func IdentityUnion[T comparable](a Set[T]) Set[T] {
	return a
}

// IdentityIntersection returns A ∩ U = A
func IdentityIntersection[T comparable](a, universe Set[T]) Set[T] {
	return a
}

// Equal returns true if two sets have the same elements
func Equal[T comparable](a, b Set[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// IsSubset returns true if a is a subset of b: A ⊆ B
func IsSubset[T comparable](a, b Set[T]) bool {
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// IsEmpty returns true if the set is empty
func IsEmpty[T comparable](s Set[T]) bool {
	return len(s) == 0
}

func ToSlice[T comparable](s Set[T]) []T {
	result := make([]T, 0, len(s))
	for k := range s {
		result = append(result, k)
	}
	return result
}
