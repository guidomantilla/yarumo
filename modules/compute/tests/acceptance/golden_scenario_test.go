package acceptance_test

import (
	"testing"

	bayesianEngine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	causalEngine "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
	"github.com/guidomantilla/yarumo/compute/engine/causal/model"
	deductiveEngine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
	deductiveRules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
	"github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
	"github.com/guidomantilla/yarumo/compute/math/logic"
)

// Section 2: Cross-Paradigm Golden Scenario — Loan Approval Pipeline

func TestAcceptance_LoanApproval_approved(t *testing.T) {
	t.Parallel()

	// Phase 1: Deductive — eligibility
	t.Run("phase 1 deductive eligibility", func(t *testing.T) {
		t.Parallel()

		r1 := deductiveRules.NewRule("eligibility",
			logic.AndF{L: logic.Var("age_ok"), R: logic.Var("no_fraud")},
			map[logic.Var]bool{"eligible": true},
		)
		r2 := deductiveRules.NewRule("can-apply",
			logic.AndF{L: logic.Var("eligible"), R: logic.Var("has_income")},
			map[logic.Var]bool{"can_apply": true},
		)

		e := deductiveEngine.NewEngine()
		result := e.Forward(
			logic.Fact{"age_ok": true, "no_fraud": true, "has_income": true},
			[]deductiveRules.Rule{r1, r2},
		)

		snap := result.Facts.Snapshot()
		if !snap["eligible"] {
			t.Fatal("expected eligible=true")
		}

		if !snap["can_apply"] {
			t.Fatal("expected can_apply=true")
		}

		if result.Steps != 2 {
			t.Fatalf("expected 2 steps, got %d", result.Steps)
		}
	})

	// Phase 2: Bayesian — P(Default) with good credit and high income
	t.Run("phase 2 bayesian default probability", func(t *testing.T) {
		t.Parallel()

		bn := makeLoanNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("CreditHistory", "good")
		ev.Observe("IncomeLevel", "high")

		eng := bayesianEngine.NewEngine()
		result := eng.Query(bn, ev, "Default")

		got := float64(result.Posterior["yes"])
		// With direct evidence: P(Default=yes|good,high) = 0.02
		assertFloat(t, "P(Default=yes|good,high)", got, 0.02, probTolerance)

		// Posterior must sum to 1
		posteriorSum := float64(result.Posterior["yes"]) + float64(result.Posterior["no"])
		assertFloat(t, "posterior sum", posteriorSum, 1.0, 1e-9)
	})

	// Phase 3: Fuzzy — risk assessment from debt ratio
	t.Run("phase 3 fuzzy risk assessment", func(t *testing.T) {
		t.Parallel()

		eng := makeLoanRiskEngine()
		result := eng.Infer(map[string]float64{"debt_ratio": 60})

		risk := result.Outputs["risk"]
		if risk < 30 || risk > 70 {
			t.Fatalf("expected risk in [30,70] for debt_ratio=60, got %f", risk)
		}

		if len(result.Outputs) == 0 {
			t.Fatal("expected at least one output from fuzzy engine")
		}
	})

	// Phase 4: Causal — what-if income increases
	t.Run("phase 4 causal what-if", func(t *testing.T) {
		t.Parallel()

		scm := model.NewSCM()
		scm.AddVariable("Income", nil, func(_ map[string]float64) float64 { return 0 })
		scm.AddVariable("DebtRatio", []string{"Income"}, func(p map[string]float64) float64 {
			return 100 - p["Income"]*0.8
		})
		scm.AddVariable("RiskScore", []string{"DebtRatio"}, func(p map[string]float64) float64 {
			return p["DebtRatio"] * 0.5
		})

		e := causalEngine.NewEngine()

		// Factual: Income=50 -> DebtRatio=60 -> RiskScore=30
		factual := e.Propagate(scm, map[string]float64{"Income": 50})
		assertFloat(t, "factual DebtRatio", factual.Values["DebtRatio"], 60, floatTolerance)
		assertFloat(t, "factual RiskScore", factual.Values["RiskScore"], 30, floatTolerance)

		if len(factual.Values) != 3 {
			t.Fatalf("expected 3 variables in propagation, got %d", len(factual.Values))
		}

		// Intervention: do(Income=80) -> DebtRatio=36 -> RiskScore=18
		counterfactual := e.Intervene(scm, map[string]float64{"Income": 80})
		assertFloat(t, "cf DebtRatio", counterfactual.Values["DebtRatio"], 36, floatTolerance)
		assertFloat(t, "cf RiskScore", counterfactual.Values["RiskScore"], 18, floatTolerance)

		if counterfactual.Values["RiskScore"] >= factual.Values["RiskScore"] {
			t.Fatalf("intervention should reduce risk: factual=%f, cf=%f",
				factual.Values["RiskScore"], counterfactual.Values["RiskScore"])
		}
	})

	// Phase 5: MCDM — rank loan options
	t.Run("phase 5 mcdm ranking", func(t *testing.T) {
		t.Parallel()

		matrix := [][]float64{
			{5, 30, 200}, // Option A
			{4, 15, 150}, // Option B
			{6, 20, 250}, // Option C
		}
		criteria := []topsis.Criterion{
			{Weight: 0.5, Benefit: false}, // rate: cost (lower is better)
			{Weight: 0.3, Benefit: true},  // term: benefit (higher is better)
			{Weight: 0.2, Benefit: true},  // amount: benefit (higher is better)
		}

		result, err := topsis.Rank(matrix, criteria)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result.Scores) != 3 {
			t.Fatalf("expected 3 scores, got %d", len(result.Scores))
		}

		for i, s := range result.Scores {
			if s < 0 || s > 1 {
				t.Fatalf("option %d score %f outside [0,1]", i, s)
			}
		}

		// Option A balances moderate rate with best term and good amount,
		// which outweighs B's rate advantage under TOPSIS normalization.
		if result.Scores[0] < result.Scores[1] || result.Scores[0] < result.Scores[2] {
			t.Fatalf("Option A should rank highest: A=%f, B=%f, C=%f",
				result.Scores[0], result.Scores[1], result.Scores[2])
		}
	})
}

func TestAcceptance_LoanApproval_rejected(t *testing.T) {
	t.Parallel()

	t.Run("fraud detected stops at phase 1", func(t *testing.T) {
		t.Parallel()

		r1 := deductiveRules.NewRule("eligibility",
			logic.AndF{L: logic.Var("age_ok"), R: logic.Var("no_fraud")},
			map[logic.Var]bool{"eligible": true},
		)

		e := deductiveEngine.NewEngine()
		result := e.Forward(
			logic.Fact{"age_ok": true, "no_fraud": false},
			[]deductiveRules.Rule{r1},
		)

		snap := result.Facts.Snapshot()
		if snap["eligible"] {
			t.Fatal("fraudulent applicant should NOT be eligible")
		}
	})
}

func TestAcceptance_LoanApproval_borderline(t *testing.T) {
	t.Parallel()

	// Phase 1: Eligible — same rules, still passes
	t.Run("phase 1 eligible borderline applicant", func(t *testing.T) {
		t.Parallel()

		r1 := deductiveRules.NewRule("eligibility",
			logic.AndF{L: logic.Var("age_ok"), R: logic.Var("no_fraud")},
			map[logic.Var]bool{"eligible": true},
		)
		r2 := deductiveRules.NewRule("can-apply",
			logic.AndF{L: logic.Var("eligible"), R: logic.Var("has_income")},
			map[logic.Var]bool{"can_apply": true},
		)

		e := deductiveEngine.NewEngine()
		result := e.Forward(
			logic.Fact{"age_ok": true, "no_fraud": true, "has_income": true},
			[]deductiveRules.Rule{r1, r2},
		)

		snap := result.Facts.Snapshot()
		if !snap["can_apply"] {
			t.Fatal("expected can_apply=true")
		}

		if result.Steps != 2 {
			t.Fatalf("expected 2 deductive steps, got %d", result.Steps)
		}
	})

	// Phase 2: Bayesian — good credit, low income -> moderate default
	t.Run("phase 2 moderate default probability", func(t *testing.T) {
		t.Parallel()

		bn := makeLoanNetwork()
		ev := evidence.NewEvidenceBase()
		ev.Observe("CreditHistory", "good")
		ev.Observe("IncomeLevel", "low")

		eng := bayesianEngine.NewEngine()
		result := eng.Query(bn, ev, "Default")

		got := float64(result.Posterior["yes"])
		// P(Default=yes|good,low) = 0.10 (direct CPT lookup)
		assertFloat(t, "P(Default=yes|good,low)", got, 0.10, probTolerance)

		sum := float64(result.Posterior["yes"]) + float64(result.Posterior["no"])
		assertFloat(t, "posterior sum", sum, 1.0, 1e-9)
	})

	// Phase 3: Fuzzy — moderate debt ratio produces medium risk
	t.Run("phase 3 medium risk assessment", func(t *testing.T) {
		t.Parallel()

		eng := makeLoanRiskEngine()
		result := eng.Infer(map[string]float64{"debt_ratio": 50})

		risk := result.Outputs["risk"]
		if risk < 30 || risk > 70 {
			t.Fatalf("expected risk in [30,70] for debt_ratio=50 (medium peak), got %f", risk)
		}
	})

	// Phase 4: Causal — moderate income, intervention improves
	t.Run("phase 4 causal intervention improves risk", func(t *testing.T) {
		t.Parallel()

		scm := model.NewSCM()
		scm.AddVariable("Income", nil, func(_ map[string]float64) float64 { return 0 })
		scm.AddVariable("DebtRatio", []string{"Income"}, func(p map[string]float64) float64 {
			return 100 - p["Income"]*0.8
		})
		scm.AddVariable("RiskScore", []string{"DebtRatio"}, func(p map[string]float64) float64 {
			return p["DebtRatio"] * 0.5
		})

		e := causalEngine.NewEngine()

		factual := e.Propagate(scm, map[string]float64{"Income": 62.5})
		assertFloat(t, "factual DebtRatio", factual.Values["DebtRatio"], 50, floatTolerance)
		assertFloat(t, "factual RiskScore", factual.Values["RiskScore"], 25, floatTolerance)

		intervened := e.Intervene(scm, map[string]float64{"Income": 75})
		assertFloat(t, "cf DebtRatio", intervened.Values["DebtRatio"], 40, floatTolerance)
		assertFloat(t, "cf RiskScore", intervened.Values["RiskScore"], 20, floatTolerance)

		if intervened.Values["RiskScore"] >= factual.Values["RiskScore"] {
			t.Fatalf("intervention should reduce risk: factual=%f, intervened=%f",
				factual.Values["RiskScore"], intervened.Values["RiskScore"])
		}
	})

	// Phase 5: MCDM — close options, verify sensitivity
	t.Run("phase 5 mcdm close ranking", func(t *testing.T) {
		t.Parallel()

		matrix := [][]float64{
			{5, 30, 200},
			{4, 15, 150},
			{6, 20, 250},
		}
		criteria := []topsis.Criterion{
			{Weight: 0.5, Benefit: false},
			{Weight: 0.3, Benefit: true},
			{Weight: 0.2, Benefit: true},
		}

		result, err := topsis.Rank(matrix, criteria)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, s := range result.Scores {
			if s < 0 || s > 1 {
				t.Fatalf("option %d score %f outside [0,1]", i, s)
			}
		}

		// Option A balances moderate rate with best term and good amount,
		// which outweighs B's rate advantage under TOPSIS normalization.
		if result.Scores[0] < result.Scores[1] || result.Scores[0] < result.Scores[2] {
			t.Fatalf("Option A should rank highest: A=%f, B=%f, C=%f",
				result.Scores[0], result.Scores[1], result.Scores[2])
		}
	})
}
