# Compute Module

Mathematical primitives and reasoning engines for decision and process systems.
Two layers: math/ provides pure mathematical foundations (propositional,
predicate, and temporal logic; graphs, statistics, fuzzy, sets, FSM, Markov
chains); engine/ composes those primitives into complete inference paradigms
(deductive, bayesian, fuzzy, causal, MCDM) with mandatory traceability. Future
engines (steps, states, montecarlo, mining) cover the process dimension.

Two Go modules:

- **math/** — pure mathematical primitives. Zero external dependencies beyond
  `common/` error pattern.
- **engine/** — reasoning engines that compose math/ primitives into complete
  inference pipelines. Each paradigm produces audit-ready traces.

```
engine/
  deductive/   bayesian/   fuzzy/   causal/   mcdm/
       \           |          |        |        /
        +---------+-----------+--------+-------+
                          math/
    logic/  fuzzy/  sets/  stats/  graph/  fsm/  markov/
```

## Quick Start

### Deductive — Forward Chaining

```go
import (
    "github.com/guidomantilla/yarumo/compute/math/logic"
    "github.com/guidomantilla/yarumo/compute/engine/deductive/engine"
    "github.com/guidomantilla/yarumo/compute/engine/deductive/rules"
)

r1 := rules.NewRule("rain-wet",
    logic.Var("rain"),
    map[logic.Var]bool{"wet_ground": true},
)
r2 := rules.NewRule("wet-slippery",
    logic.Var("wet_ground"),
    map[logic.Var]bool{"slippery": true},
)

e := engine.NewEngine()
result := e.Forward(logic.Fact{"rain": true}, []rules.Rule{r1, r2})

snap := result.Facts.Snapshot()
// snap["rain"] == true, snap["wet_ground"] == true, snap["slippery"] == true
// result.Steps == 2
```

### Bayesian — Network Inference

```go
import (
    "github.com/guidomantilla/yarumo/compute/math/stats"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
)

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

// ... define WetGrass node similarly ...

ev := evidence.NewEvidenceBase()
ev.Observe("WetGrass", "true")

eng := engine.NewEngine()
result := eng.Query(bn, ev, "Rain")
// result.Posterior["true"] > 0.2 (posterior probability of rain given wet grass)
```

### Fuzzy — Mamdani Tipping

```go
import (
    fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/engine"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/rules"
    "github.com/guidomantilla/yarumo/compute/engine/fuzzy/variable"
)

bad, _ := fuzzym.Trapezoidal(0, 0, 2, 4)
average, _ := fuzzym.Triangular(2, 5, 8)
good, _ := fuzzym.Trapezoidal(6, 8, 10, 10)

food := variable.NewVariable("food", 0, 10, []variable.Term{
    {Name: "bad", Fn: bad}, {Name: "average", Fn: average}, {Name: "good", Fn: good},
})

// ... define service input, tip output, and rules ...

eng := engine.NewEngine(
    []variable.Variable{food, service},
    []variable.Variable{tip},
    ruleSet,
)
result := eng.Infer(map[string]float64{"food": 9.0, "service": 9.0})
// result.Outputs["tip"] > 15 (high tip for good food + excellent service)
```

### Causal — do-Operator

```go
import (
    "github.com/guidomantilla/yarumo/compute/engine/causal/engine"
    "github.com/guidomantilla/yarumo/compute/engine/causal/model"
)

scm := model.NewSCM()
scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 0 })
scm.AddVariable("Z", []string{"X"}, func(p map[string]float64) float64 {
    return p["X"] * 2
})
scm.AddVariable("Y", []string{"Z"}, func(p map[string]float64) float64 {
    return p["Z"] + 3
})

e := engine.NewEngine()

// Level 1: Observation — propagate with X=5
obs := e.Propagate(scm, map[string]float64{"X": 5})
// obs.Values["Z"] == 10, obs.Values["Y"] == 13

// Level 2: Intervention — do(X=10), graph surgery cuts incoming edges to X
intervened := e.Intervene(scm, map[string]float64{"X": 10})
// intervened.Values["Z"] == 20, intervened.Values["Y"] == 23
```

### MCDM — AHP

```go
import "github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"

matrix := [][]float64{
    {1, 3, 5},
    {1.0 / 3, 1, 3},
    {1.0 / 5, 1.0 / 3, 1},
}

result, _ := ahp.Analyze(matrix)
// result.Weights ≈ [0.633, 0.260, 0.106] (Saaty priorities)
// result.ConsistencyRatio < 0.10 (consistent)
```

## Math Foundation

| Package | What it does |
|---------|-------------|
| `logic/` | Propositional logic: Formula, Eval, NNF/CNF/DNF, simplify (18 rules), TruthTable, FailCases |
| `logic/parser/` | Recursive descent parser: `"A & (B \| C)"` → Formula. Unicode + keywords |
| `logic/sat/` | DPLL SAT solver: unit propagation, pure literal elimination |
| `logic/entailment/` | Semantic entailment (A ⊨ B) with countermodel generation |
| `logic/predicate/` | Bounded quantifiers: ForAll, Exists, Count, Filter over finite collections |
| `logic/temporal/` | Bounded temporal assertions (ResponseWithin, FrequencyWithin, Eventually, Before, Elapsed, Sequence) + LTL primitives (Always, Next, Until, Release, Since) |
| `fuzzy/` | Membership functions (triangular, trapezoidal, gaussian, sigmoid, constant), t-norm/t-conorm, defuzzification (Centroid, Bisector, MeanOfMax, LargestOfMax, SmallestOfMax), Fuzzify, Clip, Scale, AggregateMax, Sample |
| `sets/` | Union, Intersection, Difference, SymmetricDifference, IsSubset, IsSuperset, Contains, Equal |
| `stats/` | 17 distributions (Normal, Exponential, Uniform, Beta, Gamma, ChiSquared, StudentT, Lognormal, Weibull, FDist, Poisson, Binomial, Geometric, Gumbel, Hypergeometric, NegativeBinomial, Pareto), hypothesis testing, RunningStats (Welford), WindowedStats, descriptive stats, Bayes theorem |
| `graph/` | Directed, undirected, DAG, bipartite graphs. Centrality, coloring, MST, matching, paths, matrix representation, multigraph, traversal, tree |
| `fsm/` | Finite state machines: transitions, guards, determinism, reachability, dead states, completeness |
| `markov/` | Markov chains: transition, steady state, classification, absorption, mean first passage, step-n, simulate |

Module: `github.com/guidomantilla/yarumo/compute/math`

## Engine Capabilities

| Paradigm | What it does | Configuration |
|----------|-------------|---------------|
| **Deductive** | Forward chaining (fixed-point) + backward chaining (goal-directed). Clone-on-attempt preserves semantics. | Strategy (PriorityOrder/FirstMatch), depth limit, rule priority |
| **Bayesian** | Exact inference on Bayesian networks via variable elimination or full enumeration. CPT validation. | Inference method (VE/Enumeration), elimination order |
| **Fuzzy** | Mamdani (fuzzify → rules → aggregate → defuzzify) and Sugeno (weighted average). | T-norm (Min/Product), defuzzification method, rule weights |
| **Causal** | Pearl's causal hierarchy — Level 1 (association/propagation), Level 2 (intervention/do-operator), counterfactuals via abduction-action-prediction. | Structural equations, graph surgery |
| **MCDM** | AHP (pairwise comparison matrix → eigenvector priorities + consistency ratio) and TOPSIS (vector normalization → ideal solution distance → closeness ranking). | Comparison matrices, benefit/cost criteria |

Every paradigm includes an `explain/` package that produces structured traces for
audit and debugging. Traces are not optional — they are always generated.

Module: `github.com/guidomantilla/yarumo/compute/engine`

## Limitations

| What | Why |
|------|-----|
| No full LTL/CTL model checking | Only bounded temporal operators + 5 LTL primitives. Full model checking is formal verification scope |
| No MCMC / Gibbs sampling | Variable elimination handles networks of 5–30 variables in microseconds |
| No parameter learning (MLE/EM) | Data science / ML scope, not decision-making inference |
| No full first-order logic (FOL) | Requires unification and resolution (Prolog scope). Bounded quantifiers cover finite-domain use cases |
| No type-2 fuzzy sets | Academic, no current consumer |
| SAT solver is DPLL, not CDCL | Sufficient for decision-system formula sizes. CDCL adds clause learning for industrial SAT |
| AHP eigenvector via power iteration | Approximate, not exact eigendecomposition. Sufficient for typical comparison matrices |
| No random sampling from distributions | Nice-to-have for Monte Carlo simulation, not yet implemented |
| Causal: Pearl levels 1–2 + basic counterfactuals | Full level 3 counterfactuals over arbitrary SCMs is doctoral-thesis scope |

## Mathematical Correctness

Every algorithm was formally verified against its reference literature
(Mendelson, Enderton, Davis-Putnam, Koller & Friedman, Pearl, Zadeh, Saaty,
Casella & Berger, Cormen (CLRS), Hopcroft, Norris, among others).

A rigorous analysis covering soundness, completeness, termination, and
edge cases for all algorithms is available in
[CORRECTNESS.md](CORRECTNESS.md).

**Verdict**: all algorithms are mathematically correct within their declared scope.

## Design Decisions

1. **math/ as an independent module** — zero external dependencies (only `common/`
   for error pattern). Can be used standalone without the engines.
2. **Five inference paradigms** — each with `engine/` + `explain/` + domain-specific
   sub-packages. Consistent structural pattern across paradigms.
3. **Explain is not optional** — every paradigm produces structured traces
   (audit trail). This is a first-class design constraint, not an afterthought.
4. **TypedError pattern** — domain errors embed `errs.TypedError` with type
   constants and factory functions across all packages.
5. **Quantifiers live in math/** — `predicate/` and `temporal/` are extensions
   of logic, not inference. They belong in the mathematical foundation.
6. **Causal bounded by design** — Pearl levels 1–2 (association + intervention)
   plus basic counterfactuals. Full level 3 is explicitly out of scope.
7. **Graph primitives in math/** — `graph/`, `fsm/`, and `markov/` are mathematical
   foundations for future process engines (states/, montecarlo/, mining/).

## Package Index

### math/ — 15 packages

| Package | Files | Description |
|---------|-------|-------------|
| `fuzzy/` | 7 | Membership functions (triangular, trapezoidal, gaussian, sigmoid, constant), t-norm/t-conorm, defuzzification, Fuzzify, Clip, Scale, AggregateMax, Sample |
| `fuzzy/examples/` | 3† | Integration tests, benchmarks, property tests |
| `graph/` | 20 | Directed, undirected, DAG, bipartite. Centrality, coloring, MST, matching, paths, matrix, multigraph, traversal, tree |
| `logic/` | 10 | Propositional logic: Formula, Eval, NNF/CNF/DNF, simplify (18 rules), satisfiability, TruthTable, FailCases |
| `logic/entailment/` | 1 | Logical entailment with countermodel |
| `logic/parser/` | 5 | Recursive descent parser: Unicode + keywords, 10 token types, operator precedence |
| `logic/predicate/` | 3 | Bounded quantifiers: ForAll, Exists, Count, Filter over finite collections |
| `logic/sat/` | 4 | DPLL SAT solver: unit propagation, pure literal elimination, CNF conversion |
| `logic/temporal/` | 3 | Bounded temporal assertions + LTL primitives (Always, Next, Until, Release, Since) |
| `logic/examples/` | 3† | Integration tests, benchmarks, property tests |
| `sets/` | 4 | Set operations: Union, Intersection, Difference, SymmetricDifference, IsSubset, IsSuperset, Contains, Equal |
| `stats/` | 27 | 17 distributions, hypothesis testing, RunningStats (Welford), WindowedStats, descriptive stats, Bayes theorem |
| `stats/examples/` | 3† | Integration tests, benchmarks, property tests |
| `fsm/` | 4 | Finite state machines: transitions, guards, determinism, reachability, dead states, completeness |
| `markov/` | 7 | Markov chains: transition, steady state, classification, absorption, mean first passage, step-n, simulate |

### engine/ — 27 packages

| Package | Files | Description |
|---------|-------|-------------|
| `bayesian/` | 5 | Error pattern: BayesianType, Error, sentinels, factories |
| `bayesian/engine/` | 5 | Variable elimination + enumeration inference |
| `bayesian/evidence/` | 1 | Observable fact management with clone support |
| `bayesian/explain/` | 3 | Trace: Initialize, Propagate, Marginalize, Complete |
| `bayesian/network/` | 2 | Bayesian network DAG: nodes, CPT, validation |
| `bayesian/examples/` | 2† | Integration tests, benchmarks |
| `causal/` | 1 | Error pattern: CausalType, Error, sentinels, factory |
| `causal/engine/` | 2 | Propagate (association), Intervene (do-operator), Counterfactual |
| `causal/explain/` | 2 | Trace: Propagation, Intervention, Counterfactual, Attribution, Complete |
| `causal/model/` | 2 | Structural Causal Model (SCM): DAG, structural equations, topological sort |
| `causal/examples/` | 2† | Integration tests, benchmarks |
| `deductive/engine/` | 6 | Forward + backward chaining with clone-on-attempt |
| `deductive/explain/` | 3 | Trace steps for deductive reasoning |
| `deductive/facts/` | 3 | FactBase: Assert, Derive, Retract, Snapshot, Clone |
| `deductive/rules/` | 5 | Rule: Name, Condition (Formula), Conclusion, Fires |
| `deductive/examples/` | 2† | Integration tests, benchmarks |
| `fuzzy/` | 1 | Error pattern: FuzzyType, Error, sentinels, factory |
| `fuzzy/engine/` | 4 | Mamdani + Sugeno inference with configurable options |
| `fuzzy/explain/` | 3 | Trace: Fuzzification, RuleEvaluation, Aggregation, Defuzzification, Complete |
| `fuzzy/rules/` | 3 | Fuzzy rules: antecedent/consequent, rule sets |
| `fuzzy/variable/` | 3 | Linguistic variables with terms and membership functions |
| `fuzzy/examples/` | 2† | Integration tests, benchmarks |
| `mcdm/` | 1 | Error pattern: MCDMType, Error, sentinels, factory |
| `mcdm/ahp/` | 3 | AHP: pairwise comparison, priority weights, consistency ratio, ranking |
| `mcdm/explain/` | 2 | MCDM trace: method, criteria, weights, rankings |
| `mcdm/topsis/` | 3 | TOPSIS: vector normalization, ideal solutions, relative closeness |
| `mcdm/examples/` | 2† | Integration tests, benchmarks |

† File counts for `examples/` packages are test files only (`_test.go`).

**Note**: `deductive/` has no root-level error pattern package unlike the other
four paradigms (bayesian/, causal/, fuzzy/, mcdm/). This is intentional, not a gap.

## Test Coverage

| Module | Packages | Coverage |
|--------|----------|----------|
| math/ | 15 | 100% |
| engine/ | 27 | 95–100% |
| **Total** | **42** | **~98%** |
