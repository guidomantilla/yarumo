package logic

// SATSolverFn is the function type for external SAT solvers.
// It returns whether the formula is satisfiable and a satisfying assignment if one exists.
type SATSolverFn func(f Formula) (satisfiable bool, assignment Fact)

var satSolver SATSolverFn //nolint:gochecknoglobals // SAT solver hook by design

// RegisterSATSolver registers an external SAT solver for use by satisfiability functions.
// When registered, IsSatisfiable, IsContradiction, and IsTautology use the solver
// instead of brute-force truth table enumeration.
func RegisterSATSolver(solver SATSolverFn) {
	satSolver = solver
}

// IsSatisfiable returns true if there exists at least one variable assignment
// that makes the formula evaluate to true.
func IsSatisfiable(f Formula) bool {
	if satSolver != nil {
		sat, _ := satSolver(f)

		return sat
	}

	return bruteForceIsSatisfiable(f)
}

// IsContradiction returns true if no variable assignment makes the formula true.
func IsContradiction(f Formula) bool {
	return !IsSatisfiable(f)
}

// IsTautology returns true if every variable assignment makes the formula true.
func IsTautology(f Formula) bool {
	// A formula is a tautology iff its negation is unsatisfiable.
	return !IsSatisfiable(NotF{F: f})
}

func bruteForceIsSatisfiable(f Formula) bool {
	vars := f.Vars()
	n := len(vars)

	for i := range 1 << n {
		assignment := make(Fact, n)

		for j, v := range vars {
			assignment[v] = (i>>j)&1 == 1
		}

		if f.Eval(assignment) {
			return true
		}
	}

	return false
}
