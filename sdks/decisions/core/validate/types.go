// Package validate provides pre-deploy validation of ruleset configurations.
package validate

import (
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

var _ Validator = (*validator)(nil)

// Validator defines the interface for ruleset validation.
type Validator interface {
	// ValidateDeductive validates a deductive ruleset configuration.
	ValidateDeductive(config *schema.DeductiveConfig) Report
	// ValidateBayesian validates a Bayesian network configuration.
	ValidateBayesian(config *schema.BayesianConfig) Report
	// ValidateFuzzy validates a fuzzy inference configuration.
	ValidateFuzzy(config *schema.FuzzyConfig) Report
	// ValidateTable validates a decision table configuration.
	ValidateTable(config *schema.TableConfig) Report
	// ValidateScorecard validates a scorecard configuration.
	ValidateScorecard(config *schema.ScorecardConfig) Report
	// ValidateTree validates a decision tree configuration.
	ValidateTree(config *schema.TreeConfig) Report
}

// Report holds the results of a ruleset validation.
type Report struct {
	// Parsed is the number of rules successfully parsed.
	Parsed int
	// Contradictions lists pairs of contradictory rules.
	Contradictions []ConflictPair
	// Redundant lists rules implied by other rules.
	Redundant []RedundantRule
	// Gaps lists input combinations where no rule fires.
	Gaps []logic.Fact
	// Simplified lists rules with their simplified equivalents.
	Simplified []SimplifiedRule
	// Errors lists structural or parse errors found during validation.
	Errors []string
	// Valid is true when no contradictions, no errors, and all rules parsed.
	Valid bool
}

// ConflictPair describes two rules that contradict each other.
type ConflictPair struct {
	RuleA  string
	RuleB  string
	Detail string
}

// RedundantRule describes a rule that is logically implied by other rules.
type RedundantRule struct {
	Rule      string
	ImpliedBy []string
}

// SimplifiedRule pairs an original rule condition with its simplified form.
type SimplifiedRule struct {
	RuleName   string
	Original   string
	Simplified string
}
