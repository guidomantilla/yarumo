# Acceptance Tests — Prompt 05: Golden Tests

## Context

Read `00-context.md` for project structure, imports, and coding standards.

## Role

You are a Go testing engineer. Generate golden test files for the acceptance test module.

## Output

Generate exactly TWO files:
1. `golden_scenario_test.go` — Section 2: Loan Approval Pipeline (3 variants)
2. `golden_files_test.go` — Section 3: Behavioral Snapshots (6 snapshots)

Place them at: `modules/compute/tests/acceptance/`

## Constraints

- Package: `package acceptance_test`
- No testify — use `t.Fatal`/`t.Fatalf`
- No table-driven tests — individual `t.Run` subtests
- `t.Parallel()` on every test and subtest
- Only import public APIs
- No inline assignments

## Helper References

From `helpers_test.go`:
- `assertFloat(t, name, got, want, tolerance)` — float comparison
- `makeRainNetwork()` — Rain-Sprinkler-WetGrass network
- `makeLoanNetwork()` — CreditHistory-Default-IncomeLevel network
- `makeLoanRiskEngine()` — fuzzy risk assessment (debt_ratio → risk)
- `makeTippingEngine(opts ...fuzzyEngine.Option)` — canonical tipping engine
- Tolerances: `floatTolerance` (1e-9), `probTolerance` (1e-6), `defuzzTolerance` (0.5), `goldenBayesian` (1e-4), `goldenFuzzy` (1e-2)

## Required Imports (both files)

```go
import (
    "math"
    "testing"

    "github.com/guidomantilla/yarumo/compute/math/logic"
    fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/math/stats"

    bayesianEngine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
    deductiveEngine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
    deductiveRules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
    fuzzyEngine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
    fuzzyRules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
    causalEngine "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
    "github.com/guidomantilla/yarumo/compute/engine/causal/model"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)
```

---

## File 1: golden_scenario_test.go

### Scenario: Loan Approval Pipeline

A loan approval system that uses all 5 paradigms in sequence:
1. Deductive: Eligibility check (age_ok AND no_fraud → eligible, eligible AND has_income → can_apply)
2. Bayesian: Default probability (CreditHistory-Default-IncomeLevel network)
3. Fuzzy: Risk assessment (debt_ratio → risk via makeLoanRiskEngine)
4. Causal: What-if (Income → DebtRatio → RiskScore, DebtRatio=100-Income*0.8, RiskScore=DebtRatio*0.5)
5. MCDM: Loan options (3 options, criteria: rate=cost(0.5), term=benefit(0.3), amount=benefit(0.2))

### Test: TestAcceptance_LoanApproval_approved

5 subtests (one per phase):
1. "phase 1 deductive eligibility" — age_ok=true, no_fraud=true, has_income=true → eligible=true, can_apply=true, Steps=2
2. "phase 2 bayesian default probability" — CreditHistory=good, IncomeLevel=high → P(Default=yes)=0.02, posterior sums to 1
3. "phase 3 fuzzy risk assessment" — debt_ratio=60, risk in [30,70], at least one output
4. "phase 4 causal what-if" — Income=50→DebtRatio=60→RiskScore=30, do(Income=80)→DebtRatio=36→RiskScore=18, intervention reduces risk
5. "phase 5 mcdm ranking" — 3 options {5,30,200},{4,15,150},{6,20,250}, scores in [0,1], Option B (lowest rate) ranks highest

```go
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

        // Trace verification: posterior must sum to 1
        posteriorSum := float64(result.Posterior["yes"]) + float64(result.Posterior["no"])
        assertFloat(t, "posterior sum", posteriorSum, 1.0, 1e-9)
    })

    // Phase 3: Fuzzy — risk assessment from debt ratio
    t.Run("phase 3 fuzzy risk assessment", func(t *testing.T) {
        t.Parallel()

        eng := makeLoanRiskEngine()
        result := eng.Infer(map[string]float64{"debt_ratio": 60})

        risk := result.Outputs["risk"]
        // debt_ratio=60: low μ=0, medium μ=(70-60)/(70-50)=0.5, high μ=(60-60)/(80-60)=0
        // Only medium rule fires at strength 0.5.
        // Defuzzified output should be near 50 (center of medium term).
        if risk < 30 || risk > 70 {
            t.Fatalf("expected risk in [30,70] for debt_ratio=60, got %f", risk)
        }

        // Verify trace: fuzzification produced expected membership degrees
        // (implementation detail: check result has at least one output)
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

        // Factual: Income=50 → DebtRatio=60 → RiskScore=30
        factual := e.Propagate(scm, map[string]float64{"Income": 50})
        assertFloat(t, "factual DebtRatio", factual.Values["DebtRatio"], 60, floatTolerance)
        assertFloat(t, "factual RiskScore", factual.Values["RiskScore"], 30, floatTolerance)

        // Trace: verify all 3 variables computed
        if len(factual.Values) != 3 {
            t.Fatalf("expected 3 variables in propagation, got %d", len(factual.Values))
        }

        // Intervention: do(Income=80) → DebtRatio=36 → RiskScore=18
        counterfactual := e.Intervene(scm, map[string]float64{"Income": 80})
        assertFloat(t, "cf DebtRatio", counterfactual.Values["DebtRatio"], 36, floatTolerance)
        assertFloat(t, "cf RiskScore", counterfactual.Values["RiskScore"], 18, floatTolerance)

        // Verify intervention reduced risk
        if counterfactual.Values["RiskScore"] >= factual.Values["RiskScore"] {
            t.Fatalf("intervention should reduce risk: factual=%f, cf=%f",
                factual.Values["RiskScore"], counterfactual.Values["RiskScore"])
        }
    })

    // Phase 5: MCDM — rank loan options
    t.Run("phase 5 mcdm ranking", func(t *testing.T) {
        t.Parallel()

        matrix := [][]float64{
            {5, 30, 200},  // Option A
            {4, 15, 150},  // Option B
            {6, 20, 250},  // Option C
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

        // All scores should be in [0,1]
        for i, s := range result.Scores {
            if s < 0 || s > 1 {
                t.Fatalf("option %d score %f outside [0,1]", i, s)
            }
        }

        // With rate as dominant criterion (0.5 weight, cost), Option B (rate=4) should rank best
        if result.Scores[1] < result.Scores[0] || result.Scores[1] < result.Scores[2] {
            t.Fatalf("Option B (lowest rate) should rank highest: A=%f, B=%f, C=%f",
                result.Scores[0], result.Scores[1], result.Scores[2])
        }
    })
}
```

### Test: TestAcceptance_LoanApproval_rejected

1 subtest:
- "fraud detected stops at phase 1" — age_ok=true, no_fraud=false → eligible NOT derived

```go
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
```

### Test: TestAcceptance_LoanApproval_borderline

Variant 3: Eligible applicant with mixed signals — good credit but low income.

Derivations:
- P(Default=yes | good, low) = 0.10 (direct CPT lookup)
- debt_ratio=50 (Income=62.5 → DebtRatio=100-62.5×0.8=50): medium μ(50)=1.0 (peak of triangular(30,50,70)), high μ=0, low μ=0 → only medium fires at 1.0
- do(Income=75) → DebtRatio=100-75×0.8=40 → RiskScore=40×0.5=20

5 subtests (one per phase):

```go
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
        // Verify trace: exactly 2 steps (eligibility + can-apply)
        if result.Steps != 2 {
            t.Fatalf("expected 2 deductive steps, got %d", result.Steps)
        }
    })

    // Phase 2: Bayesian — good credit, low income → moderate default
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

        // Posterior must sum to 1
        sum := float64(result.Posterior["yes"]) + float64(result.Posterior["no"])
        assertFloat(t, "posterior sum", sum, 1.0, 1e-9)
    })

    // Phase 3: Fuzzy — moderate debt ratio produces medium risk
    t.Run("phase 3 medium risk assessment", func(t *testing.T) {
        t.Parallel()

        eng := makeLoanRiskEngine()
        // debt_ratio=50 → medium term peaks at 50, so μ_medium=1.0, others=0
        result := eng.Infer(map[string]float64{"debt_ratio": 50})

        risk := result.Outputs["risk"]
        // Only medium rule fires at 1.0 → risk should be near 50 (medium center)
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

        // Factual: Income=62.5 → DebtRatio=50 → RiskScore=25
        factual := e.Propagate(scm, map[string]float64{"Income": 62.5})
        assertFloat(t, "factual DebtRatio", factual.Values["DebtRatio"], 50, floatTolerance)
        assertFloat(t, "factual RiskScore", factual.Values["RiskScore"], 25, floatTolerance)

        // Counterfactual: do(Income=75) → DebtRatio=40 → RiskScore=20
        // Income increases by 20% → risk drops
        intervened := e.Intervene(scm, map[string]float64{"Income": 75})
        assertFloat(t, "cf DebtRatio", intervened.Values["DebtRatio"], 40, floatTolerance)
        assertFloat(t, "cf RiskScore", intervened.Values["RiskScore"], 20, floatTolerance)

        // Verify intervention reduces risk
        if intervened.Values["RiskScore"] >= factual.Values["RiskScore"] {
            t.Fatalf("intervention should reduce risk: factual=%f, intervened=%f",
                factual.Values["RiskScore"], intervened.Values["RiskScore"])
        }
    })

    // Phase 5: MCDM — close options, verify sensitivity
    t.Run("phase 5 mcdm close ranking", func(t *testing.T) {
        t.Parallel()

        // Options designed to be close in score
        matrix := [][]float64{
            {5, 30, 200},  // Option A: medium rate, long term, medium amount
            {4, 15, 150},  // Option B: low rate, short term, low amount
            {6, 20, 250},  // Option C: high rate, medium term, high amount
        }
        criteria := []topsis.Criterion{
            {Weight: 0.5, Benefit: false}, // rate: cost
            {Weight: 0.3, Benefit: true},  // term: benefit
            {Weight: 0.2, Benefit: true},  // amount: benefit
        }

        result, err := topsis.Rank(matrix, criteria)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        // All scores must be in [0,1]
        for i, s := range result.Scores {
            if s < 0 || s > 1 {
                t.Fatalf("option %d score %f outside [0,1]", i, s)
            }
        }

        // With rate as dominant criterion (0.5 weight, cost), Option B (rate=4) should rank highest
        if result.Scores[1] < result.Scores[0] || result.Scores[1] < result.Scores[2] {
            t.Fatalf("Option B (lowest rate) should rank highest: A=%f, B=%f, C=%f",
                result.Scores[0], result.Scores[1], result.Scores[2])
        }
    })
}
```

---

## File 2: golden_files_test.go

Golden files freeze exact outputs. Failure message pattern: `"behavioral regression: ... was X, now Y"`

### Test: TestGolden_BayesianRainPosterior

4 subtests. Frozen values hand-calculated in Appendix B (B.1, B.5, B.6). Tolerance: goldenBayesian (1e-4).

```go
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

        // Explaining away: observing Sprinkler=true reduces P(Rain|WG) — see Appendix B.5
        ev := evidence.NewEvidenceBase()
        ev.Observe("WetGrass", "true")
        ev.Observe("Sprinkler", "true")
        eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
        result := eng.Query(bn, ev, "Rain")

        got := float64(result.Posterior["true"])
        if math.Abs(got-0.21569) > goldenBayesian {
            t.Fatalf("behavioral regression: P(Rain=true|WetGrass=true,Sprinkler=true) was 0.21569, now %f", got)
        }
    })

    t.Run("P(Sprinkler=true | WetGrass=true) marginal", func(t *testing.T) {
        t.Parallel()

        // Marginal query on a different variable — see Appendix B.6
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

        // Classic medical screening: low prior, high sensitivity, moderate false positive.
        // P(D=present) = 0.001, sensitivity = 0.99, false_positive = 0.05.
        // Bayes: P(D|T+) = 0.99*0.001 / (0.99*0.001 + 0.05*0.999) = 0.00099 / 0.05094 ≈ 0.01943
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
```

### Test: TestGolden_DeductiveBusinessScenario

No subtests. Rules: premium-customer (high_spend AND loyal → premium), discount-eligible (premium → discount), notify (discount → notify).
Initial: {high_spend, loyal}. Verify: all 5 facts, Steps=3, Trace non-nil with len=3, derived provenance.

```go
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

    // Trace structure verification confirms no behavioral regression in inference path recording.
    if result.Trace == nil {
        t.Fatal("behavioral regression: result.Trace is nil, expected non-nil trace")
    }
    if len(result.Trace) != 3 {
        t.Fatalf("behavioral regression: expected 3 trace steps (one per rule fired), got %d", len(result.Trace))
    }

    // Verify provenance: derived facts should be marked as Derived in the factbase.
    derivedFacts := []logic.Var{"premium", "discount", "notify"}
    for _, fact := range derivedFacts {
        if !result.Facts.IsDerived(fact) {
            t.Fatalf("behavioral regression: fact %q should be marked as Derived", fact)
        }
    }

    // Initial facts should NOT be marked as derived.
    initialFacts := []logic.Var{"high_spend", "loyal"}
    for _, fact := range initialFacts {
        if result.Facts.IsDerived(fact) {
            t.Fatalf("behavioral regression: initial fact %q should not be marked as Derived", fact)
        }
    }
}
```

### Test: TestGolden_FuzzyTipping

5 subtests:
- 3 Mamdani/Centroid cases: (1,1), (5,5), (9,9) — verify tip in [0,25] range
- "mid inputs Mamdani/Bisector" — rebuild engine with Bisector defuzz
- "mid inputs Mamdani/MeanOfMax" — rebuild engine with MeanOfMax
- "high inputs Sugeno" — rebuild engine with WithMethod(Sugeno)

Note: [TO_FREEZE] markers remain. The test currently validates range only. Once values are frozen, replace range check with exact comparison using goldenFuzzy tolerance.

For the Bisector/MeanOfMax/Sugeno subtests, the engine must be fully rebuilt inline (not using makeTippingEngine) because different options are needed.

```go
func TestGolden_FuzzyTipping(t *testing.T) {
    t.Parallel()

    eng := makeTippingEngine() // Mamdani with Centroid defuzzification

    cases := []struct {
        name    string
        food    float64
        service float64
        // [TO_FREEZE: execute and capture exact tip values]
    }{
        {"low inputs", 1.0, 1.0},
        {"mid inputs", 5.0, 5.0},
        {"high inputs", 9.0, 9.0},
    }

    for _, tc := range cases {
        t.Run(tc.name+" Mamdani/Centroid", func(t *testing.T) {
            t.Parallel()

            result := eng.Infer(map[string]float64{"food": tc.food, "service": tc.service})
            tip := result.Outputs["tip"]

            // Verify output is in valid range [0, 25]
            if tip < 0 || tip > 25 {
                t.Fatalf("behavioral regression: tip=%f outside [0,25] for food=%f, service=%f", tip, tc.food, tc.service)
            }

            // [TO_FREEZE: replace range check with exact frozen value comparison]
            // if math.Abs(tip - frozenValue) > goldenFuzzy {
            //     t.Fatalf("behavioral regression: tip was %f, now %f", frozenValue, tip)
            // }
        })
    }

    t.Run("mid inputs Mamdani/Bisector", func(t *testing.T) {
        t.Parallel()

        // Rebuild tipping engine with Bisector defuzzification.
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
            fuzzyRules.NewRule("r1", []fuzzyRules.Condition{{Variable: "food", Term: "bad"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r2", []fuzzyRules.Condition{{Variable: "service", Term: "poor"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r3", []fuzzyRules.Condition{{Variable: "food", Term: "average"}}, fuzzyRules.Consequent{Variable: "tip", Term: "medium"}),
            fuzzyRules.NewRule("r4", []fuzzyRules.Condition{{Variable: "food", Term: "good"}, {Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
            fuzzyRules.NewRule("r5", []fuzzyRules.Condition{{Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
        }
        bisectorEng := fuzzyEngine.NewEngine(
            []variable.Variable{food, service},
            []variable.Variable{tip},
            ruleSet,
            fuzzyEngine.WithDefuzzify(fuzzym.Bisector),
        )

        result := bisectorEng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
        tipVal := result.Outputs["tip"]

        // [TO_FREEZE: capture exact bisector value and replace range check]
        if tipVal < 0 || tipVal > 25 {
            t.Fatalf("behavioral regression: Bisector tip=%f outside [0,25]", tipVal)
        }
    })

    t.Run("mid inputs Mamdani/MeanOfMax", func(t *testing.T) {
        t.Parallel()

        // Rebuild tipping engine with MeanOfMax defuzzification.
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
            fuzzyRules.NewRule("r1", []fuzzyRules.Condition{{Variable: "food", Term: "bad"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r2", []fuzzyRules.Condition{{Variable: "service", Term: "poor"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r3", []fuzzyRules.Condition{{Variable: "food", Term: "average"}}, fuzzyRules.Consequent{Variable: "tip", Term: "medium"}),
            fuzzyRules.NewRule("r4", []fuzzyRules.Condition{{Variable: "food", Term: "good"}, {Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
            fuzzyRules.NewRule("r5", []fuzzyRules.Condition{{Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
        }
        momEng := fuzzyEngine.NewEngine(
            []variable.Variable{food, service},
            []variable.Variable{tip},
            ruleSet,
            fuzzyEngine.WithDefuzzify(fuzzym.MeanOfMax),
        )

        result := momEng.Infer(map[string]float64{"food": 5.0, "service": 5.0})
        tipVal := result.Outputs["tip"]

        // [TO_FREEZE: capture exact MeanOfMax value and replace range check]
        if tipVal < 0 || tipVal > 25 {
            t.Fatalf("behavioral regression: MeanOfMax tip=%f outside [0,25]", tipVal)
        }
    })

    t.Run("high inputs Sugeno", func(t *testing.T) {
        t.Parallel()

        // Rebuild tipping engine with Sugeno method.
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
            fuzzyRules.NewRule("r1", []fuzzyRules.Condition{{Variable: "food", Term: "bad"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r2", []fuzzyRules.Condition{{Variable: "service", Term: "poor"}}, fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r3", []fuzzyRules.Condition{{Variable: "food", Term: "average"}}, fuzzyRules.Consequent{Variable: "tip", Term: "medium"}),
            fuzzyRules.NewRule("r4", []fuzzyRules.Condition{{Variable: "food", Term: "good"}, {Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
            fuzzyRules.NewRule("r5", []fuzzyRules.Condition{{Variable: "service", Term: "excellent"}}, fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
        }
        sugenoEng := fuzzyEngine.NewEngine(
            []variable.Variable{food, service},
            []variable.Variable{tip},
            ruleSet,
            fuzzyEngine.WithMethod(fuzzyEngine.Sugeno),
        )

        result := sugenoEng.Infer(map[string]float64{"food": 9.0, "service": 9.0})
        tipVal := result.Outputs["tip"]

        // [TO_FREEZE: capture exact Sugeno value and replace range check]
        if tipVal < 0 || tipVal > 25 {
            t.Fatalf("behavioral regression: Sugeno tip=%f outside [0,25]", tipVal)
        }
    })
}
```

### Test: TestGolden_CausalLinearSCM

4 subtests. Frozen values: exact (deterministic arithmetic). Tolerance: floatTolerance (1e-9).

SCM: X→Z(=2X)→Y(=Z+3)

```go
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

        // Factual world: X=5, Z=2*5=10, Y=10+3=13.
        // Counterfactual do(Z=7): X keeps factual value 5 (not a descendant of Z),
        // Z=7 (intervened), Y=Z+3=7+3=10.
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
```

### Test: TestGolden_AHPConsistentMatrix

No subtests. Matrix {1,2,6; 0.5,1,3; 1/6,1/3,1}. Frozen weights [0.6,0.3,0.1] tolerance 1e-6. CR=0.0 tolerance 1e-9.

```go
func TestGolden_AHPConsistentMatrix(t *testing.T) {
    t.Parallel()

    matrix := ahp.PairwiseMatrix{
        {1, 2, 6},
        {0.5, 1, 3},
        {1.0 / 6, 1.0 / 3, 1},
    }
    result, _ := ahp.Analyze(matrix)

    frozen := []float64{0.6, 0.3, 0.1}
    for i, w := range frozen {
        if math.Abs(result.Weights[i]-w) > 1e-6 {
            t.Fatalf("behavioral regression: AHP weight[%d] was %f, now %f", i, w, result.Weights[i])
        }
    }

    // ConsistencyRatio must be exactly 0.0 for a perfectly consistent matrix.
    if math.Abs(result.ConsistencyRatio-0.0) > 1e-9 {
        t.Fatalf("behavioral regression: AHP CR was 0.0, now %f", result.ConsistencyRatio)
    }
}
```

### Test: TestGolden_TOPSISStrictDominance

3 subtests. A=[10,10,10], B=[1,1,1], all benefit equal weight.
Score(A)=1.0, Score(B)=0.0, Ranking[0]=0, Ranking[1]=1.

Note: The TOPSIS test uses `Beneficial` field (not `Benefit`) and `Weight: 1.0/3.0` for each criterion.

```go
func TestGolden_TOPSISStrictDominance(t *testing.T) {
    t.Parallel()

    // Alternative A dominates B on all criteria.
    // A=[10,10,10], B=[1,1,1], all benefit, equal weight.
    alternatives := [][]float64{
        {10, 10, 10},
        {1, 1, 1},
    }
    criteria := []topsis.Criterion{
        {Weight: 1.0 / 3.0, Beneficial: true},
        {Weight: 1.0 / 3.0, Beneficial: true},
        {Weight: 1.0 / 3.0, Beneficial: true},
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

    t.Run("ranking order", func(t *testing.T) {
        t.Parallel()

        if result.Ranking[0] != 0 {
            t.Fatalf("behavioral regression: TOPSIS best alternative was index 0, now %d", result.Ranking[0])
        }
        if result.Ranking[1] != 1 {
            t.Fatalf("behavioral regression: TOPSIS worst alternative was index 1, now %d", result.Ranking[1])
        }
    })
}
```

---

## Appendix B: Manual Derivations (include as comments)

B.1: P(Rain=t|WG=t) = 0.16038/0.44838 ≈ 0.35770
B.2: P(Default=yes) = 0.0084+0.028+0.027+0.048 = 0.1114
B.3: Causal Linear SCM: Propagate X=5→Z=10→Y=13, Intervene do(Z=7)→Y=10, CF do(X=10)→Y=23, CF do(Z=7)|X=5→Y=10
B.4: AHP weights [0.6,0.3,0.1], CR=0.0
B.5: P(Rain=t|WG=t,Spr=t) = 0.198/0.918 ≈ 0.21569
B.6: P(Spr=t|WG=t) = 0.28998/0.44838 ≈ 0.64680

## [TO_FREEZE] Resolution

There are 7 [TO_FREEZE] markers in TestGolden_FuzzyTipping. These require running the actual engine to capture exact outputs:
1. Leave range-check tests as-is for now
2. Add a comment `// [TO_FREEZE]` where exact values will go
3. After the test module compiles, run the engine manually to freeze values

## Verification
```
cd modules/compute/tests/acceptance
go vet ./...
go test -run "TestAcceptance_LoanApproval|TestGolden" -count=1 -v ./...
```
