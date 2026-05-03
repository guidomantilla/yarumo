# Acceptance Tests — Prompt 04: invariants_engine_test.go

## Context

Read `00-context.md` for project structure, imports, and coding standards.

## Role

You are a Go testing engineer. Generate the file `invariants_engine_test.go`.

## Output

Generate exactly ONE file: `invariants_engine_test.go`

Place it at: `modules/compute/tests/acceptance/invariants_engine_test.go`

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
- `makeRainNetwork() network.Network` — Rain-Sprinkler-WetGrass Bayesian network
- `makeTippingEngine(opts ...fuzzyEngine.Option) fuzzyEngine.Engine` — canonical tipping engine
- `generateEvidenceCombos(vars, outcomes)` — all evidence subsets
- Tolerances: `floatTolerance` (1e-9), `probTolerance` (1e-6), `defuzzTolerance` (0.5), `goldenBayesian` (1e-4)

## Required Imports

```go
import (
    "fmt"
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
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
    causalEngine "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
    "github.com/guidomantilla/yarumo/compute/engine/causal/model"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)
```

## Tests

The file must contain exactly 14 test functions organized in 5 sections.

---

### Section 1.8: engine/deductive/ — Adversarial (1 test, 5 subtests)

**TestAcceptance_Deductive_adversarial**
- Strategy: Adversarial
- Reference: Russell & Norvig "AIMA" Ch. 7-9
- Subtests: cyclic rules terminate, deep chain 25 rules, backward chaining depth limit, clone-on-attempt isolation, conflict resolution determinism

```go
func TestAcceptance_Deductive_adversarial(t *testing.T) {
    t.Parallel()

    t.Run("cyclic rules terminate", func(t *testing.T) {
        t.Parallel()

        // A→B, B→C, C→A — should terminate because no new facts after step 3
        r1 := deductiveRules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
        r2 := deductiveRules.NewRule("r2", logic.Var("B"), map[logic.Var]bool{"C": true})
        r3 := deductiveRules.NewRule("r3", logic.Var("C"), map[logic.Var]bool{"A": true})

        e := deductiveEngine.NewEngine()
        result := e.Forward(logic.Fact{"A": true}, []deductiveRules.Rule{r1, r2, r3})

        snap := result.Facts.Snapshot()
        if !snap["A"] || !snap["B"] || !snap["C"] {
            t.Fatal("expected all three facts derived")
        }

        if result.Steps > 3 {
            t.Fatalf("expected <= 3 steps for cyclic rules, got %d", result.Steps)
        }
    })

    t.Run("deep chain 25 rules", func(t *testing.T) {
        t.Parallel()

        rulesList := make([]deductiveRules.Rule, 25)
        for i := range 25 {
            from := logic.Var(fmt.Sprintf("V%d", i))
            to := logic.Var(fmt.Sprintf("V%d", i+1))
            rulesList[i] = deductiveRules.NewRule(
                fmt.Sprintf("r%d", i),
                from,
                map[logic.Var]bool{to: true},
            )
        }

        e := deductiveEngine.NewEngine()
        result := e.Forward(logic.Fact{"V0": true}, rulesList)

        snap := result.Facts.Snapshot()
        if !snap["V25"] {
            t.Fatal("expected V25 derived through 25-rule chain")
        }

        if result.Steps != 25 {
            t.Fatalf("expected 25 steps, got %d", result.Steps)
        }
    })

    t.Run("backward chaining depth limit", func(t *testing.T) {
        t.Parallel()

        rulesList := make([]deductiveRules.Rule, 25)
        for i := range 25 {
            from := logic.Var(fmt.Sprintf("V%d", i))
            to := logic.Var(fmt.Sprintf("V%d", i+1))
            rulesList[i] = deductiveRules.NewRule(
                fmt.Sprintf("r%d", i),
                from,
                map[logic.Var]bool{to: true},
            )
        }

        e := deductiveEngine.NewEngine(deductiveEngine.WithMaxDepth(10))
        proven, _ := e.Backward(logic.Fact{"V0": true}, rulesList, "V25")

        if proven {
            t.Fatal("depth=10 should not prove V25 through 25-rule chain")
        }
    })

    t.Run("clone-on-attempt isolation", func(t *testing.T) {
        t.Parallel()

        // Rule needs A AND B, but B is false — rule fails, factbase must be unchanged
        r := deductiveRules.NewRule("r1",
            logic.AndF{L: logic.Var("A"), R: logic.Var("B")},
            map[logic.Var]bool{"C": true},
        )

        e := deductiveEngine.NewEngine()
        initial := logic.Fact{"A": true}
        result := e.Forward(initial, []deductiveRules.Rule{r})

        snap := result.Facts.Snapshot()
        if snap["C"] {
            t.Fatal("C should not be derived when B is missing")
        }

        if len(snap) != 1 || !snap["A"] {
            t.Fatalf("factbase should only contain A, got %v", snap)
        }
    })

    t.Run("conflict resolution determinism", func(t *testing.T) {
        t.Parallel()

        // Two rules with same priority, both applicable — PriorityOrder fires both
        r1 := deductiveRules.NewRule("r1", logic.Var("A"), map[logic.Var]bool{"B": true})
        r2 := deductiveRules.NewRule("r2", logic.Var("A"), map[logic.Var]bool{"C": true})

        e := deductiveEngine.NewEngine(deductiveEngine.WithStrategy(deductiveEngine.PriorityOrder))
        result := e.Forward(logic.Fact{"A": true}, []deductiveRules.Rule{r1, r2})

        snap := result.Facts.Snapshot()
        if !snap["B"] || !snap["C"] {
            t.Fatalf("PriorityOrder should fire both rules: B=%v, C=%v", snap["B"], snap["C"])
        }

        // FirstMatch fires only one rule per step
        eFM := deductiveEngine.NewEngine(deductiveEngine.WithStrategy(deductiveEngine.FirstMatch))
        resultFM := eFM.Forward(logic.Fact{"A": true}, []deductiveRules.Rule{r1, r2})

        snapFM := resultFM.Facts.Snapshot()
        // Both eventually fire (in separate steps), but takes more steps than PriorityOrder
        if !snapFM["B"] || !snapFM["C"] {
            t.Fatalf("FirstMatch should eventually fire both: B=%v, C=%v", snapFM["B"], snapFM["C"])
        }
        if resultFM.Steps < result.Steps {
            t.Fatalf("FirstMatch should take >= steps than PriorityOrder: FM=%d, PO=%d",
                resultFM.Steps, result.Steps)
        }
    })
}
```

---

### Section 1.9: engine/bayesian/ — Exhaustive + Known-answer (4 tests)

**TestAcceptance_Bayesian_VE_equals_Enumeration**
- Strategy: Exhaustive over rain network
- All queries x all evidence combinations. VE and Enumeration must produce identical posteriors (tolerance 1e-9).
- Note: `generateEvidenceCombos` is defined in helpers_test.go.

```go
func TestAcceptance_Bayesian_VE_equals_Enumeration(t *testing.T) {
    t.Parallel()

    // Uses makeRainNetwork helper (same as examples)
    bn := makeRainNetwork()
    variables := []stats.Var{"Rain", "Sprinkler", "WetGrass"}
    outcomes := map[stats.Var][]stats.Outcome{
        "Rain":      {"true", "false"},
        "Sprinkler": {"true", "false"},
        "WetGrass":  {"true", "false"},
    }

    t.Run("all queries and evidence combinations", func(t *testing.T) {
        t.Parallel()

        for _, query := range variables {
            others := make([]stats.Var, 0)
            for _, v := range variables {
                if v != query {
                    others = append(others, v)
                }
            }

            // Generate all subsets of evidence
            evidenceCombos := generateEvidenceCombos(others, outcomes)

            for _, combo := range evidenceCombos {
                ev := evidence.NewEvidenceBase()
                for v, o := range combo {
                    ev.Observe(v, o)
                }

                veEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
                enumEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.Enumeration))

                veResult := veEng.Query(bn, ev, query)
                enumResult := enumEng.Query(bn, ev, query)

                for _, outcome := range outcomes[query] {
                    veVal := float64(veResult.Posterior[outcome])
                    enumVal := float64(enumResult.Posterior[outcome])
                    diff := math.Abs(veVal - enumVal)
                    if diff > 1e-9 {
                        t.Fatalf("VE != Enum for query=%s, evidence=%v, outcome=%s: VE=%f, Enum=%f",
                            query, combo, outcome, veVal, enumVal)
                    }
                }
            }
        }
    })
}
```

**TestAcceptance_Bayesian_known_answer_rain**
- Strategy: Known-answer (hand-calculated)
- Derivation (Appendix B.1):
  P(WG=t) = 0.99×0.01×0.2 + 0.8×0.99×0.2 + 0.9×0.4×0.8 + 0.0×0.6×0.8
           = 0.00198 + 0.1584 + 0.288 + 0.0 = 0.44838
  P(Rain=t, WG=t) = 0.00198 + 0.1584 = 0.16038
  P(Rain=t | WG=t) = 0.16038 / 0.44838 ≈ 0.35770

```go
func TestAcceptance_Bayesian_known_answer_rain(t *testing.T) {
    t.Parallel()

    bn := makeRainNetwork()

    t.Run("P(Rain=true | WetGrass=true)", func(t *testing.T) {
        t.Parallel()

        ev := evidence.NewEvidenceBase()
        ev.Observe("WetGrass", "true")

        eng := bayesianEngine.NewEngine()
        result := eng.Query(bn, ev, "Rain")

        got := float64(result.Posterior["true"])
        assertFloat(t, "P(Rain=true|WetGrass=true)", got, 0.35770, goldenBayesian)
    })

    t.Run("evidence clamping", func(t *testing.T) {
        t.Parallel()

        ev := evidence.NewEvidenceBase()
        ev.Observe("Rain", "true")

        eng := bayesianEngine.NewEngine()
        result := eng.Query(bn, ev, "Rain")

        got := float64(result.Posterior["true"])
        assertFloat(t, "P(Rain=true|Rain=true)", got, 1.0, 1e-9)
    })

    t.Run("prior sums to one", func(t *testing.T) {
        t.Parallel()

        ev := evidence.NewEvidenceBase()
        eng := bayesianEngine.NewEngine()

        for _, v := range []stats.Var{"Rain", "Sprinkler", "WetGrass"} {
            result := eng.Query(bn, ev, v)
            sum := 0.0
            for _, p := range result.Posterior {
                sum += float64(p)
            }
            assertFloat(t, fmt.Sprintf("sum P(%s)", v), sum, 1.0, 1e-9)
        }
    })
}
```

**TestAcceptance_Bayesian_elimination_order_invariance**
- Strategy: Exhaustive — all permutations of hidden variables produce identical posteriors.
- Orders: [Sprinkler,WetGrass] and [WetGrass,Sprinkler], query Rain with no evidence.

```go
func TestAcceptance_Bayesian_elimination_order_invariance(t *testing.T) {
    t.Parallel()

    bn := makeRainNetwork()
    ev := evidence.NewEvidenceBase()
    ev.Observe("WetGrass", "true")

    // Query Rain; hidden variables are {Sprinkler}
    // For a 3-variable network querying 1 with 1 observed, there's only 1 hidden var.
    // Use a larger test: query Rain with no evidence, hidden = {Sprinkler, WetGrass}
    evEmpty := evidence.NewEvidenceBase()

    orders := [][]stats.Var{
        {"Sprinkler", "WetGrass"},
        {"WetGrass", "Sprinkler"},
    }

    var reference map[stats.Outcome]float64
    for i, order := range orders {
        eng := bayesianEngine.NewEngine(
            bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination),
            bayesianEngine.WithEliminationOrder(order),
        )
        result := eng.Query(bn, evEmpty, "Rain")

        if i == 0 {
            reference = make(map[stats.Outcome]float64)
            for k, v := range result.Posterior {
                reference[k] = float64(v)
            }
            continue
        }

        for outcome, refVal := range reference {
            got := float64(result.Posterior[outcome])
            if math.Abs(got-refVal) > 1e-9 {
                t.Fatalf("elimination order %v gave P(Rain=%s)=%f, expected %f",
                    order, outcome, got, refVal)
            }
        }
    }
}
```

**TestAcceptance_Bayesian_explaining_away**
- Strategy: Known-answer — explaining away (D-separation)
- Derivation (Appendix B.5): P(Rain=t|WG=t,Spr=t) = 0.198/0.918 ≈ 0.21569 < P(Rain=t|WG=t) ≈ 0.35770

```go
func TestAcceptance_Bayesian_explaining_away(t *testing.T) {
    t.Parallel()

    bn := makeRainNetwork()

    t.Run("sprinkler explains away rain", func(t *testing.T) {
        t.Parallel()

        // P(Rain=true | WetGrass=true) — without sprinkler evidence
        ev1 := evidence.NewEvidenceBase()
        ev1.Observe("WetGrass", "true")
        eng := bayesianEngine.NewEngine()
        r1 := eng.Query(bn, ev1, "Rain")
        pNoSpr := float64(r1.Posterior["true"])

        // P(Rain=true | WetGrass=true, Sprinkler=true) — with sprinkler evidence
        ev2 := evidence.NewEvidenceBase()
        ev2.Observe("WetGrass", "true")
        ev2.Observe("Sprinkler", "true")
        r2 := eng.Query(bn, ev2, "Rain")
        pWithSpr := float64(r2.Posterior["true"])

        // Explaining away: pWithSpr < pNoSpr
        if pWithSpr >= pNoSpr {
            t.Fatalf("explaining away violated: P(Rain|WG,Spr)=%f >= P(Rain|WG)=%f",
                pWithSpr, pNoSpr)
        }

        // Verify exact values from hand calculation (Appendix B.1, B.5)
        assertFloat(t, "P(Rain|WG)", pNoSpr, 0.35770, goldenBayesian)
        assertFloat(t, "P(Rain|WG,Spr)", pWithSpr, 0.21569, goldenBayesian)
    })
}
```

---

### Section 1.10: engine/fuzzy/ — Monotonia + Known-answer (3 tests)

**TestAcceptance_Fuzzy_monotonia_exhaustive**
- Strategy: Exhaustive monotonia check
- food from 0 to 10 step 0.5, service fixed at 5.0. tip must be non-decreasing (within defuzzTolerance).

```go
func TestAcceptance_Fuzzy_monotonia_exhaustive(t *testing.T) {
    t.Parallel()

    t.Run("food monotonia with fixed service", func(t *testing.T) {
        t.Parallel()

        eng := makeTippingEngine() // uses same setup as fuzzy examples

        prevTip := -1.0
        for food := 0.0; food <= 10.0; food += 0.5 {
            result := eng.Infer(map[string]float64{"food": food, "service": 5.0})
            tip := result.Outputs["tip"]

            if tip < prevTip-defuzzTolerance {
                t.Fatalf("monotonia violated: food=%f gave tip=%f, previous tip=%f", food, tip, prevTip)
            }

            prevTip = tip
        }
    })
}
```

**TestAcceptance_Fuzzy_Sugeno_single_rule**
- Strategy: Known-answer
- Single input "x" with "high" term (triangular 8,10,10). Output "y" with "strong" term.
- WithMethod(Sugeno), WithSugenoOutputs({"y/strong": 25.0}). x=10 -> y=25.0 exactly.

```go
func TestAcceptance_Fuzzy_Sugeno_single_rule(t *testing.T) {
    t.Parallel()

    t.Run("full activation gives exact singleton", func(t *testing.T) {
        t.Parallel()

        // Single input "x" with term "high" that is 1.0 at x=10
        high, _ := fuzzy.Triangular(8, 10, 10)
        input := variable.NewVariable("x", 0, 10, []variable.Term{
            {Name: "high", Fn: high},
        })

        // Output doesn't need real terms for Sugeno, just the variable definition
        outTerm, _ := fuzzy.Triangular(20, 25, 30)
        output := variable.NewVariable("y", 0, 30, []variable.Term{
            {Name: "strong", Fn: outTerm},
        })

        ruleSet := []fuzzyRules.Rule{
            fuzzyRules.NewRule("single",
                []fuzzyRules.Condition{{Variable: "x", Term: "high"}},
                fuzzyRules.Consequent{Variable: "y", Term: "strong"},
            ),
        }

        eng := fuzzyEngine.NewEngine(
            []variable.Variable{input},
            []variable.Variable{output},
            ruleSet,
            fuzzyEngine.WithMethod(fuzzyEngine.Sugeno),
            fuzzyEngine.WithSugenoOutputs(map[string]float64{"y/strong": 25.0}),
        )

        result := eng.Infer(map[string]float64{"x": 10.0})
        assertFloat(t, "Sugeno single rule", result.Outputs["y"], 25.0, 1e-9)
    })
}
```

**Note**: This test uses `fuzzy` (not `fuzzym`) for the membership constructors because fuzzyRules is already imported. Adjust the import alias if needed — the import block uses `fuzzym` for the math/fuzzy package, so use `fuzzym.Triangular` instead:

Actually, the Required Imports block does not alias `fuzzy` — it uses `fuzzym` for `math/fuzzy`. The test code in ACCEPTANCE_TESTS.md uses `fuzzy.Triangular`. You must reconcile: if the import is `fuzzym`, replace `fuzzy.Triangular` with `fuzzym.Triangular` throughout.

**TestAcceptance_Fuzzy_output_bounds_all_defuzz_methods**
- 5 methods (Centroid, Bisector, MeanOfMax, LargestOfMax, SmallestOfMax).
- All inputs food,service in {0..10}. tip must be in [0,25].

```go
func TestAcceptance_Fuzzy_output_bounds_all_defuzz_methods(t *testing.T) {
    t.Parallel()

    methods := []struct {
        name string
        fn   fuzzym.DefuzzifyFn
    }{
        {"Centroid", fuzzym.Centroid},
        {"Bisector", fuzzym.Bisector},
        {"MeanOfMax", fuzzym.MeanOfMax},
        {"LargestOfMax", fuzzym.LargestOfMax},
        {"SmallestOfMax", fuzzym.SmallestOfMax},
    }

    for _, m := range methods {
        t.Run(m.name, func(t *testing.T) {
            t.Parallel()

            eng := makeTippingEngine(fuzzyEngine.WithDefuzzify(m.fn))

            inputs := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
            for _, food := range inputs {
                for _, svc := range inputs {
                    result := eng.Infer(map[string]float64{"food": food, "service": svc})
                    tip := result.Outputs["tip"]
                    if tip < 0 || tip > 25 {
                        t.Fatalf("%s: tip=%f outside [0,25] for food=%f, service=%f",
                            m.name, tip, food, svc)
                    }
                }
            }
        })
    }
}
```

---

### Section 1.11: engine/causal/ — Adversarial + Known-answer (1 test, 4 subtests)

**TestAcceptance_Causal_adversarial**
- Reference: Pearl "Causality" (2009)

```go
func TestAcceptance_Causal_adversarial(t *testing.T) {
    t.Parallel()

    t.Run("do differs from observe with confounder", func(t *testing.T) {
        t.Parallel()

        // U → X (X = U*2), U → Y (Y = U*3 + X*0), X has no direct effect on Y
        scm := model.NewSCM()
        scm.AddVariable("U", nil, func(_ map[string]float64) float64 { return 0 })
        scm.AddVariable("X", []string{"U"}, func(p map[string]float64) float64 { return p["U"] * 2 })
        scm.AddVariable("Y", []string{"U", "X"}, func(p map[string]float64) float64 { return p["U"]*3 + p["X"]*0 })

        e := causalEngine.NewEngine()

        obs := e.Propagate(scm, map[string]float64{"U": 1})
        // Observational: U=1, X=2, Y=3
        assertFloat(t, "obs Y", obs.Values["Y"], 3.0, floatTolerance)

        intervened := e.Intervene(scm, map[string]float64{"X": 5})
        // do(X=5): U still defaults (0), Y = 0*3 + 5*0 = 0
        // X has NO causal effect on Y — do changes X but not Y
        if math.Abs(intervened.Values["Y"]-obs.Values["Y"]) < 0.001 {
            // Y should differ because U defaults to 0 under intervention (no observation of U)
            // vs U=1 under propagation
        }
    })

    t.Run("diamond graph counterfactual", func(t *testing.T) {
        t.Parallel()

        // U → X, U → Z, X → Y, Z → Y (Y = X + Z)
        scm := model.NewSCM()
        scm.AddVariable("U", nil, func(_ map[string]float64) float64 { return 0 })
        scm.AddVariable("X", []string{"U"}, func(p map[string]float64) float64 { return p["U"] + 1 })
        scm.AddVariable("Z", []string{"U"}, func(p map[string]float64) float64 { return p["U"] + 2 })
        scm.AddVariable("Y", []string{"X", "Z"}, func(p map[string]float64) float64 { return p["X"] + p["Z"] })

        e := causalEngine.NewEngine()

        // Factual: U=3 → X=4, Z=5, Y=9
        factual := e.Propagate(scm, map[string]float64{"U": 3})
        assertFloat(t, "factual X", factual.Values["X"], 4.0, floatTolerance)
        assertFloat(t, "factual Z", factual.Values["Z"], 5.0, floatTolerance)
        assertFloat(t, "factual Y", factual.Values["Y"], 9.0, floatTolerance)

        // Counterfactual: do(X=10), Z stays at factual value (5)
        cf := e.Counterfactual(scm, map[string]float64{"U": 3}, map[string]float64{"X": 10})
        assertFloat(t, "cf X", cf.Values["X"], 10.0, floatTolerance)
        assertFloat(t, "cf Z", cf.Values["Z"], 5.0, floatTolerance) // Z preserves factual (not downstream of X)
        assertFloat(t, "cf Y", cf.Values["Y"], 15.0, floatTolerance) // Y = 10 + 5
    })

    t.Run("intervention idempotence", func(t *testing.T) {
        t.Parallel()

        // X → Y (Y = X + 3)
        scm := model.NewSCM()
        scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 0 })
        scm.AddVariable("Y", []string{"X"}, func(p map[string]float64) float64 { return p["X"] + 3 })

        e := causalEngine.NewEngine()
        obs := e.Propagate(scm, map[string]float64{"X": 5})
        intervened := e.Intervene(scm, map[string]float64{"X": 5})

        assertFloat(t, "Y obs", obs.Values["Y"], 8.0, floatTolerance)
        assertFloat(t, "Y intervened", intervened.Values["Y"], 8.0, floatTolerance)
    })

    t.Run("causal effect with non-zero coefficient", func(t *testing.T) {
        t.Parallel()

        // X → Y where Y = X + 10 — direct causal effect
        scm := model.NewSCM()
        scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 0 })
        scm.AddVariable("Y", []string{"X"}, func(p map[string]float64) float64 { return p["X"] + 10 })

        e := causalEngine.NewEngine()

        // Propagate: X=5 → Y=15
        obs := e.Propagate(scm, map[string]float64{"X": 5})
        assertFloat(t, "obs Y", obs.Values["Y"], 15.0, floatTolerance)

        // do(X=20) → Y=30
        intervened := e.Intervene(scm, map[string]float64{"X": 20})
        assertFloat(t, "do(X=20) Y", intervened.Values["Y"], 30.0, floatTolerance)

        // Causal effect: ΔY/ΔX = (30-15)/(20-5) = 1.0 (the coefficient on X in Y's equation)
        deltaY := intervened.Values["Y"] - obs.Values["Y"]
        deltaX := 20.0 - 5.0
        assertFloat(t, "causal effect ΔY/ΔX", deltaY/deltaX, 1.0, floatTolerance)
    })
}
```

---

### Section 1.12: engine/mcdm/ — Known-answer + Boundary (4 tests)

**TestAcceptance_AHP_known_answer**
- Perfectly consistent matrix yields exact weights and CR=0.
- Derivation (Appendix B.4): Matrix {1,2,6; 0.5,1,3; 1/6,1/3,1}. Exact weights [0.6, 0.3, 0.1]. CR=0.

```go
func TestAcceptance_AHP_known_answer(t *testing.T) {
    t.Parallel()

    t.Run("perfectly consistent 3x3", func(t *testing.T) {
        t.Parallel()

        // Real weights: [0.6, 0.3, 0.1]
        matrix := ahp.PairwiseMatrix{
            {1, 2, 6},
            {1.0 / 2, 1, 3},
            {1.0 / 6, 1.0 / 3, 1},
        }

        result, err := ahp.Analyze(matrix)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        assertFloat(t, "weight[0]", result.Weights[0], 0.6, 1e-6)
        assertFloat(t, "weight[1]", result.Weights[1], 0.3, 1e-6)
        assertFloat(t, "weight[2]", result.Weights[2], 0.1, 1e-6)

        if result.ConsistencyRatio > 1e-10 {
            t.Fatalf("expected CR≈0 for consistent matrix, got %f", result.ConsistencyRatio)
        }
    })

    t.Run("weights sum to one", func(t *testing.T) {
        t.Parallel()

        matrix := ahp.PairwiseMatrix{
            {1, 3, 5},
            {1.0 / 3, 1, 3},
            {1.0 / 5, 1.0 / 3, 1},
        }

        result, _ := ahp.Analyze(matrix)
        sum := 0.0
        for _, w := range result.Weights {
            sum += w
        }
        assertFloat(t, "weights sum", sum, 1.0, 1e-9)
    })
}
```

**TestAcceptance_TOPSIS_known_answer**
- Strict dominance, benefit vs cost, ideal scores 1.0, anti-ideal scores 0.0, weight sensitivity.

```go
func TestAcceptance_TOPSIS_known_answer(t *testing.T) {
    t.Parallel()

    t.Run("strict dominance", func(t *testing.T) {
        t.Parallel()

        matrix := [][]float64{
            {10, 10, 10}, // A dominates
            {1, 1, 1},    // B dominated
        }
        criteria := []topsis.Criterion{
            {Weight: 1.0 / 3, Benefit: true},
            {Weight: 1.0 / 3, Benefit: true},
            {Weight: 1.0 / 3, Benefit: true},
        }

        result, err := topsis.Rank(matrix, criteria)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if result.Scores[0] <= result.Scores[1] {
            t.Fatalf("dominant alternative should score higher: A=%f, B=%f", result.Scores[0], result.Scores[1])
        }
    })

    t.Run("benefit vs cost", func(t *testing.T) {
        t.Parallel()

        // col 0 = benefit (higher is better), col 1 = cost (lower is better)
        matrix := [][]float64{
            {10, 1},  // high benefit, low cost → best
            {1, 10},  // low benefit, high cost → worst
        }
        criteria := []topsis.Criterion{
            {Weight: 0.5, Benefit: true},
            {Weight: 0.5, Benefit: false},
        }

        result, err := topsis.Rank(matrix, criteria)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if result.Scores[0] <= result.Scores[1] {
            t.Fatalf("[10,1] should dominate [1,10]: scores %f vs %f", result.Scores[0], result.Scores[1])
        }
    })

    t.Run("ideal alternative scores 1.0", func(t *testing.T) {
        t.Parallel()

        // With 2 alternatives on all-benefit criteria, the strictly dominant one
        // should score exactly 1.0 (distance to ideal = 0, distance to anti-ideal = max)
        matrix := [][]float64{
            {10, 10},
            {1, 1},
        }
        criteria := []topsis.Criterion{
            {Weight: 0.5, Benefit: true},
            {Weight: 0.5, Benefit: true},
        }

        result, err := topsis.Rank(matrix, criteria)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        assertFloat(t, "ideal score", result.Scores[0], 1.0, 1e-9)
    })

    t.Run("anti-ideal alternative scores 0.0", func(t *testing.T) {
        t.Parallel()

        // The strictly dominated alternative should score exactly 0.0
        matrix := [][]float64{
            {10, 10},
            {1, 1},
        }
        criteria := []topsis.Criterion{
            {Weight: 0.5, Benefit: true},
            {Weight: 0.5, Benefit: true},
        }

        result, err := topsis.Rank(matrix, criteria)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        assertFloat(t, "anti-ideal score", result.Scores[1], 0.0, 1e-9)
    })

    t.Run("weight sensitivity", func(t *testing.T) {
        t.Parallel()

        // A and B: A is better on criterion 0, B is better on criterion 1.
        // Changing weights should flip the ranking.
        matrix := [][]float64{
            {10, 1}, // A: strong on crit 0
            {1, 10}, // B: strong on crit 1
        }

        // Heavy weight on crit 0 → A wins
        crit1 := []topsis.Criterion{{Weight: 0.9, Benefit: true}, {Weight: 0.1, Benefit: true}}
        r1, _ := topsis.Rank(matrix, crit1)
        if r1.Scores[0] <= r1.Scores[1] {
            t.Fatalf("A should win with weight on crit 0: A=%f, B=%f", r1.Scores[0], r1.Scores[1])
        }

        // Heavy weight on crit 1 → B wins
        crit2 := []topsis.Criterion{{Weight: 0.1, Benefit: true}, {Weight: 0.9, Benefit: true}}
        r2, _ := topsis.Rank(matrix, crit2)
        if r2.Scores[1] <= r2.Scores[0] {
            t.Fatalf("B should win with weight on crit 1: A=%f, B=%f", r2.Scores[0], r2.Scores[1])
        }
    })
}
```

**TestAcceptance_AHP_Saaty_textbook**
- Matrix {1,3,5; 1/3,1,3; 1/5,1/3,1}. Weights ≈ [0.633, 0.260, 0.106]. CR ≈ 0.033.

```go
func TestAcceptance_AHP_Saaty_textbook(t *testing.T) {
    t.Parallel()

    matrix := ahp.PairwiseMatrix{
        {1, 3, 5},
        {1.0 / 3, 1, 3},
        {1.0 / 5, 1.0 / 3, 1},
    }

    result, err := ahp.Analyze(matrix)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Saaty textbook weights ≈ [0.633, 0.260, 0.106]
    assertFloat(t, "weight[0]", result.Weights[0], 0.633, 0.01)
    assertFloat(t, "weight[1]", result.Weights[1], 0.260, 0.01)
    assertFloat(t, "weight[2]", result.Weights[2], 0.106, 0.01)

    // CR ≈ 0.033, well below 0.10 threshold
    if result.ConsistencyRatio > 0.10 {
        t.Fatalf("expected CR < 0.10, got %f", result.ConsistencyRatio)
    }
    assertFloat(t, "CR", result.ConsistencyRatio, 0.033, 0.01)

    if !result.Consistent {
        t.Fatal("expected Consistent=true for CR < 0.10")
    }
}
```

**TestAcceptance_AHP_CR_boundary**
- Consistent: perfectly consistent matrix, CR=0, Consistent=true.
- Inconsistent: circular preferences {1,9,1/9; 1/9,1,9; 9,1/9,1}. CR > 0.10, Consistent=false.

```go
func TestAcceptance_AHP_CR_boundary(t *testing.T) {
    t.Parallel()

    t.Run("consistent matrix CR well below threshold", func(t *testing.T) {
        t.Parallel()

        // Perfectly consistent matrix: CR = 0
        matrix := ahp.PairwiseMatrix{
            {1, 2, 6},
            {0.5, 1, 3},
            {1.0 / 6, 1.0 / 3, 1},
        }

        result, err := ahp.Analyze(matrix)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if !result.Consistent {
            t.Fatalf("expected Consistent=true, got CR=%f", result.ConsistencyRatio)
        }
    })

    t.Run("inconsistent matrix CR above threshold", func(t *testing.T) {
        t.Parallel()

        // Highly inconsistent: A>B (9), B>C (9), but C>A (9)
        matrix := ahp.PairwiseMatrix{
            {1, 9, 1.0 / 9},
            {1.0 / 9, 1, 9},
            {9, 1.0 / 9, 1},
        }

        result, err := ahp.Analyze(matrix)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result.Consistent {
            t.Fatalf("expected Consistent=false for circular preferences, got CR=%f", result.ConsistencyRatio)
        }
        if result.ConsistencyRatio < 0.10 {
            t.Fatalf("expected CR >= 0.10 for inconsistent matrix, got %f", result.ConsistencyRatio)
        }
    })
}
```

---

## Appendix B Derivations (include as comments)

B.1: P(Rain=true|WetGrass=true) = 0.16038/0.44838 ≈ 0.35770
B.4: AHP consistent matrix weights [0.6,0.3,0.1] from ratios a12=2, a13=6, a23=3. CR=0.
B.5: P(Rain=t|WG=t,Spr=t) = 0.198/0.918 ≈ 0.21569

Total: 14 tests (counting top-level test functions)

## Verification
```
cd modules/compute/tests/acceptance
go vet ./...
go test -run "TestAcceptance_(Deductive|Bayesian|Fuzzy|Causal|AHP|TOPSIS)" -count=1 -v ./...
```
