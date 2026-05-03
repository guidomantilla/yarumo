package evaluate

import (
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"
)

// DeductiveBinder translates domain data into propositional facts for deductive inference.
type DeductiveBinder[D any] interface {
	// BindDeductive converts domain data to propositional facts.
	BindDeductive(domain D) logic.Fact
}

// BayesianBinder translates domain data into an evidence base for Bayesian inference.
type BayesianBinder[D any] interface {
	// BindBayesian converts domain data to an evidence base.
	BindBayesian(domain D) evidence.EvidenceBase
}

// FuzzyBinder translates domain data into crisp input values for fuzzy inference.
type FuzzyBinder[D any] interface {
	// BindFuzzy converts domain data to crisp input values.
	BindFuzzy(domain D) map[string]float64
}

// ExpressionBinder translates domain data into an expression evaluation context.
type ExpressionBinder[D any] interface {
	// BindExpression converts domain data to an expression context.
	BindExpression(domain D) cexpressions.Context
}

// Binder combines all four paradigm binders for convenience.
// Use individual binder interfaces when only one paradigm is needed.
type Binder[D any] interface {
	DeductiveBinder[D]
	BayesianBinder[D]
	FuzzyBinder[D]
	ExpressionBinder[D]
}
