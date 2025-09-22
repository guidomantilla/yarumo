package props

import (
	"fmt"
	"sort"
)

// Var represents a propositional variable.
type Var string

// Fact is a partial valuation of propositional variables.
type Fact map[Var]bool

// Formula is the common interface of all propositional formulas.
type Formula interface {
	String() string
	Eval(Fact) bool
	Vars() []string
}

// --- Atomic formulas ---

type TrueF struct{}

type FalseF struct{}

func (TrueF) String() string      { return "⊤" }
func (FalseF) String() string     { return "⊥" }
func (TrueF) Eval(Fact) bool      { return true }
func (FalseF) Eval(Fact) bool     { return false }
func (TrueF) Vars() []string      { return nil }
func (FalseF) Vars() []string     { return nil }
func (v Var) String() string      { return string(v) }
func (v Var) Eval(f Fact) bool    { return f[v] }
func (v Var) Vars() []string      { return []string{string(v)} }

// --- Composite formulas ---

type NotF struct{ F Formula }

type AndF struct{ L, R Formula }

type OrF struct{ L, R Formula }

type ImplF struct{ L, R Formula }

type IffF struct{ L, R Formula }

type GroupF struct{ Inner Formula }

func (f NotF) String() string  { return fmt.Sprintf("!%s", f.F.String()) }
func (f AndF) String() string  { return fmt.Sprintf("(%s & %s)", f.L.String(), f.R.String()) }
func (f OrF) String() string   { return fmt.Sprintf("(%s | %s)", f.L.String(), f.R.String()) }
func (f ImplF) String() string { return fmt.Sprintf("(%s => %s)", f.L.String(), f.R.String()) }
func (f IffF) String() string  { return fmt.Sprintf("(%s <=> %s)", f.L.String(), f.R.String()) }
func (g GroupF) String() string { return fmt.Sprintf("(%s)", g.Inner.String()) }

func (f NotF) Eval(m Fact) bool   { return !f.F.Eval(m) }
func (f AndF) Eval(m Fact) bool   { return f.L.Eval(m) && f.R.Eval(m) }
func (f OrF) Eval(m Fact) bool    { return f.L.Eval(m) || f.R.Eval(m) }
func (f ImplF) Eval(m Fact) bool  { return !f.L.Eval(m) || f.R.Eval(m) }
func (f IffF) Eval(m Fact) bool   { return f.L.Eval(m) == f.R.Eval(m) }
func (g GroupF) Eval(m Fact) bool { return g.Inner.Eval(m) }

func (f NotF) Vars() []string   { return f.F.Vars() }
func (f AndF) Vars() []string   { return union(f.L.Vars(), f.R.Vars()) }
func (f OrF) Vars() []string    { return union(f.L.Vars(), f.R.Vars()) }
func (f ImplF) Vars() []string  { return union(f.L.Vars(), f.R.Vars()) }
func (f IffF) Vars() []string   { return union(f.L.Vars(), f.R.Vars()) }
func (g GroupF) Vars() []string { return g.Inner.Vars() }

// --- Utilities ---

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

func Equivalent(a, b Formula) bool {
	ta := TruthTable(a)
	tb := TruthTable(b)
	if len(ta) != len(tb) { return false }
	for i := range ta {
		if ta[i]["result"] != tb[i]["result"] {
			return false
		}
	}
	return true
}

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

func IsContradiction(f Formula) bool { return !IsSatisfiable(f) }

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

func FailCases(f Formula) []map[string]bool {
	var out []map[string]bool
	for _, row := range TruthTable(f) {
		if !row["result"] {
			out = append(out, row)
		}
	}
	return out
}

// --- Transformations ---

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
		return f
	}
}

func ToCNF(f Formula) Formula {
	f = ToNNF(f)
	switch x := f.(type) {
	case AndF:
		return AndF{L: ToCNF(x.L), R: ToCNF(x.R)}
	case OrF:
		l, lok := x.L.(AndF)
		r, rok := x.R.(AndF)
		switch {
		case lok:
			return AndF{L: ToCNF(OrF{L: l.L, R: x.R}), R: ToCNF(OrF{L: l.R, R: x.R})}
		case rok:
			return AndF{L: ToCNF(OrF{L: x.L, R: r.L}), R: ToCNF(OrF{L: x.L, R: r.R})}
		default:
			return OrF{L: ToCNF(x.L), R: ToCNF(x.R)}
		}
	default:
		return f
	}
}

func ToDNF(f Formula) Formula {
	f = ToNNF(f)
	switch x := f.(type) {
	case OrF:
		return OrF{L: ToDNF(x.L), R: ToDNF(x.R)}
	case AndF:
		l, lok := x.L.(OrF)
		r, rok := x.R.(OrF)
		switch {
		case lok:
			return OrF{L: ToDNF(AndF{L: l.L, R: x.R}), R: ToDNF(AndF{L: l.R, R: x.R})}
		case rok:
			return OrF{L: ToDNF(AndF{L: x.L, R: r.L}), R: ToDNF(AndF{L: x.L, R: r.R})}
		default:
			return AndF{L: ToDNF(x.L), R: ToDNF(x.R)}
		}
	default:
		return f
	}
}

// Simplify reduces a formula applying simple algebraic rules until no change.
func Simplify(f Formula) Formula {
	prev := Formula(nil)
	cur := simplifyOnce(f)
	for !structuralEqual(cur, prev) {
		prev = cur
		cur = simplifyOnce(cur)
	}
	return cur
}

func simplifyOnce(f Formula) Formula {
	switch x := f.(type) {
	case GroupF:
		return Simplify(x.Inner)
	case TrueF, FalseF, Var:
		return x
	case NotF:
		inner := Simplify(x.F)
		switch y := inner.(type) {
		case TrueF:
			return FalseF{}
		case FalseF:
			return TrueF{}
		case NotF:
			return Simplify(y.F)
		default:
			return NotF{F: inner}
		}
	case AndF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		if isFalse(L) || isFalse(R) { return FalseF{} }
		if isTrue(L) { return R }
		if isTrue(R) { return L }
		if structuralEqual(L, R) { return L }
		if isNegationOf(L, R) || isNegationOf(R, L) { return FalseF{} }
		if rOr, ok := R.(OrF); ok { if structuralEqual(L, rOr.L) || structuralEqual(L, rOr.R) { return L } }
		if lOr, ok := L.(OrF); ok { if structuralEqual(R, lOr.L) || structuralEqual(R, lOr.R) { return R } }
		return AndF{L: L, R: R}
	case OrF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		if isTrue(L) || isTrue(R) { return TrueF{} }
		if isFalse(L) { return R }
		if isFalse(R) { return L }
		if structuralEqual(L, R) { return L }
		if isNegationOf(L, R) || isNegationOf(R, L) { return TrueF{} }
		if rAnd, ok := R.(AndF); ok { if structuralEqual(L, rAnd.L) || structuralEqual(L, rAnd.R) { return L } }
		if lAnd, ok := L.(AndF); ok { if structuralEqual(R, lAnd.L) || structuralEqual(R, lAnd.R) { return R } }
		return OrF{L: L, R: R}
	case ImplF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		if isFalse(L) { return TrueF{} }
		if isTrue(L) { return R }
		if isTrue(R) { return TrueF{} }
		if isFalse(R) { return NotF{F: L} }
		if structuralEqual(L, R) { return TrueF{} }
		return ImplF{L: L, R: R}
	case IffF:
		L := Simplify(x.L)
		R := Simplify(x.R)
		if structuralEqual(L, R) { return TrueF{} }
		if isTrue(L) { return R }
		if isTrue(R) { return L }
		if isFalse(L) { return NotF{F: R} }
		if isFalse(R) { return NotF{F: L} }
		return IffF{L: L, R: R}
	}
	return f
}

func isTrue(f Formula) bool  { _, ok := f.(TrueF); return ok }
func isFalse(f Formula) bool { _, ok := f.(FalseF); return ok }

func isNegationOf(a, b Formula) bool {
	na, ok := a.(NotF)
	if !ok { return false }
	return structuralEqual(na.F, b)
}

func structuralEqual(a, b Formula) bool {
	if a == nil && b == nil { return true }
	if a == nil || b == nil { return false }
	switch x := a.(type) {
	case Var:
		y, ok := b.(Var); return ok && x == y
	case TrueF:
		_, ok := b.(TrueF); return ok
	case FalseF:
		_, ok := b.(FalseF); return ok
	case NotF:
		y, ok := b.(NotF); return ok && structuralEqual(x.F, y.F)
	case AndF:
		y, ok := b.(AndF); return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case OrF:
		y, ok := b.(OrF); return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case ImplF:
		y, ok := b.(ImplF); return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case IffF:
		y, ok := b.(IffF); return ok && structuralEqual(x.L, y.L) && structuralEqual(x.R, y.R)
	case GroupF:
		y, ok := b.(GroupF); return ok && structuralEqual(x.Inner, y.Inner)
	default:
		return false
	}
}

// Version returns the version/snapshot of the props package.
func Version() string { return "logic2/props@phase1" }
