package entailment

import (
	p "github.com/guidomantilla/yarumo/common/maths/logic/props"
	"github.com/guidomantilla/yarumo/common/maths/logic/sat"
)

// Entails implements entailment by refutation over propositional formulas.
// KB ⊨ phi  iff  (∧KB) ∧ ¬phi is UNSAT.
func Entails(KB []p.Formula, phi p.Formula) bool {
	target := p.AndF{L: andAll(KB), R: p.NotF{F: phi}}
	// Optionally simplify before satisfiability check
	simpl := p.Simplify(target)
	return !p.IsSatisfiable(simpl)
}

// EntailsWithCounterModel performs entailment by refutation using SAT/DPLL and, when
// the entailment does not hold, returns a countermodel (assignment) for KB ∧ ¬phi.
// If SAT is not available, it falls back to a boolean result with nil assignment.
func EntailsWithCounterModel(KB []p.Formula, phi p.Formula) (bool, sat.Assignment) {
	target := p.AndF{L: andAll(KB), R: p.NotF{F: phi}}
	target = p.Simplify(target).(p.AndF)

	cnf, err := sat.FromFormulaToCNF(target)
	if err != nil {
		// Fallback: just return boolean entailment, no model
		return !p.IsSatisfiable(target), nil
	}
	ok, asg := sat.DPLL(cnf, nil)
	if ok {
		// Satisfiable: entailment does not hold; return countermodel
		return false, asg
	}
	// Unsatisfiable: entailment holds; no countermodel
	return true, nil
}

// andAll builds the conjunction of all formulas in KB. Empty KB yields True.
func andAll(fs []p.Formula) p.Formula {
	if len(fs) == 0 {
		return p.TrueF{}
	}
	acc := fs[0]
	for i := 1; i < len(fs); i++ {
		acc = p.AndF{L: acc, R: fs[i]}
	}
	return acc
}
