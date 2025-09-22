package engine

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"

// FactBase stores boolean facts.
type FactBase map[props.Var]bool

// Get returns (value, ok).
func (fb FactBase) Get(v props.Var) (bool, bool) { val, ok := fb[v]; return val, ok }

// Set assigns a fact value.
func (fb FactBase) Set(v props.Var, val bool) { fb[v] = val }

// Retract removes a fact.
func (fb FactBase) Retract(v props.Var) { delete(fb, v) }

// Merge incorporates facts from another FactBase.
func (fb FactBase) Merge(other FactBase) {
	for k, v := range other {
		fb[k] = v
	}
}

// Rule represents a propositional rule with a literal consequence (MVP).
type Rule struct {
	ID   string
	When props.Formula
	Then props.Var
}

// Engine contains current rules and facts.
type Engine struct {
	Facts FactBase
	Rules []Rule
}

// Assert sets a fact to true.
func (e *Engine) Assert(v props.Var) { if e.Facts == nil { e.Facts = FactBase{} }; e.Facts[v] = true }

// Retract removes a fact.
func (e *Engine) Retract(v props.Var) { if e.Facts != nil { delete(e.Facts, v) } }

// FireOnce evaluates rules in a single pass and returns fired rule IDs.
// Phase 0: no-op (returns empty slice).
func (e *Engine) FireOnce() (fired []string) { return nil }

// RunToFixpoint iterates FireOnce until convergence or maxIters is reached.
// Phase 0: no-op (returns empty slice).
func (e *Engine) RunToFixpoint(maxIters int) (fired []string) { return nil }

// Query evaluates a goal against the current facts.
// Phase 0: always returns false, nil.
func (e *Engine) Query(goal props.Formula) (bool, *Explain) { return false, nil }

// Explain is a minimal structure for traces.
type Explain struct {
	ID    string
	Expr  string
	Value bool
	Why   string
	Kids  []*Explain
}
