package props

import (
	"sort"
)

// TruthTable returns all valuations with a special key "result" set to the formula value.
func TruthTable(f Formula) []map[string]bool {
	vars := f.Vars()
	n := len(vars)
	m := 1 << n

	rows := make([]map[string]bool, 0, 1<<n)
	for i := range m {
		row := make(map[string]bool, n+1)
		facts := make(Fact, n)

		for j, v := range vars {
			val := (i>>j)&1 == 1
			facts[Var(v)] = val
			row[v] = val
		}

		row["result"] = f.Eval(facts)
		rows = append(rows, row)
	}

	return rows
}

// Equivalent returns true if the two formulas are equivalent.
func Equivalent(a, b Formula) bool {
	ta := TruthTable(a)

	tb := TruthTable(b)
	if len(ta) != len(tb) {
		return false
	}

	for i := range ta {
		if ta[i]["result"] != tb[i]["result"] {
			return false
		}
	}

	return true
}

// SATThreshold controls when to switch from truth-table to SAT.
// If number of variables is greater than SATThreshold, and a SAT solver is registered,
// IsSatisfiable will use SAT; otherwise it falls back to truth-table.
var SATThreshold = 12

// RegisterSATSolver lets another package (logic/sat) register a SAT-based satisfiability checker.
// The function should return (ok bool, result bool). When ok==true, result is the satisfiability answer.
// When ok==false, IsSatisfiable will fall back to truth-table.
var satSolver func(Formula) (bool, bool)

func RegisterSATSolver(fn func(Formula) (bool, bool)) {
	satSolver = fn
}

// IsSatisfiable returns true if the formula is satisfiable using the configured policy.
func IsSatisfiable(f Formula) bool {
	vars := f.Vars()

	n := len(vars)
	if n > SATThreshold && satSolver != nil {
		if ok, res := satSolver(f); ok {
			return res
		}
	}
	// Fallback: brute-force truth table
	m := 1 << n
	for i := range m {
		facts := make(Fact, n)
		for j, v := range vars {
			facts[Var(v)] = (i>>j)&1 == 1
		}

		if f.Eval(facts) {
			return true
		}
	}

	return false
}

// IsContradiction returns true if the formula is contradictory.
func IsContradiction(f Formula) bool { return !IsSatisfiable(f) }

// IsTautology returns true if the formula is tautological.
func IsTautology(f Formula) bool {
	return !IsSatisfiable(NotF{F: f})
}

// FailCases returns the rows of the truth table that evaluate to false.
func FailCases(f Formula) []map[string]bool {
	var out []map[string]bool

	for _, row := range TruthTable(f) {
		if !row["result"] {
			out = append(out, row)
		}
	}

	return out
}

// Version returns the version/snapshot of the props package.
func Version() string { return "logic/props@phase1" }

func union(a, b []string) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for _, x := range a {
		set[x] = struct{}{}
	}

	for _, x := range b {
		set[x] = struct{}{}
	}

	out := make([]string, 0, len(set))
	for x := range set {
		out = append(out, x)
	}

	sort.Strings(out)

	return out
}
