package sat

import "github.com/guidomantilla/yarumo/common/maths/logic/props"

// Lit represents a literal (a variable or its negation).
type Lit struct {
	Var props.Var
	Neg bool
}

func (l Lit) Negated() Lit { return Lit{Var: l.Var, Neg: !l.Neg} }

// Clause is a disjunction of literals.
type Clause []Lit

// CNF is a conjunction of clauses.
type CNF []Clause

// Assignment maps variables to boolean values.
type Assignment map[props.Var]bool
