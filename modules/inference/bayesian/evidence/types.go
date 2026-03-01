// Package evidence provides an observable evidence base for Bayesian inference.
package evidence

import (
	"maps"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/probability"
)

// EvidenceBase defines the interface for managing observed evidence.
type EvidenceBase interface {
	// Observe sets an observed outcome for a variable.
	Observe(variable probability.Var, outcome probability.Outcome)
	// Retract removes an observation.
	Retract(variable probability.Var)
	// Get returns the observed outcome for a variable.
	Get(variable probability.Var) (probability.Outcome, bool)
	// Observed returns all observations as an assignment.
	Observed() probability.Assignment
	// Len returns the number of observations.
	Len() int
	// Clone returns a deep copy of the evidence base.
	Clone() EvidenceBase
}

var _ EvidenceBase = (*evidenceBase)(nil)

type evidenceBase struct {
	entries map[probability.Var]probability.Outcome
}

// NewEvidenceBase creates an empty evidence base.
func NewEvidenceBase() EvidenceBase {
	return &evidenceBase{
		entries: make(map[probability.Var]probability.Outcome),
	}
}

// NewEvidenceBaseFrom creates an evidence base from an existing assignment.
func NewEvidenceBaseFrom(assignment probability.Assignment) EvidenceBase {
	entries := make(map[probability.Var]probability.Outcome, len(assignment))
	maps.Copy(entries, assignment)

	return &evidenceBase{entries: entries}
}

func (eb *evidenceBase) Observe(variable probability.Var, outcome probability.Outcome) {
	cassert.NotNil(eb, "evidenceBase is nil")

	eb.entries[variable] = outcome
}

func (eb *evidenceBase) Retract(variable probability.Var) {
	cassert.NotNil(eb, "evidenceBase is nil")

	delete(eb.entries, variable)
}

func (eb *evidenceBase) Get(variable probability.Var) (probability.Outcome, bool) {
	cassert.NotNil(eb, "evidenceBase is nil")

	outcome, ok := eb.entries[variable]

	return outcome, ok
}

func (eb *evidenceBase) Observed() probability.Assignment {
	cassert.NotNil(eb, "evidenceBase is nil")

	result := make(probability.Assignment, len(eb.entries))
	maps.Copy(result, eb.entries)

	return result
}

func (eb *evidenceBase) Len() int {
	cassert.NotNil(eb, "evidenceBase is nil")

	return len(eb.entries)
}

func (eb *evidenceBase) Clone() EvidenceBase {
	cassert.NotNil(eb, "evidenceBase is nil")

	cloned := make(map[probability.Var]probability.Outcome, len(eb.entries))
	maps.Copy(cloned, eb.entries)

	return &evidenceBase{entries: cloned}
}
