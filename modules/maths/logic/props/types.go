package props

import (
	"fmt"
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

func (TrueF) String() string { return "⊤" }

func (FalseF) String() string { return "⊥" }

func (TrueF) Eval(Fact) bool { return true }

func (FalseF) Eval(Fact) bool { return false }

func (TrueF) Vars() []string { return nil }

func (FalseF) Vars() []string { return nil }

func (v Var) String() string { return string(v) }

func (v Var) Eval(f Fact) bool { return f[v] }

func (v Var) Vars() []string { return []string{string(v)} }

// --- Composite formulas ---

type NotF struct{ F Formula }

type AndF struct{ L, R Formula }

type OrF struct{ L, R Formula }

type ImplF struct{ L, R Formula }

type IffF struct{ L, R Formula }

type GroupF struct{ Inner Formula }

func (f NotF) String() string { return "!" + f.F.String() }

func (f AndF) String() string { return fmt.Sprintf("(%s & %s)", f.L.String(), f.R.String()) }

func (f OrF) String() string { return fmt.Sprintf("(%s | %s)", f.L.String(), f.R.String()) }

func (f ImplF) String() string { return fmt.Sprintf("(%s => %s)", f.L.String(), f.R.String()) }

func (f IffF) String() string { return fmt.Sprintf("(%s <=> %s)", f.L.String(), f.R.String()) }

func (g GroupF) String() string { return fmt.Sprintf("(%s)", g.Inner.String()) }

func (f NotF) Eval(m Fact) bool { return !f.F.Eval(m) }

func (f AndF) Eval(m Fact) bool { return f.L.Eval(m) && f.R.Eval(m) }

func (f OrF) Eval(m Fact) bool { return f.L.Eval(m) || f.R.Eval(m) }

func (f ImplF) Eval(m Fact) bool { return !f.L.Eval(m) || f.R.Eval(m) }

func (f IffF) Eval(m Fact) bool { return f.L.Eval(m) == f.R.Eval(m) }

func (g GroupF) Eval(m Fact) bool { return g.Inner.Eval(m) }

func (f NotF) Vars() []string { return f.F.Vars() }

func (f AndF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f OrF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f ImplF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f IffF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (g GroupF) Vars() []string { return g.Inner.Vars() }
