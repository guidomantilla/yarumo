package evaluate

import (
	"context"
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestCascadePipeline_Execute(t *testing.T) {
	t.Parallel()

	deductiveRuleSet := &schema.RuleSet{
		Name: "compliance",
		Deductive: &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "has-invoicing",
					Condition:  "facturacion",
					Conclusion: map[string]bool{"cumple_facturacion": true},
				},
				{
					Name:       "no-invoicing",
					Condition:  "not facturacion",
					Conclusion: map[string]bool{"cumple_facturacion": false},
				},
			},
		},
	}

	bayesianRuleSet := &schema.RuleSet{
		Name: "risk",
		Bayesian: &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "compliance",
					Outcomes: []string{"compliant", "non_compliant"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"compliant": 0.7, "non_compliant": 0.3}},
					},
				},
				{
					Variable: "audit",
					Parents:  []string{"compliance"},
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{
							Given:         map[string]string{"compliance": "compliant"},
							Probabilities: map[string]float64{"yes": 0.1, "no": 0.9},
						},
						{
							Given:         map[string]string{"compliance": "non_compliant"},
							Probabilities: map[string]float64{"yes": 0.8, "no": 0.2},
						},
					},
				},
			},
		},
	}

	t.Run("deductive to bayesian cascade", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "compliance", Paradigm: Deductive, RuleSet: deductiveRuleSet},
			{Name: "risk", Paradigm: Bayesian, RuleSet: bayesianRuleSet, Query: "audit"},
		}

		converter := func(result Result) (any, error) {
			eb := evidence.NewEvidenceBase()

			cVal, ok := result.Outcome.Facts["cumple_facturacion"]
			if ok && cVal {
				eb.Observe(stats.Var("compliance"), stats.Outcome("compliant"))
			} else {
				eb.Observe(stats.Var("compliance"), stats.Outcome("non_compliant"))
			}

			return eb, nil
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{converter})

		initialFacts := logic.Fact{"facturacion": true}
		cascadeResult, err := pipeline.Execute(context.Background(), initialFacts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cascadeResult.Stages) != 2 {
			t.Fatalf("expected 2 stages, got %d", len(cascadeResult.Stages))
		}

		if cascadeResult.Final.Paradigm != Bayesian {
			t.Fatalf("expected bayesian final, got %s", cascadeResult.Final.Paradigm)
		}

		if len(cascadeResult.Final.Outcome.Distribution) == 0 {
			t.Fatal("expected non-empty distribution")
		}

		if cascadeResult.Explanation == "" {
			t.Fatal("expected non-empty cascade explanation")
		}
	})

	t.Run("converter error", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Deductive, RuleSet: deductiveRuleSet},
			{Name: "s2", Paradigm: Bayesian, RuleSet: bayesianRuleSet, Query: "audit"},
		}

		converter := func(_ Result) (any, error) {
			return nil, errors.New("convert failed")
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{converter})

		_, err := pipeline.Execute(context.Background(), logic.Fact{"facturacion": true})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrCascadeFailed) {
			t.Fatal("expected ErrCascadeFailed")
		}
	})

	t.Run("wrong input type", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Deductive, RuleSet: deductiveRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), "not a fact map")

		if err == nil {
			t.Fatal("expected error for wrong input type")
		}

		if !errors.Is(err, ErrTypeMismatch) {
			t.Fatal("expected ErrTypeMismatch")
		}
	})

	t.Run("bayesian wrong input type", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Bayesian, RuleSet: bayesianRuleSet, Query: "audit"},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), "not evidence")

		if err == nil {
			t.Fatal("expected error for wrong input type")
		}
	})

	t.Run("fuzzy wrong input type", func(t *testing.T) {
		t.Parallel()

		fuzzyRuleSet := &schema.RuleSet{
			Name: "fuzzy",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "triangular", Params: []float64{0, 0.5, 1}}}},
				},
				OutputVars: []schema.FuzzyVarDef{
					{Name: "y", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "b", Type: "triangular", Params: []float64{0, 0.5, 1}}}},
				},
				Rules: []schema.FuzzyRuleDef{
					{Name: "r1", Conditions: []schema.FuzzyConditionDef{{Variable: "x", Term: "a"}}, Consequent: schema.FuzzyConsequentDef{Variable: "y", Term: "b"}},
				},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Fuzzy, RuleSet: fuzzyRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), "not a float map")

		if err == nil {
			t.Fatal("expected error for wrong input type")
		}
	})

	t.Run("missing stage config", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Deductive, RuleSet: &schema.RuleSet{Name: "empty"}},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), logic.Fact{"a": true})

		if err == nil {
			t.Fatal("expected error for missing config")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})

	t.Run("unsupported paradigm in stage", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Paradigm(99), RuleSet: &schema.RuleSet{Name: "x"}},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), logic.Fact{})

		if err == nil {
			t.Fatal("expected error for unsupported paradigm")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})

	t.Run("missing bayesian config in stage", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Bayesian, RuleSet: &schema.RuleSet{Name: "empty"}, Query: "x"},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})
		eb := evidence.NewEvidenceBase()

		_, err := pipeline.Execute(context.Background(), eb)

		if err == nil {
			t.Fatal("expected error for missing bayesian config")
		}
	})

	t.Run("missing fuzzy config in stage", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Fuzzy, RuleSet: &schema.RuleSet{Name: "empty"}},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), map[string]float64{"x": 1})

		if err == nil {
			t.Fatal("expected error for missing fuzzy config")
		}
	})

	t.Run("fuzzy stage success", func(t *testing.T) {
		t.Parallel()

		fuzzyRuleSet := &schema.RuleSet{
			Name: "fuzzy",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{
						Name: "temp",
						Min:  0,
						Max:  100,
						Terms: []schema.FuzzyTermDef{
							{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
							{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
						},
					},
				},
				OutputVars: []schema.FuzzyVarDef{
					{
						Name: "speed",
						Min:  0,
						Max:  100,
						Terms: []schema.FuzzyTermDef{
							{Name: "slow", Type: "triangular", Params: []float64{0, 0, 50}},
							{Name: "fast", Type: "triangular", Params: []float64{50, 100, 100}},
						},
					},
				},
				Rules: []schema.FuzzyRuleDef{
					{
						Name:       "r1",
						Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "hot"}},
						Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"},
					},
					{
						Name:       "r2",
						Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "cold"}},
						Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "slow"},
					},
				},
			},
		}

		stages := []CascadeStage{
			{Name: "fuzz", Paradigm: Fuzzy, RuleSet: fuzzyRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		result, err := pipeline.Execute(context.Background(), map[string]float64{"temp": 75})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Final.Paradigm != Fuzzy {
			t.Fatalf("expected fuzzy, got %s", result.Final.Paradigm)
		}

		if len(result.Final.Outcome.Outputs) == 0 {
			t.Fatal("expected non-empty outputs")
		}
	})

	t.Run("deductive stage parse error", func(t *testing.T) {
		t.Parallel()

		badRuleSet := &schema.RuleSet{
			Name: "bad",
			Deductive: &schema.DeductiveConfig{
				Rules: []schema.DeductiveRuleDef{
					{Name: "bad", Condition: "(((", Conclusion: map[string]bool{"x": true}},
				},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Deductive, RuleSet: badRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), logic.Fact{"a": true})

		if err == nil {
			t.Fatal("expected error for parse failure")
		}
	})

	t.Run("bayesian stage build error", func(t *testing.T) {
		t.Parallel()

		badRuleSet := &schema.RuleSet{
			Name: "bad",
			Bayesian: &schema.BayesianConfig{
				Nodes: []schema.BayesianNodeDef{
					{
						Variable: "rain",
						Outcomes: []string{"yes", "no"},
						CPT:      []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.5, "no": 0.8}}},
					},
				},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Bayesian, RuleSet: badRuleSet, Query: "rain"},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})
		eb := evidence.NewEvidenceBase()

		_, err := pipeline.Execute(context.Background(), eb)

		if err == nil {
			t.Fatal("expected error for network build failure")
		}
	})

	t.Run("fuzzy stage build var error", func(t *testing.T) {
		t.Parallel()

		badRuleSet := &schema.RuleSet{
			Name: "bad",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "unknown", Params: []float64{0}}}},
				},
				OutputVars: []schema.FuzzyVarDef{},
				Rules:      []schema.FuzzyRuleDef{},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Fuzzy, RuleSet: badRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), map[string]float64{"x": 0.5})

		if err == nil {
			t.Fatal("expected error for fuzzy var build failure")
		}
	})

	t.Run("fuzzy stage output var error", func(t *testing.T) {
		t.Parallel()

		badRuleSet := &schema.RuleSet{
			Name: "bad",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "triangular", Params: []float64{0, 0.5, 1}}}},
				},
				OutputVars: []schema.FuzzyVarDef{
					{Name: "y", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "b", Type: "unknown", Params: []float64{0}}}},
				},
				Rules: []schema.FuzzyRuleDef{},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Fuzzy, RuleSet: badRuleSet},
		}

		pipeline := NewCascadePipeline(stages, []StageConverter{})

		_, err := pipeline.Execute(context.Background(), map[string]float64{"x": 0.5})

		if err == nil {
			t.Fatal("expected error for fuzzy output var build failure")
		}
	})

	t.Run("deductive stage explain error", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Deductive, RuleSet: deductiveRuleSet},
		}

		explainer := &failingExplainer{err: errors.New("explain failed")}
		pipeline := NewCascadePipeline(stages, []StageConverter{}, WithExplainer(explainer))

		_, err := pipeline.Execute(context.Background(), logic.Fact{"facturacion": true})

		if err == nil {
			t.Fatal("expected explain error")
		}
	})

	t.Run("bayesian stage explain error", func(t *testing.T) {
		t.Parallel()

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Bayesian, RuleSet: bayesianRuleSet, Query: "audit"},
		}

		explainer := &failingExplainer{err: errors.New("explain failed")}
		pipeline := NewCascadePipeline(stages, []StageConverter{}, WithExplainer(explainer))
		eb := evidence.NewEvidenceBase()
		eb.Observe(stats.Var("compliance"), stats.Outcome("compliant"))

		_, err := pipeline.Execute(context.Background(), eb)

		if err == nil {
			t.Fatal("expected explain error")
		}
	})

	t.Run("fuzzy stage explain error", func(t *testing.T) {
		t.Parallel()

		fuzzyRuleSet := &schema.RuleSet{
			Name: "fuzzy",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "temp", Min: 0, Max: 100, Terms: []schema.FuzzyTermDef{
						{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
						{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
					}},
				},
				OutputVars: []schema.FuzzyVarDef{
					{Name: "speed", Min: 0, Max: 100, Terms: []schema.FuzzyTermDef{
						{Name: "slow", Type: "triangular", Params: []float64{0, 0, 50}},
						{Name: "fast", Type: "triangular", Params: []float64{50, 100, 100}},
					}},
				},
				Rules: []schema.FuzzyRuleDef{
					{Name: "r1", Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "hot"}}, Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"}},
				},
			},
		}

		stages := []CascadeStage{
			{Name: "s1", Paradigm: Fuzzy, RuleSet: fuzzyRuleSet},
		}

		explainer := &failingExplainer{err: errors.New("explain failed")}
		pipeline := NewCascadePipeline(stages, []StageConverter{}, WithExplainer(explainer))

		_, err := pipeline.Execute(context.Background(), map[string]float64{"temp": 75})

		if err == nil {
			t.Fatal("expected explain error")
		}
	})
}
