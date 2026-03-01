// Package variable provides linguistic variable definitions for fuzzy inference.
package variable

import fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

// Term is a named fuzzy set within a linguistic variable.
type Term struct {
	Name string
	Fn   fuzzym.MembershipFn
}

// Variable represents a linguistic variable with a domain range and fuzzy terms.
type Variable interface {
	// Name returns the variable identifier.
	Name() string
	// Min returns the lower bound of the domain.
	Min() float64
	// Max returns the upper bound of the domain.
	Max() float64
	// Terms returns all terms defined for this variable.
	Terms() []Term
	// Term returns the named term if it exists.
	Term(name string) (Term, bool)
	// Fuzzify evaluates all terms for the given crisp input.
	Fuzzify(x float64) map[string]fuzzym.Degree
	// Resolution returns the number of sampling points for defuzzification.
	Resolution() int
}

var _ Variable = (*variable)(nil)
