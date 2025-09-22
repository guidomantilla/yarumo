package sat

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"

// Lit represents a literal (a variable or its negation).
type Lit struct {
	Var props.Var
	Neg bool
}

// Clause is a disjunction of literals.
type Clause []Lit

// CNF is a conjunction of clauses.
type CNF []Clause

// Assignment maps variables to boolean values.
type Assignment map[props.Var]bool

// FromFormulaToCNF converts a formula into a flattened CNF.
// Phase 0: stub that returns an error until Phase 2.
func FromFormulaToCNF(f props.Formula) (CNF, error) { return nil, ErrNotImplemented }

// DPLL solves satisfiability over a CNF.
// Phase 0: stub that returns false, nil until Phase 2.
func DPLL(cnf CNF, asg Assignment) (bool, Assignment) { return false, nil }
