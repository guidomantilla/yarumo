# Acceptance Tests — Shared Context

This file provides shared context for all acceptance test prompts (01-07).
It does NOT generate any code.

## Project Structure

```
modules/compute/
├── math/          ← Go module (go.mod)
│   ├── logic/
│   │   ├── sat/
│   │   ├── entailment/
│   │   ├── predicate/
│   │   └── temporal/
│   ├── fuzzy/
│   ├── sets/
│   └── stats/
└── engine/        ← Go module (depends on math/ and common/)
    ├── deductive/
    │   └── engine/, rules/
    ├── bayesian/
    │   └── engine/, evidence/, network/
    ├── fuzzy/
    │   └── engine/, rules/, variable/
    ├── causal/
    │   └── engine/, model/
    └── mcdm/
        ├── ahp/
        └── topsis/
```

## Test Module

Acceptance tests live in a separate Go module:

```
modules/compute/tests/acceptance/
├── go.mod
├── helpers_test.go             ← prompt 01
├── invariants_logic_test.go    ← prompt 02
├── invariants_numeric_test.go  ← prompt 03
├── invariants_engine_test.go   ← prompt 04
├── golden_scenario_test.go     ← prompt 05
├── golden_files_test.go        ← prompt 05
├── error_contracts_test.go     ← prompt 06
└── performance_test.go         ← prompt 07
```

### go.mod

```
module github.com/guidomantilla/yarumo/compute/tests/acceptance

go 1.25.5

require (
    github.com/guidomantilla/yarumo/compute/math v0.0.0
    github.com/guidomantilla/yarumo/compute/engine v0.0.0
)

replace (
    github.com/guidomantilla/yarumo/compute/math => ../../math
    github.com/guidomantilla/yarumo/compute/engine => ../../engine
    github.com/guidomantilla/yarumo/core/common => ../../../common
)
```

### Package

```go
package acceptance_test
```

All files use external test package (black-box). Only public APIs are imported.

## Import Aliases

```go
import (
    // stdlib
    "fmt"
    "math"
    "testing"
    "time"

    // math/
    "github.com/guidomantilla/yarumo/compute/math/logic"
    "github.com/guidomantilla/yarumo/compute/math/logic/entailment"
    "github.com/guidomantilla/yarumo/compute/math/logic/predicate"
    "github.com/guidomantilla/yarumo/compute/math/logic/sat"
    "github.com/guidomantilla/yarumo/compute/math/logic/temporal"
    "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    fuzzym "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/math/stats"

    // engine/
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
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)
```

## Coding Standards

- No testify — use `t.Fatal` / `t.Fatalf` only
- No table-driven tests — individual `t.Run` subtests
- `t.Parallel()` on every test function and every subtest
- Descriptive test names in English
- No inline assignments (`if err := fn(); err != nil` forbidden)
- Doc comments end with periods

## Tolerances (defined in helpers_test.go)

```go
const (
    floatTolerance  = 1e-9   // general floating-point arithmetic
    probTolerance   = 1e-6   // probability marginalization accumulates error
    defuzzTolerance = 0.5    // defuzzification discretization introduces error
    goldenBayesian  = 1e-4   // golden files for Bayesian posteriors
    goldenFuzzy     = 1e-2   // golden files for fuzzy outputs
)
```

## Existing Tests (DO NOT Duplicate)

| Existing Test | File | Why Not Duplicated |
|--------------|------|-------------------|
| TestProperties_foundational (11 subtests) | logic/examples/properties_test.go | Covers specific identity/De Morgan laws — acceptance tests add EXHAUSTIVE checks |
| TestProperties_identity (11 subtests) | logic/examples/properties_test.go | Covers absorption, complement — acceptance tests verify over full corpus |
| TestProperties_tnorms (9 subtests) | fuzzy/examples/properties_test.go | Spot-checks 3 axioms — acceptance tests verify over FULL 21x21 grid |
| TestProperties_tconorms (7 subtests) | fuzzy/examples/properties_test.go | Same — spot-check vs grid |
| TestEnumerationWithEvidence | bayesian/examples | Tests 1 evidence combo — acceptance tests cover ALL combos |
| TestVEMatchesEnumeration | bayesian/examples | Tests 1 query — acceptance tests cover all variables x evidence |
| TestEngine_Propagate_basic | causal/engine | Tests linear SCM — acceptance tests add confounders and diamond |
| TestAnalyze_basic | mcdm/ahp | Tests Saaty matrix — acceptance tests add consistent matrix with exact weights |
| TestRank_basic | mcdm/topsis | Tests 3 alternatives — acceptance tests add dominance and benefit/cost |
| TestMamdaniTipping (3 subtests) | fuzzy/examples | Tests range checks — acceptance tests add monotonia and Sugeno single-rule |

## Helper Functions (defined in helpers_test.go)

Available helpers that other test files can use:

- `generateFormulas(depth int, vars []logic.Var) []logic.Formula` — builds all formulas up to depth
- `fuzzyGrid(step float64) []float64` — generates [0.0, step, ..., 1.0]
- `assertFloat(t *testing.T, name string, got, want, tolerance float64)` — float comparison helper
- `makeRainNetwork() network.Network` — Rain-Sprinkler-WetGrass Bayesian network
- `makeLoanNetwork() network.Network` — CreditHistory-Default-IncomeLevel network
- `makeLoanRiskEngine() fuzzyEngine.Engine` — fuzzy risk assessment engine
- `makeTippingEngine(opts ...fuzzyEngine.Option) fuzzyEngine.Engine` — canonical tipping engine
- `generateEvidenceCombos(vars, outcomes)` — all evidence subsets
- Structural checkers: `isCNF`, `isDNF`, `isClause`, etc.
