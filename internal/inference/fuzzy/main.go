package main

import (
	"fmt"

	fuzzym "github.com/guidomantilla/yarumo/maths/fuzzy"

	"github.com/guidomantilla/yarumo/inference/fuzzy/engine"
	"github.com/guidomantilla/yarumo/inference/fuzzy/rules"
	"github.com/guidomantilla/yarumo/inference/fuzzy/variable"
)

func main() {
	tippingMamdani()
	tippingSugeno()
	compareMethods()
	customOperators()
	ruleWeights()
	traceInspection()
}

// tippingMamdani shows the classic tipping problem solved with Mamdani inference.
// Input: food quality + service quality → Output: tip percentage
func tippingMamdani() {
	fmt.Println("=== Tipping Problem (Mamdani) ===")

	// Define input variables with linguistic terms
	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "average", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	service := variable.NewVariable("service", 0, 10, []variable.Term{
		{Name: "poor", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "acceptable", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "excellent", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	// Define output variable
	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 5, 10)},
		{Name: "medium", Fn: fuzzym.Triangular(10, 15, 20)},
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	// Define fuzzy rules
	ruleSet := []rules.Rule{
		// IF food is bad OR service is poor THEN tip is low
		rules.NewRule("bad-or-poor",
			[]rules.Condition{
				{Variable: "food", Term: "bad"},
				{Variable: "service", Term: "poor"},
			},
			rules.Consequent{Variable: "tip", Term: "low"},
			rules.WithOperator(rules.Or),
		),
		// IF service is acceptable THEN tip is medium
		rules.NewRule("acceptable-service",
			[]rules.Condition{
				{Variable: "service", Term: "acceptable"},
			},
			rules.Consequent{Variable: "tip", Term: "medium"},
		),
		// IF food is good AND service is excellent THEN tip is high
		rules.NewRule("good-and-excellent",
			[]rules.Condition{
				{Variable: "food", Term: "good"},
				{Variable: "service", Term: "excellent"},
			},
			rules.Consequent{Variable: "tip", Term: "high"},
			rules.WithOperator(rules.And),
		),
	}

	// Create Mamdani engine (default method)
	eng := engine.NewEngine(
		[]variable.Variable{food, service},
		[]variable.Variable{tip},
		ruleSet,
	)

	// Test different scenarios
	scenarios := []struct {
		food    float64
		service float64
		label   string
	}{
		{1.0, 1.0, "Bad food, poor service"},
		{5.0, 5.0, "Average food, acceptable service"},
		{9.0, 9.0, "Good food, excellent service"},
		{3.0, 8.0, "Below-average food, great service"},
		{8.0, 2.0, "Great food, poor service"},
	}

	for _, s := range scenarios {
		result := eng.Infer(map[string]float64{"food": s.food, "service": s.service})
		fmt.Printf("  %-40s → tip = %.1f%%\n", s.label, result.Outputs["tip"])
	}
	fmt.Println()
}

// tippingSugeno shows the same problem solved with Sugeno inference.
// Sugeno uses singleton outputs (crisp values) instead of fuzzy sets.
func tippingSugeno() {
	fmt.Println("=== Tipping Problem (Sugeno) ===")

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "average", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	service := variable.NewVariable("service", 0, 10, []variable.Term{
		{Name: "poor", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "acceptable", Fn: fuzzym.Triangular(2, 5, 8)},
		{Name: "excellent", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	// Sugeno output variable still needs terms for rule matching,
	// but the actual output values are singletons defined separately
	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 5, 10)},
		{Name: "medium", Fn: fuzzym.Triangular(10, 15, 20)},
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("bad-or-poor",
			[]rules.Condition{
				{Variable: "food", Term: "bad"},
				{Variable: "service", Term: "poor"},
			},
			rules.Consequent{Variable: "tip", Term: "low"},
			rules.WithOperator(rules.Or),
		),
		rules.NewRule("acceptable-service",
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

	// Sugeno engine: output = weighted average of singleton values
	eng := engine.NewEngine(
		[]variable.Variable{food, service},
		[]variable.Variable{tip},
		ruleSet,
		engine.WithMethod(engine.Sugeno),
		engine.WithSugenoOutputs(map[string]float64{
			"tip/low":    5.0,  // low tip = 5%
			"tip/medium": 15.0, // medium tip = 15%
			"tip/high":   25.0, // high tip = 25%
		}),
	)

	result := eng.Infer(map[string]float64{"food": 9.0, "service": 9.0})
	fmt.Printf("  Good food + excellent service: tip = %.1f%%\n", result.Outputs["tip"])

	result = eng.Infer(map[string]float64{"food": 1.0, "service": 1.0})
	fmt.Printf("  Bad food + poor service:       tip = %.1f%%\n", result.Outputs["tip"])

	result = eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
	fmt.Printf("  Average input:                 tip = %.1f%%\n", result.Outputs["tip"])
	fmt.Println()
}

// compareMethods shows Mamdani vs Sugeno side by side.
func compareMethods() {
	fmt.Println("=== Mamdani vs Sugeno ===")

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 5, 10)},
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{{Variable: "food", Term: "bad"}},
			rules.Consequent{Variable: "tip", Term: "low"},
		),
		rules.NewRule("r2",
			[]rules.Condition{{Variable: "food", Term: "good"}},
			rules.Consequent{Variable: "tip", Term: "high"},
		),
	}

	mamdaniEng := engine.NewEngine(
		[]variable.Variable{food}, []variable.Variable{tip}, ruleSet,
	)
	sugenoEng := engine.NewEngine(
		[]variable.Variable{food}, []variable.Variable{tip}, ruleSet,
		engine.WithMethod(engine.Sugeno),
		engine.WithSugenoOutputs(map[string]float64{"tip/low": 5.0, "tip/high": 25.0}),
	)

	for _, val := range []float64{1.0, 3.0, 5.0, 7.0, 9.0} {
		input := map[string]float64{"food": val}
		m := mamdaniEng.Infer(input).Outputs["tip"]
		s := sugenoEng.Infer(input).Outputs["tip"]
		fmt.Printf("  food=%.0f → Mamdani=%.1f%%, Sugeno=%.1f%%\n", val, m, s)
	}

	fmt.Println()
	fmt.Println("  Mamdani: fuzzy set outputs → defuzzification (centroid)")
	fmt.Println("  Sugeno:  singleton outputs → weighted average (faster)")
	fmt.Println()
}

// customOperators shows how to change t-norms and defuzzification methods.
func customOperators() {
	fmt.Println("=== Custom Operators ===")

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	service := variable.NewVariable("service", 0, 10, []variable.Term{
		{Name: "excellent", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("r1",
			[]rules.Condition{
				{Variable: "food", Term: "good"},
				{Variable: "service", Term: "excellent"},
			},
			rules.Consequent{Variable: "tip", Term: "high"},
		),
	}

	input := map[string]float64{"food": 7.0, "service": 7.0}

	// Default: Min t-norm + Centroid defuzzification
	defaultEng := engine.NewEngine(
		[]variable.Variable{food, service}, []variable.Variable{tip}, ruleSet,
	)

	// Product t-norm: softer AND (multiplies degrees instead of taking minimum)
	productEng := engine.NewEngine(
		[]variable.Variable{food, service}, []variable.Variable{tip}, ruleSet,
		engine.WithTNorm(fuzzym.Product),
	)

	// Bisector defuzzification: divides area in half instead of centroid
	bisectorEng := engine.NewEngine(
		[]variable.Variable{food, service}, []variable.Variable{tip}, ruleSet,
		engine.WithDefuzzify(fuzzym.Bisector),
	)

	fmt.Printf("  Min + Centroid:     %.1f%%\n", defaultEng.Infer(input).Outputs["tip"])
	fmt.Printf("  Product + Centroid: %.1f%%\n", productEng.Infer(input).Outputs["tip"])
	fmt.Printf("  Min + Bisector:     %.1f%%\n", bisectorEng.Infer(input).Outputs["tip"])
	fmt.Println()
}

// ruleWeights shows how weights modulate rule influence.
func ruleWeights() {
	fmt.Println("=== Rule Weights ===")

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	input := map[string]float64{"food": 9.0}

	// Full weight (default = 1.0)
	fullRule := rules.NewRule("full",
		[]rules.Condition{{Variable: "food", Term: "good"}},
		rules.Consequent{Variable: "tip", Term: "high"},
	)

	// Half weight (0.5) — expert says this rule is less certain
	halfRule := rules.NewRule("half",
		[]rules.Condition{{Variable: "food", Term: "good"}},
		rules.Consequent{Variable: "tip", Term: "high"},
		rules.WithWeight(0.5),
	)

	fullEng := engine.NewEngine([]variable.Variable{food}, []variable.Variable{tip}, []rules.Rule{fullRule})
	halfEng := engine.NewEngine([]variable.Variable{food}, []variable.Variable{tip}, []rules.Rule{halfRule})

	fmt.Printf("  Weight 1.0: tip = %.1f%%\n", fullEng.Infer(input).Outputs["tip"])
	fmt.Printf("  Weight 0.5: tip = %.1f%%\n", halfEng.Infer(input).Outputs["tip"])
	fmt.Println("  => Lower weight reduces the rule's influence on the output")
	fmt.Println()
}

// traceInspection shows how to inspect the inference trace step by step.
func traceInspection() {
	fmt.Println("=== Trace Inspection ===")

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: fuzzym.Trapezoidal(0, 0, 2, 4)},
		{Name: "good", Fn: fuzzym.Trapezoidal(6, 8, 10, 10)},
	})

	tip := variable.NewVariable("tip", 0, 30, []variable.Term{
		{Name: "low", Fn: fuzzym.Triangular(0, 5, 10)},
		{Name: "high", Fn: fuzzym.Triangular(20, 25, 30)},
	})

	ruleSet := []rules.Rule{
		rules.NewRule("bad-food",
			[]rules.Condition{{Variable: "food", Term: "bad"}},
			rules.Consequent{Variable: "tip", Term: "low"},
		),
		rules.NewRule("good-food",
			[]rules.Condition{{Variable: "food", Term: "good"}},
			rules.Consequent{Variable: "tip", Term: "high"},
		),
	}

	eng := engine.NewEngine(
		[]variable.Variable{food}, []variable.Variable{tip}, ruleSet,
	)

	result := eng.Infer(map[string]float64{"food": 7.0})

	// The trace shows every step: fuzzification, rule evaluation, defuzzification
	fmt.Println(result.Trace.String())

	// Access individual parts
	fmt.Printf("Output: tip = %.1f%%\n", result.Outputs["tip"])
	fmt.Printf("Trace steps: %d\n", len(result.Trace.Steps))
	fmt.Printf("Trace outputs: %d\n", len(result.Trace.Outputs))
}
