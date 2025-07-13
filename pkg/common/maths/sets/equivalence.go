package sets

/*
// Equivalence represents an equivalence relation on a set of type T.
type Equivalence[T comparable] struct {
	relation  *Relation[T, T]
	partition []Set[T]
}

// NewEquivalenceFromRelation creates a new Equivalence instance from a Relation.
func NewEquivalenceFromRelation[T comparable](r *Relation[T, T]) (*Equivalence[T], error) {
	if !r.IsEquivalenceRelation() {
		return nil, fmt.Errorf("relation is not an equivalence relation")
	}
	partition, err := PartitionFromRelation(r)
	if err != nil {
		return nil, err
	}
	return &Equivalence[T]{relation: r, partition: partition}, nil
}

// NewEquivalenceFromPartition crates a new Equivalence instance from a partition of sets.
func NewEquivalenceFromPartition[T comparable](p []Set[T]) *Equivalence[T] {
	r := RelationFromPartition(p)
	return &Equivalence[T]{relation: r, partition: p}
}

// Relation returns the equivalence relation R ⊆ A × A
func (e *Equivalence[T]) Relation() *Relation[T, T] {
	return e.relation
}

// Partition returns the partition of sets that represents the equivalence classes.
func (e *Equivalence[T]) Partition() []Set[T] {
	return e.partition
}

// ClassOf returns the equivalence class of an element x.
func (e *Equivalence[T]) ClassOf(x T) (Set[T], bool) {
	for _, class := range e.partition {
		if class.Contains(x) {
			return class, true
		}
	}
	return pointer.Zero[Set[T]](), false
}
*/
