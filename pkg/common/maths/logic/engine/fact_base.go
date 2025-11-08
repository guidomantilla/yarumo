package engine

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic/props"

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
