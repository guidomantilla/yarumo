package examples

import (
	"math"
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/engine"
	"github.com/guidomantilla/yarumo/inference/fuzzy/explain"
	"github.com/guidomantilla/yarumo/inference/fuzzy/rules"
	"github.com/guidomantilla/yarumo/inference/fuzzy/variable"
)

// Classic tipping problem: food quality + service → tip percentage.

func makeFoodInput() variable.Variable {
	return variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "average", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})
}

func makeServiceInput() variable.Variable {
	return variable.NewVariable("service", 0, 10, []variable.Term{
		{Name: "poor", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "acceptable", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "excellent", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})
}

func makeTipOutput() variable.Variable {
	return variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 5, 10)},
		{Name: "medium", Fn: fuzzym.Triangular(10, 15, 20)},
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})
}

func makeTippingRules() []rules.Rule {
	return []rules.Rule{
		rules.NewRule("bad-or-poor",
			[]rules.Condition{
				{Variable: "food", Term: "bad"},
				{Variable: "service", Term: "poor"},
			},
			rules.Consequent{Variable: "tip", Term: "low"},
			rules.WithOperator(rules.Or),
		),
		rules.NewRule("average-service",
			[]rules.Condition{
				{Variable: "service", Term: "acceptable"},
			},
			rules.Consequent{Variable: "tip", Term: "medium"},
		),
		rules.NewRule("good-and-excellent",
			[]rules.Condition{
				{Variable: "food", Term: "good"},
				{Variable: "service", Term: "excellent"},
			},
			rules.Consequent{Variable: "tip", Term: "high"},
			rules.WithOperator(rules.And),
		),
	}
}

func TestMamdaniTipping(t *testing.T) {
	t.Parallel()

	t.Run("bad food and poor service gives low tip", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 1.0, "service": 1.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip > 15 {
			t.Fatalf("expected low tip for bad food+poor service, got %f", tip)
		}
	})

	t.Run("good food and excellent service gives high tip", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 9.0, "service": 9.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip < 15 {
			t.Fatalf("expected high tip for good food+excellent service, got %f", tip)
		}
	})

	t.Run("average input gives medium tip", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip < 5 || tip > 25 {
			t.Fatalf("expected moderate tip, got %f", tip)
		}
	})
}

func TestSugenoTipping(t *testing.T) {
	t.Parallel()

	t.Run("sugeno produces numeric output", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
			engine.WithMethod(engine.Sugeno),
			engine.WithSugenoOutputs(map[string]float64{
				"tip/low":    5.0,
				"tip/medium": 15.0,
				"tip/high":   25.0,
			}),
		)

		result := eng.Infer(map[string]float64{"food": 9.0, "service": 9.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip < 15 {
			t.Fatalf("expected high sugeno tip, got %f", tip)
		}
	})

	t.Run("sugeno low inputs give low tip", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
			engine.WithMethod(engine.Sugeno),
			engine.WithSugenoOutputs(map[string]float64{
				"tip/low":    5.0,
				"tip/medium": 15.0,
				"tip/high":   25.0,
			}),
		)

		result := eng.Infer(map[string]float64{"food": 1.0, "service": 1.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip > 15 {
			t.Fatalf("expected low sugeno tip, got %f", tip)
		}
	})
}

func TestMamdaniVsSugeno(t *testing.T) {
	t.Parallel()

	t.Run("both methods agree on direction", func(t *testing.T) {
		t.Parallel()

		inputs := []variable.Variable{makeFoodInput(), makeServiceInput()}
		outputs := []variable.Variable{makeTipOutput()}
		ruleSet := makeTippingRules()

		mamdaniEng := engine.NewEngine(inputs, outputs, ruleSet)
		sugenoEng := engine.NewEngine(inputs, outputs, ruleSet,
			engine.WithMethod(engine.Sugeno),
			engine.WithSugenoOutputs(map[string]float64{
				"tip/low":    5.0,
				"tip/medium": 15.0,
				"tip/high":   25.0,
			}),
		)

		lowInput := map[string]float64{"food": 1.0, "service": 1.0}
		highInput := map[string]float64{"food": 9.0, "service": 9.0}

		mamdaniLow := mamdaniEng.Infer(lowInput).Outputs["tip"]
		mamdaniHigh := mamdaniEng.Infer(highInput).Outputs["tip"]
		sugenoLow := sugenoEng.Infer(lowInput).Outputs["tip"]
		sugenoHigh := sugenoEng.Infer(highInput).Outputs["tip"]

		if mamdaniLow >= mamdaniHigh {
			t.Fatalf("expected mamdani low < high, got %f >= %f", mamdaniLow, mamdaniHigh)
		}

		if sugenoLow >= sugenoHigh {
			t.Fatalf("expected sugeno low < high, got %f >= %f", sugenoLow, sugenoHigh)
		}
	})
}

func TestCustomTNorm(t *testing.T) {
	t.Parallel()

	t.Run("product t-norm produces different results", func(t *testing.T) {
		t.Parallel()

		inputs := []variable.Variable{makeFoodInput(), makeServiceInput()}
		outputs := []variable.Variable{makeTipOutput()}
		ruleSet := makeTippingRules()

		minEng := engine.NewEngine(inputs, outputs, ruleSet)
		productEng := engine.NewEngine(inputs, outputs, ruleSet,
			engine.WithTNorm(fuzzym.Product),
		)

		input := map[string]float64{"food": 7.0, "service": 7.0}
		minTip := minEng.Infer(input).Outputs["tip"]
		productTip := productEng.Infer(input).Outputs["tip"]

		if math.Abs(minTip-productTip) < 0.001 {
			t.Fatalf("expected different results with product t-norm, both got %f", minTip)
		}
	})
}

func TestRuleWeights(t *testing.T) {
	t.Parallel()

	t.Run("lower weight reduces rule influence", func(t *testing.T) {
		t.Parallel()

		inputs := []variable.Variable{makeFoodInput(), makeServiceInput()}
		outputs := []variable.Variable{makeTipOutput()}

		fullWeightRules := []rules.Rule{
			rules.NewRule("good-excellent",
				[]rules.Condition{
					{Variable: "food", Term: "good"},
					{Variable: "service", Term: "excellent"},
				},
				rules.Consequent{Variable: "tip", Term: "high"},
			),
		}

		halfWeightRules := []rules.Rule{
			rules.NewRule("good-excellent",
				[]rules.Condition{
					{Variable: "food", Term: "good"},
					{Variable: "service", Term: "excellent"},
				},
				rules.Consequent{Variable: "tip", Term: "high"},
				rules.WithWeight(0.5),
			),
		}

		input := map[string]float64{"food": 9.0, "service": 9.0}

		fullEng := engine.NewEngine(inputs, outputs, fullWeightRules)
		halfEng := engine.NewEngine(inputs, outputs, halfWeightRules)

		fullTip := fullEng.Infer(input).Outputs["tip"]
		halfTip := halfEng.Infer(input).Outputs["tip"]

		if halfTip >= fullTip {
			t.Fatalf("expected half weight to produce lower tip, got %f >= %f", halfTip, fullTip)
		}
	})
}

func TestTracePhases(t *testing.T) {
	t.Parallel()

	t.Run("trace contains all phases", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		phases := make(map[explain.Phase]bool)

		for _, step := range result.Trace.Steps {
			phases[step.Phase] = true
		}

		if !phases[explain.Fuzzification] {
			t.Fatal("expected fuzzification phase in trace")
		}

		if !phases[explain.RuleEvaluation] {
			t.Fatal("expected rule evaluation phase in trace")
		}

		if !phases[explain.Defuzzification] {
			t.Fatal("expected defuzzification phase in trace")
		}
	})
}

func TestTraceString(t *testing.T) {
	t.Parallel()

	t.Run("trace string is non-empty", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		traceStr := result.Trace.String()
		if traceStr == "" {
			t.Fatal("expected non-empty trace string")
		}
	})
}

func TestTraceOutputs(t *testing.T) {
	t.Parallel()

	t.Run("trace outputs match result outputs", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		if len(result.Trace.Outputs) == 0 {
			t.Fatal("expected trace outputs")
		}

		traceOutput := result.Trace.Outputs[0]

		resultOutput := result.Outputs[traceOutput.Variable]
		if math.Abs(traceOutput.CrispValue-resultOutput) > 0.001 {
			t.Fatalf("trace output %f != result output %f", traceOutput.CrispValue, resultOutput)
		}
	})
}

func TestVariableOperations(t *testing.T) {
	t.Parallel()

	t.Run("fuzzify returns degrees for all terms", func(t *testing.T) {
		t.Parallel()

		food := makeFoodInput()
		degrees := food.Fuzzify(5.0)

		if len(degrees) != 3 {
			t.Fatalf("expected 3 term degrees, got %d", len(degrees))
		}

		if degrees["average"] <= 0 {
			t.Fatalf("expected positive membership for average at 5.0, got %f", float64(degrees["average"]))
		}
	})

	t.Run("term lookup", func(t *testing.T) {
		t.Parallel()

		food := makeFoodInput()

		term, ok := food.Term("good")
		if !ok {
			t.Fatal("expected term 'good' to exist")
		}

		if term.Name != "good" {
			t.Fatalf("expected name=good, got %s", term.Name)
		}
	})
}

func TestDefuzzificationMethods(t *testing.T) {
	t.Parallel()

	t.Run("bisector produces valid output", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
			engine.WithDefuzzify(fuzzym.Bisector),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip < 0 || tip > 30 {
			t.Fatalf("expected tip in [0,30], got %f", tip)
		}
	})

	t.Run("mean-of-max produces valid output", func(t *testing.T) {
		t.Parallel()

		eng := engine.NewEngine(
			[]variable.Variable{makeFoodInput(), makeServiceInput()},
			[]variable.Variable{makeTipOutput()},
			makeTippingRules(),
			engine.WithDefuzzify(fuzzym.MeanOfMax),
		)

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})

		tip, ok := result.Outputs["tip"]
		if !ok {
			t.Fatal("expected tip output")
		}

		if tip < 0 || tip > 30 {
			t.Fatalf("expected tip in [0,30], got %f", tip)
		}
	})
}
