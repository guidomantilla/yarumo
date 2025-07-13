package sets

// CartesianProduct returns the Cartesian product of two sets A and B: A × B
func CartesianProduct[A comparable, B comparable](a Set[A], b Set[B]) *PairSet[A, B] {
	result := &PairSet[A, B]{pairs: make(map[string]Pair[A, B])}
	for _, x := range a.Elements() {
		for _, y := range b.Elements() {
			result.Add(Pair[A, B]{First: x, Second: y})
		}
	}
	return result
}

// Pair represents a pair of elements, where First is of type A and Second is of type B.
type Pair[A, B any] struct {
	First  A
	Second B
}

// PairSet represents a set of unique pairs, where each pair is a combination of elements from two sets A and B.
type PairSet[A comparable, B comparable] struct {
	pairs map[string]Pair[A, B]
}

func NewPairSet[A comparable, B comparable]() *PairSet[A, B] {
	return &PairSet[A, B]{pairs: make(map[string]Pair[A, B])}
}

// Add a new pair to the set if it is not already present.
func (ps *PairSet[A, B]) Add(pair Pair[A, B]) {
	key := SerializePairSet(pair.First, pair.Second)
	ps.pairs[key] = pair
}

// Get retrieves a specific pair from the set by its elements.
func (ps *PairSet[A, B]) Get(first A, second B) (Pair[A, B], bool) {
	key := SerializePairSet(first, second)
	pair, exists := ps.pairs[key]
	return pair, exists
}

// Contains checks if the set contains a specific pair.
func (ps *PairSet[A, B]) Contains(pair Pair[A, B]) bool {
	_, ok := ps.pairs[SerializePairSet(pair.First, pair.Second)]
	return ok
}

// Len returns the number of unique pairs in the set.
func (ps *PairSet[A, B]) Len() int {
	return len(ps.pairs)
}

// Elements return a slice containing all unique pairs in the set.
func (ps *PairSet[A, B]) Elements() []Pair[A, B] {
	result := make([]Pair[A, B], 0, len(ps.pairs))
	for _, pair := range ps.pairs {
		result = append(result, pair)
	}
	return result
}

/*
 *
 */

// CartesianPower returns Aⁿ: the set of all n-tuples where each element is from set A
func CartesianPower[T comparable](a Set[T], n int) *TupleSet[T] {
	result := &TupleSet[T]{tuples: make(map[string][]T)}

	if n == 0 {
		result.Add([]T{}) // Solo la tupla vacía
		return result
	}

	elements := a.Elements()

	var build func(prefix []T, depth int)
	build = func(prefix []T, depth int) {
		if depth == 0 {
			result.Add(prefix)
			return
		}
		for _, x := range elements {
			build(append(prefix, x), depth-1)
		}
	}

	build([]T{}, n)
	return result
}

// TupleSet represents a set of unique tuples, where each tuple is a slice of elements of type T.
type TupleSet[T comparable] struct {
	tuples map[string][]T
}

// Add a new tuple to the set if it is not already present.
func (ts *TupleSet[T]) Add(tuple []T) {
	key := SerializeTupleSet(tuple)
	if _, exists := ts.tuples[key]; !exists {
		cpy := make([]T, len(tuple))
		copy(cpy, tuple)
		ts.tuples[key] = cpy
	}
}

// Get retrieves a specific tuple from the set by its elements.
func (ts *TupleSet[T]) Get(tuple []T) ([]T, bool) {
	key := SerializeTupleSet(tuple)
	t, exists := ts.tuples[key]
	if !exists {
		return nil, false
	}
	// Return a copy of the tuple to avoid external modifications
	cpy := make([]T, len(t))
	copy(cpy, t)
	return cpy, true
}

// Contains checks if the set contains a specific tuple.
func (ts *TupleSet[T]) Contains(tuple []T) bool {
	_, ok := ts.tuples[SerializeTupleSet(tuple)]
	return ok
}

// Len return the number of unique tuples in the set.
func (ts *TupleSet[T]) Len() int {
	return len(ts.tuples)
}

// Elements return a slice containing all unique tuples in the set.
func (ts *TupleSet[T]) Elements() [][]T {
	out := make([][]T, 0, len(ts.tuples))
	for _, t := range ts.tuples {
		out = append(out, t)
	}
	return out
}
