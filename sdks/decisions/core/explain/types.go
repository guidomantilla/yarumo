// Package explain provides the Explainer interface and a default template-based implementation.
package explain

import (
	"context"
)

// DeductiveExplainer generates human-readable explanations for deductive inference results.
type DeductiveExplainer interface {
	// ExplainDeductive generates an explanation for a deductive inference trace.
	ExplainDeductive(ctx context.Context, trace DeductiveTrace) (string, error)
}

// BayesianExplainer generates human-readable explanations for Bayesian inference results.
type BayesianExplainer interface {
	// ExplainBayesian generates an explanation for a Bayesian inference trace.
	ExplainBayesian(ctx context.Context, trace BayesianTrace) (string, error)
}

// FuzzyExplainer generates human-readable explanations for fuzzy inference results.
type FuzzyExplainer interface {
	// ExplainFuzzy generates an explanation for a fuzzy inference trace.
	ExplainFuzzy(ctx context.Context, trace FuzzyTrace) (string, error)
}

// TableExplainer generates human-readable explanations for decision table results.
type TableExplainer interface {
	// ExplainTable generates an explanation for a decision table trace.
	ExplainTable(ctx context.Context, trace TableTrace) (string, error)
}

// ScorecardExplainer generates human-readable explanations for scorecard results.
type ScorecardExplainer interface {
	// ExplainScorecard generates an explanation for a scorecard trace.
	ExplainScorecard(ctx context.Context, trace ScoreTrace) (string, error)
}

// TreeExplainer generates human-readable explanations for decision tree results.
type TreeExplainer interface {
	// ExplainTree generates an explanation for a decision tree trace.
	ExplainTree(ctx context.Context, trace TreeTrace) (string, error)
}

var _ Explainer = (*templateExplainer)(nil)

// Explainer generates human-readable explanations from inference traces.
// The SDK provides a default template-based implementation; applications may implement this
// interface with AI-enhanced explanations (e.g., via Claude API).
type Explainer interface {
	DeductiveExplainer
	BayesianExplainer
	FuzzyExplainer
	TableExplainer
	ScorecardExplainer
	TreeExplainer
}

// Locale selects the language for template-based explanations.
type Locale string

// Supported locales.
const (
	Spanish Locale = "es"
	English Locale = "en"
)
