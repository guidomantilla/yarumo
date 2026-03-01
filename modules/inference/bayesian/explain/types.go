// Package explain provides explanation trace types for Bayesian inference.
package explain

import "github.com/guidomantilla/yarumo/maths/probability"

// Phase identifies a stage in the Bayesian inference process.
type Phase int

const (
	// Initialize marks the setup phase.
	Initialize Phase = iota
	// Propagate marks the message propagation phase.
	Propagate
	// Marginalize marks the variable marginalization phase.
	Marginalize
	// Complete marks the final result phase.
	Complete
)

// Factor describes a factor involved in an inference step.
type Factor struct {
	Variables []probability.Var
	Size      int
}

// Step represents a single step in the Bayesian inference trace.
type Step struct {
	Number  int
	Phase   Phase
	Message string
	Factor  Factor
}

// Posterior holds the computed posterior distribution for a query variable.
type Posterior struct {
	Variable     probability.Var
	Distribution probability.Distribution
}

// Trace is an ordered sequence of Bayesian inference steps.
type Trace struct {
	Steps      []Step
	Query      probability.Var
	Evidence   probability.Assignment
	Posteriors []Posterior
}
