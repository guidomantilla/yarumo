package evaluate

import (
	"context"
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestRunDeductive(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{Name: "r1", Condition: "a and b", Conclusion: map[string]bool{"c": true}},
			},
		}

		result, err := runDeductive(context.Background(), config, logic.Fact{"a": true, "b": true}, explain.NewTemplateExplainer(explain.English))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Deductive {
			t.Fatalf("expected deductive, got %s", result.Paradigm)
		}

		cVal, ok := result.Outcome.Facts["c"]
		if !ok || !cVal {
			t.Fatal("expected c=true in outcome")
		}

		if result.Explanation == "" {
			t.Fatal("expected non-empty explanation")
		}
	})

	t.Run("parse error", func(t *testing.T) {
		t.Parallel()

		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{Name: "bad", Condition: "(((", Conclusion: map[string]bool{"x": true}},
			},
		}

		_, err := runDeductive(context.Background(), config, logic.Fact{"a": true}, explain.NewTemplateExplainer(explain.English))
		if err == nil {
			t.Fatal("expected error for parse failure")
		}
	})

	t.Run("explain error", func(t *testing.T) {
		t.Parallel()

		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{Name: "r1", Condition: "a", Conclusion: map[string]bool{"b": true}},
			},
		}

		_, err := runDeductive(context.Background(), config, logic.Fact{"a": true}, &failingExplainer{err: errors.New("explain failed")})
		if err == nil {
			t.Fatal("expected explain error")
		}

		if !errors.Is(err, ErrExplainFailed) {
			t.Fatal("expected ErrExplainFailed")
		}
	})
}

func TestRunBayesian(t *testing.T) {
	t.Parallel()

	validConfig := &schema.BayesianConfig{
		Nodes: []schema.BayesianNodeDef{
			{
				Variable: "rain",
				Outcomes: []string{"yes", "no"},
				CPT:      []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}}},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		eb := evidence.NewEvidenceBase()

		result, err := runBayesian(context.Background(), validConfig, eb, "rain", explain.NewTemplateExplainer(explain.English))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Bayesian {
			t.Fatalf("expected bayesian, got %s", result.Paradigm)
		}

		if len(result.Outcome.Distribution) == 0 {
			t.Fatal("expected non-empty distribution")
		}
	})

	t.Run("build error", func(t *testing.T) {
		t.Parallel()

		badConfig := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT:      []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.5, "no": 0.8}}},
				},
			},
		}

		eb := evidence.NewEvidenceBase()

		_, err := runBayesian(context.Background(), badConfig, eb, "rain", explain.NewTemplateExplainer(explain.English))
		if err == nil {
			t.Fatal("expected error for build failure")
		}
	})

	t.Run("explain error", func(t *testing.T) {
		t.Parallel()

		eb := evidence.NewEvidenceBase()

		_, err := runBayesian(context.Background(), validConfig, eb, "rain", &failingExplainer{err: errors.New("explain failed")})
		if err == nil {
			t.Fatal("expected explain error")
		}

		if !errors.Is(err, ErrExplainFailed) {
			t.Fatal("expected ErrExplainFailed")
		}
	})
}

func TestRunFuzzy(t *testing.T) {
	t.Parallel()

	validConfig := &schema.FuzzyConfig{
		InputVars: []schema.FuzzyVarDef{
			{
				Name: "temp", Min: 0, Max: 100,
				Terms: []schema.FuzzyTermDef{
					{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
					{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
				},
			},
		},
		OutputVars: []schema.FuzzyVarDef{
			{
				Name: "speed", Min: 0, Max: 100,
				Terms: []schema.FuzzyTermDef{
					{Name: "slow", Type: "triangular", Params: []float64{0, 0, 50}},
					{Name: "fast", Type: "triangular", Params: []float64{50, 100, 100}},
				},
			},
		},
		Rules: []schema.FuzzyRuleDef{
			{Name: "r1", Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "hot"}}, Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"}},
			{Name: "r2", Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "cold"}}, Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "slow"}},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		result, err := runFuzzy(context.Background(), validConfig, map[string]float64{"temp": 75}, explain.NewTemplateExplainer(explain.English))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Fuzzy {
			t.Fatalf("expected fuzzy, got %s", result.Paradigm)
		}

		if len(result.Outcome.Outputs) == 0 {
			t.Fatal("expected non-empty outputs")
		}
	})

	t.Run("input var build error", func(t *testing.T) {
		t.Parallel()

		badConfig := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "unknown", Params: []float64{0}}}},
			},
			OutputVars: []schema.FuzzyVarDef{},
			Rules:      []schema.FuzzyRuleDef{},
		}

		_, err := runFuzzy(context.Background(), badConfig, map[string]float64{"x": 0.5}, explain.NewTemplateExplainer(explain.English))
		if err == nil {
			t.Fatal("expected error for input var build failure")
		}
	})

	t.Run("output var build error", func(t *testing.T) {
		t.Parallel()

		badConfig := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "triangular", Params: []float64{0, 0.5, 1}}}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "y", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "b", Type: "unknown", Params: []float64{0}}}},
			},
			Rules: []schema.FuzzyRuleDef{},
		}

		_, err := runFuzzy(context.Background(), badConfig, map[string]float64{"x": 0.5}, explain.NewTemplateExplainer(explain.English))
		if err == nil {
			t.Fatal("expected error for output var build failure")
		}
	})

	t.Run("explain error", func(t *testing.T) {
		t.Parallel()

		_, err := runFuzzy(context.Background(), validConfig, map[string]float64{"temp": 75}, &failingExplainer{err: errors.New("explain failed")})
		if err == nil {
			t.Fatal("expected explain error")
		}

		if !errors.Is(err, ErrExplainFailed) {
			t.Fatal("expected ErrExplainFailed")
		}
	})
}

func TestDispatchParadigm_TypeMismatch(t *testing.T) {
	t.Parallel()

	t.Run("wrong ruleset type", func(t *testing.T) {
		t.Parallel()

		_, err := dispatchParadigm(context.Background(), Deductive, "not a ruleset", nil, "", explainerSet{})

		if err == nil {
			t.Fatal("expected error for wrong ruleset type")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("deductive wrong input", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test", Deductive: &schema.DeductiveConfig{}}

		_, err := dispatchParadigm(context.Background(), Deductive, rs, "not facts", "", explainerSet{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("bayesian wrong input", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test", Bayesian: &schema.BayesianConfig{}}

		_, err := dispatchParadigm(context.Background(), Bayesian, rs, "not evidence", "", explainerSet{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("fuzzy wrong input", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test", Fuzzy: &schema.FuzzyConfig{}}

		_, err := dispatchParadigm(context.Background(), Fuzzy, rs, "not map", "", explainerSet{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("unsupported paradigm", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test"}

		_, err := dispatchParadigm(context.Background(), Paradigm(99), rs, nil, "", explainerSet{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})

	t.Run("model paradigm rejected", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test"}

		_, err := dispatchParadigm(context.Background(), Table, rs, nil, "", explainerSet{})

		if err == nil {
			t.Fatal("expected error for model paradigm in dispatchParadigm")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})
}

func TestDispatchModelParadigm(t *testing.T) {
	t.Parallel()

	t.Run("wrong ruleset type", func(t *testing.T) {
		t.Parallel()

		_, err := dispatchModelParadigm(context.Background(), Table, "not a ruleset", nil, &Options{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("unsupported model paradigm", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test"}

		_, err := dispatchModelParadigm(context.Background(), Paradigm(99), rs, nil, &Options{})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})

	t.Run("inference paradigm rejected", func(t *testing.T) {
		t.Parallel()

		rs := &schema.RuleSet{Name: "test"}

		_, err := dispatchModelParadigm(context.Background(), Deductive, rs, nil, &Options{})

		if err == nil {
			t.Fatal("expected error for inference paradigm in dispatchModelParadigm")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})
}
