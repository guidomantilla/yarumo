package variable

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"
)

type variable struct {
	name       string
	lo         float64
	hi         float64
	terms      []Term
	resolution int
}

// NewVariable creates a linguistic variable with the given name, domain range, and terms.
func NewVariable(name string, lo, hi float64, terms []Term, opts ...Option) Variable {
	cassert.NotEmpty(name, "variable name is empty")
	cassert.True(lo < hi, "variable min must be less than max")
	cassert.NotEmpty(terms, "variable terms are empty")

	options := NewOptions(opts...)

	copied := make([]Term, len(terms))
	copy(copied, terms)

	return &variable{
		name:       name,
		lo:         lo,
		hi:         hi,
		terms:      copied,
		resolution: options.resolution,
	}
}

// Name returns the variable identifier.
func (v *variable) Name() string {
	cassert.NotNil(v, "variable is nil")

	return v.name
}

// Min returns the lower bound of the domain.
func (v *variable) Min() float64 {
	cassert.NotNil(v, "variable is nil")

	return v.lo
}

// Max returns the upper bound of the domain.
func (v *variable) Max() float64 {
	cassert.NotNil(v, "variable is nil")

	return v.hi
}

// Terms returns all terms defined for this variable.
func (v *variable) Terms() []Term {
	cassert.NotNil(v, "variable is nil")

	copied := make([]Term, len(v.terms))
	copy(copied, v.terms)

	return copied
}

// Term returns the named term if it exists.
func (v *variable) Term(name string) (Term, bool) {
	cassert.NotNil(v, "variable is nil")

	for _, t := range v.terms {
		if t.Name == name {
			return t, true
		}
	}

	return Term{}, false
}

// Fuzzify evaluates all terms for the given crisp input.
func (v *variable) Fuzzify(x float64) map[string]fuzzym.Degree {
	cassert.NotNil(v, "variable is nil")

	result := make(map[string]fuzzym.Degree, len(v.terms))

	for _, t := range v.terms {
		result[t.Name] = t.Fn(x)
	}

	return result
}

// Resolution returns the number of sampling points for defuzzification.
func (v *variable) Resolution() int {
	cassert.NotNil(v, "variable is nil")

	return v.resolution
}
