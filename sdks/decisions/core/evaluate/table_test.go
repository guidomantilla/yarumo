package evaluate

import (
	"context"
	"errors"
	"testing"

	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestRunTable_FirstPolicy(t *testing.T) {
	t.Parallel()

	t.Run("single match", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "first",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"age > 18"}, Outputs: map[string]any{"approved": true}},
				{Name: "r2", Conditions: []string{"age > 65"}, Outputs: map[string]any{"approved": false}},
			},
		}
		exprCtx := cexpressions.Context{"age": 25}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Table {
			t.Fatalf("expected Table paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Table == nil {
			t.Fatal("expected non-nil table outcome")
		}

		if len(result.Outcome.Table.MatchedRules) != 1 {
			t.Fatalf("expected 1 matched rule, got %d", len(result.Outcome.Table.MatchedRules))
		}

		if result.Outcome.Table.MatchedRules[0] != "r1" {
			t.Fatalf("expected r1, got %s", result.Outcome.Table.MatchedRules[0])
		}
	})

	t.Run("multiple matches returns first", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "first",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
				{Name: "r2", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 2}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Table.Outputs["val"] != 1 {
			t.Fatalf("expected val=1, got %v", result.Outcome.Table.Outputs["val"])
		}
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "first",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 100"}, Outputs: map[string]any{"val": 1}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		_, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err == nil {
			t.Fatal("expected error for no match")
		}

		if !errors.Is(err, ErrNoMatch) {
			t.Fatalf("expected ErrNoMatch, got %v", err)
		}
	})
}

func TestRunTable_UniquePolicy(t *testing.T) {
	t.Parallel()

	t.Run("exactly one match", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "unique",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 10"}, Outputs: map[string]any{"val": "a"}},
				{Name: "r2", Conditions: []string{"x < 5"}, Outputs: map[string]any{"val": "b"}},
			},
		}
		exprCtx := cexpressions.Context{"x": 3}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Table.MatchedRules[0] != "r2" {
			t.Fatalf("expected r2, got %s", result.Outcome.Table.MatchedRules[0])
		}
	})

	t.Run("multiple matches error", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "unique",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": "a"}},
				{Name: "r2", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": "b"}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		_, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err == nil {
			t.Fatal("expected error for multiple matches")
		}

		if !errors.Is(err, ErrMultipleMatches) {
			t.Fatalf("expected ErrMultipleMatches, got %v", err)
		}
	})

	t.Run("no match error", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "unique",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 100"}, Outputs: map[string]any{"val": "a"}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		_, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if !errors.Is(err, ErrNoMatch) {
			t.Fatalf("expected ErrNoMatch, got %v", err)
		}
	})
}

func TestRunTable_CollectPolicy(t *testing.T) {
	t.Parallel()

	t.Run("multiple matches merged", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "collect",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"a": 1}},
				{Name: "r2", Conditions: []string{"x > 0"}, Outputs: map[string]any{"b": 2}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Outcome.Table.MatchedRules) != 2 {
			t.Fatalf("expected 2 matched rules, got %d", len(result.Outcome.Table.MatchedRules))
		}

		if result.Outcome.Table.Outputs["a"] != 1 {
			t.Fatalf("expected a=1, got %v", result.Outcome.Table.Outputs["a"])
		}

		if result.Outcome.Table.Outputs["b"] != 2 {
			t.Fatalf("expected b=2, got %v", result.Outcome.Table.Outputs["b"])
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "collect",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 100"}, Outputs: map[string]any{"a": 1}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Outcome.Table.MatchedRules) != 0 {
			t.Fatalf("expected 0 matched rules, got %d", len(result.Outcome.Table.MatchedRules))
		}
	})
}

func TestRunTable_PriorityPolicy(t *testing.T) {
	t.Parallel()

	t.Run("highest priority wins", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "priority",
			Rules: []schema.TableRuleDef{
				{Name: "low", Priority: 1, Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": "low"}},
				{Name: "high", Priority: 10, Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": "high"}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		result, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Outcome.Table.Outputs["val"] != "high" {
			t.Fatalf("expected val=high, got %v", result.Outcome.Table.Outputs["val"])
		}
	})

	t.Run("no match error", func(t *testing.T) {
		t.Parallel()

		config := &schema.TableConfig{
			HitPolicy: "priority",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Priority: 1, Conditions: []string{"x > 100"}, Outputs: map[string]any{"val": 1}},
			},
		}
		exprCtx := cexpressions.Context{"x": 5}

		_, err := runTable(context.Background(), config, exprCtx, NewOptions())

		if !errors.Is(err, ErrNoMatch) {
			t.Fatalf("expected ErrNoMatch, got %v", err)
		}
	})
}

func TestRunTable_DefaultPolicy(t *testing.T) {
	t.Parallel()

	config := &schema.TableConfig{
		Rules: []schema.TableRuleDef{
			{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	result, err := runTable(context.Background(), config, exprCtx, NewOptions())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Outcome.Table.Outputs["val"] != 1 {
		t.Fatalf("expected val=1, got %v", result.Outcome.Table.Outputs["val"])
	}
}

func TestRunTable_InvalidHitPolicy(t *testing.T) {
	t.Parallel()

	config := &schema.TableConfig{
		HitPolicy: "invalid",
		Rules: []schema.TableRuleDef{
			{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	_, err := runTable(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for invalid hit policy")
	}

	if !errors.Is(err, ErrInvalidHitPolicy) {
		t.Fatalf("expected ErrInvalidHitPolicy, got %v", err)
	}
}

func TestRunTable_ConditionError(t *testing.T) {
	t.Parallel()

	config := &schema.TableConfig{
		Rules: []schema.TableRuleDef{
			{Name: "r1", Conditions: []string{"((("}, Outputs: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	_, err := runTable(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for bad condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunTable_ConditionNonBool(t *testing.T) {
	t.Parallel()

	config := &schema.TableConfig{
		Rules: []schema.TableRuleDef{
			{Name: "r1", Conditions: []string{"x + 1"}, Outputs: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	_, err := runTable(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for non-bool condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunTable_ExplainError(t *testing.T) {
	t.Parallel()

	config := &schema.TableConfig{
		Rules: []schema.TableRuleDef{
			{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	failExplainer := &failingTableExplainer{err: errors.New("explain boom")}
	opts := NewOptions(WithTableExplainer(failExplainer))

	_, err := runTable(context.Background(), config, exprCtx, opts)

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatalf("expected ErrExplainFailed, got %v", err)
	}
}

type failingTableExplainer struct {
	err error
}

func (e *failingTableExplainer) ExplainTable(_ context.Context, _ explain.TableTrace) (string, error) {
	return "", e.err
}

// Verify interface compliance.
var _ explain.TableExplainer = (*failingTableExplainer)(nil)
