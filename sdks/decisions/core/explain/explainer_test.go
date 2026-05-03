package explain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"text/template"
)

func TestNewTemplateExplainer(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		e := NewTemplateExplainer(English)
		if e == nil {
			t.Fatal("expected non-nil explainer")
		}

		result, err := e.ExplainDeductive(context.Background(), DeductiveTrace{Steps: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "step") {
			t.Fatalf("expected english output, got: %s", result)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		e := NewTemplateExplainer(Spanish)
		if e == nil {
			t.Fatal("expected non-nil explainer")
		}

		result, err := e.ExplainDeductive(context.Background(), DeductiveTrace{Steps: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "paso") {
			t.Fatalf("expected spanish output, got: %s", result)
		}
	})

	t.Run("invalid defaults to english", func(t *testing.T) {
		t.Parallel()

		e := NewTemplateExplainer("fr")
		if e == nil {
			t.Fatal("expected non-nil explainer")
		}

		result, err := e.ExplainDeductive(context.Background(), DeductiveTrace{Steps: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "step") {
			t.Fatalf("expected english default output, got: %s", result)
		}
	})
}

func TestTemplateExplainer_ExplainDeductive(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		trace := DeductiveTrace{
			Steps: 1,
			Reasons: []DeductiveReason{
				{Variable: "wet", Value: true, RuleName: "rain-wet", Step: 1},
			},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainDeductive(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "step") {
			t.Fatalf("expected 'step' in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := DeductiveTrace{
			Steps: 1,
			Reasons: []DeductiveReason{
				{Variable: "wet", Value: true, RuleName: "rain-wet", Step: 1},
			},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainDeductive(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "paso") {
			t.Fatalf("expected 'paso' in explanation, got: %s", explanation)
		}
	})

	t.Run("no derived facts", func(t *testing.T) {
		t.Parallel()

		trace := DeductiveTrace{
			Steps:   0,
			Reasons: nil,
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainDeductive(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "0 step") {
			t.Fatalf("expected '0 step' in explanation, got: %s", explanation)
		}
	})
}

func TestTemplateExplainer_ExplainBayesian(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		trace := BayesianTrace{
			Query: "win",
			Factors: []BayesianFactor{
				{Outcome: "yes", Probability: 0.65},
				{Outcome: "no", Probability: 0.35},
			},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainBayesian(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "win") {
			t.Fatalf("expected 'win' in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := BayesianTrace{
			Query: "win",
			Factors: []BayesianFactor{
				{Outcome: "yes", Probability: 0.65},
			},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainBayesian(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Posterior para") {
			t.Fatalf("expected 'Posterior para' in explanation, got: %s", explanation)
		}
	})
}

func TestTemplateExplainer_ExplainFuzzy(t *testing.T) {
	t.Parallel()

	t.Run("english with memberships", func(t *testing.T) {
		t.Parallel()

		trace := FuzzyTrace{
			Outputs: []FuzzyOutput{
				{Variable: "speed", Value: 75.5},
			},
			Memberships: []FuzzyMembership{
				{Variable: "temp", Term: "hot", Degree: 0.8},
			},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainFuzzy(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "speed") {
			t.Fatalf("expected 'speed' in explanation, got: %s", explanation)
		}

		if !strings.Contains(explanation, "temp/hot") {
			t.Fatalf("expected 'temp/hot' in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := FuzzyTrace{
			Outputs: []FuzzyOutput{
				{Variable: "speed", Value: 50},
			},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainFuzzy(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Salidas fuzzy") {
			t.Fatalf("expected 'Salidas fuzzy' in explanation, got: %s", explanation)
		}
	})
}

func TestTemplateExplainer_ExplainTable(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		trace := TableTrace{
			HitPolicy: "first",
			MatchedRules: []TableMatchEntry{
				{RuleName: "approve", Outputs: map[string]any{"result": "approved"}},
			},
			Outputs: map[string]any{"result": "approved"},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainTable(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "first") {
			t.Fatalf("expected 'first' in explanation, got: %s", explanation)
		}

		if !strings.Contains(explanation, "approve") {
			t.Fatalf("expected 'approve' in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := TableTrace{
			HitPolicy:    "unique",
			MatchedRules: []TableMatchEntry{{RuleName: "r1", Outputs: map[string]any{"x": 1}}},
			Outputs:      map[string]any{"x": 1},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainTable(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Tabla") {
			t.Fatalf("expected 'Tabla' in explanation, got: %s", explanation)
		}
	})

	t.Run("no matches", func(t *testing.T) {
		t.Parallel()

		trace := TableTrace{
			HitPolicy:    "first",
			MatchedRules: nil,
			Outputs:      nil,
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainTable(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "0 rule") {
			t.Fatalf("expected '0 rule' in explanation, got: %s", explanation)
		}
	})
}

func TestTemplateExplainer_ExplainScorecard(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		trace := ScoreTrace{
			BaseScore:  100,
			TotalScore: 175,
			Breakdown: []ScoreEntry{
				{Attribute: "income", Points: 50, Weight: 1.5, Weighted: 75},
			},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainScorecard(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "175.00") {
			t.Fatalf("expected '175.00' in explanation, got: %s", explanation)
		}

		if !strings.Contains(explanation, "income") {
			t.Fatalf("expected 'income' in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := ScoreTrace{
			BaseScore:  0,
			TotalScore: 50,
			Breakdown:  []ScoreEntry{{Attribute: "age", Points: 50, Weight: 1, Weighted: 50}},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainScorecard(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Puntaje") {
			t.Fatalf("expected 'Puntaje' in explanation, got: %s", explanation)
		}
	})

	t.Run("empty breakdown", func(t *testing.T) {
		t.Parallel()

		trace := ScoreTrace{
			BaseScore:  100,
			TotalScore: 100,
			Breakdown:  nil,
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainScorecard(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "100.00") {
			t.Fatalf("expected '100.00' in explanation, got: %s", explanation)
		}
	})
}

func TestTemplateExplainer_ExplainTree(t *testing.T) {
	t.Parallel()

	t.Run("english", func(t *testing.T) {
		t.Parallel()

		trace := TreeTrace{
			Path: []TreeStep{
				{Condition: "income > 50000", Result: true},
				{Condition: "credit > 700", Result: false},
			},
			Outputs: map[string]any{"risk": "medium"},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainTree(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Tree decision") {
			t.Fatalf("expected 'Tree decision' in explanation, got: %s", explanation)
		}

		if !strings.Contains(explanation, "income > 50000") {
			t.Fatalf("expected condition in explanation, got: %s", explanation)
		}
	})

	t.Run("spanish", func(t *testing.T) {
		t.Parallel()

		trace := TreeTrace{
			Path:    []TreeStep{{Condition: "x > 10", Result: true}},
			Outputs: map[string]any{"y": 1},
		}

		e := NewTemplateExplainer(Spanish)
		explanation, err := e.ExplainTree(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Decision de arbol") {
			t.Fatalf("expected 'Decision de arbol' in explanation, got: %s", explanation)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()

		trace := TreeTrace{
			Path:    nil,
			Outputs: map[string]any{"result": "default"},
		}

		e := NewTemplateExplainer(English)
		explanation, err := e.ExplainTree(context.Background(), trace)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(explanation, "Tree decision") {
			t.Fatalf("expected 'Tree decision' in explanation, got: %s", explanation)
		}
	})
}

func TestRender(t *testing.T) {
	t.Parallel()

	t.Run("execute error", func(t *testing.T) {
		t.Parallel()

		tmpl := template.Must(template.New("bad").Parse(`{{.MissingMethod}}`))

		_, err := render(tmpl, "a string has no MissingMethod")

		if err == nil {
			t.Fatal("expected template execute error")
		}

		if !errors.Is(err, ErrRenderFailed) {
			t.Fatalf("expected ErrRenderFailed, got: %v", err)
		}
	})
}

func TestErrRender(t *testing.T) {
	t.Parallel()

	t.Run("wraps sentinel", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("test cause")
		err := ErrRender(cause)

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrRenderFailed) {
			t.Fatal("expected error to wrap ErrRenderFailed")
		}

		if !errors.Is(err, cause) {
			t.Fatal("expected error to wrap cause")
		}

		var typed *Error
		ok := errors.As(err, &typed)

		if !ok {
			t.Fatal("expected error to be *Error")
		}

		if typed.Type != ExplainType {
			t.Fatalf("expected type %s, got %s", ExplainType, typed.Type)
		}
	})
}
