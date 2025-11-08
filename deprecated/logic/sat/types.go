package sat

import "github.com/guidomantilla/yarumo/deprecated/logic/propositions"

// Lit represents a literal: a (possibly negated) propositional variable.
type Lit struct {
	Var propositions.Var
	Neg bool
}

func (l Lit) Negated() Lit { return Lit{Var: l.Var, Neg: !l.Neg} }

// Clause is a disjunction of literals.
type Clause []Lit

// CNF is a conjunction of clauses.
type CNF []Clause

// Assignment maps variables to boolean values.
type Assignment map[propositions.Var]bool
