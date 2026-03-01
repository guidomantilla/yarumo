package sat

import (
	"maps"

	"github.com/guidomantilla/yarumo/maths/logic"
)

// Solve determines whether the given CNF formula is satisfiable using DPLL.
// If satisfiable, it returns true and a satisfying assignment.
func Solve(cnf CNF) (bool, logic.Fact) {
	assignment := make(logic.Fact)

	sat := dpll(cnf, assignment)
	if !sat {
		return false, nil
	}

	return true, assignment
}

func dpll(clauses CNF, assignment logic.Fact) bool {
	clauses = unitPropagate(clauses, assignment)

	if hasEmptyClause(clauses) {
		return false
	}

	if len(clauses) == 0 {
		return true
	}

	clauses = pureLiteralEliminate(clauses, assignment)

	if len(clauses) == 0 {
		return true
	}

	chosen := clauses[0][0]

	posAssign := copyFact(assignment)
	posAssign[chosen.V] = !chosen.Neg

	if dpll(propagate(clauses, chosen), posAssign) {
		copyInto(assignment, posAssign)

		return true
	}

	negLit := Lit{V: chosen.V, Neg: !chosen.Neg}

	negAssign := copyFact(assignment)
	negAssign[negLit.V] = !negLit.Neg

	if dpll(propagate(clauses, negLit), negAssign) {
		copyInto(assignment, negAssign)

		return true
	}

	return false
}

func unitPropagate(clauses CNF, assignment logic.Fact) CNF {
	for {
		unit, found := findUnit(clauses)
		if !found {
			return clauses
		}

		assignment[unit.V] = !unit.Neg
		clauses = propagate(clauses, unit)
	}
}

func findUnit(clauses CNF) (Lit, bool) {
	for _, clause := range clauses {
		if len(clause) == 1 {
			return clause[0], true
		}
	}

	return Lit{}, false
}

func pureLiteralEliminate(clauses CNF, assignment logic.Fact) CNF {
	for {
		lit, found := findPureLiteral(clauses)
		if !found {
			return clauses
		}

		assignment[lit.V] = !lit.Neg
		clauses = propagate(clauses, lit)
	}
}

func findPureLiteral(clauses CNF) (Lit, bool) {
	pos := make(map[logic.Var]bool)
	neg := make(map[logic.Var]bool)

	for _, clause := range clauses {
		for _, lit := range clause {
			if lit.Neg {
				neg[lit.V] = true
			} else {
				pos[lit.V] = true
			}
		}
	}

	for v := range pos {
		if !neg[v] {
			return Lit{V: v}, true
		}
	}

	for v := range neg {
		if !pos[v] {
			return Lit{V: v, Neg: true}, true
		}
	}

	return Lit{}, false
}

func propagate(clauses CNF, lit Lit) CNF {
	var result CNF

	for _, clause := range clauses {
		if containsLit(clause, lit) {
			continue
		}

		newClause := removeLit(clause, Lit{V: lit.V, Neg: !lit.Neg})
		result = append(result, newClause)
	}

	return result
}

func containsLit(clause Clause, lit Lit) bool {
	for _, l := range clause {
		if l.V == lit.V && l.Neg == lit.Neg {
			return true
		}
	}

	return false
}

func removeLit(clause Clause, lit Lit) Clause {
	var result Clause

	for _, l := range clause {
		if l.V == lit.V && l.Neg == lit.Neg {
			continue
		}

		result = append(result, l)
	}

	if result == nil {
		return Clause{}
	}

	return result
}

func hasEmptyClause(clauses CNF) bool {
	for _, clause := range clauses {
		if len(clause) == 0 {
			return true
		}
	}

	return false
}

func copyFact(f logic.Fact) logic.Fact {
	cp := make(logic.Fact, len(f))
	maps.Copy(cp, f)

	return cp
}

func copyInto(dst, src logic.Fact) {
	maps.Copy(dst, src)
}
