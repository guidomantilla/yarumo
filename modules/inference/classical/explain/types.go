// Package explain provides explanation trace types for rule-based inference.
package explain

import "github.com/guidomantilla/yarumo/maths/logic"

// Origin describes how a fact was established.
type Origin int

const (
	// Asserted indicates a user-provided fact.
	Asserted Origin = iota
	// Derived indicates a fact produced by a rule.
	Derived
)

// Provenance records the origin of a single fact value.
type Provenance struct {
	Variable logic.Var
	Value    bool
	Origin   Origin
	RuleName string // Empty if Asserted.
	Step     int    // Zero if Asserted.
}

// Step represents a single inference step in a trace.
type Step struct {
	Number      int
	RuleName    string
	Condition   logic.Formula
	FactsBefore logic.Fact
	Produced    map[logic.Var]bool
}

// Trace is an ordered sequence of inference steps.
type Trace struct {
	Steps []Step
	Goal  logic.Var // Empty for forward chaining.
}
