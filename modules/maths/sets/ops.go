package sets

// Union returns A ∪ B: all elements from both A and B
func Union[T comparable](a, b Set[T]) Set[T] {
	result := New[T]()
	for _, x := range a.Elements() {
		result.Add(x)
	}

	for _, y := range b.Elements() {
		result.Add(y)
	}

	return result
}

// Intersection returns A ∩ B: elements that are in both A and B
func Intersection[T comparable](a, b Set[T]) Set[T] {
	result := New[T]()

	for _, x := range a.Elements() {
		if b.Contains(x) {
			result.Add(x)
		}
	}

	return result
}

// Difference returns A \ B: elements in A that are not in B
func Difference[T comparable](a, b Set[T]) Set[T] {
	result := New[T]()

	for _, x := range a.Elements() {
		if !b.Contains(x) {
			result.Add(x)
		}
	}

	return result
}

// Complement devuelve el complemento de A respecto al universo U
func Complement[T comparable](universe, a Set[T]) Set[T] {
	return Difference(universe, a)
}

// Equal checks if two sets are equal: A = B if they have the same elements
func Equal[T comparable](a, b Set[T]) bool {
	if a.Cardinality() != b.Cardinality() {
		return false
	}

	for _, x := range a.Elements() {
		if !b.Contains(x) {
			return false
		}
	}

	return true
}

// Subsets return all subsets of A with exactly k elements
func Subsets[T comparable](A Set[T], k int) []Set[T] {
	elements := A.Elements()

	var (
		result []Set[T]
		comb   []T
	)

	var backtrack func(start int)

	backtrack = func(start int) {
		if len(comb) == k {
			result = append(result, New(comb...))
			return
		}

		for i := start; i < len(elements); i++ {
			comb = append(comb, elements[i])
			backtrack(i + 1)

			comb = comb[:len(comb)-1]
		}
	}

	backtrack(0)

	return result
}

// IsSubset checks if A ⊆ B: all elements of A are in B
func IsSubset[T comparable](a, b Set[T]) bool {
	for _, x := range a.Elements() {
		if !b.Contains(x) {
			return false
		}
	}

	return true
}

// IsProperSubset checks if A ⊂ B: A ⊆ B ∧ A ≠ B
func IsProperSubset[T comparable](a, b Set[T]) bool {
	return IsSubset(a, b) && !Equal(a, b)
}

// IsSuperset checks if A ⊇ B: all elements of B are in A
func IsSuperset[T comparable](a, b Set[T]) bool {
	return IsSubset(b, a)
}

// IsProperSuperset checks if A ⊃ B: A ⊇ B ∧ A ≠ B
func IsProperSuperset[T comparable](a, b Set[T]) bool {
	return IsProperSubset(b, a)
}

/*
sets/
├── set.go                 → definición de Set[T]
├── ops.go                 → unión, intersección, diferencia, igualdad, subconjuntos
├── universe.go            → universo y complemento con validación
├── product.go             → production cartesiano
├── power.go               → conjunto potencia
├── relation.go            → relaciones binarias y propiedades
├── function.go            → funciones, composición, inyecciones, etc.
├── axiomatics.go          → construcciones axiomáticas avanzadas (opcional)
└── util.go                → helpers para serialización, visualización, etc.
*/
