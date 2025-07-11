package maths

// Predicate is a function that evaluates a value of type T and returns true or false
type Predicate[T any] func(T) bool

// And returns a predicate that is the logical AND of two predicates
func (p Predicate[T]) And(other Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return p(t) && other(t)
	}
}

// Or returns a predicate that is the logical OR of two predicates
func (p Predicate[T]) Or(other Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return p(t) || other(t)
	}
}

// Not returns a predicate that is the logical NOT of the input predicate
func (p Predicate[T]) Not() Predicate[T] {
	return func(t T) bool {
		return !p(t)
	}
}

// Implies returns a predicate that is logically equivalent to P ⇒ Q ≡ ¬P ∨ Q
func (p Predicate[T]) Implies(q Predicate[T]) Predicate[T] {
	return func(t T) bool {
		return !p(t) || q(t)
	}
}

// Contrapositive returns a predicate that is logically equivalent to P ⇒ Q ≡ ¬Q ⇒ ¬P
func (p Predicate[T]) Contrapositive(q Predicate[T]) Predicate[T] {
	return q.Not().Implies(p.Not())
}

// Iff (if and only if) returns a predicate that is logically equivalent to P ⇔ Q ≡ (P ⇒ Q) ∧ (Q ⇒ P)
func (p Predicate[T]) Iff(q Predicate[T]) Predicate[T] {
	return p.Implies(q).And(q.Implies(p))
}
