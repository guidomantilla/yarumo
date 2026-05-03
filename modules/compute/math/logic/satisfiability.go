package logic

import "sync"

// SATSolverFn is the function type for external SAT solvers.
// It returns whether the formula is satisfiable and a satisfying assignment if one exists.
type SATSolverFn func(f Formula) (satisfiable bool, assignment Fact)

var (
	satMu     sync.RWMutex //nolint:gochecknoglobals // SAT solver hook by design
	satSolver SATSolverFn  //nolint:gochecknoglobals // SAT solver hook by design
)

// RegisterSATSolver registers an external SAT solver for use by satisfiability functions.
// When registered, IsSatisfiable, IsContradiction, and IsTautology use the solver
// instead of brute-force truth table enumeration.
func RegisterSATSolver(solver SATSolverFn) {
	satMu.Lock()

	satSolver = solver

	satMu.Unlock()
}

func loadSATSolver() SATSolverFn {
	satMu.RLock()

	s := satSolver

	satMu.RUnlock()

	return s
}

// IsSatisfiable returns true if there exists at least one variable assignment
// that makes the formula evaluate to true.
func IsSatisfiable(f Formula) bool {
	_, found := FindSatisfyingAssignment(f)

	return found
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
