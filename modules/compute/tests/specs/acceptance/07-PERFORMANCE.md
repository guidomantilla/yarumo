# Acceptance Tests -- Prompt 07: performance_test.go

## Context
Read `00-context.md` for project structure, imports, and coding standards.

## Role
You are a Go testing engineer. Generate the file `performance_test.go`.

## Output
Generate exactly ONE file: `performance_test.go`
Place it at: `modules/compute/tests/acceptance/performance_test.go`

## Constraints
- Package: `package acceptance_test`
- No testify -- use `t.Fatal`/`t.Fatalf`
- No table-driven tests -- individual `t.Run` subtests
- `t.Parallel()` on every test and subtest
- Only import public APIs
- No inline assignments

## Helper References
From `helpers_test.go`:
- `makeRainNetwork() network.Network` -- for VE benchmark

## Required Imports
```go
import (
    "fmt"
    "testing"
    "time"

    "github.com/guidomantilla/yarumo/compute/math/logic"
    "github.com/guidomantilla/yarumo/compute/math/logic/sat"
    fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/math/stats"

    "github.com/guidomantilla/yarumo/compute/engine/bayesian"
    bayesianEngine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
    deductiveEngine "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
    deductiveRules "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
    fuzzyEngine "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
    fuzzyRules "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
)
```

## Design Pattern

Performance tests use termination timeouts, NOT absolute times. The pattern:
```go
done := make(chan struct{})
go func() {
    defer close(done)
    // ... work ...
}()
select {
case <-done:
case <-time.After(N * time.Second):
    t.Fatal("performance regression: ... did not terminate within Ns")
}
```

## Tests

### Section 5.1: Termination Under Pressure (4 tests)

**TestPerformance_DPLL_terminates**
- 1 subtest: "15 variable cyclic implication chain"
- 15 variables V0..V14
- Build formula: V0 AND (V0=>V1) AND (V1=>V2) AND ... AND (V13=>V14) AND (V14=>V0)
- Convert to CNF via `logic.ToCNF`, then to sat.CNF via `sat.FromFormula`, solve with `sat.Solve`
- Timeout: 5 seconds

```go
vars := make([]logic.Var, 15)
for i := range vars {
    vars[i] = logic.Var(fmt.Sprintf("V%d", i))
}

var f logic.Formula = vars[0]
for i := 0; i < 15; i++ {
    next := vars[(i+1)%15]
    f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: next}}
}

done := make(chan struct{})
go func() {
    defer close(done)
    cnf := sat.FromFormula(logic.ToCNF(f))
    sat.Solve(cnf)
}()
select {
case <-done:
case <-time.After(5 * time.Second):
    t.Fatal("performance regression: DPLL on 15-var cycle did not terminate within 5s")
}
```

**TestPerformance_ForwardChaining_200rules_terminates**
- 1 subtest: "200 sequential rules"
- 200 rules: V0->V1, V1->V2, ..., V199->V200
- Each rule: condition is the source Var, conclusion sets the target Var to true
- Initial facts: V0=true
- Uses `deductiveEngine.NewEngine()` and `Forward(initialFacts, ruleSet)`
- Timeout: 5 seconds

```go
ruleSet := make([]deductiveRules.Rule, 200)
for i := 0; i < 200; i++ {
    src := logic.Var(fmt.Sprintf("V%d", i))
    dst := logic.Var(fmt.Sprintf("V%d", i+1))
    ruleSet[i] = deductiveRules.NewRule(
        fmt.Sprintf("r%d", i),
        src,
        map[logic.Var]bool{dst: true},
    )
}
initial := logic.Fact{logic.Var("V0"): true}
eng := deductiveEngine.NewEngine()

done := make(chan struct{})
go func() {
    defer close(done)
    eng.Forward(initial, ruleSet)
}()
select {
case <-done:
case <-time.After(5 * time.Second):
    t.Fatal("performance regression: forward chaining 200 rules did not terminate within 5s")
}
```

**TestPerformance_ForwardChaining_cyclic_terminates**
- 1 subtest: "50 cyclic rules"
- 50 rules: V0->V1, V1->V2, ..., V49->V0
- Initial facts: V0=true
- Timeout: 5 seconds

```go
ruleSet := make([]deductiveRules.Rule, 50)
for i := 0; i < 50; i++ {
    src := logic.Var(fmt.Sprintf("V%d", i))
    dst := logic.Var(fmt.Sprintf("V%d", (i+1)%50))
    ruleSet[i] = deductiveRules.NewRule(
        fmt.Sprintf("r%d", i),
        src,
        map[logic.Var]bool{dst: true},
    )
}
initial := logic.Fact{logic.Var("V0"): true}
eng := deductiveEngine.NewEngine()

done := make(chan struct{})
go func() {
    defer close(done)
    eng.Forward(initial, ruleSet)
}()
select {
case <-done:
case <-time.After(5 * time.Second):
    t.Fatal("performance regression: forward chaining 50 cyclic rules did not terminate within 5s")
}
```

**TestPerformance_Bayesian_VE_terminates**
- 1 subtest: "10 node chain"
- 10-node chain: V0->V1->...->V9, each binary (outcomes: "t", "f")
- Root V0: P(t)=0.5, P(f)=0.5
- Each child Vi: P(t|parent=t)=0.9, P(f|parent=t)=0.1, P(t|parent=f)=0.3, P(f|parent=f)=0.7
- Evidence: V9="t", Query: "V0"
- Uses `bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))`
- Timeout: 10 seconds

```go
bn := network.NewNetwork()

rootCPT := bayesian.NewCPT("V0", nil)
rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
bn.AddNode(network.Node{
    Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
})

for i := 1; i < 10; i++ {
    name := stats.Var(fmt.Sprintf("V%d", i))
    parent := stats.Var(fmt.Sprintf("V%d", i-1))
    cpt := bayesian.NewCPT(name, []stats.Var{parent})
    cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
    cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
    bn.AddNode(network.Node{
        Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
        Outcomes: []stats.Outcome{"t", "f"},
    })
}

ev := evidence.NewEvidenceBase()
ev.Observe("V9", "t")
eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

done := make(chan struct{})
go func() {
    defer close(done)
    eng.Query(bn, ev, "V0")
}()
select {
case <-done:
case <-time.After(10 * time.Second):
    t.Fatal("performance regression: VE on 10-node chain did not terminate within 10s")
}
```

### Section 5.2: Scaling (3 benchmarks + 1 test)

**BenchmarkScaling_ForwardChaining**
- Sizes: 10, 100
- Pattern: `b.Run(fmt.Sprintf("%d_rules", n), ...)`
- Build n sequential rules V0->V1->...->Vn, initial V0=true
- Uses `b.ResetTimer()` after setup

```go
func BenchmarkScaling_ForwardChaining(b *testing.B) {
    for _, n := range []int{10, 100} {
        b.Run(fmt.Sprintf("%d_rules", n), func(b *testing.B) {
            ruleSet := make([]deductiveRules.Rule, n)
            for i := 0; i < n; i++ {
                src := logic.Var(fmt.Sprintf("V%d", i))
                dst := logic.Var(fmt.Sprintf("V%d", i+1))
                ruleSet[i] = deductiveRules.NewRule(
                    fmt.Sprintf("r%d", i),
                    src,
                    map[logic.Var]bool{dst: true},
                )
            }
            initial := logic.Fact{logic.Var("V0"): true}
            eng := deductiveEngine.NewEngine()

            b.ResetTimer()
            for b.Loop() {
                eng.Forward(initial, ruleSet)
            }
        })
    }
}
```

**BenchmarkScaling_DPLL**
- Sizes: 5, 10
- Chain formula: x0 AND (x0=>x1) AND (x1=>x2) AND ...
- Uses `b.ResetTimer()` after setup

```go
func BenchmarkScaling_DPLL(b *testing.B) {
    for _, n := range []int{5, 10} {
        b.Run(fmt.Sprintf("%d_vars", n), func(b *testing.B) {
            vars := make([]logic.Var, n)
            for i := range vars {
                vars[i] = logic.Var(fmt.Sprintf("x%d", i))
            }

            var f logic.Formula = vars[0]
            for i := 0; i < n-1; i++ {
                f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
            }
            cnf := sat.FromFormula(logic.ToCNF(f))

            b.ResetTimer()
            for b.Loop() {
                sat.Solve(cnf)
            }
        })
    }
}
```

**BenchmarkScaling_VE**
- Sizes: 3, 5
- Chain network V0->V1->...->V(n-1), evidence on last variable, query V0
- Same CPT structure as termination test

```go
func BenchmarkScaling_VE(b *testing.B) {
    for _, n := range []int{3, 5} {
        b.Run(fmt.Sprintf("%d_vars", n), func(b *testing.B) {
            bn := network.NewNetwork()

            rootCPT := bayesian.NewCPT("V0", nil)
            rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
            bn.AddNode(network.Node{
                Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
            })

            for i := 1; i < n; i++ {
                name := stats.Var(fmt.Sprintf("V%d", i))
                parent := stats.Var(fmt.Sprintf("V%d", i-1))
                cpt := bayesian.NewCPT(name, []stats.Var{parent})
                cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
                cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
                bn.AddNode(network.Node{
                    Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
                    Outcomes: []stats.Outcome{"t", "f"},
                })
            }

            lastVar := stats.Var(fmt.Sprintf("V%d", n-1))
            ev := evidence.NewEvidenceBase()
            ev.Observe(lastVar, "t")
            eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))

            b.ResetTimer()
            for b.Loop() {
                eng.Query(bn, ev, "V0")
            }
        })
    }
}
```

**TestPerformance_Scaling_ratios**
- 3 subtests using `testing.Benchmark`

```go
func TestPerformance_Scaling_ratios(t *testing.T) {
    t.Parallel()

    t.Run("forward chaining ratio", func(t *testing.T) {
        t.Parallel()

        small := testing.Benchmark(func(b *testing.B) {
            ruleSet := make([]deductiveRules.Rule, 10)
            for i := 0; i < 10; i++ {
                src := logic.Var(fmt.Sprintf("V%d", i))
                dst := logic.Var(fmt.Sprintf("V%d", i+1))
                ruleSet[i] = deductiveRules.NewRule(
                    fmt.Sprintf("r%d", i), src, map[logic.Var]bool{dst: true},
                )
            }
            initial := logic.Fact{logic.Var("V0"): true}
            eng := deductiveEngine.NewEngine()
            b.ResetTimer()
            for b.Loop() {
                eng.Forward(initial, ruleSet)
            }
        })

        large := testing.Benchmark(func(b *testing.B) {
            ruleSet := make([]deductiveRules.Rule, 100)
            for i := 0; i < 100; i++ {
                src := logic.Var(fmt.Sprintf("V%d", i))
                dst := logic.Var(fmt.Sprintf("V%d", i+1))
                ruleSet[i] = deductiveRules.NewRule(
                    fmt.Sprintf("r%d", i), src, map[logic.Var]bool{dst: true},
                )
            }
            initial := logic.Fact{logic.Var("V0"): true}
            eng := deductiveEngine.NewEngine()
            b.ResetTimer()
            for b.Loop() {
                eng.Forward(initial, ruleSet)
            }
        })

        ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
        if ratio > 150 {
            t.Fatalf("forward chaining scaling ratio too high: %.1f (want < 150)", ratio)
        }
    })

    t.Run("DPLL ratio", func(t *testing.T) {
        t.Parallel()

        small := testing.Benchmark(func(b *testing.B) {
            vars := make([]logic.Var, 5)
            for i := range vars {
                vars[i] = logic.Var(fmt.Sprintf("x%d", i))
            }
            var f logic.Formula = vars[0]
            for i := 0; i < 4; i++ {
                f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
            }
            cnf := sat.FromFormula(logic.ToCNF(f))
            b.ResetTimer()
            for b.Loop() {
                sat.Solve(cnf)
            }
        })

        large := testing.Benchmark(func(b *testing.B) {
            vars := make([]logic.Var, 10)
            for i := range vars {
                vars[i] = logic.Var(fmt.Sprintf("x%d", i))
            }
            var f logic.Formula = vars[0]
            for i := 0; i < 9; i++ {
                f = logic.AndF{L: f, R: logic.ImplF{L: vars[i], R: vars[i+1]}}
            }
            cnf := sat.FromFormula(logic.ToCNF(f))
            b.ResetTimer()
            for b.Loop() {
                sat.Solve(cnf)
            }
        })

        ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
        if ratio > 100 {
            t.Fatalf("DPLL scaling ratio too high: %.1f (want < 100)", ratio)
        }
    })

    t.Run("VE ratio", func(t *testing.T) {
        t.Parallel()

        buildChainNetwork := func(n int) (network.Network, evidence.EvidenceBase) {
            bn := network.NewNetwork()

            rootCPT := bayesian.NewCPT("V0", nil)
            rootCPT.Set(stats.Assignment{}, stats.Distribution{"t": 0.5, "f": 0.5})
            bn.AddNode(network.Node{
                Variable: "V0", CPT: rootCPT, Outcomes: []stats.Outcome{"t", "f"},
            })

            for i := 1; i < n; i++ {
                name := stats.Var(fmt.Sprintf("V%d", i))
                parent := stats.Var(fmt.Sprintf("V%d", i-1))
                cpt := bayesian.NewCPT(name, []stats.Var{parent})
                cpt.Set(stats.Assignment{parent: "t"}, stats.Distribution{"t": 0.9, "f": 0.1})
                cpt.Set(stats.Assignment{parent: "f"}, stats.Distribution{"t": 0.3, "f": 0.7})
                bn.AddNode(network.Node{
                    Variable: name, Parents: []stats.Var{parent}, CPT: cpt,
                    Outcomes: []stats.Outcome{"t", "f"},
                })
            }

            lastVar := stats.Var(fmt.Sprintf("V%d", n-1))
            ev := evidence.NewEvidenceBase()
            ev.Observe(lastVar, "t")

            return bn, ev
        }

        small := testing.Benchmark(func(b *testing.B) {
            bn, ev := buildChainNetwork(3)
            eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
            b.ResetTimer()
            for b.Loop() {
                eng.Query(bn, ev, "V0")
            }
        })

        large := testing.Benchmark(func(b *testing.B) {
            bn, ev := buildChainNetwork(5)
            eng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
            b.ResetTimer()
            for b.Loop() {
                eng.Query(bn, ev, "V0")
            }
        })

        ratio := float64(large.NsPerOp()) / float64(small.NsPerOp())
        if ratio > 100 {
            t.Fatalf("VE scaling ratio too high: %.1f (want < 100)", ratio)
        }
    })
}
```

### Section 5.3: Algorithm Comparison (2 tests)

**TestPerformance_VE_vs_Enumeration**
- 1 subtest: "both produce same result and neither is 10x slower"
- Uses `makeRainNetwork()`, evidence WetGrass="true", query "Rain"
- 100 iterations each, measure elapsed time
- Verify posteriors match within `probTolerance`
- Ratio must be between 0.1 and 10

```go
func TestPerformance_VE_vs_Enumeration(t *testing.T) {
    t.Parallel()

    t.Run("both produce same result and neither is 10x slower", func(t *testing.T) {
        t.Parallel()

        bn := makeRainNetwork()
        ev := evidence.NewEvidenceBase()
        ev.Observe("WetGrass", "true")

        veEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.VariableElimination))
        enumEng := bayesianEngine.NewEngine(bayesianEngine.WithAlgorithm(bayesianEngine.Enumeration))

        veStart := time.Now()
        var veResult bayesianEngine.Result
        for i := 0; i < 100; i++ {
            veResult = veEng.Query(bn, ev, "Rain")
        }
        veElapsed := time.Since(veStart)

        enumStart := time.Now()
        var enumResult bayesianEngine.Result
        for i := 0; i < 100; i++ {
            enumResult = enumEng.Query(bn, ev, "Rain")
        }
        enumElapsed := time.Since(enumStart)

        // Verify posteriors match.
        for outcome, veProb := range veResult.Posterior {
            enumProb := enumResult.Posterior[outcome]
            diff := veProb - enumProb
            if diff < 0 {
                diff = -diff
            }
            if diff > probTolerance {
                t.Fatalf("posterior mismatch for %s: VE=%.9f, Enum=%.9f", outcome, veProb, enumProb)
            }
        }

        // Verify neither is 10x slower.
        ratio := float64(veElapsed) / float64(enumElapsed)
        if ratio < 0.1 || ratio > 10 {
            t.Fatalf("VE/Enum time ratio out of bounds: %.2f (want 0.1..10)", ratio)
        }
    })
}
```

**TestPerformance_Mamdani_vs_Sugeno**
- 1 subtest: "neither method is 10x slower than the other"
- Simple 2-rule engine built inline (NOT using makeTippingEngine):
  - food variable: bad (trapezoidal 0,0,2,4), good (trapezoidal 6,8,10,10)
  - tip variable: low (trapezoidal 0,0,5,10), high (trapezoidal 15,20,25,25)
  - 2 rules: food bad -> tip low, food good -> tip high
  - Mamdani engine: `fuzzyEngine.NewEngine(inputs, outputs, ruleSet)`
  - Sugeno engine: `fuzzyEngine.NewEngine(inputs, outputs, ruleSet, fuzzyEngine.WithMethod(fuzzyEngine.Sugeno), fuzzyEngine.WithSugenoOutputs(map[string]float64{"tip/low": 5.0, "tip/high": 20.0}))`
- Input: food=7.0
- 100 iterations each, ratio must be between 0.1 and 10

```go
func TestPerformance_Mamdani_vs_Sugeno(t *testing.T) {
    t.Parallel()

    t.Run("neither method is 10x slower than the other", func(t *testing.T) {
        t.Parallel()

        bad, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
        good, _ := fuzzym.Trapezoidal(6, 8, 10, 10)
        food := variable.NewVariable("food", 0, 10, []variable.Term{
            {Name: "bad", Fn: bad}, {Name: "good", Fn: good},
        })

        lowTip, _ := fuzzym.Trapezoidal(0, 0, 5, 10)
        highTip, _ := fuzzym.Trapezoidal(15, 20, 25, 25)
        tip := variable.NewVariable("tip", 0, 25, []variable.Term{
            {Name: "low", Fn: lowTip}, {Name: "high", Fn: highTip},
        })

        ruleSet := []fuzzyRules.Rule{
            fuzzyRules.NewRule("r1",
                []fuzzyRules.Condition{{Variable: "food", Term: "bad"}},
                fuzzyRules.Consequent{Variable: "tip", Term: "low"}),
            fuzzyRules.NewRule("r2",
                []fuzzyRules.Condition{{Variable: "food", Term: "good"}},
                fuzzyRules.Consequent{Variable: "tip", Term: "high"}),
        }

        inputs := []variable.Variable{food}
        outputs := []variable.Variable{tip}

        mamdaniEng := fuzzyEngine.NewEngine(inputs, outputs, ruleSet)
        sugenoEng := fuzzyEngine.NewEngine(inputs, outputs, ruleSet,
            fuzzyEngine.WithMethod(fuzzyEngine.Sugeno),
            fuzzyEngine.WithSugenoOutputs(map[string]float64{
                "tip/low":  5.0,
                "tip/high": 20.0,
            }),
        )

        input := map[string]float64{"food": 7.0}

        mamdaniStart := time.Now()
        for i := 0; i < 100; i++ {
            mamdaniEng.Infer(input)
        }
        mamdaniElapsed := time.Since(mamdaniStart)

        sugenoStart := time.Now()
        for i := 0; i < 100; i++ {
            sugenoEng.Infer(input)
        }
        sugenoElapsed := time.Since(sugenoStart)

        ratio := float64(mamdaniElapsed) / float64(sugenoElapsed)
        if ratio < 0.1 || ratio > 10 {
            t.Fatalf("Mamdani/Sugeno time ratio out of bounds: %.2f (want 0.1..10)", ratio)
        }
    })
}
```

## Verification
```
cd modules/compute/tests/acceptance
go vet ./...
go test -run TestPerformance -count=1 -v ./...
go test -bench=BenchmarkScaling -count=1 ./...
```
