package examples

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/causal/engine"
	"github.com/guidomantilla/yarumo/compute/engine/causal/explain"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

// makeEconomicSCM builds: Education → Income → Spending.
// Education is exogenous; Income = 2*Education; Spending = 0.6*Income.
func makeEconomicSCM() model.SCM {
	scm := model.NewSCM()

	scm.AddVariable("Education", nil, func(parents map[string]float64) float64 {
		return parents["Education"]
	})

	scm.AddVariable("Income", []string{"Education"}, func(parents map[string]float64) float64 {
		return 2 * parents["Education"]
	})

	scm.AddVariable("Spending", []string{"Income"}, func(parents map[string]float64) float64 {
		return 0.6 * parents["Income"]
	})

	return scm
}

// makeHealthSCM builds: Exercise → Fitness → Health ← Diet.
// Fitness = 1.5*Exercise; Health = 0.5*Fitness + 0.5*Diet.
func makeHealthSCM() model.SCM {
	scm := model.NewSCM()

	scm.AddVariable("Exercise", nil, func(parents map[string]float64) float64 {
		return parents["Exercise"]
	})

	scm.AddVariable("Diet", nil, func(parents map[string]float64) float64 {
		return parents["Diet"]
	})

	scm.AddVariable("Fitness", []string{"Exercise"}, func(parents map[string]float64) float64 {
		return 1.5 * parents["Exercise"]
	})

	scm.AddVariable("Health", []string{"Fitness", "Diet"}, func(parents map[string]float64) float64 {
		return 0.5*parents["Fitness"] + 0.5*parents["Diet"]
	})

	return scm
}

func TestPropagation(t *testing.T) {
	t.Parallel()

	t.Run("linear chain propagation", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Propagate(scm, map[string]float64{"Education": 10})

		income := result.Values["Income"]
		if math.Abs(income-20) > 0.01 {
			t.Fatalf("expected Income=20, got %f", income)
		}

		spending := result.Values["Spending"]
		if math.Abs(spending-12) > 0.01 {
			t.Fatalf("expected Spending=12, got %f", spending)
		}
	})

	t.Run("multiple roots propagation", func(t *testing.T) {
		t.Parallel()

		scm := makeHealthSCM()
		eng := engine.NewEngine()
		result := eng.Propagate(scm, map[string]float64{"Exercise": 8, "Diet": 6})

		fitness := result.Values["Fitness"]
		if math.Abs(fitness-12) > 0.01 {
			t.Fatalf("expected Fitness=12, got %f", fitness)
		}

		health := result.Values["Health"]

		expected := 0.5*12 + 0.5*6
		if math.Abs(health-expected) > 0.01 {
			t.Fatalf("expected Health=%f, got %f", expected, health)
		}
	})
}

func TestIntervention(t *testing.T) {
	t.Parallel()

	t.Run("do-operator changes downstream", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()

		observed := eng.Propagate(scm, map[string]float64{"Education": 10})
		intervened := eng.Intervene(scm, map[string]float64{"Income": 50})

		observedSpending := observed.Values["Spending"]
		intervenedSpending := intervened.Values["Spending"]

		if math.Abs(intervenedSpending-30) > 0.01 {
			t.Fatalf("expected Spending=30 after do(Income=50), got %f", intervenedSpending)
		}

		if math.Abs(observedSpending-intervenedSpending) < 0.01 {
			t.Fatal("expected intervention to change spending")
		}
	})

	t.Run("intervention bypasses structural equation", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()

		// Observing Education=10 gives Income=20, Spending=12.
		observed := eng.Propagate(scm, map[string]float64{"Education": 10})

		// Intervening on Income=50 sets it directly, ignoring the Education→Income equation.
		intervened := eng.Intervene(scm, map[string]float64{"Income": 50})

		observedIncome := observed.Values["Income"]
		intervenedIncome := intervened.Values["Income"]

		if math.Abs(intervenedIncome-50) > 0.01 {
			t.Fatalf("expected Income=50 after do(Income=50), got %f", intervenedIncome)
		}

		if math.Abs(observedIncome-intervenedIncome) < 0.01 {
			t.Fatal("expected intervention to override observed income")
		}

		spending := intervened.Values["Spending"]
		if math.Abs(spending-30) > 0.01 {
			t.Fatalf("expected Spending=30, got %f", spending)
		}
	})
}

func TestCounterfactual(t *testing.T) {
	t.Parallel()

	t.Run("what if education were different", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()

		factual := map[string]float64{"Education": 10}
		hypothetical := map[string]float64{"Education": 15}

		result := eng.Counterfactual(scm, factual, hypothetical)

		income := result.Values["Income"]
		if math.Abs(income-30) > 0.01 {
			t.Fatalf("expected Income=30 under do(Education=15), got %f", income)
		}

		spending := result.Values["Spending"]
		if math.Abs(spending-18) > 0.01 {
			t.Fatalf("expected Spending=18, got %f", spending)
		}
	})

	t.Run("counterfactual differs from factual", func(t *testing.T) {
		t.Parallel()

		scm := makeHealthSCM()
		eng := engine.NewEngine()

		factual := map[string]float64{"Exercise": 5, "Diet": 7}
		hypothetical := map[string]float64{"Exercise": 10}

		factualResult := eng.Propagate(scm, factual)
		cfResult := eng.Counterfactual(scm, factual, hypothetical)

		factualHealth := factualResult.Values["Health"]
		cfHealth := cfResult.Values["Health"]

		if cfHealth <= factualHealth {
			t.Fatalf("expected counterfactual health > factual, got %f <= %f", cfHealth, factualHealth)
		}
	})
}

func TestTraceInspection(t *testing.T) {
	t.Parallel()

	t.Run("propagation trace has steps", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Propagate(scm, map[string]float64{"Education": 10})

		if len(result.Trace.Steps) == 0 {
			t.Fatal("expected non-empty trace steps")
		}

		hasPropagation := false
		hasComplete := false

		for _, step := range result.Trace.Steps {
			if step.Phase == explain.Propagation {
				hasPropagation = true
			}

			if step.Phase == explain.Complete {
				hasComplete = true
			}
		}

		if !hasPropagation {
			t.Fatal("expected propagation phase in trace")
		}

		if !hasComplete {
			t.Fatal("expected complete phase in trace")
		}
	})

	t.Run("intervention trace has intervention phase", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Intervene(scm, map[string]float64{"Income": 50})

		hasIntervention := false

		for _, step := range result.Trace.Steps {
			if step.Phase == explain.Intervention {
				hasIntervention = true
			}
		}

		if !hasIntervention {
			t.Fatal("expected intervention phase in trace")
		}
	})

	t.Run("counterfactual trace has counterfactual phase", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Counterfactual(scm,
			map[string]float64{"Education": 10},
			map[string]float64{"Education": 15},
		)

		hasCounterfactual := false

		for _, step := range result.Trace.Steps {
			if step.Phase == explain.Counterfactual {
				hasCounterfactual = true
			}
		}

		if !hasCounterfactual {
			t.Fatal("expected counterfactual phase in trace")
		}
	})
}

func TestTraceString(t *testing.T) {
	t.Parallel()

	t.Run("trace string is non-empty", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Propagate(scm, map[string]float64{"Education": 10})

		traceStr := result.Trace.String()
		if traceStr == "" {
			t.Fatal("expected non-empty trace string")
		}
	})
}

func TestTraceOutputs(t *testing.T) {
	t.Parallel()

	t.Run("trace outputs match result values", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		eng := engine.NewEngine()
		result := eng.Propagate(scm, map[string]float64{"Education": 10})

		if len(result.Trace.Outputs) == 0 {
			t.Fatal("expected trace outputs")
		}

		for k, traceVal := range result.Trace.Outputs {
			resultVal, ok := result.Values[k]
			if !ok {
				t.Fatalf("expected key %s in result values", k)
			}

			if math.Abs(traceVal-resultVal) > 0.001 {
				t.Fatalf("trace output %s=%f != result value %f", k, traceVal, resultVal)
			}
		}
	})
}

func TestSCMOperations(t *testing.T) {
	t.Parallel()

	t.Run("validate detects missing parents", func(t *testing.T) {
		t.Parallel()

		scm := model.NewSCM()

		scm.AddVariable("A", nil, func(_ map[string]float64) float64 { return 1 })
		scm.AddVariable("C", []string{"B"}, func(p map[string]float64) float64 { return p["B"] })

		err := scm.Validate()
		if err == nil {
			t.Fatal("expected validation error for missing parent B")
		}
	})

	t.Run("variables returns topological order", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()
		vars := scm.Variables()

		if len(vars) != 3 {
			t.Fatalf("expected 3 variables, got %d", len(vars))
		}

		eduIdx := -1
		spendIdx := -1

		for i, v := range vars {
			if v == "Education" {
				eduIdx = i
			}

			if v == "Spending" {
				spendIdx = i
			}
		}

		if eduIdx >= spendIdx {
			t.Fatalf("expected Education before Spending, got indices %d, %d", eduIdx, spendIdx)
		}
	})

	t.Run("children and parents", func(t *testing.T) {
		t.Parallel()

		scm := makeEconomicSCM()

		children := scm.Children("Education")
		if len(children) != 1 || children[0] != "Income" {
			t.Fatalf("expected children=[Income], got %v", children)
		}

		parents := scm.Parents("Income")
		if len(parents) != 1 || parents[0] != "Education" {
			t.Fatalf("expected parents=[Education], got %v", parents)
		}
	})
}
