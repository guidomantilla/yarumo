// Package entailment provides logical entailment checking for propositional formulas.
package entailment

import "github.com/guidomantilla/yarumo/maths/logic"

// Entails returns true if the conclusion follows logically from the premises.
// Every assignment satisfying all premises must also satisfy the conclusion.
func Entails(premises []logic.Formula, conclusion logic.Formula) bool {
	impl := logic.ImplF{L: buildConjunction(premises), R: conclusion}

	return logic.IsTautology(impl)
}

// EntailsWithCounterModel checks entailment and returns a countermodel if it fails.
// A countermodel is a variable assignment where all premises are true but the conclusion is false.
func EntailsWithCounterModel(premises []logic.Formula, conclusion logic.Formula) (bool, logic.Fact) {
	conj := buildConjunction(premises)
	negated := logic.AndF{L: conj, R: logic.NotF{F: conclusion}}

	vars := negated.Vars()
	n := len(vars)

	for i := range 1 << n {
		assignment := make(logic.Fact, n)

		for j, v := range vars {
			assignment[v] = (i>>j)&1 == 1
		}

		if negated.Eval(assignment) {
			return false, assignment
		}
	}

	return true, nil
}

func buildConjunction(formulas []logic.Formula) logic.Formula {
	if len(formulas) == 0 {
		return logic.TrueF{}
	}

	result := formulas[0]

	for _, f := range formulas[1:] {
		result = logic.AndF{L: result, R: f}
	}

	return result
}
