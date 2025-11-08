package sets

// Relation represents a mathematical relation R ⊆ A × B, where A and B are sets.
type Relation[A comparable, B comparable] struct {
	pairs PairSet[A, B]
}

// NewRelation creates a new Relation instance with an empty PairSet.
func NewRelation[A comparable, B comparable]() *Relation[A, B] {
	pairSet := NewPairSet[A, B]()
	return &Relation[A, B]{pairs: *pairSet}
}

// Add the pair (a, b) to the relation R
func (r *Relation[A, B]) Add(a A, b B) {
	r.pairs.Add(Pair[A, B]{First: a, Second: b})
}

// Get retrieves the pair (a, b) from the relation R if it exists
func (r *Relation[A, B]) Get(a A, b B) (Pair[A, B], bool) {
	key := SerializePairSet(a, b)
	pair, exists := r.pairs.pairs[key]
	return pair, exists
}

// Contains checks if the relation R contains the pair (a, b): (a, b) ∈ R
func (r *Relation[A, B]) Contains(a A, b B) bool {
	return r.pairs.Contains(Pair[A, B]{First: a, Second: b})
}

// Remove deletes the pair (a, b) from the relation R
func (r *Relation[A, B]) Remove(a A, b B) {
	key := SerializePairSet(a, b)
	delete(r.pairs.pairs, key)
}

// Len returns the number of pairs in the relation R
func (r *Relation[A, B]) Len() int {
	return r.pairs.Len()
}

// Elements return a slice containing all pairs (a, b) in the relation R
func (r *Relation[A, B]) Elements() []Pair[A, B] {
	return r.pairs.Elements()
}

// Domain returns the set of all elements such a ∈ A such that there exists (a, b) ∈ R
func (r *Relation[A, B]) Domain() Set[A] {
	domain := New[A]()
	for _, p := range r.pairs.Elements() {
		domain.Add(p.First)
	}
	return domain
}

// Codomain returns the set of all elements b ∈ B such that there exists (a, b) ∈ R
func (r *Relation[A, B]) Codomain() Set[B] {
	codomain := New[B]()
	for _, p := range r.pairs.Elements() {
		codomain.Add(p.Second)
	}
	return codomain
}

// Image return the set of all elements b ∈ B such that (a, b) ∈ R for a given a ∈ A
func (r *Relation[A, B]) Image(a A) Set[B] {
	image := New[B]()
	for _, p := range r.pairs.Elements() {
		if p.First == a {
			image.Add(p.Second)
		}
	}
	return image
}

/*
type RelationOn[T comparable] = Relation[T, T]

func NewRelationOn[T comparable]() *RelationOn[T] {
	return &RelationOn[T]{pairs: *NewPairSet[T, T]()}
}

// IsReflexive checks if the relation R is reflexive: (a, a) ∈ R for all a ∈ A
func (r *RelationOn[T]) IsReflexive() bool {
	domain := r.Domain()
	for _, x := range domain.Elements() {
		if !r.Contains(x, x) {
			return false
		}
	}
	return true
}

// IsSymmetric checks if the relation R is symmetric: (a, b) ∈ R ⇒ (b, a) ∈ R
func (r *RelationOn[T]) IsSymmetric() bool {
	for _, p := range r.Elements() {
		if !r.Contains(p.Second, p.First) {
			return false
		}
	}
	return true
}

// IsAntisymmetric checks if the relation R is antisymmetric: (a, b) ∈ R and (b, a) ∈ R ⇒ a = b
func (r *RelationOn[T]) IsAntisymmetric() bool {
	for _, p := range r.Elements() {
		if p.First != p.Second && r.Contains(p.Second, p.First) {
			return false
		}
	}
	return true
}

// IsTransitive checks if the relation R is transitive: (a, b) ∈ R and (b, c) ∈ R ⇒ (a, c) ∈ R
func (r *RelationOn[T]) IsTransitive() bool {
	for _, p1 := range r.Elements() {
		for _, p2 := range r.Elements() {
			if p1.Second == p2.First {
				if !r.Contains(p1.First, p2.Second) {
					return false
				}
			}
		}
	}
	return true
}

// IsEquivalenceRelation checks if the relation R is an equivalence relation: reflex
func (r *Relation[T, T]) IsEquivalenceRelation() bool {
	return r.IsReflexive() && r.IsSymmetric() && r.IsTransitive()
}
*/
