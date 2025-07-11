package bags

// Bag is a multiset of elements of type T, storing element counts
// For example: {a: 2, b: 1, c: 3} represents multiset {a, a, b, c, c, c}
type Bag[T comparable] map[T]int

// Add inserts one or more occurrences of an item
func Add[T comparable](bag Bag[T], item T, count int) Bag[T] {
	if count <= 0 {
		return bag
	}
	if _, ok := bag[item]; !ok {
		bag[item] = 0
	}
	bag[item] += count
	return bag
}

// Remove removes up to 'count' occurrences of an item
func Remove[T comparable](bag Bag[T], item T, count int) Bag[T] {
	if current, ok := bag[item]; ok {
		if count >= current {
			delete(bag, item)
		} else {
			bag[item] -= count
		}
	}
	return bag
}

// Count returns the number of times an item appears in the bag
func Count[T comparable](bag Bag[T], item T) int {
	if count, ok := bag[item]; ok {
		return count
	}
	return 0
}

// Size returns the total number of elements in the bag (including duplicates)
func Size[T comparable](b Bag[T]) int {
	size := 0
	for _, count := range b {
		size += count
	}
	return size
}

// UniqueSize returns the number of distinct elements in the bag
func UniqueSize[T comparable](b Bag[T]) int {
	if b == nil {
		return 0
	}
	return len(b)
}

// Union returns a new bag with the sum of counts from both bags
func Union[T comparable](a, b Bag[T]) Bag[T] {
	out := make(Bag[T])
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] += v
	}
	return out
}

// Intersection returns a new bag with the minimum counts from both bags
func Intersection[T comparable](a, b Bag[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range a {
		if otherCount, ok := b[item]; ok {
			out[item] = min(count, otherCount)
		}
	}
	return out
}

// Difference returns a new bag with counts from a minus b, clamped at zero
func Difference[T comparable](a, b Bag[T]) Bag[T] {
	out := make(Bag[T])
	for k, vA := range a {
		vB := b[k]
		if diff := vA - vB; diff > 0 {
			out[k] = diff
		}
	}
	return out
}

// Complement returns the complement of a in the given universe: universe - a
func Complement[T comparable](a, universe Bag[T]) Bag[T] {
	return Difference(universe, a)
}

// IsSubset returns true if all counts in a are less than or equal to those in b
func IsSubset[T comparable](a, b Bag[T]) bool {
	for k, vA := range a {
		if vB := b[k]; vA > vB {
			return false
		}
	}
	return true
}

// Equal returns true if two bags contain the same element counts
func Equal[T comparable](a, b Bag[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func ToSlice[T comparable](b Bag[T]) []T {
	result := make([]T, 0, Size(b))
	for item, count := range b {
		for i := 0; i < count; i++ {
			result = append(result, item)
		}
	}
	return result
}
