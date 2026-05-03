package adapters

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestAdaptFuzzyVars(t *testing.T) {
	t.Parallel()

	t.Run("triangular", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "temp",
				Min:  0,
				Max:  100,
				Terms: []schema.FuzzyTermDef{
					{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
					{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
				},
			},
		}

		vars, err := AdaptFuzzyVars(defs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(vars) != 1 {
			t.Fatalf("expected 1 var, got %d", len(vars))
		}

		if vars[0].Name() != "temp" {
			t.Fatalf("expected temp, got %s", vars[0].Name())
		}
	})

	t.Run("trapezoidal", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "speed",
				Min:  0,
				Max:  200,
				Terms: []schema.FuzzyTermDef{
					{Name: "slow", Type: "trapezoidal", Params: []float64{0, 0, 30, 60}},
				},
			},
		}

		vars, err := AdaptFuzzyVars(defs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(vars) != 1 {
			t.Fatalf("expected 1 var, got %d", len(vars))
		}
	})

	t.Run("gaussian", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "pressure",
				Min:  0,
				Max:  100,
				Terms: []schema.FuzzyTermDef{
					{Name: "normal", Type: "gaussian", Params: []float64{50, 10}},
				},
			},
		}

		vars, err := AdaptFuzzyVars(defs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(vars) != 1 {
			t.Fatalf("expected 1 var, got %d", len(vars))
		}
	})

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "x",
				Min:  0,
				Max:  1,
				Terms: []schema.FuzzyTermDef{
					{Name: "a", Type: "sigmoid", Params: []float64{0.5, 10}},
				},
			},
		}

		_, err := AdaptFuzzyVars(defs)
		if err == nil {
			t.Fatal("expected error for unknown type")
		}

		if !errors.Is(err, ErrAdaptVariablesFailed) {
			t.Fatalf("expected ErrAdaptVariablesFailed, got: %v", err)
		}
	})

	t.Run("wrong param count triangular", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "x",
				Min:  0,
				Max:  1,
				Terms: []schema.FuzzyTermDef{
					{Name: "a", Type: "triangular", Params: []float64{0, 1}},
				},
			},
		}

		_, err := AdaptFuzzyVars(defs)
		if err == nil {
			t.Fatal("expected error for wrong param count")
		}

		if !errors.Is(err, ErrInvalidParamCount) {
			t.Fatalf("expected ErrInvalidParamCount, got: %v", err)
		}
	})

	t.Run("wrong param count trapezoidal", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "x",
				Min:  0,
				Max:  1,
				Terms: []schema.FuzzyTermDef{
					{Name: "a", Type: "trapezoidal", Params: []float64{0, 1, 2}},
				},
			},
		}

		_, err := AdaptFuzzyVars(defs)
		if err == nil {
			t.Fatal("expected error for wrong param count")
		}
	})

	t.Run("wrong param count gaussian", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyVarDef{
			{
				Name: "x",
				Min:  0,
				Max:  1,
				Terms: []schema.FuzzyTermDef{
					{Name: "a", Type: "gaussian", Params: []float64{0.5}},
				},
			},
		}

		_, err := AdaptFuzzyVars(defs)
		if err == nil {
			t.Fatal("expected error for wrong param count")
		}
	})
}

func TestAdaptFuzzyRules(t *testing.T) {
	t.Parallel()

	t.Run("valid rules", func(t *testing.T) {
		t.Parallel()

		defs := []schema.FuzzyRuleDef{
			{
				Name: "r1",
				Conditions: []schema.FuzzyConditionDef{
					{Variable: "temp", Term: "hot"},
				},
				Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"},
			},
			{
				Name: "r2",
				Conditions: []schema.FuzzyConditionDef{
					{Variable: "temp", Term: "cold"},
					{Variable: "pressure", Term: "high"},
				},
				Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "slow"},
				Operator:   "or",
				Weight:     0.8,
			},
		}

		rules := AdaptFuzzyRules(defs)

		if len(rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(rules))
		}

		if rules[0].Name() != "r1" {
			t.Fatalf("expected r1, got %s", rules[0].Name())
		}
	})
}

func TestAdaptFuzzyOpts(t *testing.T) {
	t.Parallel()

	t.Run("default mamdani", func(t *testing.T) {
		t.Parallel()

		opts := AdaptFuzzyOpts(&schema.FuzzyConfig{})
		if len(opts) != 0 {
			t.Fatalf("expected 0 opts, got %d", len(opts))
		}
	})

	t.Run("sugeno with outputs", func(t *testing.T) {
		t.Parallel()

		opts := AdaptFuzzyOpts(&schema.FuzzyConfig{
			Method:        "sugeno",
			SugenoOutputs: map[string]float64{"speed/fast": 80},
		})

		if len(opts) != 2 {
			t.Fatalf("expected 2 opts, got %d", len(opts))
		}
	})

	t.Run("sugeno without outputs", func(t *testing.T) {
		t.Parallel()

		opts := AdaptFuzzyOpts(&schema.FuzzyConfig{
			Method: "sugeno",
		})

		if len(opts) != 1 {
			t.Fatalf("expected 1 opt, got %d", len(opts))
		}
	})
}
