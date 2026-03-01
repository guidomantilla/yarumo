// Package rules provides rule definitions for propositional inference.
package rules

import "github.com/guidomantilla/yarumo/maths/logic"

// Rule represents a production rule with a condition and conclusion.
type Rule interface {
	// Name returns the rule identifier.
	Name() string
	// Priority returns the rule priority (lower = higher priority).
	Priority() int
	// Condition returns the propositional formula that must be satisfied.
	Condition() logic.Formula
	// Conclusion returns the variable assignments produced when the rule fires.
	Conclusion() map[logic.Var]bool
	// Fires reports whether the rule condition is satisfied by the given facts.
	Fires(facts logic.Fact) bool
	// Produces reports whether the rule would derive new information from the given facts.
	Produces(facts logic.Fact) bool
}

var _ Rule = (*rule)(nil)
