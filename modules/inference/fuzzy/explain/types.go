// Package explain provides explanation trace types for fuzzy inference.
package explain

import fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

// Phase identifies a stage in the fuzzy inference process.
type Phase int

const (
	// Fuzzification marks the input fuzzification phase.
	Fuzzification Phase = iota
	// RuleEvaluation marks the rule evaluation phase.
	RuleEvaluation
	// Aggregation marks the output aggregation phase.
	Aggregation
	// Defuzzification marks the defuzzification phase.
	Defuzzification
	// Complete marks the final result phase.
	Complete
)

// Membership records a fuzzification result.
type Membership struct {
	Variable string
	Term     string
	Degree   fuzzym.Degree
}

// Activation records a rule activation result.
type Activation struct {
	RuleName string
	Strength fuzzym.Degree
	Output   string
	Term     string
}

// Step represents a single step in the fuzzy inference trace.
type Step struct {
	Number      int
	Phase       Phase
	Message     string
	Memberships []Membership
	Activations []Activation
}

// Output holds the crisp output value for a variable.
type Output struct {
	Variable   string
	CrispValue float64
}

// Trace is an ordered sequence of fuzzy inference steps.
type Trace struct {
	Steps   []Step
	Inputs  map[string]float64
	Outputs []Output
}
