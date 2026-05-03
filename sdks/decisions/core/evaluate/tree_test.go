package evaluate

import (
	"context"
	"errors"
	"testing"

	cexpressions "github.com/guidomantilla/yarumo/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestRunTree_BasicTraversal(t *testing.T) {
	t.Parallel()

	t.Run("left branch to leaf", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "income > 50000",
				True:      &schema.TreeNodeDef{Output: map[string]any{"risk": "low"}},
				False:     &schema.TreeNodeDef{Output: map[string]any{"risk": "high"}},
			},
		}
		exprCtx := cexpressions.Context{"income": 60000}

		result, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Tree {
			t.Fatalf("expected Tree paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Tree == nil {
			t.Fatal("expected non-nil tree outcome")
		}

		if result.Outcome.Tree.Outputs["risk"] != "low" {
			t.Fatalf("expected risk=low, got %v", result.Outcome.Tree.Outputs["risk"])
		}

		if len(result.Outcome.Tree.Path) != 1 {
			t.Fatalf("expected 1 path step, got %d", len(result.Outcome.Tree.Path))
		}
	})

	t.Run("right branch to leaf", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "income > 50000",
				True:      &schema.TreeNodeDef{Output: map[string]any{"risk": "low"}},
				False:     &schema.TreeNodeDef{Output: map[string]any{"risk": "high"}},
			},
		}
		exprCtx := cexpressions.Context{"income": 30000}

		result, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Tree.Outputs["risk"] != "high" {
			t.Fatalf("expected risk=high, got %v", result.Outcome.Tree.Outputs["risk"])
		}
	})

	t.Run("deep traversal", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "a > 0",
				True: &schema.TreeNodeDef{
					Condition: "b > 0",
					True:      &schema.TreeNodeDef{Output: map[string]any{"result": "both"}},
					False:     &schema.TreeNodeDef{Output: map[string]any{"result": "a_only"}},
				},
				False: &schema.TreeNodeDef{Output: map[string]any{"result": "none"}},
			},
		}
		exprCtx := cexpressions.Context{"a": 5, "b": 10}

		result, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Tree.Outputs["result"] != "both" {
			t.Fatalf("expected result=both, got %v", result.Outcome.Tree.Outputs["result"])
		}

		if len(result.Outcome.Tree.Path) != 2 {
			t.Fatalf("expected 2 path steps, got %d", len(result.Outcome.Tree.Path))
		}
	})

	t.Run("root is leaf", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Output: map[string]any{"default": true},
			},
		}
		exprCtx := cexpressions.Context{}

		result, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Tree.Outputs["default"] != true {
			t.Fatalf("expected default=true, got %v", result.Outcome.Tree.Outputs["default"])
		}

		if len(result.Outcome.Tree.Path) != 0 {
			t.Fatalf("expected 0 path steps, got %d", len(result.Outcome.Tree.Path))
		}
	})
}

func TestRunTree_MissingBranch(t *testing.T) {
	t.Parallel()

	t.Run("missing true branch", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "x > 0",
				True:      nil,
				False:     &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		_, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err == nil {
			t.Fatal("expected error for missing true branch")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatalf("expected ErrMissingConfig, got %v", err)
		}
	})

	t.Run("missing false branch", func(t *testing.T) {
		t.Parallel()

		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "x > 0",
				True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
				False:     nil,
			},
		}
		exprCtx := cexpressions.Context{"x": -5}

		_, err := runTree(context.Background(), config, exprCtx, NewOptions())

		if err == nil {
			t.Fatal("expected error for missing false branch")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatalf("expected ErrMissingConfig, got %v", err)
		}
	})
}

func TestRunTree_MissingCondition(t *testing.T) {
	t.Parallel()

	config := &schema.TreeConfig{
		Root: schema.TreeNodeDef{
			Condition: "",
			True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{}

	_, err := runTree(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for missing condition")
	}

	if !errors.Is(err, ErrMissingConfig) {
		t.Fatalf("expected ErrMissingConfig, got %v", err)
	}
}

func TestRunTree_ConditionError(t *testing.T) {
	t.Parallel()

	config := &schema.TreeConfig{
		Root: schema.TreeNodeDef{
			Condition: "(((",
			True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
			False:     &schema.TreeNodeDef{Output: map[string]any{"val": 2}},
		},
	}
	exprCtx := cexpressions.Context{}

	_, err := runTree(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for bad condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunTree_NonBoolCondition(t *testing.T) {
	t.Parallel()

	config := &schema.TreeConfig{
		Root: schema.TreeNodeDef{
			Condition: "x + 1",
			True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
			False:     &schema.TreeNodeDef{Output: map[string]any{"val": 2}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	_, err := runTree(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for non-bool condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunTree_ExplainError(t *testing.T) {
	t.Parallel()

	config := &schema.TreeConfig{
		Root: schema.TreeNodeDef{
			Output: map[string]any{"val": 1},
		},
	}
	exprCtx := cexpressions.Context{}

	failExplainer := &failingTreeExplainer{err: errors.New("explain boom")}
	opts := NewOptions(WithTreeExplainer(failExplainer))

	_, err := runTree(context.Background(), config, exprCtx, opts)

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatalf("expected ErrExplainFailed, got %v", err)
	}
}

type failingTreeExplainer struct {
	err error
}

func (e *failingTreeExplainer) ExplainTree(_ context.Context, _ explain.TreeTrace) (string, error) {
	return "", e.err
}

// Verify interface compliance.
var _ explain.TreeExplainer = (*failingTreeExplainer)(nil)
