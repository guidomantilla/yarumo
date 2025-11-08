package entailment

import (
	p "github.com/guidomantilla/yarumo/deprecated/logic/propositions"
	"github.com/guidomantilla/yarumo/deprecated/logic/sat"
)

// Result captures the outcome of an entailment check.
// If Entails is false, CounterModel provides a valuation that satisfies KB ∧ ¬phi.
type Result struct {
	Entails      bool
	CounterModel map[p.Var]bool
}

// Entails implements entailment by refutation over propositional formulas.
// KB ⊨ phi  iff  (∧KB) ∧ ¬phi is UNSAT.
func Entails(KB []p.Formula, phi p.Formula) bool {
	conj := andAll(KB)
	target := p.AndF{L: conj, R: p.NotF{F: phi}}
	// Light simplification to reduce size before SAT check (and also helps resolution fallback)
	target = p.Simplify(target).(p.AndF)
	return !p.IsSatisfiable(target)
}

// EntailsWithCounterModel performs entailment by refutation using SAT/DPLL and, when
// the entailment does not hold, returns a countermodel (assignment) for KB ∧ ¬phi.
func EntailsWithCounterModel(KB []p.Formula, phi p.Formula) (Result, error) {
	conj := andAll(KB)
	target := p.AndF{L: conj, R: p.NotF{F: phi}}
	// Simplify before CNF conversion for robustness
	target = p.Simplify(target).(p.AndF)

	cnf, err := sat.FromFormulaToCNF(target)
	if err != nil {
		// Fallback to boolean result only if CNF aplanado no está disponible
		return Result{Entails: !p.IsSatisfiable(target)}, nil
	}
	ok, asg := sat.DPLL(cnf, nil)
	if ok {
		m := make(map[p.Var]bool, len(asg))
		for v, val := range asg {
			m[v] = val
		}
		return Result{Entails: false, CounterModel: m}, nil
	}
	return Result{Entails: true}, nil
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
