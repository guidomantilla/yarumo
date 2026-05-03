package acceptance_test

import (
	"maps"
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
	fuzzyEngine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	fuzzyRules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
	"github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
	fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

const (
	floatTolerance  = 1e-9 // general floating-point arithmetic
	probTolerance   = 1e-6 // probability marginalization accumulates error
	defuzzTolerance = 0.5  // defuzzification discretization introduces error
	goldenBayesian  = 1e-4 // golden files for Bayesian posteriors
	goldenFuzzy     = 1e-2 // golden files for fuzzy outputs
)

func assertFloat(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()

	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s: got %.12f, want %.12f (tolerance %e)", name, got, want, tolerance)
	}
}

// generateFormulas builds all formulas up to the given depth using the provided variables.
// depth=0: each Var, TrueF{}, FalseF{}
// depth=n: NotF{f} for each f at depth n-1, BinOp{f1,f2} for each f1,f2 at depth<=n-1.
func generateFormulas(depth int, vars []logic.Var) []logic.Formula {
	if depth < 0 {
		return nil
	}

	// depth 0: atoms
	atoms := make([]logic.Formula, 0, len(vars)+2)
	for _, v := range vars {
		atoms = append(atoms, v)
	}

	atoms = append(atoms, logic.TrueF{}, logic.FalseF{})

	if depth == 0 {
		return atoms
	}

	// Build incrementally
	byDepth := make([][]logic.Formula, depth+1)
	byDepth[0] = atoms

	for d := 1; d <= depth; d++ {
		var level []logic.Formula

		// Collect all formulas at depth < d
		var prior []logic.Formula
		for dd := range d {
			prior = append(prior, byDepth[dd]...)
		}

		// NotF of depth d-1
		for _, f := range byDepth[d-1] {
			level = append(level, logic.NotF{F: f})
		}

		// BinOps: one operand from depth d-1, other from depth <= d-1
		for _, f1 := range byDepth[d-1] {
			for _, f2 := range prior {
				level = append(level, logic.AndF{L: f1, R: f2})
				level = append(level, logic.OrF{L: f1, R: f2})
				level = append(level, logic.ImplF{L: f1, R: f2})
				level = append(level, logic.IffF{L: f1, R: f2})
			}
		}

		// Also: prior x depth d-1 (reversed operand order for non-commutative ops)
		for _, f1 := range prior {
			for _, f2 := range byDepth[d-1] {
				// Skip if f1 is also at d-1 (already covered above)
				level = append(level, logic.ImplF{L: f1, R: f2})
			}
		}

		byDepth[d] = level

		// Safety: if total exceeds 5000, stop
		total := 0
		for dd := 0; dd <= d; dd++ {
			total += len(byDepth[dd])
		}

		if total > 5000 {
			break
		}
	}

	var all []logic.Formula
	for _, formulas := range byDepth {
		all = append(all, formulas...)
	}

	return all
}

// fuzzyGrid generates values from 0.0 to 1.0 with the given step.
// step=0.05 produces 21 values: [0.0, 0.05, 0.10, ..., 1.0].
func fuzzyGrid(step float64) []float64 {
	var vals []float64

	for v := 0.0; v <= 1.0+step/2; v += step {
		if v > 1.0 {
			v = 1.0
		}

		vals = append(vals, v)
	}

	return vals
}

// makeRainNetwork builds the Rain-Sprinkler-WetGrass Bayesian network from Koller & Friedman.
func makeRainNetwork() network.Network {
	bn := network.NewNetwork()

	rainCPT := bayesian.NewCPT("Rain", nil)
	rainCPT.Set(stats.Assignment{}, stats.Distribution{"true": 0.2, "false": 0.8})
	bn.AddNode(network.Node{
		Variable: "Rain", CPT: rainCPT, Outcomes: []stats.Outcome{"true", "false"},
	})

	sprinklerCPT := bayesian.NewCPT("Sprinkler", []stats.Var{"Rain"})
	sprinklerCPT.Set(stats.Assignment{"Rain": "true"}, stats.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(stats.Assignment{"Rain": "false"}, stats.Distribution{"true": 0.4, "false": 0.6})
	bn.AddNode(network.Node{
		Variable: "Sprinkler", Parents: []stats.Var{"Rain"}, CPT: sprinklerCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	wetGrassCPT := bayesian.NewCPT("WetGrass", []stats.Var{"Rain", "Sprinkler"})
	wetGrassCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "true"}, stats.Distribution{"true": 0.99, "false": 0.01})
	wetGrassCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "false"}, stats.Distribution{"true": 0.8, "false": 0.2})
	wetGrassCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "true"}, stats.Distribution{"true": 0.9, "false": 0.1})
	wetGrassCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "false"}, stats.Distribution{"true": 0.0, "false": 1.0})
	bn.AddNode(network.Node{
		Variable: "WetGrass", Parents: []stats.Var{"Rain", "Sprinkler"}, CPT: wetGrassCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	return bn
}

// makeLoanNetwork builds the CreditHistory-Default-IncomeLevel network for the loan approval scenario.
func makeLoanNetwork() network.Network {
	bn := network.NewNetwork()

	chCPT := bayesian.NewCPT("CreditHistory", nil)
	chCPT.Set(stats.Assignment{}, stats.Distribution{"good": 0.7, "bad": 0.3})
	bn.AddNode(network.Node{
		Variable: "CreditHistory", CPT: chCPT, Outcomes: []stats.Outcome{"good", "bad"},
	})

	ilCPT := bayesian.NewCPT("IncomeLevel", nil)
	ilCPT.Set(stats.Assignment{}, stats.Distribution{"high": 0.6, "low": 0.4})
	bn.AddNode(network.Node{
		Variable: "IncomeLevel", CPT: ilCPT, Outcomes: []stats.Outcome{"high", "low"},
	})

	defaultCPT := bayesian.NewCPT("Default", []stats.Var{"CreditHistory", "IncomeLevel"})
	defaultCPT.Set(stats.Assignment{"CreditHistory": "good", "IncomeLevel": "high"}, stats.Distribution{"yes": 0.02, "no": 0.98})
	defaultCPT.Set(stats.Assignment{"CreditHistory": "good", "IncomeLevel": "low"}, stats.Distribution{"yes": 0.10, "no": 0.90})
	defaultCPT.Set(stats.Assignment{"CreditHistory": "bad", "IncomeLevel": "high"}, stats.Distribution{"yes": 0.15, "no": 0.85})
	defaultCPT.Set(stats.Assignment{"CreditHistory": "bad", "IncomeLevel": "low"}, stats.Distribution{"yes": 0.40, "no": 0.60})
	bn.AddNode(network.Node{
		Variable: "Default", Parents: []stats.Var{"CreditHistory", "IncomeLevel"}, CPT: defaultCPT,
		Outcomes: []stats.Outcome{"yes", "no"},
	})

	return bn
}

// makeLoanRiskEngine builds a fuzzy risk assessment engine for the loan approval scenario.
func makeLoanRiskEngine() fuzzyEngine.Engine {
	lowDR, _ := fuzzym.Trapezoidal(0, 0, 20, 40)
	medDR, _ := fuzzym.Triangular(30, 50, 70)
	highDR, _ := fuzzym.Trapezoidal(60, 80, 100, 100)

	debtRatio := variable.NewVariable("debt_ratio", 0, 100, []variable.Term{
		{Name: "low", Fn: lowDR}, {Name: "medium", Fn: medDR}, {Name: "high", Fn: highDR},
	})

	lowR, _ := fuzzym.Trapezoidal(0, 0, 20, 40)
	medR, _ := fuzzym.Triangular(30, 50, 70)
	highR, _ := fuzzym.Trapezoidal(60, 80, 100, 100)

	risk := variable.NewVariable("risk", 0, 100, []variable.Term{
		{Name: "low", Fn: lowR}, {Name: "medium", Fn: medR}, {Name: "high", Fn: highR},
	})

	ruleSet := []fuzzyRules.Rule{
		fuzzyRules.NewRule("low-low", []fuzzyRules.Condition{{Variable: "debt_ratio", Term: "low"}},
			fuzzyRules.Consequent{Variable: "risk", Term: "low"}),
		fuzzyRules.NewRule("med-med", []fuzzyRules.Condition{{Variable: "debt_ratio", Term: "medium"}},
			fuzzyRules.Consequent{Variable: "risk", Term: "medium"}),
		fuzzyRules.NewRule("high-high", []fuzzyRules.Condition{{Variable: "debt_ratio", Term: "high"}},
			fuzzyRules.Consequent{Variable: "risk", Term: "high"}),
	}

	return fuzzyEngine.NewEngine(
		[]variable.Variable{debtRatio},
		[]variable.Variable{risk},
		ruleSet,
	)
}

// makeTippingEngine builds the canonical tipping fuzzy engine (Mamdani).
func makeTippingEngine(opts ...fuzzyEngine.Option) fuzzyEngine.Engine {
	bad, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
	average, _ := fuzzym.Triangular(2, 5, 8)
	good, _ := fuzzym.Trapezoidal(6, 8, 10, 10)

	food := variable.NewVariable("food", 0, 10, []variable.Term{
		{Name: "bad", Fn: bad}, {Name: "average", Fn: average}, {Name: "good", Fn: good},
	})

	poor, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
	acceptable, _ := fuzzym.Triangular(2, 5, 8)
	excellent, _ := fuzzym.Trapezoidal(6, 8, 10, 10)

	service := variable.NewVariable("service", 0, 10, []variable.Term{
		{Name: "poor", Fn: poor}, {Name: "acceptable", Fn: acceptable}, {Name: "excellent", Fn: excellent},
	})

	lowTip, _ := fuzzym.Trapezoidal(0, 0, 5, 10)
	medTip, _ := fuzzym.Triangular(5, 12.5, 20)
	highTip, _ := fuzzym.Trapezoidal(15, 20, 25, 25)

	tip := variable.NewVariable("tip", 0, 25, []variable.Term{
		{Name: "low", Fn: lowTip}, {Name: "medium", Fn: medTip}, {Name: "high", Fn: highTip},
	})

	ruleSet := []fuzzyRules.Rule{
		fuzzyRules.NewRule("r1",
			[]fuzzyRules.Condition{{Variable: "food", Term: "bad"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
		fuzzyRules.NewRule("r2",
			[]fuzzyRules.Condition{{Variable: "service", Term: "poor"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
		fuzzyRules.NewRule("r3",
			[]fuzzyRules.Condition{{Variable: "food", Term: "average"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "medium"}),
		fuzzyRules.NewRule("r4",
			[]fuzzyRules.Condition{{Variable: "food", Term: "good"}, {Variable: "service", Term: "excellent"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
		fuzzyRules.NewRule("r5",
			[]fuzzyRules.Condition{{Variable: "service", Term: "excellent"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
		fuzzyRules.NewRule("r6",
			[]fuzzyRules.Condition{{Variable: "food", Term: "good"}},
			fuzzyRules.Consequent{Variable: "tip", Term: "medium"}),
	}

	return fuzzyEngine.NewEngine(
		[]variable.Variable{food, service},
		[]variable.Variable{tip},
		ruleSet,
		opts...,
	)
}

// generateEvidenceCombos generates all possible evidence subsets for the given variables.
func generateEvidenceCombos(vars []stats.Var, outcomes map[stats.Var][]stats.Outcome) []map[stats.Var]stats.Outcome {
	combos := []map[stats.Var]stats.Outcome{{}} // start with empty evidence

	for _, v := range vars {
		var expanded []map[stats.Var]stats.Outcome
		for _, combo := range combos {
			// Keep without this variable
			expanded = append(expanded, combo)
			// Add each outcome
			for _, o := range outcomes[v] {
				newCombo := make(map[stats.Var]stats.Outcome)
				maps.Copy(newCombo, combo)

				newCombo[v] = o
				expanded = append(expanded, newCombo)
			}
		}

		combos = expanded
	}

	return combos
}

// isCNF checks that a formula is in Conjunctive Normal Form.
func isCNF(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.Var, logic.TrueF, logic.FalseF:
		return true
	case logic.NotF:
		_, ok := v.F.(logic.Var)
		return ok
	case logic.OrF:
		return isClause(v)
	case logic.AndF:
		return isCNFConjunct(v.L) && isCNFConjunct(v.R)
	default:
		return false
	}
}

func isClause(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.Var, logic.TrueF, logic.FalseF:
		return true
	case logic.NotF:
		_, ok := v.F.(logic.Var)
		return ok
	case logic.OrF:
		return isClause(v.L) && isClause(v.R)
	default:
		return false
	}
}

func isCNFConjunct(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.AndF:
		return isCNFConjunct(v.L) && isCNFConjunct(v.R)
	default:
		return isClause(f)
	}
}

func isDNF(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.Var, logic.TrueF, logic.FalseF:
		return true
	case logic.NotF:
		_, ok := v.F.(logic.Var)
		return ok
	case logic.AndF:
		return isDNFConjunct(v)
	case logic.OrF:
		return isDNFDisjunct(v.L) && isDNFDisjunct(v.R)
	default:
		return false
	}
}

func isDNFConjunct(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.Var, logic.TrueF, logic.FalseF:
		return true
	case logic.NotF:
		_, ok := v.F.(logic.Var)
		return ok
	case logic.AndF:
		return isDNFConjunct(v.L) && isDNFConjunct(v.R)
	default:
		return false
	}
}

func isDNFDisjunct(f logic.Formula) bool {
	switch v := f.(type) {
	case logic.OrF:
		return isDNFDisjunct(v.L) && isDNFDisjunct(v.R)
	default:
		return isDNFConjunct(f)
	}
}
