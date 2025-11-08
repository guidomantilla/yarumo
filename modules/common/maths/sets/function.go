package sets

import "fmt"

// Function representa una función f: A → B
type Function[A comparable, B comparable] struct {
	rel *Relation[A, B]
}

// NewFunction crea una función vacía
func NewFunction[A comparable, B comparable]() *Function[A, B] {
	return &Function[A, B]{rel: NewRelation[A, B]()}
}

// Set define f(a) = b. Si a ya existía, reemplaza su valor.
func (f *Function[A, B]) Set(a A, b B) {
	// Eliminar cualquier (a, _) existente
	for _, p := range f.rel.Elements() {
		if p.First == a {
			f.rel.Remove(p.First, p.Second)
		}
	}
	f.rel.Add(a, b)
}

// Get the value b such that f(a) = b.
func (f *Function[A, B]) Get(a A) (B, bool) {
	for _, p := range f.rel.Elements() {
		if p.First == a {
			return p.Second, true
		}
	}
	var zero B
	return zero, false
}

// Domain returns the domain (all a such that there exists b with f(a) = b)
func (f *Function[A, B]) Domain() Set[A] {
	return f.rel.Domain()
}

// Codomain returns the codomain (all b such that there exists an a with f(a) = b)
func (f *Function[A, B]) Codomain() Set[B] {
	return f.rel.Codomain()
}

// Image returns the set of all elements b ∈ B such that f(a) = b for a given a ∈ A
func (f *Function[A, B]) Image(a A) Set[B] {
	image := New[B]()
	if b, ok := f.Get(a); ok {
		image.Add(b)
	}
	return image
}

/*
 *
 */

// FunctionFromSet represents a function f: A → B defined on specific sets domain A and codomain B.
type FunctionFromSet[A comparable, B comparable] struct {
	function    *Function[A, B]
	domainSet   Set[A] // conjunto A
	codomainSet Set[B] // conjunto B declarado (para sobreyectividad)
}

// NewFunctionFromSet creates a new function defined on the given sets domain and codomain.
func NewFunctionFromSet[A comparable, B comparable](domain Set[A], codomain Set[B]) *FunctionFromSet[A, B] {
	return &FunctionFromSet[A, B]{
		function:    NewFunction[A, B](),
		domainSet:   domain,
		codomainSet: codomain,
	}
}

// Set defines f(a) = b. Requires that  ∈ A y b ∈ B. If a is not in the domain, it returns an error.
func (f *FunctionFromSet[A, B]) Set(a A, b B) error {
	if !f.domainSet.Contains(a) {
		return fmt.Errorf("a = %v not in domain", a)
	}
	if !f.codomainSet.Contains(b) {
		return fmt.Errorf("b = %v not in codomain", b)
	}
	f.function.Set(a, b)
	return nil
}

// Get returns f(a)
func (f *FunctionFromSet[A, B]) Get(a A) (B, bool) {
	return f.function.Get(a)
}

// IsTotal checks if the function is total, meaning it is defined for every element in the domain set.
func (f *FunctionFromSet[A, B]) IsTotal() bool {
	for _, a := range f.domainSet.Elements() {
		if _, ok := f.function.Get(a); !ok {
			return false
		}
	}
	return true
}

// IsInjective checks if the function is injective, meaning that for every b ∈ codomainSet, there is at most one a ∈ domainSet such that f(a) = b.
//
// f(a₁) = f(a₂) ⇒ a₁ = a₂
func (f *FunctionFromSet[A, B]) IsInjective() bool {
	inverse := make(map[B]A)
	for _, a := range f.domainSet.Elements() {
		b, ok := f.function.Get(a)
		if !ok {
			continue
		}
		if prevA, seen := inverse[b]; seen && prevA != a {
			return false
		}
		inverse[b] = a
	}
	return true
}

// IsSurjective checks if the function is surjective, meaning that for every b ∈ codomainSet, there exists at least one a ∈ domainSet such that f(a) = b.
func (f *FunctionFromSet[A, B]) IsSurjective() bool {
	seen := New[B]()
	for _, a := range f.domainSet.Elements() {
		b, ok := f.function.Get(a)
		if ok {
			seen.Add(b)
		}
	}
	return IsSubset(f.codomainSet, seen)
}

// IsBijective ⇔ Injective ∧ Surjective
func (f *FunctionFromSet[A, B]) IsBijective() bool {
	return f.IsInjective() && f.IsSurjective()
}

// Inverse returns f⁻¹: B → A, if f is Bijective
func (f *FunctionFromSet[A, B]) Inverse() (*FunctionFromSet[B, A], error) {
	if !f.IsBijective() {
		return nil, fmt.Errorf("function is not bijective")
	}

	inverse := NewFunctionFromSet[B, A](f.codomainSet, f.domainSet)

	for _, a := range f.domainSet.Elements() {
		b, ok := f.Get(a)
		if !ok {
			return nil, fmt.Errorf("function is not total")
		}
		if err := inverse.Set(b, a); err != nil {
			return nil, err
		}
	}

	return inverse, nil
}

// IsLeftInverse checks that (f⁻¹ ∘ f)(a) = a ∀a ∈ A
func (f *FunctionFromSet[A, B]) IsLeftInverse(inv *FunctionFromSet[B, A]) bool {
	for _, a := range f.domainSet.Elements() {
		b, ok1 := f.Get(a)
		if !ok1 {
			return false
		}
		a2, ok2 := inv.Get(b)
		if !ok2 || a2 != a {
			return false
		}
	}
	return true
}

// IsRightInverse checks that (f ∘ f⁻¹)(b) = b ∀b ∈ B
func (f *FunctionFromSet[A, B]) IsRightInverse(inv *FunctionFromSet[B, A]) bool {
	for _, b := range f.codomainSet.Elements() {
		a, ok1 := inv.Get(b)
		if !ok1 {
			return false
		}
		b2, ok2 := f.Get(a)
		if !ok2 || b2 != b {
			return false
		}
	}
	return true
}

/*
 *
 */

// Compose takes f: A → B and g: B → C and returns g∘f: A → C
func Compose[A comparable, B comparable, C comparable](f *FunctionFromSet[A, B], g *FunctionFromSet[B, C]) (*FunctionFromSet[A, C], error) {
	if !IsSubset(f.codomainSet, g.domainSet) {
		return nil, fmt.Errorf("codomain of f is not contained in domain of g")
	}

	composed := NewFunctionFromSet[A, C](f.domainSet, g.codomainSet)
	for _, a := range f.domainSet.Elements() {
		b, ok1 := f.Get(a)
		if !ok1 {
			return nil, fmt.Errorf("f is not total on %v", a)
		}
		c, ok2 := g.Get(b)
		if !ok2 {
			return nil, fmt.Errorf("g is not defined at %v", b)
		}
		if err := composed.Set(a, c); err != nil {
			return nil, err
		}
	}

	return composed, nil
}
