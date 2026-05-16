package acceptance

import (
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	bayesianEngine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
	causalEngine "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
	deductiveEngine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	deductiveExplain "github.com/guidomantilla/yarumo/compute/engine/deductive/explain"
	deductiveRules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
	fuzzyEngine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
	fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/logic"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// Section 3: Behavioral Snapshots — Golden Files

// Appendix B derivations:
// B.1: P(Rain=t|WG=t) = 0.16038/0.44838 ~ 0.35770
// B.5: P(Rain=t|WG=t,Spr=t) = 0.00198/0.28998 ~ 0.00683
// B.6: P(Spr=t|WG=t) = 0.28998/0.44838 ~ 0.64680

func TestGolden_BayesianRainPosterior(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	t.Run("P(Rain=true | WetGrass=true)", func(t *testing.T) {
		t.Parallel()

		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		eng := bayesianEngine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		got := float64(result.Posterior["true"])
		if math.Abs(got-0.35770) > goldenBayesian {
			t.Fatalf("behavioral regression: P(Rain=true|WetGrass=true) was 0.35770, now %f", got)
		}
	})

	t.Run("P(Rain=true | WetGrass=true, Sprinkler=true) explaining away", func(t *testing.T) {
		t.Parallel()

		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")
		ev.Observe("Sprinkler", "true")

		eng := bayesianEngine.NewEngine()
		result := eng.Query(bn, ev, "Rain")

		got := float64(result.Posterior["true"])
		if math.Abs(got-0.00683) > goldenBayesian {
			t.Fatalf("behavioral regression: P(Rain=true|WetGrass=true,Sprinkler=true) was 0.00683, now %f", got)
		}
	})

	t.Run("P(Sprinkler=true | WetGrass=true) marginal", func(t *testing.T) {
		t.Parallel()

		ev := evidence.NewEvidenceBase()
		ev.Observe("WetGrass", "true")

		eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
		result := eng.Query(bn, ev, "Sprinkler")

		got := float64(result.Posterior["true"])
		if math.Abs(got-0.64680) > goldenBayesian {
			t.Fatalf("behavioral regression: P(Sprinkler=true|WetGrass=true) was 0.64680, now %f", got)
		}
	})

	t.Run("medical network P(Disease=present | TestResult=positive)", func(t *testing.T) {
		t.Parallel()

		medBN := network.NewNetwork()

		diseaseCPT := bayesian.NewCPT("Disease", nil)
		diseaseCPT.Set(stats.Assignment{}, stats.Distribution{"present": 0.001, "absent": 0.999})
		medBN.AddNode(network.Node{
			Variable: "Disease", CPT: diseaseCPT, Outcomes: []stats.Outcome{"present", "absent"},
		})

		testCPT := bayesian.NewCPT("TestResult", []stats.Var{"Disease"})
		testCPT.Set(stats.Assignment{"Disease": "present"}, stats.Distribution{"positive": 0.99, "negative": 0.01})
		testCPT.Set(stats.Assignment{"Disease": "absent"}, stats.Distribution{"positive": 0.05, "negative": 0.95})
		medBN.AddNode(network.Node{
			Variable: "TestResult", Parents: []stats.Var{"Disease"}, CPT: testCPT,
			Outcomes: []stats.Outcome{"positive", "negative"},
		})

		ev := evidence.NewEvidenceBase()
		ev.Observe("TestResult", "positive")

		eng := bayesianEngine.NewEngine()
		result := eng.Query(medBN, ev, "Disease")

		got := float64(result.Posterior["present"])
		if math.Abs(got-0.01943) > goldenBayesian {
			t.Fatalf("behavioral regression: P(Disease=present|TestResult=positive) was 0.01943, now %f", got)
		}
	})
}

func TestGolden_DeductiveBusinessScenario(t *testing.T) {
	t.Parallel()

	premium := deductiveRules.NewRule("premium-customer",
		logic.AndF{L: logic.Var("high_spend"), R: logic.Var("loyal")},
		map[logic.Var]bool{"premium": true},
	)
	discount := deductiveRules.NewRule("discount-eligible",
		logic.Var("premium"),
		map[logic.Var]bool{"discount": true},
	)
	notification := deductiveRules.NewRule("notify",
		logic.Var("discount"),
		map[logic.Var]bool{"notify": true},
	)

	e := deductiveEngine.NewEngine()
	result := e.Forward(logic.Fact{"high_spend": true, "loyal": true},
		[]deductiveRules.Rule{premium, discount, notification})

	snap := result.Facts.Snapshot()

	expected := map[logic.Var]bool{"high_spend": true, "loyal": true, "premium": true, "discount": true, "notify": true}
	for k, v := range expected {
		if snap[k] != v {
			t.Fatalf("behavioral regression: expected %s=%v, got %v", k, v, snap[k])
		}
	}

	if result.Steps != 3 {
		t.Fatalf("behavioral regression: expected 3 steps, got %d", result.Steps)
	}

	// Trace structure verification: Trace is a struct with Steps field.
	if len(result.Trace.Steps) != 3 {
		t.Fatalf("behavioral regression: expected 3 trace steps (one per rule fired), got %d", len(result.Trace.Steps))
	}

	// Verify provenance: derived facts should have Origin == Derived.
	derivedFacts := []logic.Var{"premium", "discount", "notify"}
	for _, fact := range derivedFacts {
		prov, ok := result.Facts.Provenance(fact)
		if !ok {
			t.Fatalf("behavioral regression: no provenance for derived fact %q", fact)
		}

		if prov.Origin != deductiveExplain.Derived {
			t.Fatalf("behavioral regression: fact %q should have Origin=Derived, got %v", fact, prov.Origin)
		}
	}

	// Initial facts should have Origin == Asserted.
	initialFacts := []logic.Var{"high_spend", "loyal"}
	for _, fact := range initialFacts {
		prov, ok := result.Facts.Provenance(fact)
		if !ok {
			t.Fatalf("behavioral regression: no provenance for initial fact %q", fact)
		}

		if prov.Origin != deductiveExplain.Asserted {
			t.Fatalf("behavioral regression: initial fact %q should have Origin=Asserted, got %v", fact, prov.Origin)
		}
	}
}

func TestGolden_FuzzyTipping(t *testing.T) {
	t.Parallel()

	eng := makeTippingEngine() // Mamdani with Centroid defuzzification

	t.Run("low inputs Mamdani/Centroid", func(t *testing.T) {
		t.Parallel()

		result := eng.Infer(map[string]float64{"food": 1.0, "service": 1.0})
		tip := result.Outputs["tip"]

		if tip < 0 || tip > 25 {
			t.Fatalf("behavioral regression: tip=%f outside [0,25] for food=1.0, service=1.0", tip)
		}
		// [TO_FREEZE: replace range check with exact frozen value comparison]
	})

	t.Run("mid inputs Mamdani/Centroid", func(t *testing.T) {
		t.Parallel()

		result := eng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
		tip := result.Outputs["tip"]

		if tip < 0 || tip > 25 {
			t.Fatalf("behavioral regression: tip=%f outside [0,25] for food=5.0, service=5.0", tip)
		}
		// [TO_FREEZE: replace range check with exact frozen value comparison]
	})

	t.Run("high inputs Mamdani/Centroid", func(t *testing.T) {
		t.Parallel()

		result := eng.Infer(map[string]float64{"food": 9.0, "service": 9.0})
		tip := result.Outputs["tip"]

		if tip < 0 || tip > 25 {
			t.Fatalf("behavioral regression: tip=%f outside [0,25] for food=9.0, service=9.0", tip)
		}
		// [TO_FREEZE: replace range check with exact frozen value comparison]
	})

	t.Run("mid inputs Mamdani/Bisector", func(t *testing.T) {
		t.Parallel()

		bisectorEng := makeTippingEngine(fuzzyEngine.WithDefuzzify(fuzzym.Bisector))

		result := bisectorEng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
		tipVal := result.Outputs["tip"]

		// [TO_FREEZE: capture exact bisector value and replace range check]
		if tipVal < 0 || tipVal > 25 {
			t.Fatalf("behavioral regression: Bisector tip=%f outside [0,25]", tipVal)
		}
	})

	t.Run("mid inputs Mamdani/MeanOfMax", func(t *testing.T) {
		t.Parallel()

		momEng := makeTippingEngine(fuzzyEngine.WithDefuzzify(fuzzym.MeanOfMax))

		result := momEng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
		tipVal := result.Outputs["tip"]

		// [TO_FREEZE: capture exact MeanOfMax value and replace range check]
		if tipVal < 0 || tipVal > 25 {
			t.Fatalf("behavioral regression: MeanOfMax tip=%f outside [0,25]", tipVal)
		}
	})

	t.Run("high inputs Sugeno", func(t *testing.T) {
		t.Parallel()

		sugenoEng := makeTippingEngine(fuzzyEngine.WithMethod(fuzzyEngine.Sugeno))

		result := sugenoEng.Infer(map[string]float64{"food": 9.0, "service": 9.0})
		tipVal := result.Outputs["tip"]

		// [TO_FREEZE: capture exact Sugeno value and replace range check]
		if tipVal < 0 || tipVal > 25 {
			t.Fatalf("behavioral regression: Sugeno tip=%f outside [0,25]", tipVal)
		}
	})
}

// Appendix B.3: Causal Linear SCM
// Propagate X=5 -> Z=10 -> Y=13
// Intervene do(Z=7) -> Y=10
// CF do(X=10) -> Y=23
// CF do(Z=7)|X=5 -> Y=10

func TestGolden_CausalLinearSCM(t *testing.T) {
	t.Parallel()

	scm := model.NewSCM()
	scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 0 })
	scm.AddVariable("Z", []string{"X"}, func(p map[string]float64) float64 { return p["X"] * 2 })
	scm.AddVariable("Y", []string{"Z"}, func(p map[string]float64) float64 { return p["Z"] + 3 })

	e := causalEngine.NewEngine()

	t.Run("propagate", func(t *testing.T) {
		t.Parallel()

		result := e.Propagate(scm, map[string]float64{"X": 5})
		if math.Abs(result.Values["Y"]-13.0) > floatTolerance {
			t.Fatalf("behavioral regression: Propagate(X=5) Y was 13.0, now %f", result.Values["Y"])
		}
	})

	t.Run("intervene", func(t *testing.T) {
		t.Parallel()

		result := e.Intervene(scm, map[string]float64{"Z": 7})
		if math.Abs(result.Values["Y"]-10.0) > floatTolerance {
			t.Fatalf("behavioral regression: Intervene(Z=7) Y was 10.0, now %f", result.Values["Y"])
		}
	})

	t.Run("counterfactual do(X=10) given factual X=5", func(t *testing.T) {
		t.Parallel()

		result := e.Counterfactual(scm, map[string]float64{"X": 5}, map[string]float64{"X": 10})
		if math.Abs(result.Values["Y"]-23.0) > floatTolerance {
			t.Fatalf("behavioral regression: Counterfactual Y was 23.0, now %f", result.Values["Y"])
		}
	})

	t.Run("counterfactual do(Z=7) given factual X=5", func(t *testing.T) {
		t.Parallel()

		result := e.Counterfactual(scm, map[string]float64{"X": 5}, map[string]float64{"Z": 7})

		if math.Abs(result.Values["X"]-5.0) > floatTolerance {
			t.Fatalf("behavioral regression: Counterfactual(do(Z=7)|X=5) X was 5.0, now %f", result.Values["X"])
		}

		if math.Abs(result.Values["Z"]-7.0) > floatTolerance {
			t.Fatalf("behavioral regression: Counterfactual(do(Z=7)|X=5) Z was 7.0, now %f", result.Values["Z"])
		}

		if math.Abs(result.Values["Y"]-10.0) > floatTolerance {
			t.Fatalf("behavioral regression: Counterfactual(do(Z=7)|X=5) Y was 10.0, now %f", result.Values["Y"])
		}
	})
}

// Appendix B.4: AHP weights [0.6, 0.3, 0.1], CR=0.0

func TestGolden_AHPConsistentMatrix(t *testing.T) {
	t.Parallel()

	matrix := ahp.PairwiseMatrix{
		{1, 2, 6},
		{0.5, 1, 3},
		{1.0 / 6, 1.0 / 3, 1},
	}

	result, err := ahp.Analyze(matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frozen := []float64{0.6, 0.3, 0.1}
	for i, w := range frozen {
		if math.Abs(result.Weights[i]-w) > 1e-6 {
			t.Fatalf("behavioral regression: AHP weight[%d] was %f, now %f", i, w, result.Weights[i])
		}
	}

	if math.Abs(result.ConsistencyRatio-0.0) > 1e-9 {
		t.Fatalf("behavioral regression: AHP CR was 0.0, now %f", result.ConsistencyRatio)
	}
}

func TestGolden_TOPSISStrictDominance(t *testing.T) {
	t.Parallel()

	// Alternative A dominates B on all criteria.
	alternatives := [][]float64{
		{10, 10, 10},
		{1, 1, 1},
	}
	criteria := []topsis.Criterion{
		{Weight: 1.0 / 3.0, Benefit: true},
		{Weight: 1.0 / 3.0, Benefit: true},
		{Weight: 1.0 / 3.0, Benefit: true},
	}

	result, err := topsis.Rank(alternatives, criteria)
	if err != nil {
		t.Fatalf("behavioral regression: TOPSIS Rank returned unexpected error: %v", err)
	}

	t.Run("dominant alternative score is 1.0", func(t *testing.T) {
		t.Parallel()

		if math.Abs(result.Scores[0]-1.0) > 1e-9 {
			t.Fatalf("behavioral regression: TOPSIS Score(A) was 1.0, now %f", result.Scores[0])
		}
	})

	t.Run("dominated alternative score is 0.0", func(t *testing.T) {
		t.Parallel()

		if math.Abs(result.Scores[1]-0.0) > 1e-9 {
			t.Fatalf("behavioral regression: TOPSIS Score(B) was 0.0, now %f", result.Scores[1])
		}
	})

	t.Run("ranking order by score", func(t *testing.T) {
		t.Parallel()

		// No Ranking field in Result, verify via Scores: A should have higher score than B
		if result.Scores[0] <= result.Scores[1] {
			t.Fatalf("behavioral regression: dominant alternative A (score=%f) should outrank B (score=%f)",
				result.Scores[0], result.Scores[1])
		}
	})
}
