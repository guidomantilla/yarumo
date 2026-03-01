// Package engine provides forward and backward chaining inference over propositional rules.
package engine

import (
	"github.com/guidomantilla/yarumo/maths/logic"

	"github.com/guidomantilla/yarumo/inference/classical/explain"
	"github.com/guidomantilla/yarumo/inference/classical/facts"
	"github.com/guidomantilla/yarumo/inference/classical/rules"
)

// Strategy defines the conflict resolution strategy for forward chaining.
type Strategy int

const (
	// PriorityOrder fires all applicable rules by priority each pass.
	PriorityOrder Strategy = iota
	// FirstMatch fires only the first applicable rule per pass.
	FirstMatch
)

// Result holds the outcome of an inference run.
type Result struct {
	Facts facts.FactBase
	Trace explain.Trace
	Steps int
}

// Engine defines the interface for a rule-based inference engine.
type Engine interface {
	// Forward runs forward chaining from the given initial facts and rule set.
	Forward(initialFacts logic.Fact, ruleSet []rules.Rule) Result
	// Backward attempts to prove the goal variable via backward chaining.
	Backward(initialFacts logic.Fact, ruleSet []rules.Rule, goal logic.Var) (bool, Result)
}

var _ Engine = (*engine)(nil)
