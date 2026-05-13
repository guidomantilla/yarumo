package evaluate

import (
	cexpressions "github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
)

// Options holds configuration for a Service.
type Options struct {
	deductiveExplainer explain.DeductiveExplainer
	bayesianExplainer  explain.BayesianExplainer
	fuzzyExplainer     explain.FuzzyExplainer
	tableExplainer     explain.TableExplainer
	scorecardExplainer explain.ScorecardExplainer
	treeExplainer      explain.TreeExplainer
	auditLog           Log
	expressionOpts     []cexpressions.Option
}

// Option is a functional option for configuring Service Options.
type Option func(*Options)

// NewOptions creates Options from the given functional options.
func NewOptions(opts ...Option) *Options {
	defaultExplainer := explain.NewTemplateExplainer(explain.English)

	o := &Options{
		deductiveExplainer: defaultExplainer,
		bayesianExplainer:  defaultExplainer,
		fuzzyExplainer:     defaultExplainer,
		tableExplainer:     defaultExplainer,
		scorecardExplainer: defaultExplainer,
		treeExplainer:      defaultExplainer,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithExplainer sets all six paradigm explainers to the given Explainer.
func WithExplainer(e explain.Explainer) Option {
	return func(o *Options) {
		if e != nil {
			o.deductiveExplainer = e
			o.bayesianExplainer = e
			o.fuzzyExplainer = e
			o.tableExplainer = e
			o.scorecardExplainer = e
			o.treeExplainer = e
		}
	}
}

// WithDeductiveExplainer sets the explainer for deductive inference results.
func WithDeductiveExplainer(e explain.DeductiveExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.deductiveExplainer = e
		}
	}
}

// WithBayesianExplainer sets the explainer for Bayesian inference results.
func WithBayesianExplainer(e explain.BayesianExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.bayesianExplainer = e
		}
	}
}

// WithFuzzyExplainer sets the explainer for fuzzy inference results.
func WithFuzzyExplainer(e explain.FuzzyExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.fuzzyExplainer = e
		}
	}
}

// WithTableExplainer sets the explainer for decision table results.
func WithTableExplainer(e explain.TableExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.tableExplainer = e
		}
	}
}

// WithScorecardExplainer sets the explainer for scorecard results.
func WithScorecardExplainer(e explain.ScorecardExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.scorecardExplainer = e
		}
	}
}

// WithTreeExplainer sets the explainer for decision tree results.
func WithTreeExplainer(e explain.TreeExplainer) Option {
	return func(o *Options) {
		if e != nil {
			o.treeExplainer = e
		}
	}
}

// WithAuditLog sets the AuditLog implementation. If nil, auditing is disabled.
func WithAuditLog(l Log) Option {
	return func(o *Options) {
		if l != nil {
			o.auditLog = l
		}
	}
}

// WithExpressionFunc registers a custom function for use in expressions.
func WithExpressionFunc(name string, fn cexpressions.Func) Option {
	return func(o *Options) {
		if fn != nil {
			o.expressionOpts = append(o.expressionOpts, cexpressions.WithFunc(name, fn))
		}
	}
}

func (o *Options) explainers() explainerSet {
	return explainerSet{
		deductive: o.deductiveExplainer,
		bayesian:  o.bayesianExplainer,
		fuzzy:     o.fuzzyExplainer,
		table:     o.tableExplainer,
		scorecard: o.scorecardExplainer,
		tree:      o.treeExplainer,
	}
}
