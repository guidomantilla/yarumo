package examples

import (
	"fmt"
	"testing"

	fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"

	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
)

func makeBenchInputs() []variable.Variable {
	foodBad, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
	foodAvg, _ := fuzzym.Triangular(2, 5, 8)
	foodGood, _ := fuzzym.Trapezoidal(6, 8, 10, 10)

	svcPoor, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
	svcAccept, _ := fuzzym.Triangular(2, 5, 8)
	svcExcellent, _ := fuzzym.Trapezoidal(6, 8, 10, 10)

	return []variable.Variable{
		variable.NewVariable("food", 0, 10, []variable.Term{
			{Name: "bad", Fn: foodBad},
			{Name: "average", Fn: foodAvg},
			{Name: "good", Fn: foodGood},
		}),
		variable.NewVariable("service", 0, 10, []variable.Term{
			{Name: "poor", Fn: svcPoor},
			{Name: "acceptable", Fn: svcAccept},
			{Name: "excellent", Fn: svcExcellent},
		}),
	}
}

func makeBenchOutputs() []variable.Variable {
	low, _ := fuzzym.Triangular(0, 5, 10)
	medium, _ := fuzzym.Triangular(10, 15, 20)
	high, _ := fuzzym.Triangular(20, 25, 30)

	return []variable.Variable{
		variable.NewVariable("tip", 0, 30, []variable.Term{
			{Name: "low", Fn: low},
			{Name: "medium", Fn: medium},
			{Name: "high", Fn: high},
		}),
	}
}

func makeBenchRules(n int) []rules.Rule {
	foodTerms := []string{"bad", "average", "good"}
	serviceTerms := []string{"poor", "acceptable", "excellent"}
	tipTerms := []string{"low", "medium", "high"}

	ruleSet := make([]rules.Rule, 0, n)

	for i := range n {
		fi := i % len(foodTerms)
		si := (i / len(foodTerms)) % len(serviceTerms)
		ti := (fi + si) % len(tipTerms)
		name := fmt.Sprintf("r%d", i)

		r := rules.NewRule(name,
			[]rules.Condition{
				{Variable: "food", Term: foodTerms[fi]},
				{Variable: "service", Term: serviceTerms[si]},
			},
			rules.Consequent{Variable: "tip", Term: tipTerms[ti]},
		)
		ruleSet = append(ruleSet, r)
	}

	return ruleSet
}

func BenchmarkMamdani9Rules(b *testing.B) {
	eng := engine.NewEngine(
		makeBenchInputs(),
		makeBenchOutputs(),
		makeBenchRules(9),
	)
	input := map[string]float64{"food": 5.0, "service": 5.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Infer(input)
	}
}

func BenchmarkSugeno9Rules(b *testing.B) {
	eng := engine.NewEngine(
		makeBenchInputs(),
		makeBenchOutputs(),
		makeBenchRules(9),
		engine.WithMethod(engine.Sugeno),
		engine.WithSugenoOutputs(map[string]float64{
			"tip/low":    5.0,
			"tip/medium": 15.0,
			"tip/high":   25.0,
		}),
	)
	input := map[string]float64{"food": 5.0, "service": 5.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Infer(input)
	}
}

func BenchmarkMamdaniHighResolution(b *testing.B) {
	eng := engine.NewEngine(
		makeBenchInputs(),
		makeBenchOutputs(),
		makeBenchRules(9),
		engine.WithResolution(500),
	)
	input := map[string]float64{"food": 5.0, "service": 5.0}

	b.ResetTimer()

	for b.Loop() {
		eng.Infer(input)
	}
}
