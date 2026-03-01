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
	n := len(vars)
	rows := make([]Row, 0, 1<<n)

	for i := range 1 << n {
		assignment := make(Fact, n)

		for j, v := range vars {
			assignment[v] = (i>>j)&1 == 1
		}

		rows = append(rows, Row{
			Assignment: assignment,
			Result:     f.Eval(assignment),
		})
	}

	return rows
}

// Equivalent returns true if two formulas produce the same truth value
// for every possible variable assignment.
func Equivalent(a, b Formula) bool {
	vars := mergeVars(a.Vars(), b.Vars())
	n := len(vars)

	for i := range 1 << n {
		assignment := make(Fact, n)

		for j, v := range vars {
			assignment[v] = (i>>j)&1 == 1
		}

		if a.Eval(assignment) != b.Eval(assignment) {
			return false
		}
	}

	return true
}

// FailCases returns all variable assignments for which the formula evaluates to false.
func FailCases(f Formula) []Fact {
	vars := f.Vars()
	n := len(vars)

	var fails []Fact

	for i := range 1 << n {
		assignment := make(Fact, n)

		for j, v := range vars {
			assignment[v] = (i>>j)&1 == 1
		}

		if !f.Eval(assignment) {
			fails = append(fails, assignment)
		}
	}

	return fails
}
