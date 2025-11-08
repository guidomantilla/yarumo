package sat

import (
	p "github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"
)

// DPLL determines satisfiability of a CNF. Returns (ok, model if ok).
func DPLL(cnf CNF, asg Assignment) (bool, Assignment) {
	if asg == nil {
		asg = make(Assignment)
	}

	cnf = removeTautologies(cnf)

	for {
		// 1) Unit propagation
		if unit, ok := findUnitClause(cnf); ok {
			asg[unit.Var] = !unit.Neg
			cnf, _ = assign(cnf, unit)
			continue
		}
		// 2) Pure literal elimination
		if pure, ok := findPureLiteral(cnf); ok {
			asg[pure.Var] = !pure.Neg
			cnf, _ = assign(cnf, pure)
			continue
		}
		break
	}

	if len(cnf) == 0 {
		return true, asg
	}
	if hasEmptyClause(cnf) {
		return false, nil
	}

	// 3) Branching (choose literal from shortest clause)
	lit := chooseLiteral(cnf)

	// Assume lit = true
	cnf1, _ := assign(copyCNF(cnf), lit)
	if ok, asg1 := DPLL(cnf1, copyAsg(asg)); ok {
		asg1[lit.Var] = !lit.Neg
		return true, asg1
	}
	// Assume lit = false (i.e., ¬lit = true)
	cnf2, _ := assign(copyCNF(cnf), lit.Negated())
	if ok, asg2 := DPLL(cnf2, copyAsg(asg)); ok {
		asg2[lit.Var] = lit.Neg
		return true, asg2
	}
	return false, nil
}

// --- helpers ---

func assign(cnf CNF, lit Lit) (CNF, bool) {
	out := make(CNF, 0, len(cnf))
	satisfied := false
	for _, c := range cnf {
		if clauseHasLit(c, lit) {
			satisfied = true
			continue // clause satisfied, drop it
		}
		neg := lit.Negated()
		c2 := removeLit(c, neg)
		out = append(out, c2)
	}
	return out, satisfied
}

func clauseHasLit(c Clause, l Lit) bool {
	for _, x := range c {
		if x.Var == l.Var && x.Neg == l.Neg {
			return true
		}
	}
	return false
}

func removeLit(c Clause, l Lit) Clause {
	out := c[:0]
	for _, x := range c {
		if !(x.Var == l.Var && x.Neg == l.Neg) {
			out = append(out, x)
		}
	}
	return out
}

func hasEmptyClause(cnf CNF) bool {
	for _, c := range cnf {
		if len(c) == 0 {
			return true
		}
	}
	return false
}

func findUnitClause(cnf CNF) (Lit, bool) {
	for _, c := range cnf {
		if len(c) == 1 {
			return c[0], true
		}
	}
	return Lit{}, false
}

func findPureLiteral(cnf CNF) (Lit, bool) {
	seenPos := map[p.Var]bool{}
	seenNeg := map[p.Var]bool{}
	for _, c := range cnf {
		for _, l := range c {
			if l.Neg {
				seenNeg[l.Var] = true
			} else {
				seenPos[l.Var] = true
			}
		}
	}
	for v := range seenPos {
		if !seenNeg[v] {
			return Lit{Var: v, Neg: false}, true
		}
	}
	for v := range seenNeg {
		if !seenPos[v] {
			return Lit{Var: v, Neg: true}, true
		}
	}
	return Lit{}, false
}

func chooseLiteral(cnf CNF) Lit {
	bestIdx := -1
	bestLen := int(^uint(0) >> 1)
	for i, c := range cnf {
		if len(c) < bestLen && len(c) > 0 {
			bestLen = len(c)
			bestIdx = i
		}
	}
	if bestIdx >= 0 && len(cnf[bestIdx]) > 0 {
		return cnf[bestIdx][0]
	}
	for _, c := range cnf {
		if len(c) > 0 {
			return c[0]
		}
	}
	return Lit{}
}

func copyCNF(cnf CNF) CNF {
	out := make(CNF, len(cnf))
	for i, c := range cnf {
		cc := make(Clause, len(c))
		copy(cc, c)
		out[i] = cc
	}
	return out
}

func copyAsg(a Assignment) Assignment {
	b := make(Assignment, len(a))
	for k, v := range a {
		b[k] = v
	}
	return b
}

func removeTautologies(cnf CNF) CNF {
	out := make(CNF, 0, len(cnf))
	for _, c := range cnf {
		if isTautologyClause(c) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func isTautologyClause(c Clause) bool {
	m := make(map[p.Var]struct{ pos, neg bool })
	for _, l := range c {
		st := m[l.Var]
		if l.Neg {
			st.neg = true
		} else {
			st.pos = true
		}
		m[l.Var] = st
	}
	for _, st := range m {
		if st.pos && st.neg {
			return true
		}
	}
	return false
}
