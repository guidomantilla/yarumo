package engine

import (
	"math"
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/rules"
	"github.com/guidomantilla/yarumo/inference/fuzzy/variable"
)

func makeTemperatureInput() variable.Variable {
	return variable.NewVariable("temperature", 0, 100, []variable.Term{
		{Name: "cold", Fn: fuzzym.Triangular(0, 0, 50)},
		{Name: "warm", Fn: fuzzym.Triangular(20, 50, 80)},
		{Name: "hot", Fn: fuzzym.Triangular(50, 100, 100)},
	})
}

func makeSpeedOutput() variable.Variable {
	return variable.NewVariable("speed", 0, 100, []variable.Term{
		{Name: "slow", Fn: fuzzym.Triangular(0, 0, 50)},
		{Name: "medium", Fn: fuzzym.Triangular(20, 50, 80)},
		{Name: "fast", Fn: fuzzym.Triangular(50, 100, 100)},
	})
}

func makeBasicRules() []rules.Rule {
	return []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "cold"}},
			rules.Consequent{Variable: "speed", Term: "slow"},
		),
		rules.NewRule("r2",
			[]rules.Condition{{Variable: "temperature", Term: "warm"}},
			rules.Consequent{Variable: "speed", Term: "medium"},
		),
		rules.NewRule("r3",
			[]rules.Condition{{Variable: "temperature", Term: "hot"}},
			rules.Consequent{Variable: "speed", Term: "fast"},
		),
	}
}

func TestNewEngine(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
	)

	if eng == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestEngine_Infer_mamdani_cold(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
	)

	result := eng.Infer(map[string]float64{"temperature": 10.0})

	speed, ok := result.Outputs["speed"]
	if !ok {
		t.Fatal("expected speed output")
	}

	// Cold input should produce low speed.
	if speed > 50 {
		t.Fatalf("expected low speed for cold temp, got %f", speed)
	}
}

func TestEngine_Infer_mamdani_hot(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	speed, ok := result.Outputs["speed"]
	if !ok {
		t.Fatal("expected speed output")
	}

	// Hot input should produce high speed.
	if speed < 50 {
		t.Fatalf("expected high speed for hot temp, got %f", speed)
	}
}

func TestEngine_Infer_mamdani_trace(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
	)

	result := eng.Infer(map[string]float64{"temperature": 50.0})

	if len(result.Trace.Steps) == 0 {
		t.Fatal("expected non-empty trace")
	}

	if len(result.Trace.Outputs) == 0 {
		t.Fatal("expected trace outputs")
	}

	if result.Trace.Inputs["temperature"] != 50.0 {
		t.Fatalf("expected temperature=50.0 in trace, got %f", result.Trace.Inputs["temperature"])
	}
}

func TestEngine_Infer_sugeno(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
		WithMethod(Sugeno),
		WithSugenoOutputs(map[string]float64{
			"speed/slow":   10.0,
			"speed/medium": 50.0,
			"speed/fast":   90.0,
		}),
	)

	result := eng.Infer(map[string]float64{"temperature": 50.0})

	speed, ok := result.Outputs["speed"]
	if !ok {
		t.Fatal("expected speed output")
	}

	// At temp=50: warm=1.0 (peak), cold=0.0, hot=0.0.
	// Weighted average = (0*10 + 1*50 + 0*90) / (0+1+0) = 50.
	if math.Abs(speed-50.0) > 1.0 {
		t.Fatalf("expected speed ≈ 50.0, got %f", speed)
	}
}

func TestEngine_Infer_sugeno_mixedInputs(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
		WithMethod(Sugeno),
		WithSugenoOutputs(map[string]float64{
			"speed/slow":   10.0,
			"speed/medium": 50.0,
			"speed/fast":   90.0,
		}),
	)

	// At temp=75: warm=Triangular(20,50,80)(75)=0.166, hot=Triangular(50,100,100)(75)=0.5.
	result := eng.Infer(map[string]float64{"temperature": 75.0})

	speed, ok := result.Outputs["speed"]
	if !ok {
		t.Fatal("expected speed output")
	}

	// Should be weighted toward fast.
	if speed < 50 {
		t.Fatalf("expected speed > 50 for hot temp, got %f", speed)
	}
}

func TestEngine_Infer_missingInput(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
	)

	// Missing temperature input.
	result := eng.Infer(map[string]float64{"humidity": 0.5})

	// Should complete without panic; speed should be 0 (no rules fire).
	if result.Outputs == nil {
		t.Fatal("expected non-nil outputs")
	}
}

func TestEngine_Infer_mamdani_multipleConditions_and(t *testing.T) {
	t.Parallel()

	humidityInput := variable.NewVariable("humidity", 0, 100, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 0, 50)},
		{Name: "high", Fn: fuzzym.Triangular(50, 100, 100)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{
				{Variable: "temperature", Term: "hot"},
				{Variable: "humidity", Term: "high"},
			},
			rules.Consequent{Variable: "speed", Term: "fast"},
			rules.WithOperator(rules.And),
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput(), humidityInput},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0, "humidity": 90.0})

	if result.Outputs["speed"] <= 0 {
		t.Fatalf("expected non-zero speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_mamdani_multipleConditions_or(t *testing.T) {
	t.Parallel()

	humidityInput := variable.NewVariable("humidity", 0, 100, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 0, 50)},
		{Name: "high", Fn: fuzzym.Triangular(50, 100, 100)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{
				{Variable: "temperature", Term: "hot"},
				{Variable: "humidity", Term: "high"},
			},
			rules.Consequent{Variable: "speed", Term: "fast"},
			rules.WithOperator(rules.Or),
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput(), humidityInput},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0, "humidity": 10.0})

	// OR: max of hot(90)=0.8 and high(10)=0.0 → 0.8. Should produce non-zero speed.
	if result.Outputs["speed"] <= 0 {
		t.Fatalf("expected non-zero speed with OR, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_mamdani_ruleWeight(t *testing.T) {
	t.Parallel()

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "hot"}},
			rules.Consequent{Variable: "speed", Term: "fast"},
			rules.WithWeight(0.5),
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	// Weight=0.5 should reduce the output compared to weight=1.0.
	if result.Outputs["speed"] <= 0 {
		t.Fatalf("expected non-zero speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_sugeno_noMatchingSugenoOutput(t *testing.T) {
	t.Parallel()

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		makeBasicRules(),
		WithMethod(Sugeno),
		WithSugenoOutputs(map[string]float64{}), // No Sugeno outputs defined.
	)

	result := eng.Infer(map[string]float64{"temperature": 50.0})

	// No matching Sugeno outputs, so crisp value should be 0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_mamdani_unknownTerm(t *testing.T) {
	t.Parallel()

	// Rule references a term that doesn't exist in the output variable.
	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "hot"}},
			rules.Consequent{Variable: "speed", Term: "turbo"}, // "turbo" not defined
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	// Should complete without panic.
	if result.Outputs == nil {
		t.Fatal("expected non-nil outputs")
	}
}

func TestEngine_Infer_mamdani_zeroStrength(t *testing.T) {
	t.Parallel()

	// Rule with condition for "cold" but input is at 100 (cold=0).
	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "cold"}},
			rules.Consequent{Variable: "speed", Term: "slow"},
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 100.0})

	// cold(100)=0 so no rules fire → speed=0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed for zero-strength rule, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_mamdani_unknownConditionVariable(t *testing.T) {
	t.Parallel()

	// Rule references a variable not in inputs.
	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "pressure", Term: "high"}},
			rules.Consequent{Variable: "speed", Term: "fast"},
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	// pressure not fuzzified → degree=0 → speed=0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_mamdani_unknownConditionTerm(t *testing.T) {
	t.Parallel()

	// Rule references a term that doesn't exist in the input variable.
	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "freezing"}},
			rules.Consequent{Variable: "speed", Term: "slow"},
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 10.0})

	// "freezing" term not in temperature → degree=0 → speed=0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_evaluateRule_emptyConditions(t *testing.T) {
	t.Parallel()

	eng := &engine{
		options: NewOptions(),
	}

	fuzzified := map[string]map[string]fuzzym.Degree{
		"temperature": {"hot": 0.8},
	}

	// A mock rule interface that returns empty conditions.
	r := &emptyConditionsRule{}
	strength := eng.evaluateRule(r, fuzzified)

	if strength != 0 {
		t.Fatalf("expected 0 for empty conditions, got %f", float64(strength))
	}
}

func TestEngine_Infer_mamdani_ruleForDifferentOutput(t *testing.T) {
	t.Parallel()

	// Rule produces for "power" but we only have "speed" as output variable.
	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "hot"}},
			rules.Consequent{Variable: "power", Term: "high"},
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	// Rule produces for "power" not "speed", so speed output should be 0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed, got %f", result.Outputs["speed"])
	}
}

func TestEngine_Infer_sugeno_ruleForDifferentOutput(t *testing.T) {
	t.Parallel()

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "temperature", Term: "hot"}},
			rules.Consequent{Variable: "power", Term: "high"},
		),
	}

	eng := NewEngine(
		[]variable.Variable{makeTemperatureInput()},
		[]variable.Variable{makeSpeedOutput()},
		ruleSet,
		WithMethod(Sugeno),
		WithSugenoOutputs(map[string]float64{"power/high": 100.0}),
	)

	result := eng.Infer(map[string]float64{"temperature": 90.0})

	// Rule produces for "power" not "speed", so speed should be 0.
	if result.Outputs["speed"] != 0 {
		t.Fatalf("expected 0 speed, got %f", result.Outputs["speed"])
	}
}

func TestAlgorithm_constants(t *testing.T) {
	t.Parallel()

	if Mamdani != 0 {
		t.Fatalf("expected Mamdani=0, got %d", Mamdani)
	}

	if Sugeno != 1 {
		t.Fatalf("expected Sugeno=1, got %d", Sugeno)
	}
}

// emptyConditionsRule is a test helper implementing rules.Rule with empty conditions.
type emptyConditionsRule struct{}

func (r *emptyConditionsRule) Name() string { return "empty" }

func (r *emptyConditionsRule) Conditions() []rules.Condition { return nil }

func (r *emptyConditionsRule) Operator() rules.Operator { return rules.And }

func (r *emptyConditionsRule) Consequent() rules.Consequent {
	return rules.Consequent{Variable: "speed", Term: "fast"}
}

func (r *emptyConditionsRule) Weight() float64 { return 1.0 }
