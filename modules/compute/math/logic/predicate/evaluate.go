package predicate

import (
	"slices"

	"github.com/guidomantilla/yarumo/compute/math/logic"
)

// ForAll returns true if predicate evaluates to true for every element in the collection.
// Returns true for an empty collection (vacuous truth).
func ForAll(collection Collection, predicate logic.Formula) (bool, error) {
	if predicate == nil {
		return false, ErrNilPredicate
	}

	for _, element := range collection {
		if !predicate.Eval(element) {
			return false, nil
		}
	}

	return true, nil
}

// Exists returns true if predicate evaluates to true for at least one element.
// Returns false for an empty collection (vacuous falsity).
func Exists(collection Collection, predicate logic.Formula) (bool, error) {
	if predicate == nil {
		return false, ErrNilPredicate
	}

	if len(collection) == 0 {
		return false, nil
	}

	return slices.ContainsFunc(collection, predicate.Eval), nil
}

// Count returns the number of elements satisfying the predicate.
// Returns 0 for an empty collection.
func Count(collection Collection, predicate logic.Formula) (int, error) {
	if predicate == nil {
		return 0, ErrNilPredicate
	}

	n := 0

	for _, element := range collection {
		if predicate.Eval(element) {
			n++
		}
	}

	return n, nil
}

// Filter returns elements satisfying the predicate.
// Returns nil for an empty collection.
func Filter(collection Collection, predicate logic.Formula) ([]logic.Fact, error) {
	if predicate == nil {
		return nil, ErrNilPredicate
	}

	var result []logic.Fact

	for _, element := range collection {
		if predicate.Eval(element) {
			result = append(result, element)
		}
	}

	return result, nil
}
