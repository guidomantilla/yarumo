package sets

// AxiomEmptySet verifica que el conjunto vacío está contenido en todo conjunto
func AxiomEmptySet[T comparable](s Set[T]) bool {
	return IsSubset(New[T](), s)
}

// AxiomExtensionality verifica que dos conjuntos son iguales si tienen los mismos elementos
func AxiomExtensionality[T comparable](a, b Set[T]) bool {
	return a.Cardinality() == b.Cardinality() && IsSubset(a, b)
}

// AxiomPairing verifica que existe un conjunto que contiene exactamente a x e y
func AxiomPairing[T comparable](x, y T) Set[T] {
	s := New[T]()
	s.Add(x)
	s.Add(y)
	return s
}

// AxiomUnion verifica que ⋃S contiene todos los elementos de los subconjuntos de S
func AxiomUnion[T comparable](sets []Set[T]) Set[T] {
	union := New[T]()
	for _, s := range sets {
		for _, e := range s.Elements() {
			union.Add(e)
		}
	}
	return union
}

// AxiomPowerSet verifica que PowerSet(A) contiene todos los subconjuntos de A
func AxiomPowerSet[T comparable](a Set[T]) *PowerSet[T] {
	return NewPowerSet(a)
}

// CommutativeUnion verifica A ∪ B = B ∪ A
func CommutativeUnion[T comparable](a, b Set[T]) bool {
	return Equal(Union(a, b), Union(b, a))
}

// CommutativeIntersection verifica A ∩ B = B ∩ A
func CommutativeIntersection[T comparable](a, b Set[T]) bool {
	return Equal(Intersection(a, b), Intersection(b, a))
}

// AssociativeUnion verifica (A ∪ B) ∪ C = A ∪ (B ∪ C)
func AssociativeUnion[T comparable](a, b, c Set[T]) bool {
	left := Union(Union(a, b), c)
	right := Union(a, Union(b, c))
	return Equal(left, right)
}

// DistributiveIntersectionOverUnion verifica A ∩ (B ∪ C) = (A ∩ B) ∪ (A ∩ C)
func DistributiveIntersectionOverUnion[T comparable](a, b, c Set[T]) bool {
	left := Intersection(a, Union(b, c))
	right := Union(Intersection(a, b), Intersection(a, c))
	return Equal(left, right)
}

// DoubleComplement verifica que (U \ (U \ A)) = A
func DoubleComplement[T comparable](u, a Set[T]) bool {
	return Equal(Complement(u, Complement(u, a)), a)
}

// DeMorganUnion verifica que U \ (A ∪ B) = (U \ A) ∩ (U \ B)
func DeMorganUnion[T comparable](u, a, b Set[T]) bool {
	left := Complement(u, Union(a, b))
	right := Intersection(Complement(u, a), Complement(u, b))
	return Equal(left, right)
}

// DeMorganIntersection verifica que U \ (A ∩ B) = (U \ A) ∪ (U \ B)
func DeMorganIntersection[T comparable](u, a, b Set[T]) bool {
	left := Complement(u, Intersection(a, b))
	right := Union(Complement(u, a), Complement(u, b))
	return Equal(left, right)
}
