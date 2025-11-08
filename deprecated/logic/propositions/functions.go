package propositions

import (
	"sort"
)

func TruthTable(f Formula) []Fact {
	vars := f.Vars()
	n := len(vars)
	rows := make([]Fact, 0)
	for i := 0; i < 1<<n; i++ {
		row := make(Fact)
		for j, v := range vars {
			row[Var(v)] = (i>>j)&1 == 1
		}
		row["result"] = f.Eval(row)
		rows = append(rows, row)
	}
	return rows
}

func Equivalent(a, b Formula) bool {
	ttA := TruthTable(a)
	ttB := TruthTable(b)
	if len(ttA) != len(ttB) {
		return false
	}
	for i := range ttA {
		if ttA[i]["result"] != ttB[i]["result"] {
			return false
		}
	}
	return true
}

func IsSatisfiable(f Formula) bool {
	return resolution(f)
}

func IsContradiction(f Formula) bool {
	return !resolution(f)
}

func IsTautology(f Formula) bool {
	tt := TruthTable(f)
	for _, row := range tt {
		if !row["result"] {
			return false
		}
	}
	return true
}

func FailCases(f Formula) []Fact {
	failCases := make([]Fact, 0)
	tt := TruthTable(f)
	for _, row := range tt {
		if !row["result"] {
			failCases = append(failCases, row)
		}
	}
	return failCases
}

//

func union(a, b []string) []string {
	set := make(map[string]struct{})
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, x := range b {
		set[x] = struct{}{}
	}
	var out []string
	for x := range set {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

//

func ToNNF(f Formula) Formula {
	switch x := f.(type) {
	case NotF:
		switch inner := x.F.(type) {
		case AndF:
			return OrF{L: ToNNF(NotF{F: inner.L}), R: ToNNF(NotF{F: inner.R})}
		case OrF:
			return AndF{L: ToNNF(NotF{F: inner.L}), R: ToNNF(NotF{F: inner.R})}
		case NotF:
			return ToNNF(inner.F)
		case ImplF:
			return AndF{L: ToNNF(inner.L), R: ToNNF(NotF{F: inner.R})}
		case IffF:
			return OrF{
				L: AndF{L: ToNNF(inner.L), R: ToNNF(NotF{F: inner.R})},
				R: AndF{L: ToNNF(NotF{F: inner.L}), R: ToNNF(inner.R)},
			}
		default:
			return NotF{F: ToNNF(inner)}
		}
	case AndF:
		return AndF{L: ToNNF(x.L), R: ToNNF(x.R)}
	case OrF:
		return OrF{L: ToNNF(x.L), R: ToNNF(x.R)}
	case ImplF:
		return OrF{L: ToNNF(NotF{F: x.L}), R: ToNNF(x.R)}
	case IffF:
		return AndF{
			L: OrF{L: ToNNF(NotF{F: x.L}), R: ToNNF(x.R)},
			R: OrF{L: ToNNF(NotF{F: x.R}), R: ToNNF(x.L)},
		}
	case GroupF:
		return ToNNF(x.Inner)
	default:
		return x
	}
}

func ToCNF(f Formula) Formula {
	f = ToNNF(f)
	switch x := f.(type) {
	case AndF:
		return AndF{L: ToCNF(x.L), R: ToCNF(x.R)}
	case OrF:
		// distribución
		l, lok := x.L.(AndF)
		r, rok := x.R.(AndF)
		switch {
		case lok:
			return AndF{
				L: ToCNF(OrF{L: l.L, R: x.R}),
				R: ToCNF(OrF{L: l.R, R: x.R}),
			}
		case rok:
			return AndF{
				L: ToCNF(OrF{L: x.L, R: r.L}),
				R: ToCNF(OrF{L: x.L, R: r.R}),
			}
		default:
			return OrF{L: ToCNF(x.L), R: ToCNF(x.R)}
		}
	default:
		return x
	}
}

func ToDNF(f Formula) Formula {
	f = ToNNF(f)
	switch x := f.(type) {
	case OrF:
		return OrF{L: ToDNF(x.L), R: ToDNF(x.R)}
	case AndF:
		// distribución
		l, lok := x.L.(OrF)
		r, rok := x.R.(OrF)
		switch {
		case lok:
			return OrF{
				L: ToDNF(AndF{L: l.L, R: x.R}),
				R: ToDNF(AndF{L: l.R, R: x.R}),
			}
		case rok:
			return OrF{
				L: ToDNF(AndF{L: x.L, R: r.L}),
				R: ToDNF(AndF{L: x.L, R: r.R}),
			}
		default:
			return AndF{L: ToDNF(x.L), R: ToDNF(x.R)}
		}
	default:
		return x
	}
}
