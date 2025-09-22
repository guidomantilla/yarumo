package props

// Var represents a propositional variable.
type Var string

// Fact is a partial valuation of propositional variables.
// In Phase 1 it will be used to evaluate formulas.
type Fact map[Var]bool

// Formula is the common interface of all propositional formulas.
// In Phase 1 it will include real methods (String, Eval, Vars).
// For now it is a marker to allow other packages to compile.
type Formula interface{
	isFormula()
}

// Minimal implementation to mark concrete types as formulas in the future.
// During Phase 1 this will be replaced by concrete types (And, Or, Not, etc.).

type baseFormula struct{}

func (baseFormula) isFormula() {}

// Version returns the version/snapshot of the props package.
func Version() string { return "logic2/props@phase0" }
