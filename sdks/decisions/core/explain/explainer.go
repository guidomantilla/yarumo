package explain

import (
	"bytes"
	"context"
	"text/template"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// templateExplainer generates explanations using pre-compiled Go text/template templates.
type templateExplainer struct {
	deductive *template.Template
	bayesian  *template.Template
	fuzzy     *template.Template
	table     *template.Template
	scorecard *template.Template
	tree      *template.Template
}

// NewTemplateExplainer creates a new Explainer for the given locale.
func NewTemplateExplainer(locale Locale) Explainer {
	if locale != Spanish && locale != English {
		locale = English
	}

	return &templateExplainer{
		deductive: deductiveTemplates[locale],
		bayesian:  bayesianTemplates[locale],
		fuzzy:     fuzzyTemplates[locale],
		table:     tableTemplates[locale],
		scorecard: scorecardTemplates[locale],
		tree:      treeTemplates[locale],
	}
}

// ExplainDeductive generates an explanation for a deductive inference trace.
func (e *templateExplainer) ExplainDeductive(_ context.Context, trace DeductiveTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.deductive, trace)
}

// ExplainBayesian generates an explanation for a Bayesian inference trace.
func (e *templateExplainer) ExplainBayesian(_ context.Context, trace BayesianTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.bayesian, trace)
}

// ExplainFuzzy generates an explanation for a fuzzy inference trace.
func (e *templateExplainer) ExplainFuzzy(_ context.Context, trace FuzzyTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.fuzzy, trace)
}

// ExplainTable generates an explanation for a decision table trace.
func (e *templateExplainer) ExplainTable(_ context.Context, trace TableTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.table, trace)
}

// ExplainScorecard generates an explanation for a scorecard trace.
func (e *templateExplainer) ExplainScorecard(_ context.Context, trace ScoreTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.scorecard, trace)
}

// ExplainTree generates an explanation for a decision tree trace.
func (e *templateExplainer) ExplainTree(_ context.Context, trace TreeTrace) (string, error) {
	cassert.NotNil(e, "explainer is nil")

	return render(e.tree, trace)
}

func render(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer

	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", ErrRender(err)
	}

	return buf.String(), nil
}
