package sets

// NewPowerSet generates the power set of a given set a.
func NewPowerSet[T comparable](a Set[T]) *PowerSet[T] {
	elements := a.Elements()
	n := len(elements)

	ps := &PowerSet[T]{subsets: make(map[string]Set[T])}

	m := 1 << n
	for i := range m {
		sub := New[T]()

		for j := range n {
			if i&(1<<j) != 0 {
				sub.Add(elements[j])
			}
		}

		ps.Add(sub)
	}

	return ps
}

// PowerSet represents a set of all subsets of a given set.
type PowerSet[T comparable] struct {
	subsets map[string]Set[T]
}

// Add a new subset to the power set.
func (ps *PowerSet[T]) Add(sub Set[T]) {
	key := SerializeSet(sub)
	ps.subsets[key] = sub
}

// Get retrieves a specific subset from the power set by its serialized key.
func (ps *PowerSet[T]) Get(sub Set[T]) (Set[T], bool) {
	key := SerializeSet(sub)
	s, ok := ps.subsets[key]

	return s, ok
}

// Contains checks if the power set contains a specific subset.
func (ps *PowerSet[T]) Contains(sub Set[T]) bool {
	key := SerializeSet(sub)
	_, ok := ps.subsets[key]

	return ok
}

// Len returns the number of unique subsets in the power set.
func (ps *PowerSet[T]) Len() int {
	return len(ps.subsets)
}

// Elements return a slice containing all unique subsets in the power set.
func (ps *PowerSet[T]) Elements() []Set[T] {
	result := make([]Set[T], 0, len(ps.subsets))
	for _, s := range ps.subsets {
		result = append(result, s)
	}

	return result
}

// FilterByCardinality returns a new PowerSet containing only subsets with exactly k elements.
func (ps *PowerSet[T]) FilterByCardinality(k int) *PowerSet[T] {
	filtered := &PowerSet[T]{subsets: make(map[string]Set[T])}

	for _, subset := range ps.Elements() {
		if subset.Cardinality() == k {
			filtered.Add(subset)
		}
	}

	return filtered
}

// FilterByRange returns a new PowerSet containing only subsets with a cardinality within the specified range [min, max].
func (ps *PowerSet[T]) FilterByRange(min, max int) *PowerSet[T] {
	filtered := &PowerSet[T]{subsets: make(map[string]Set[T])}

	for _, subset := range ps.Elements() {
		card := subset.Cardinality()
		if card >= min && card <= max {
			filtered.Add(subset)
		}
	}

	return filtered
}
