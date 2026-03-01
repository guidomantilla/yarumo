// Package sat provides a SAT solver based on the DPLL algorithm
// for propositional logic formulas in Conjunctive Normal Form.
package sat

import "github.com/guidomantilla/yarumo/maths/logic"

// Lit represents a literal: a propositional variable with polarity.
type Lit struct {
	V   logic.Var
	Neg bool
}

// Clause represents a disjunction (OR) of literals.
type Clause []Lit

// CNF represents a formula in Conjunctive Normal Form: a conjunction (AND) of clauses.
type CNF []Clause
