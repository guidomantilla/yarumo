package facts

import (
	"maps"
	"slices"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
)

type entry struct {
	value    bool
	origin   explain.Origin
	ruleName string
	step     int
}

type factBase struct {
	entries map[logic.Var]entry
}

// NewFactBase creates an empty fact base.
func NewFactBase() FactBase {
	return &factBase{
		entries: make(map[logic.Var]entry),
	}
}

// NewFactBaseFrom creates a fact base pre-populated with the given facts as asserted.
func NewFactBaseFrom(initial logic.Fact) FactBase {
	fb := &factBase{
		entries: make(map[logic.Var]entry, len(initial)),
	}

	for v, val := range initial {
		fb.entries[v] = entry{value: val, origin: explain.Asserted}
	}

	return fb
}

// Assert sets a user-provided fact.
func (fb *factBase) Assert(variable logic.Var, value bool) {
	cassert.NotNil(fb, "factBase is nil")

	fb.entries[variable] = entry{value: value, origin: explain.Asserted}
}

// AssertAll sets multiple user-provided facts.
func (fb *factBase) AssertAll(facts logic.Fact) {
	cassert.NotNil(fb, "factBase is nil")

	for v, val := range facts {
		fb.Assert(v, val)
	}
}

// Derive sets a fact produced by a rule at the given step.
func (fb *factBase) Derive(variable logic.Var, value bool, ruleName string, step int) {
	cassert.NotNil(fb, "factBase is nil")

	fb.entries[variable] = entry{
		value:    value,
		origin:   explain.Derived,
		ruleName: ruleName,
		step:     step,
	}
}

// Retract removes a fact entirely.
func (fb *factBase) Retract(variable logic.Var) {
	cassert.NotNil(fb, "factBase is nil")

	delete(fb.entries, variable)
}

// Get returns the value and existence of a fact.
func (fb *factBase) Get(variable logic.Var) (bool, bool) {
	cassert.NotNil(fb, "factBase is nil")

	e, ok := fb.entries[variable]
	if !ok {
		return false, false
	}

	return e.value, true
}

// Snapshot returns a copy of all facts as a logic.Fact for formula evaluation.
func (fb *factBase) Snapshot() logic.Fact {
	cassert.NotNil(fb, "factBase is nil")

	result := make(logic.Fact, len(fb.entries))

	for v, e := range fb.entries {
		result[v] = e.value
	}

	return result
}

// Provenance returns the provenance record for a specific variable.
func (fb *factBase) Provenance(variable logic.Var) (explain.Provenance, bool) {
	cassert.NotNil(fb, "factBase is nil")

	e, ok := fb.entries[variable]
	if !ok {
		return explain.Provenance{}, false
	}

	return explain.Provenance{
		Variable: variable,
		Value:    e.value,
		Origin:   e.origin,
		RuleName: e.ruleName,
		Step:     e.step,
	}, true
}

// AllProvenance returns provenance records for all known facts, sorted by variable name.
func (fb *factBase) AllProvenance() []explain.Provenance {
	cassert.NotNil(fb, "factBase is nil")

	keys := make([]logic.Var, 0, len(fb.entries))

	for v := range fb.entries {
		keys = append(keys, v)
	}

	slices.Sort(keys)

	result := make([]explain.Provenance, 0, len(keys))

	for _, v := range keys {
		p, _ := fb.Provenance(v)
		result = append(result, p)
	}

	return result
}

// Len returns the number of known facts.
func (fb *factBase) Len() int {
	cassert.NotNil(fb, "factBase is nil")

	return len(fb.entries)
}

// Clone returns a deep copy of the fact base.
func (fb *factBase) Clone() FactBase {
	cassert.NotNil(fb, "factBase is nil")

	cloned := make(map[logic.Var]entry, len(fb.entries))
	maps.Copy(cloned, fb.entries)

	return &factBase{entries: cloned}
}
