package logic

// Row represents a single row in a truth table.
type Row struct {
	Assignment Fact
	Result     bool
}

// TruthTable generates the complete truth table for a formula.
// Each row contains an assignment and its evaluation result.
func TruthTable(f Formula) []Row {
	vars := f.Vars()
	rows := make([]Row, 0, 1<<len(vars))

	eachAssignment(vars, func(a Fact) bool {
		rows = append(rows, Row{
			Assignment: a,
			Result:     f.Eval(a),
		})

		return true
	})

	return rows
}

// Equivalent returns true if two formulas produce the same truth value
// for every possible variable assignment.
func Equivalent(a, b Formula) bool {
	vars := mergeVars(a.Vars(), b.Vars())
	equal := true

	eachAssignment(vars, func(f Fact) bool {
		if a.Eval(f) != b.Eval(f) {
			equal = false

			return false
		}

		return true
	})

	return equal
}

// FailCases returns all variable assignments for which the formula evaluates to false.
func FailCases(f Formula) []Fact {
	vars := f.Vars()

	var fails []Fact

	eachAssignment(vars, func(a Fact) bool {
		if !f.Eval(a) {
			fails = append(fails, a)
		}

		return true
	})

	return fails
}

// FindSatisfyingAssignment returns the first assignment that makes the formula true.
// Uses the registered SAT solver if available, otherwise brute-force enumeration.
func FindSatisfyingAssignment(f Formula) (Fact, bool) {
	solver := loadSATSolver()
	if solver != nil {
		sat, assignment := solver(f)

		return assignment, sat
	}

	return bruteForceFindSatisfying(f)
}

func bruteForceFindSatisfying(f Formula) (Fact, bool) {
	vars := f.Vars()

	var found Fact

	eachAssignment(vars, func(a Fact) bool {
		if f.Eval(a) {
			found = a

			return false
		}

		return true
	})

	return found, found != nil
}
