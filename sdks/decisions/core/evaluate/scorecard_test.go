package evaluate

import (
	"context"
	"errors"
	"math"
	"testing"

	cexpressions "github.com/guidomantilla/yarumo/core/common/expressions"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestRunScorecard_BasicScoring(t *testing.T) {
	t.Parallel()

	t.Run("single attribute match", func(t *testing.T) {
		t.Parallel()

		config := &schema.ScorecardConfig{
			BaseScore: 100,
			Attributes: []schema.ScorecardAttributeDef{
				{
					Name:   "income",
					Weight: 1.0,
					Bins: []schema.ScorecardBinDef{
						{Condition: "income > 50000", Points: 50},
						{Condition: "income > 30000", Points: 30},
					},
				},
			},
		}
		exprCtx := cexpressions.Context{"income": 60000}

		result, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Scorecard {
			t.Fatalf("expected Scorecard paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Score == nil {
			t.Fatal("expected non-nil score outcome")
		}

		if math.Abs(result.Outcome.Score.TotalScore-150) > 0.001 {
			t.Fatalf("expected 150, got %f", result.Outcome.Score.TotalScore)
		}
	})

	t.Run("multiple attributes", func(t *testing.T) {
		t.Parallel()

		config := &schema.ScorecardConfig{
			BaseScore: 0,
			Attributes: []schema.ScorecardAttributeDef{
				{
					Name:   "age",
					Weight: 2.0,
					Bins: []schema.ScorecardBinDef{
						{Condition: "age > 30", Points: 10},
					},
				},
				{
					Name:   "income",
					Weight: 1.5,
					Bins: []schema.ScorecardBinDef{
						{Condition: "income > 50000", Points: 20},
					},
				},
			},
		}
		exprCtx := cexpressions.Context{"age": 35, "income": 60000}

		result, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := (10.0 * 2.0) + (20.0 * 1.5)
		if math.Abs(result.Outcome.Score.TotalScore-expected) > 0.001 {
			t.Fatalf("expected %f, got %f", expected, result.Outcome.Score.TotalScore)
		}
	})

	t.Run("no bin match skips attribute", func(t *testing.T) {
		t.Parallel()

		config := &schema.ScorecardConfig{
			BaseScore: 100,
			Attributes: []schema.ScorecardAttributeDef{
				{
					Name:   "age",
					Weight: 1.0,
					Bins: []schema.ScorecardBinDef{
						{Condition: "age > 100", Points: 50},
					},
				},
			},
		}
		exprCtx := cexpressions.Context{"age": 25}

		result, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if math.Abs(result.Outcome.Score.TotalScore-100) > 0.001 {
			t.Fatalf("expected 100, got %f", result.Outcome.Score.TotalScore)
		}
	})

	t.Run("first matching bin wins", func(t *testing.T) {
		t.Parallel()

		config := &schema.ScorecardConfig{
			BaseScore: 0,
			Attributes: []schema.ScorecardAttributeDef{
				{
					Name:   "score",
					Weight: 1.0,
					Bins: []schema.ScorecardBinDef{
						{Condition: "score > 90", Points: 100},
						{Condition: "score > 70", Points: 75},
						{Condition: "score > 50", Points: 50},
					},
				},
			},
		}
		exprCtx := cexpressions.Context{"score": 80}

		result, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if math.Abs(result.Outcome.Score.TotalScore-75) > 0.001 {
			t.Fatalf("expected 75, got %f", result.Outcome.Score.TotalScore)
		}
	})
}

func TestRunScorecard_ConditionError(t *testing.T) {
	t.Parallel()

	config := &schema.ScorecardConfig{
		Attributes: []schema.ScorecardAttributeDef{
			{
				Name:   "bad",
				Weight: 1.0,
				Bins: []schema.ScorecardBinDef{
					{Condition: "(((", Points: 10},
				},
			},
		},
	}
	exprCtx := cexpressions.Context{}

	_, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for bad condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunScorecard_NonBoolCondition(t *testing.T) {
	t.Parallel()

	config := &schema.ScorecardConfig{
		Attributes: []schema.ScorecardAttributeDef{
			{
				Name:   "bad",
				Weight: 1.0,
				Bins: []schema.ScorecardBinDef{
					{Condition: "x + 1", Points: 10},
				},
			},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	_, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

	if err == nil {
		t.Fatal("expected error for non-bool condition")
	}

	if !errors.Is(err, ErrConditionEval) {
		t.Fatalf("expected ErrConditionEval, got %v", err)
	}
}

func TestRunScorecard_ExplainError(t *testing.T) {
	t.Parallel()

	config := &schema.ScorecardConfig{
		BaseScore: 100,
		Attributes: []schema.ScorecardAttributeDef{
			{
				Name:   "x",
				Weight: 1.0,
				Bins: []schema.ScorecardBinDef{
					{Condition: "x > 0", Points: 10},
				},
			},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	failExplainer := &failingScorecardExplainer{err: errors.New("explain boom")}
	opts := NewOptions(WithScorecardExplainer(failExplainer))

	_, err := runScorecard(context.Background(), config, exprCtx, opts)

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatalf("expected ErrExplainFailed, got %v", err)
	}
}

func TestRunScorecard_Breakdown(t *testing.T) {
	t.Parallel()

	config := &schema.ScorecardConfig{
		BaseScore: 0,
		Attributes: []schema.ScorecardAttributeDef{
			{
				Name:   "a",
				Weight: 2.0,
				Bins: []schema.ScorecardBinDef{
					{Condition: "x > 0", Points: 10},
				},
			},
		},
	}
	exprCtx := cexpressions.Context{"x": 5}

	result, err := runScorecard(context.Background(), config, exprCtx, NewOptions())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Outcome.Score.Breakdown["a"]
	if !ok {
		t.Fatal("expected 'a' in breakdown")
	}

	if math.Abs(val-20) > 0.001 {
		t.Fatalf("expected 20, got %f", val)
	}
}

type failingScorecardExplainer struct {
	err error
}

func (e *failingScorecardExplainer) ExplainScorecard(_ context.Context, _ explain.ScoreTrace) (string, error) {
	return "", e.err
}

// Verify interface compliance.
var _ explain.ScorecardExplainer = (*failingScorecardExplainer)(nil)
