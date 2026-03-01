package sat

import "github.com/guidomantilla/yarumo/maths/logic"

// FromFormula converts a logic.Formula into a sat.CNF structure.
// The formula should be in CNF form (via logic.ToCNF) for correct results.
func FromFormula(f logic.Formula) CNF {
	var clauses CNF

	collectClauses(f, &clauses)

	return clauses
}

func collectClauses(f logic.Formula, clauses *CNF) {
	switch v := f.(type) {
	case logic.AndF:
		collectClauses(v.L, clauses)
		collectClauses(v.R, clauses)
	case logic.TrueF:
		// True as a top-level conjunct is trivially satisfied; skip.
	case logic.FalseF:
		// False as a top-level conjunct produces an empty clause (unsatisfiable).
		*clauses = append(*clauses, Clause{})
	default:
		clause, tautology := buildClause(f)
		if !tautology {
			*clauses = append(*clauses, clause)
		}
	}
}

func buildClause(f logic.Formula) (Clause, bool) {
	var lits Clause

	taut := gatherLiterals(f, &lits)

	return lits, taut
}

func gatherLiterals(f logic.Formula, lits *Clause) bool {
	switch v := f.(type) {
	case logic.OrF:
		if gatherLiterals(v.L, lits) {
			return true
		}

		return gatherLiterals(v.R, lits)
	case logic.Var:
		*lits = append(*lits, Lit{V: v})
	case logic.NotF:
		variable, ok := v.F.(logic.Var)
		if ok {
			*lits = append(*lits, Lit{V: variable, Neg: true})
		}
	case logic.TrueF:
		return true
	case logic.FalseF:
		// False contributes nothing to a disjunction.
	}

	return false
}
