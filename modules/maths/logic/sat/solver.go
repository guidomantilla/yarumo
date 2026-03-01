package sat

import "github.com/guidomantilla/yarumo/maths/logic"

// Solver returns a logic.SATSolverFn that uses the DPLL algorithm.
func Solver() logic.SATSolverFn {
	return func(f logic.Formula) (bool, logic.Fact) {
		cnf := FromFormula(logic.ToCNF(f))

		return Solve(cnf)
	}
}
