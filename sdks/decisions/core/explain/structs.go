package explain

// DeductiveTrace holds extracted trace data from a deductive inference result.
type DeductiveTrace struct {
	// Steps is the number of forward-chaining steps executed.
	Steps int
	// Reasons lists the derived facts with their provenance.
	Reasons []DeductiveReason
}

// DeductiveReason describes one derived fact in a deductive inference.
type DeductiveReason struct {
	// Variable is the name of the derived propositional variable.
	Variable string
	// Value is the truth value assigned to the variable.
	Value bool
	// RuleName is the rule that derived this fact.
	RuleName string
	// Step is the forward-chaining step number where derivation occurred.
	Step int
}

// BayesianTrace holds extracted trace data from a Bayesian inference result.
type BayesianTrace struct {
	// Query is the target variable that was queried.
	Query string
	// Factors lists the posterior probabilities for each outcome.
	Factors []BayesianFactor
}

// BayesianFactor describes one outcome in a Bayesian posterior distribution.
type BayesianFactor struct {
	// Outcome is the name of the outcome.
	Outcome string
	// Probability is the posterior probability of this outcome.
	Probability float64
}

// FuzzyTrace holds extracted trace data from a fuzzy inference result.
type FuzzyTrace struct {
	// Outputs lists the crisp output values.
	Outputs []FuzzyOutput
	// Memberships lists the fuzzification results.
	Memberships []FuzzyMembership
}

// FuzzyOutput describes one output variable value from fuzzy inference.
type FuzzyOutput struct {
	// Variable is the name of the output variable.
	Variable string
	// Value is the defuzzified crisp value.
	Value float64
}

// FuzzyMembership describes one fuzzification result.
type FuzzyMembership struct {
	// Variable is the name of the input variable.
	Variable string
	// Term is the linguistic term.
	Term string
	// Degree is the membership degree.
	Degree float64
}

// TableTrace holds extracted trace data from a decision table evaluation.
type TableTrace struct {
	// HitPolicy is the hit policy used for the table evaluation.
	HitPolicy string
	// MatchedRules lists the rules that matched during evaluation.
	MatchedRules []TableMatchEntry
	// Outputs holds the final merged output values.
	Outputs map[string]any
}

// TableMatchEntry describes one matched rule in a decision table evaluation.
type TableMatchEntry struct {
	// RuleName is the name of the matched rule.
	RuleName string
	// Outputs holds the output values produced by this rule.
	Outputs map[string]any
}

// ScoreTrace holds extracted trace data from a scorecard evaluation.
type ScoreTrace struct {
	// BaseScore is the initial base score.
	BaseScore float64
	// TotalScore is the final computed score.
	TotalScore float64
	// Breakdown lists each attribute contribution.
	Breakdown []ScoreEntry
}

// ScoreEntry describes one attribute contribution in a scorecard evaluation.
type ScoreEntry struct {
	// Attribute is the name of the scorecard attribute.
	Attribute string
	// Points is the raw points from the matched bin.
	Points float64
	// Weight is the attribute weight.
	Weight float64
	// Weighted is the weighted contribution (Points * Weight).
	Weighted float64
}

// TreeTrace holds extracted trace data from a decision tree evaluation.
type TreeTrace struct {
	// Path lists the conditions evaluated along the tree traversal.
	Path []TreeStep
	// Outputs holds the output values from the reached leaf node.
	Outputs map[string]any
}

// TreeStep describes one condition evaluation in a decision tree traversal.
type TreeStep struct {
	// Condition is the expression that was evaluated.
	Condition string
	// Result is the boolean result of the condition evaluation.
	Result bool
}
