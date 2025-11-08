package sat

import (
	"fmt"

	p "github.com/guidomantilla/yarumo/deprecated/logic/propositions"
)

// FromFormulaToCNF converts any formula to CNF (as a list of clauses) using
// the existing ToCNF transformation, after a light simplification.
func FromFormulaToCNF(f p.Formula) (CNF, error) {
	// Best-effort simplification before converting to clauses
	// (Simplify already exists in the project under propositions).
	f = p.Simplify(f)
	cnfF := p.ToCNF(f)
	return toClauses(cnfF)
}

// toClauses converts a CNF-structured formula into a CNF value.
// Accepts: And of clauses, single clause (Or of literals), or a literal.
func toClauses(f p.Formula) (CNF, error) {
	switch x := f.(type) {
	case p.AndF:
		l, err := toClauses(x.L)
		if err != nil {
			return nil, err
		}
		r, err := toClauses(x.R)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	case p.OrF:
		cl, err := flattenOr(x)
		if err != nil {
			return nil, err
		}
		// If the clause is tautological (contains A and ¬A), we can drop it later in DPLL.
		return CNF{cl}, nil
	case p.NotF:
		if v, ok := x.F.(p.Var); ok {
			return CNF{Clause{{Var: v, Neg: true}}}, nil
		}
		return nil, fmt.Errorf("NOT inner is not a Var in CNF: %T", x.F)
	case p.Var:
		return CNF{Clause{{Var: x, Neg: false}}}, nil
	case p.TrueF:
		// True in CNF is a conjunction with no clauses; neutral element.
		return CNF{}, nil
	case p.FalseF:
		// False in CNF is an empty clause in the conjunction (unsatisfiable).
		return CNF{Clause{}}, nil
	case p.GroupF:
		return toClauses(x.Inner)
	default:
		return nil, fmt.Errorf("formula is not in expected CNF structure: %T", f)
	}
}

// flattenOr flattens an Or tree of literals into a clause.
func flattenOr(f p.Formula) (Clause, error) {
	switch x := f.(type) {
	case p.OrF:
		l, err := flattenOr(x.L)
		if err != nil {
			return nil, err
		}
		r, err := flattenOr(x.R)
		if err != nil {
			return nil, err
		}
		return append(l, r...), nil
	case p.NotF:
		if v, ok := x.F.(p.Var); ok {
			return Clause{{Var: v, Neg: true}}, nil
		}
		return nil, fmt.Errorf("NOT inner is not a Var in clause: %T", x.F)
	case p.Var:
		return Clause{{Var: x, Neg: false}}, nil
	case p.TrueF:
		// A tautological literal can be represented as a special case: return empty and let caller handle
		return Clause{}, nil
	case p.FalseF:
		// Ignore False in disjunctions: (F ∨ A) == A
		return Clause{}, nil
	case p.GroupF:
		return flattenOr(x.Inner)
	default:
		return nil, fmt.Errorf("non-literal in clause: %T", f)
	}
}
