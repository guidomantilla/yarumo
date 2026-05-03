// Package evaluate provides unified decision evaluation over six paradigms:
// deductive, Bayesian, fuzzy, table, scorecard, and tree.
package evaluate

import (
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// Paradigm identifies the reasoning paradigm for a decision.
type Paradigm int

const (
	// Deductive uses propositional logic with forward or backward chaining.
	Deductive Paradigm = iota
	// Bayesian uses probabilistic inference over a Bayesian network.
	Bayesian
	// Fuzzy uses fuzzy inference with linguistic variables.
	Fuzzy
	// Table uses a decision table with hit policies.
	Table
	// Scorecard uses a weighted scorecard with bin matching.
	Scorecard
	// Tree uses a binary decision tree with expression conditions.
	Tree
)

// paradigm name constants.
const (
	paradigmDeductive = "deductive"
	paradigmBayesian  = "bayesian"
	paradigmFuzzy     = "fuzzy"
	paradigmTable     = "table"
	paradigmScorecard = "scorecard"
	paradigmTree      = "tree"
	paradigmUnknown   = "unknown"
)

// String returns the string representation of a Paradigm.
func (p Paradigm) String() string {
	switch p {
	case Deductive:
		return paradigmDeductive
	case Bayesian:
		return paradigmBayesian
	case Fuzzy:
		return paradigmFuzzy
	case Table:
		return paradigmTable
	case Scorecard:
		return paradigmScorecard
	case Tree:
		return paradigmTree
	default:
		return paradigmUnknown
	}
}

// Request holds the input for a decision execution.
type Request[D any] struct {
	// Domain is the application-specific data to bind to the engine.
	Domain D
	// RuleSetName identifies which ruleset to load from the repository.
	RuleSetName string
	// RuleSetVersion identifies which version of the ruleset to load.
	RuleSetVersion string
	// Paradigm selects the reasoning paradigm to execute.
	Paradigm Paradigm
	// Query is the target variable for Bayesian inference (ignored for other paradigms).
	Query string
	// Metadata carries optional context for auditing and tracing.
	Metadata map[string]any
}

// Result holds the outcome of a single-paradigm decision execution.
// Note: Result and CascadeResult are intentionally kept as separate types.
// They serve distinct purposes: Result is a single-paradigm outcome, while
// CascadeResult aggregates multiple stages with inter-stage explanations.
type Result struct {
	// Outcome contains the paradigm-specific decision output.
	Outcome Outcome
	// Explanation is a human-readable explanation of the decision.
	Explanation string
	// Paradigm identifies which reasoning paradigm produced this result.
	Paradigm Paradigm
}

// Outcome holds the paradigm-specific output of a decision.
type Outcome struct {
	// Facts holds derived variable values from deductive inference.
	Facts map[logic.Var]bool
	// Distribution holds the posterior probability distribution from Bayesian inference.
	Distribution stats.Distribution
	// Outputs holds crisp output values from fuzzy inference.
	Outputs map[string]float64
	// Table holds the decision table output (non-nil when paradigm is Table).
	Table *TableOutcome
	// Score holds the scorecard output (non-nil when paradigm is Scorecard).
	Score *ScoreOutcome
	// Tree holds the decision tree output (non-nil when paradigm is Tree).
	Tree *TreeOutcome
}

// TableOutcome holds the result of a decision table evaluation.
type TableOutcome struct {
	// MatchedRules lists the names of rules that matched.
	MatchedRules []string
	// Outputs holds the merged output values.
	Outputs map[string]any
}

// ScoreOutcome holds the result of a scorecard evaluation.
type ScoreOutcome struct {
	// TotalScore is the final computed score.
	TotalScore float64
	// Breakdown maps attribute name to weighted points.
	Breakdown map[string]float64
}

// TreeOutcome holds the result of a decision tree evaluation.
type TreeOutcome struct {
	// Path lists the conditions evaluated along the tree traversal.
	Path []string
	// Outputs holds the output values from the reached leaf node.
	Outputs map[string]any
}

// CascadeResult holds the aggregated results of a cascade pipeline execution.
type CascadeResult struct {
	// Stages holds the result of each individual stage.
	Stages []StageResult
	// Final is the result of the last stage.
	Final Result
	// Explanation is a combined human-readable explanation.
	Explanation string
}

// StageResult holds the result of a single cascade stage.
type StageResult struct {
	// Name identifies the stage.
	Name string
	// Result is the decision result of this stage.
	Result Result
}

// StageConverter converts the result of a cascade stage to input for the next stage.
// The returned value must be logic.Fact, evidence.EvidenceBase, or map[string]float64
// depending on the next stage paradigm.
type StageConverter func(previous Result) (any, error)

// Valid hit policies for decision tables.
const (
	HitPolicyFirst    = "first"
	HitPolicyUnique   = "unique"
	HitPolicyCollect  = "collect"
	HitPolicyPriority = "priority"
)
