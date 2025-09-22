package sat

import (
	"fmt"

	p "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

// FromFormulaToCNF converts any formula to flattened CNF (list of clauses)
// using Simplify + ToCNF and then extracting clauses/literals.
func FromFormulaToCNF(f p.Formula) (CNF, error) {
	f = p.Simplify(f)
	cnfF := p.ToCNF(f)
	return toClauses(cnfF)
}

// toClauses converts a CNF-shaped formula tree into a CNF structure.
// Accepts: And of clauses, single clause (Or of literals), or a literal/constant.
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
		return CNF{cl}, nil
	case p.NotF:
		if v, ok := x.F.(p.Var); ok {
			return CNF{Clause{{Var: v, Neg: true}}}, nil
		}
		return nil, fmt.Errorf("NOT inner is not a Var in CNF: %T", x.F)
	case p.Var:
		return CNF{Clause{{Var: x, Neg: false}}}, nil
	case p.TrueF:
		// True is neutral (conjunction of zero clauses)
		return CNF{}, nil
	case p.FalseF:
		// False is an empty clause (unsat)
		return CNF{Clause{}}, nil
	case p.GroupF:
		return toClauses(x.Inner)
	default:
		return nil, fmt.Errorf("formula is not in expected CNF structure: %T", f)
	}
}

// flattenOr flattens a disjunction of literals into a single clause.
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
		// Clause with True is tautological; represent as empty and let caller drop later
		return Clause{}, nil
	case p.FalseF:
		// False in disjunction can be ignored
		return Clause{}, nil
	case p.GroupF:
		return flattenOr(x.Inner)
	default:
		return nil, fmt.Errorf("non-literal in clause: %T", f)
	}
}
