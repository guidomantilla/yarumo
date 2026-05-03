// Package logic provides propositional logic primitives: formulas, evaluation,
// transformations, simplification, and satisfiability analysis.
package logic

// Var is a propositional variable.
type Var string

// Fact is a partial truth assignment mapping variables to boolean values.
type Fact map[Var]bool

// Formula defines the interface for propositional formulas.
type Formula interface {
	// String returns the canonical string representation of the formula.
	String() string
	// Eval evaluates the formula against the given fact assignment.
	Eval(facts Fact) bool
	// Vars returns the sorted, deduplicated list of variables in the formula.
	Vars() []Var
}
