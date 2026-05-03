package validate

import (
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/logic/sat"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestValidator_ValidateDeductive(t *testing.T) {
	t.Parallel()

	t.Run("valid rules no issues", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a and b",
					Conclusion: map[string]bool{"c": true},
				},
				{
					Name:       "r2",
					Condition:  "c",
					Conclusion: map[string]bool{"d": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if report.Parsed != 2 {
			t.Fatalf("expected 2 parsed, got %d", report.Parsed)
		}

		if len(report.Contradictions) != 0 {
			t.Fatalf("expected 0 contradictions, got %d", len(report.Contradictions))
		}
	})

	t.Run("parse error", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "bad",
					Condition:  "(((",
					Conclusion: map[string]bool{"x": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if report.Valid {
			t.Fatal("expected invalid report")
		}

		if len(report.Errors) == 0 {
			t.Fatal("expected parse error in Errors")
		}
	})

	t.Run("contradiction detected", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "approve",
					Condition:  "eligible",
					Conclusion: map[string]bool{"result": true},
				},
				{
					Name:       "reject",
					Condition:  "eligible",
					Conclusion: map[string]bool{"result": false},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Contradictions) == 0 {
			t.Fatal("expected contradictions")
		}

		if report.Valid {
			t.Fatal("expected invalid due to contradictions")
		}
	})

	t.Run("coverage gaps detected", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a and b",
					Conclusion: map[string]bool{"c": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Gaps) == 0 {
			t.Fatal("expected coverage gaps")
		}
	})

	t.Run("simplification detected", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "double-neg",
					Condition:  "not not a",
					Conclusion: map[string]bool{"b": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Simplified) == 0 {
			t.Fatal("expected simplification suggestions")
		}
	})

	t.Run("empty rules", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{},
		}

		report := v.ValidateDeductive(config)

		if report.Parsed != 0 {
			t.Fatalf("expected 0 parsed, got %d", report.Parsed)
		}

		if !report.Valid {
			t.Fatal("expected valid for empty rules")
		}
	})

	t.Run("single rule no redundancy", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "only",
					Condition:  "a",
					Conclusion: map[string]bool{"b": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Redundant) != 0 {
			t.Fatalf("expected 0 redundant, got %d", len(report.Redundant))
		}
	})

	t.Run("redundancy detected", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a",
					Conclusion: map[string]bool{"b": true},
				},
				{
					Name:       "r2",
					Condition:  "a and c",
					Conclusion: map[string]bool{"b": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Redundant) == 0 {
			t.Fatal("expected redundancy detected")
		}
	})

	t.Run("rule with empty conclusion", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a",
					Conclusion: map[string]bool{},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if report.Parsed != 1 {
			t.Fatalf("expected 1 parsed, got %d", report.Parsed)
		}
	})

	t.Run("rule with false conclusion", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a",
					Conclusion: map[string]bool{"b": false},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if report.Parsed != 1 {
			t.Fatalf("expected 1 parsed, got %d", report.Parsed)
		}
	})

	t.Run("multi conclusion rule", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "multi",
					Condition:  "a",
					Conclusion: map[string]bool{"b": true, "c": false},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if report.Parsed != 1 {
			t.Fatalf("expected 1 parsed, got %d", report.Parsed)
		}
	})

	t.Run("no contradiction when same conclusion", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.DeductiveConfig{
			Rules: []schema.DeductiveRuleDef{
				{
					Name:       "r1",
					Condition:  "a",
					Conclusion: map[string]bool{"result": true},
				},
				{
					Name:       "r2",
					Condition:  "b",
					Conclusion: map[string]bool{"result": true},
				},
			},
		}

		report := v.ValidateDeductive(config)

		if len(report.Contradictions) != 0 {
			t.Fatalf("expected 0 contradictions, got %d", len(report.Contradictions))
		}
	})
}

const (
	unknownRef   = "nonexistent"
	sugenoMethod = "sugeno"
)

func TestValidator_ValidateFuzzy(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()

		report := v.ValidateFuzzy(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 5 {
			t.Fatalf("expected 5 parsed (2 vars + 1 output + 2 rules), got %d", report.Parsed)
		}
	})

	t.Run("empty variable name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 5}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid")
		}
	})

	t.Run("min greater than max", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 10, Max: 0, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 5}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to min >= max")
		}
	})

	t.Run("no terms", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to no terms")
		}
	})

	t.Run("unknown membership type", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "sigmoid", Params: []float64{5, 1}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to unknown membership type")
		}
	})

	t.Run("wrong param count", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 5}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to wrong param count")
		}
	})

	t.Run("rule references unknown variable", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Conditions[0].Variable = unknownRef

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to unknown variable reference")
		}
	})

	t.Run("rule references unknown term", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Conditions[0].Term = unknownRef

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to unknown term reference")
		}
	})

	t.Run("consequent references input variable", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Consequent.Variable = "service"

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to consequent referencing input variable")
		}
	})

	t.Run("consequent references unknown variable", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Consequent.Variable = unknownRef

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to unknown consequent variable")
		}
	})

	t.Run("consequent references unknown term", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Consequent.Term = unknownRef

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to unknown consequent term")
		}
	})

	t.Run("duplicate variable name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.InputVars = append(config.InputVars, config.InputVars[0])

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to duplicate variable name")
		}
	})

	t.Run("duplicate term name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 5}},
					{Name: "low", Type: "triangular", Params: []float64{5, 10, 10}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to duplicate term name")
		}
	})

	t.Run("sugeno skips empty output var name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Method = sugenoMethod
		config.SugenoOutputs = map[string]float64{"tip": 15.0}
		config.OutputVars = append(config.OutputVars, schema.FuzzyVarDef{
			Name: "", Min: 0, Max: 1,
			Terms: []schema.FuzzyTermDef{{Name: "x", Type: "triangular", Params: []float64{0, 0, 1}}},
		})

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to empty output var name")
		}
	})

	t.Run("sugeno missing output value", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Method = sugenoMethod
		config.SugenoOutputs = map[string]float64{}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to missing sugeno output")
		}
	})

	t.Run("sugeno with output values", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Method = sugenoMethod
		config.SugenoOutputs = map[string]float64{"tip": 15.0}

		report := v.ValidateFuzzy(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}
	})

	t.Run("rule with empty name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Name = ""

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to empty rule name")
		}
	})

	t.Run("rule with no conditions", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Conditions = nil

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to no conditions")
		}
	})

	t.Run("invalid weight", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Weight = 1.5

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to weight > 1")
		}
	})

	t.Run("empty term name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{
			InputVars: []schema.FuzzyVarDef{
				{Name: "x", Min: 0, Max: 10, Terms: []schema.FuzzyTermDef{
					{Name: "", Type: "triangular", Params: []float64{0, 0, 5}},
				}},
			},
			OutputVars: []schema.FuzzyVarDef{
				{Name: "out", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 0.5}},
				}},
			},
		}

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to empty term name")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.FuzzyConfig{}

		report := v.ValidateFuzzy(config)

		if report.Parsed != 0 {
			t.Fatalf("expected 0 parsed, got %d", report.Parsed)
		}

		if !report.Valid {
			t.Fatal("expected valid for empty config")
		}
	})

	t.Run("negative weight", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := validFuzzyConfig()
		config.Rules[0].Weight = -0.5

		report := v.ValidateFuzzy(config)

		if report.Valid {
			t.Fatal("expected invalid due to negative weight")
		}
	})
}

func validFuzzyConfig() *schema.FuzzyConfig {
	return &schema.FuzzyConfig{
		InputVars: []schema.FuzzyVarDef{
			{
				Name: "service",
				Min:  0, Max: 10,
				Terms: []schema.FuzzyTermDef{
					{Name: "poor", Type: "triangular", Params: []float64{0, 0, 5}},
					{Name: "good", Type: "triangular", Params: []float64{0, 5, 10}},
					{Name: "excellent", Type: "triangular", Params: []float64{5, 10, 10}},
				},
			},
			{
				Name: "food",
				Min:  0, Max: 10,
				Terms: []schema.FuzzyTermDef{
					{Name: "bad", Type: "trapezoidal", Params: []float64{0, 0, 2, 5}},
					{Name: "good", Type: "trapezoidal", Params: []float64{5, 8, 10, 10}},
				},
			},
		},
		OutputVars: []schema.FuzzyVarDef{
			{
				Name: "tip",
				Min:  0, Max: 30,
				Terms: []schema.FuzzyTermDef{
					{Name: "low", Type: "triangular", Params: []float64{0, 0, 15}},
					{Name: "medium", Type: "triangular", Params: []float64{0, 15, 30}},
					{Name: "high", Type: "triangular", Params: []float64{15, 30, 30}},
				},
			},
		},
		Rules: []schema.FuzzyRuleDef{
			{
				Name: "r1",
				Conditions: []schema.FuzzyConditionDef{
					{Variable: "service", Term: "poor"},
					{Variable: "food", Term: "bad"},
				},
				Consequent: schema.FuzzyConsequentDef{Variable: "tip", Term: "low"},
			},
			{
				Name: "r2",
				Conditions: []schema.FuzzyConditionDef{
					{Variable: "service", Term: "excellent"},
					{Variable: "food", Term: "good"},
				},
				Consequent: schema.FuzzyConsequentDef{Variable: "tip", Term: "high"},
			},
		},
	}
}

func TestValidator_ValidateBayesian(t *testing.T) {
	t.Parallel()

	t.Run("valid network", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
					},
				},
				{
					Variable: "wet",
					Parents:  []string{"rain"},
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{
							Given:         map[string]string{"rain": "yes"},
							Probabilities: map[string]float64{"yes": 0.9, "no": 0.1},
						},
						{
							Given:         map[string]string{"rain": "no"},
							Probabilities: map[string]float64{"yes": 0.2, "no": 0.8},
						},
					},
				},
			},
		}

		report := v.ValidateBayesian(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 2 {
			t.Fatalf("expected 2 parsed, got %d", report.Parsed)
		}
	})

	t.Run("invalid CPT", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"yes": 0.5, "no": 0.8}},
					},
				},
			},
		}

		report := v.ValidateBayesian(config)

		if report.Valid {
			t.Fatal("expected invalid due to CPT that doesn't sum to 1")
		}

		if len(report.Errors) == 0 {
			t.Fatal("expected network errors")
		}
	})
}

func TestValidator_ValidateTable(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			HitPolicy: "first",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
				{Name: "r2", Conditions: []string{"x < 0"}, Outputs: map[string]any{"val": 2}},
			},
		}

		report := v.ValidateTable(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 2 {
			t.Fatalf("expected 2 parsed, got %d", report.Parsed)
		}
	})

	t.Run("invalid hit policy", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			HitPolicy: "invalid",
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to hit policy")
		}
	})

	t.Run("empty rules", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to no rules")
		}
	})

	t.Run("empty rule name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{
				{Name: "", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to empty name")
		}
	})

	t.Run("no conditions", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: nil, Outputs: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to no conditions")
		}
	})

	t.Run("bad condition parse", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"((("}, Outputs: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to parse error")
		}
	})

	t.Run("no outputs", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{}},
			},
		}

		report := v.ValidateTable(config)

		if report.Valid {
			t.Fatal("expected invalid due to no outputs")
		}
	})

	t.Run("default hit policy valid", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TableConfig{
			Rules: []schema.TableRuleDef{
				{Name: "r1", Conditions: []string{"x > 0"}, Outputs: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTable(config)

		if !report.Valid {
			t.Fatalf("expected valid with default hit policy, errors: %v", report.Errors)
		}
	})
}

func TestValidator_ValidateScorecard(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			BaseScore: 100,
			Attributes: []schema.ScorecardAttributeDef{
				{
					Name:   "income",
					Weight: 1.5,
					Bins: []schema.ScorecardBinDef{
						{Condition: "income > 50000", Points: 50},
					},
				},
			},
		}

		report := v.ValidateScorecard(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 1 {
			t.Fatalf("expected 1 parsed, got %d", report.Parsed)
		}
	})

	t.Run("no attributes", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to no attributes")
		}
	})

	t.Run("empty attribute name", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{
				{Name: "", Weight: 1.0, Bins: []schema.ScorecardBinDef{{Condition: "x > 0", Points: 10}}},
			},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to empty name")
		}
	})

	t.Run("zero weight", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{
				{Name: "x", Weight: 0, Bins: []schema.ScorecardBinDef{{Condition: "x > 0", Points: 10}}},
			},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to zero weight")
		}
	})

	t.Run("negative weight", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{
				{Name: "x", Weight: -1, Bins: []schema.ScorecardBinDef{{Condition: "x > 0", Points: 10}}},
			},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to negative weight")
		}
	})

	t.Run("no bins", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{
				{Name: "x", Weight: 1.0, Bins: []schema.ScorecardBinDef{}},
			},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to no bins")
		}
	})

	t.Run("bad bin condition", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.ScorecardConfig{
			Attributes: []schema.ScorecardAttributeDef{
				{Name: "x", Weight: 1.0, Bins: []schema.ScorecardBinDef{{Condition: "(((", Points: 10}}},
			},
		}

		report := v.ValidateScorecard(config)

		if report.Valid {
			t.Fatal("expected invalid due to parse error")
		}
	})
}

func TestValidator_ValidateTree(t *testing.T) {
	t.Parallel()

	t.Run("valid tree", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "x > 0",
				True:      &schema.TreeNodeDef{Output: map[string]any{"val": "a"}},
				False:     &schema.TreeNodeDef{Output: map[string]any{"val": "b"}},
			},
		}

		report := v.ValidateTree(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 3 {
			t.Fatalf("expected 3 parsed (1 internal + 2 leaves), got %d", report.Parsed)
		}
	})

	t.Run("root is leaf", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Output: map[string]any{"default": true},
			},
		}

		report := v.ValidateTree(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 1 {
			t.Fatalf("expected 1 parsed, got %d", report.Parsed)
		}
	})

	t.Run("missing condition on internal node", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "",
				True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTree(config)

		if report.Valid {
			t.Fatal("expected invalid due to missing condition")
		}
	})

	t.Run("bad condition parse", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "(((",
				True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
				False:     &schema.TreeNodeDef{Output: map[string]any{"val": 2}},
			},
		}

		report := v.ValidateTree(config)

		if report.Valid {
			t.Fatal("expected invalid due to parse error")
		}
	})

	t.Run("missing true branch", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "x > 0",
				True:      nil,
				False:     &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
			},
		}

		report := v.ValidateTree(config)

		if report.Valid {
			t.Fatal("expected invalid due to missing true branch")
		}
	})

	t.Run("missing false branch", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "x > 0",
				True:      &schema.TreeNodeDef{Output: map[string]any{"val": 1}},
				False:     nil,
			},
		}

		report := v.ValidateTree(config)

		if report.Valid {
			t.Fatal("expected invalid due to missing false branch")
		}
	})

	t.Run("exceeds max depth", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())

		// Build a chain deeper than maxTreeDepth (100).
		node := &schema.TreeNodeDef{Output: map[string]any{"val": "leaf"}}
		for range 102 {
			node = &schema.TreeNodeDef{
				Condition: "x > 0",
				True:      node,
				False:     &schema.TreeNodeDef{Output: map[string]any{"val": "f"}},
			}
		}

		config := &schema.TreeConfig{Root: *node}

		report := v.ValidateTree(config)

		if report.Valid {
			t.Fatal("expected invalid due to max depth exceeded")
		}
	})

	t.Run("deep nested tree", func(t *testing.T) {
		t.Parallel()

		v := NewValidator(sat.Solver())
		config := &schema.TreeConfig{
			Root: schema.TreeNodeDef{
				Condition: "a > 0",
				True: &schema.TreeNodeDef{
					Condition: "b > 0",
					True:      &schema.TreeNodeDef{Output: map[string]any{"val": "both"}},
					False:     &schema.TreeNodeDef{Output: map[string]any{"val": "a_only"}},
				},
				False: &schema.TreeNodeDef{Output: map[string]any{"val": "none"}},
			},
		}

		report := v.ValidateTree(config)

		if !report.Valid {
			t.Fatalf("expected valid, errors: %v", report.Errors)
		}

		if report.Parsed != 5 {
			t.Fatalf("expected 5 parsed, got %d", report.Parsed)
		}
	})
}
