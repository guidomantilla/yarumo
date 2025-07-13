package sets

import "fmt"

// PartitionFromRelation generates a partition of the domain of the relation R into equivalence classes.
func PartitionFromRelation[T comparable](r *Relation[T, T]) ([]Set[T], error) {
	if !r.IsEquivalenceRelation() {
		return nil, fmt.Errorf("relation is not an equivalence relation")
	}

	visited := make(map[T]bool)
	var classes []Set[T]

	elements := r.Domain()
	for _, a := range elements.Elements() {
		if visited[a] {
			continue
		}

		class := New[T]()
		for _, b := range elements.Elements() {
			if r.Contains(a, b) {
				class.Add(b)
				visited[b] = true
			}
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// RelationFromPartition generates a relation from a partition of equivalence classes.
func RelationFromPartition[T comparable](partition []Set[T]) *Relation[T, T] {
	relOn := NewRelationOn[T]()
	rel := (*Relation[T, T])(relOn)

	for _, class := range partition {
		elements := class.Elements()
		for i := 0; i < len(elements); i++ {
			for j := 0; j < len(elements); j++ {
				rel.Add(elements[i], elements[j])
			}
		}
	}

	return rel
}
