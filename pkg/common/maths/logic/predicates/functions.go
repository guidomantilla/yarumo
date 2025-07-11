package predicates

type Predicate[T any] func(T) bool

// And returns a predicate that is the logical AND of two predicates
func And[T any](p1, p2 Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return p1(t) && p2(t)
	}
}

// Or returns a predicate that is the logical OR of two predicates
func Or[T any](p1, p2 Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return p1(t) || p2(t)
	}
}

// Not returns a predicate that is the logical NOT of the input predicate
func Not[T any](p Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return !p(t)
	}
}

// Implies returns a predicate that is logically equivalent to P ⇒ Q ≡ ¬P ∨ Q
func Implies[T any](p, q Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return !p(t) || q(t)
	}
}

// Contrapositive returns a predicate that is logically equivalent to P ⇒ Q ≡ ¬Q ⇒ ¬P
func Contrapositive[T any](p, q Predicate[T]) Predicate[T] {
	return Implies(Not(q), Not(p))
}

// Iff (if and only if) returns a predicate that is logically equivalent to P ⇔ Q ≡ (P ⇒ Q) ∧ (Q ⇒ P)
func Iff[T any](p, q Predicate[T]) Predicate[T] {
	return And(Implies(p, q), Implies(q, p))
}

// DeMorganAnd returns a predicate that is logically equivalent to ¬(P ∧ Q) ≡ ¬P ∨ ¬Q
func DeMorganAnd[T any](p, q Predicate[T]) Predicate[T] {
	return Or(Not(p), Not(q))
}

// DeMorganOr returns a predicate that is logically equivalent to ¬(P ∨ Q) ≡ ¬P ∧ ¬Q
func DeMorganOr[T any](p, q Predicate[T]) Predicate[T] {
	return And(Not(p), Not(q))
}

// Contradiction returns a predicate that is always false if both P and ¬P are true
func Contradiction[T any](p Predicate[T]) Predicate[T] {
	return And(p, Not(p))
}

// ExcludedMiddle returns a predicate that is always true: P ∨ ¬P
func ExcludedMiddle[T any](p Predicate[T]) Predicate[T] {
	return Or(p, Not(p))
}

// Filter applies a predicate to a slice and returns a new slice with elements that match the predicate
func Filter[T any](items []T, pred Predicate[T]) []T {
	var out []T
	for _, item := range items {
		if pred(item) {
			out = append(out, item)
		}
	}
	return out
}

// True returns a predicate that always returns true
func True[T any]() Predicate[T] {
	return func(T) bool { return true }
}

// False returns a predicate that always returns false
func False[T any]() Predicate[T] {
	return func(T) bool { return false }
}
