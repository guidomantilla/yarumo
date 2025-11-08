package sets

// U represents a universe set that contains all elements of type T.
type U[T comparable] struct {
	Set[T]
}

// NewU creates a new universe set from the provided set.
func NewU[T comparable](a Set[T]) U[T] {
	return U[T]{Set: a}
}

// ValidateSubset checks if the provided subset is a valid subset of the universe set.
//
// checks if A ⊆ U: all elements of A are in U
func (u U[T]) ValidateSubset(a Set[T]) bool {
	return IsSubset(a, u.Set)
}

// Complement returns the complement of the provided subset in the universe set.
//
// The complement is defined as U \ A: all elements in U that are not in A.
func (u U[T]) Complement(a Set[T]) Set[T] {
	return Difference(u.Set, a)
}
