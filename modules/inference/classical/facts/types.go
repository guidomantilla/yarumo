// Package facts provides a fact base with provenance tracking for rule-based inference.
package facts

import (
	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
)

// FactBase defines the interface for managing known facts with provenance tracking.
type FactBase interface {
	// Assert sets a user-provided fact.
	Assert(variable logic.Var, value bool)
	// AssertAll sets multiple user-provided facts.
	AssertAll(facts logic.Fact)
	// Derive sets a fact produced by a rule at the given step.
	Derive(variable logic.Var, value bool, ruleName string, step int)
	// Retract removes a fact entirely.
	Retract(variable logic.Var)
	// Get returns the value and existence of a fact.
	Get(variable logic.Var) (value bool, known bool)
	// Snapshot returns a copy of all facts as a logic.Fact for formula evaluation.
	Snapshot() logic.Fact
	// Provenance returns the provenance record for a specific variable.
	Provenance(variable logic.Var) (explain.Provenance, bool)
	// AllProvenance returns provenance records for all known facts.
	AllProvenance() []explain.Provenance
	// Len returns the number of known facts.
	Len() int
	// Clone returns a deep copy of the fact base.
	Clone() FactBase
}

var _ FactBase = (*factBase)(nil)
