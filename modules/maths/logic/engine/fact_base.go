package engine

import p "github.com/guidomantilla/yarumo/maths/logic/props"

// FactBase stores boolean facts.
type FactBase map[p.Var]bool

// Get returns (value, ok).
func (fb FactBase) Get(v p.Var) (bool, bool) { val, ok := fb[v]; return val, ok }

// Set assigns a fact value.
func (fb FactBase) Set(v p.Var, val bool) { fb[v] = val }

// Retract removes a fact.
func (fb FactBase) Retract(v p.Var) { delete(fb, v) }

// Merge incorporates facts from another FactBase.
func (fb FactBase) Merge(other FactBase) {
	for k, v := range other {
		fb[k] = v
	}
}
