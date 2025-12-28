package sat

import p "github.com/guidomantilla/yarumo/maths/logic/props"

// Solver is an explicit SAT solver hook that callers can register into props
// using props.RegisterSATSolver(Solver). It avoids package-level side effects
// (init functions) and keeps dependencies explicit.
func Solver(f p.Formula) (bool, bool) {
	cnf, err := FromFormulaToCNF(f)
	if err != nil {
		return false, false
	}

	ok, _ := DPLL(cnf, nil)

	return true, ok
}
