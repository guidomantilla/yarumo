package props

import "sort"

// TruthTable returns all valuations with a special key "result" set to the formula value.
func TruthTable(f Formula) []map[string]bool {
	vars := f.Vars()
	n := len(vars)
	rows := make([]map[string]bool, 0, 1<<n)
	for i := 0; i < (1 << n); i++ {
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

// IsSatisfiable returns true if the formula is satisfiable.
func IsSatisfiable(f Formula) bool {
	// Provisional policy (Phase 1): truth-table evaluation
	vars := f.Vars()
	n := len(vars)
	for i := 0; i < (1 << n); i++ {
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
	vars := f.Vars()
	n := len(vars)
	for i := 0; i < (1 << n); i++ {
		facts := make(Fact, n)
		for j, v := range vars {
			facts[Var(v)] = (i>>j)&1 == 1
		}
		if !f.Eval(facts) {
			return false
		}
	}
	return true
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
func Version() string { return "logic2/props@phase1" }

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
