package evaluate

import (
	"context"
	"testing"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.deductiveExplainer == nil {
			t.Fatal("expected default deductive explainer")
		}

		if opts.bayesianExplainer == nil {
			t.Fatal("expected default bayesian explainer")
		}

		if opts.fuzzyExplainer == nil {
			t.Fatal("expected default fuzzy explainer")
		}

		if opts.tableExplainer == nil {
			t.Fatal("expected default table explainer")
		}

		if opts.scorecardExplainer == nil {
			t.Fatal("expected default scorecard explainer")
		}

		if opts.treeExplainer == nil {
			t.Fatal("expected default tree explainer")
		}

		if opts.auditLog != nil {
			t.Fatal("expected nil auditLog by default")
		}
	})

	t.Run("with explainer sets all six", func(t *testing.T) {
		t.Parallel()

		custom := explain.NewTemplateExplainer(explain.Spanish)
		opts := NewOptions(WithExplainer(custom))

		if opts.deductiveExplainer != custom {
			t.Fatal("expected custom deductive explainer")
		}

		if opts.bayesianExplainer != custom {
			t.Fatal("expected custom bayesian explainer")
		}

		if opts.fuzzyExplainer != custom {
			t.Fatal("expected custom fuzzy explainer")
		}

		if opts.tableExplainer != custom {
			t.Fatal("expected custom table explainer")
		}

		if opts.scorecardExplainer != custom {
			t.Fatal("expected custom scorecard explainer")
		}

		if opts.treeExplainer != custom {
			t.Fatal("expected custom tree explainer")
		}
	})

	t.Run("with nil explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithExplainer(nil))

		if opts.deductiveExplainer == nil {
			t.Fatal("expected default explainer when nil passed")
		}
	})

	t.Run("with nil audit log is noop", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAuditLog(nil))

		if opts.auditLog != nil {
			t.Fatal("expected nil auditLog")
		}
	})

	t.Run("with deductive explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testDeductiveExplainer{}
		opts := NewOptions(WithDeductiveExplainer(custom))

		if opts.deductiveExplainer != custom {
			t.Fatal("expected custom deductive explainer")
		}
	})

	t.Run("with bayesian explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testBayesianExplainer{}
		opts := NewOptions(WithBayesianExplainer(custom))

		if opts.bayesianExplainer != custom {
			t.Fatal("expected custom bayesian explainer")
		}
	})

	t.Run("with fuzzy explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testFuzzyExplainer{}
		opts := NewOptions(WithFuzzyExplainer(custom))

		if opts.fuzzyExplainer != custom {
			t.Fatal("expected custom fuzzy explainer")
		}
	})

	t.Run("with table explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testTableExplainer{}
		opts := NewOptions(WithTableExplainer(custom))

		if opts.tableExplainer != custom {
			t.Fatal("expected custom table explainer")
		}
	})

	t.Run("with scorecard explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testScorecardExplainer{}
		opts := NewOptions(WithScorecardExplainer(custom))

		if opts.scorecardExplainer != custom {
			t.Fatal("expected custom scorecard explainer")
		}
	})

	t.Run("with tree explainer", func(t *testing.T) {
		t.Parallel()

		custom := &testTreeExplainer{}
		opts := NewOptions(WithTreeExplainer(custom))

		if opts.treeExplainer != custom {
			t.Fatal("expected custom tree explainer")
		}
	})

	t.Run("with nil deductive explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDeductiveExplainer(nil))

		if opts.deductiveExplainer == nil {
			t.Fatal("expected default deductive explainer when nil passed")
		}
	})

	t.Run("with nil bayesian explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithBayesianExplainer(nil))

		if opts.bayesianExplainer == nil {
			t.Fatal("expected default bayesian explainer when nil passed")
		}
	})

	t.Run("with nil fuzzy explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithFuzzyExplainer(nil))

		if opts.fuzzyExplainer == nil {
			t.Fatal("expected default fuzzy explainer when nil passed")
		}
	})

	t.Run("with nil table explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTableExplainer(nil))

		if opts.tableExplainer == nil {
			t.Fatal("expected default table explainer when nil passed")
		}
	})

	t.Run("with nil scorecard explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithScorecardExplainer(nil))

		if opts.scorecardExplainer == nil {
			t.Fatal("expected default scorecard explainer when nil passed")
		}
	})

	t.Run("with nil tree explainer keeps default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTreeExplainer(nil))

		if opts.treeExplainer == nil {
			t.Fatal("expected default tree explainer when nil passed")
		}
	})

	t.Run("explainers returns explainerSet", func(t *testing.T) {
		t.Parallel()

		custom := explain.NewTemplateExplainer(explain.Spanish)
		opts := NewOptions(WithExplainer(custom))
		es := opts.explainers()

		if es.deductive != custom {
			t.Fatal("expected custom deductive in explainerSet")
		}

		if es.bayesian != custom {
			t.Fatal("expected custom bayesian in explainerSet")
		}

		if es.fuzzy != custom {
			t.Fatal("expected custom fuzzy in explainerSet")
		}

		if es.table != custom {
			t.Fatal("expected custom table in explainerSet")
		}

		if es.scorecard != custom {
			t.Fatal("expected custom scorecard in explainerSet")
		}

		if es.tree != custom {
			t.Fatal("expected custom tree in explainerSet")
		}
	})

	t.Run("with expression func", func(t *testing.T) {
		t.Parallel()

		fn := func(_ ...any) (any, error) { return 0, nil }
		opts := NewOptions(WithExpressionFunc("test", fn))

		if len(opts.expressionOpts) != 1 {
			t.Fatalf("expected 1 expression opt, got %d", len(opts.expressionOpts))
		}
	})

	t.Run("with nil expression func", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithExpressionFunc("test", nil))

		if len(opts.expressionOpts) != 0 {
			t.Fatalf("expected 0 expression opts, got %d", len(opts.expressionOpts))
		}
	})
}

// test doubles for segregated explainers.

type testDeductiveExplainer struct{}

func (e *testDeductiveExplainer) ExplainDeductive(_ context.Context, _ explain.DeductiveTrace) (string, error) {
	return "deductive", nil
}

type testBayesianExplainer struct{}

func (e *testBayesianExplainer) ExplainBayesian(_ context.Context, _ explain.BayesianTrace) (string, error) {
	return "bayesian", nil
}

type testFuzzyExplainer struct{}

func (e *testFuzzyExplainer) ExplainFuzzy(_ context.Context, _ explain.FuzzyTrace) (string, error) {
	return "fuzzy", nil
}

type testTableExplainer struct{}

func (e *testTableExplainer) ExplainTable(_ context.Context, _ explain.TableTrace) (string, error) {
	return "table", nil
}

type testScorecardExplainer struct{}

func (e *testScorecardExplainer) ExplainScorecard(_ context.Context, _ explain.ScoreTrace) (string, error) {
	return "scorecard", nil
}

type testTreeExplainer struct{}

func (e *testTreeExplainer) ExplainTree(_ context.Context, _ explain.TreeTrace) (string, error) {
	return "tree", nil
}

var _ explain.DeductiveExplainer = (*testDeductiveExplainer)(nil)

var _ explain.BayesianExplainer = (*testBayesianExplainer)(nil)

var _ explain.FuzzyExplainer = (*testFuzzyExplainer)(nil)

var _ explain.TableExplainer = (*testTableExplainer)(nil)

var _ explain.ScorecardExplainer = (*testScorecardExplainer)(nil)

var _ explain.TreeExplainer = (*testTreeExplainer)(nil)
