# Acceptance Tests -- Prompt 06: error_contracts_test.go

## Context

Read `00-context.md` for project structure, imports, and coding standards.

## Role

You are a Go testing engineer. Generate the file `error_contracts_test.go`.

## Output

Generate exactly ONE file: `error_contracts_test.go`

Place it at: `modules/compute/tests/acceptance/error_contracts_test.go`

## Constraints

- Package: `package acceptance_test`
- No testify -- use `t.Fatal`/`t.Fatalf`
- No table-driven tests -- individual `t.Run` subtests
- `t.Parallel()` on every test and subtest
- Only import public APIs
- No inline assignments

## Helper References

No helpers needed -- error contract tests are self-contained.

## Required Imports

```go
import (
    "testing"

    "github.com/guidomantilla/yarumo/compute/math/fuzzy"
    "github.com/guidomantilla/yarumo/compute/math/logic/predicate"
    "github.com/guidomantilla/yarumo/compute/math/stats"

    "github.com/guidomantilla/yarumo/compute/engine/bayesian"
    "github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
    "github.com/guidomantilla/yarumo/compute/engine/causal/model"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/ahp"
    "github.com/guidomantilla/yarumo/compute/engine/mcdm/topsis"
)
```

## Error Contract Pattern

Each test verifies:
1. Invalid input -> error returned (failure: "error contract violated: ...")
2. Valid input -> no error (failure: "error contract violated: valid ... should not error")

## Tests

The file must contain exactly 10 test functions organized in 6 sections.

---

### Section 4.1: math/fuzzy/ -- Membership Constructors

**TestErrorContract_Fuzzy_Triangular** (4 subtests):
- "a greater than b": Triangular(5,3,7) -> error
- "b greater than c": Triangular(1,5,3) -> error
- "degenerate a equals c": Triangular(5,5,5) -> error
- "valid parameters": Triangular(1,5,9) -> no error

```go
func TestErrorContract_Fuzzy_Triangular(t *testing.T) {
    t.Parallel()

    t.Run("a greater than b", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Triangular(5, 3, 7)
        if err == nil {
            t.Fatal("error contract violated: Triangular(5,3,7) should return error")
        }
    })

    t.Run("b greater than c", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Triangular(1, 5, 3)
        if err == nil {
            t.Fatal("error contract violated: Triangular(1,5,3) should return error")
        }
    })

    t.Run("degenerate a equals c", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Triangular(5, 5, 5)
        if err == nil {
            t.Fatal("error contract violated: Triangular(5,5,5) should return error")
        }
    })

    t.Run("valid parameters", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Triangular(1, 5, 9)
        if err != nil {
            t.Fatalf("error contract violated: valid Triangular should not error: %v", err)
        }
    })
}
```

**TestErrorContract_Fuzzy_Trapezoidal** (3 subtests):
- "b less than a": Trapezoidal(5,3,7,9) -> error
- "d less than a": Trapezoidal(1,2,2,1) -> error
- "valid parameters": Trapezoidal(0,2,8,10) -> no error

```go
func TestErrorContract_Fuzzy_Trapezoidal(t *testing.T) {
    t.Parallel()

    t.Run("b less than a", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Trapezoidal(5, 3, 7, 9)
        if err == nil {
            t.Fatal("error contract violated: Trapezoidal(5,3,7,9) should return error")
        }
    })

    t.Run("d less than a", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Trapezoidal(1, 2, 2, 1)
        if err == nil {
            t.Fatal("error contract violated: Trapezoidal(1,2,2,1) should return error")
        }
    })

    t.Run("valid parameters", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Trapezoidal(0, 2, 8, 10)
        if err != nil {
            t.Fatalf("error contract violated: valid Trapezoidal should not error: %v", err)
        }
    })
}
```

**TestErrorContract_Fuzzy_Gaussian** (2 subtests):
- "zero sigma": Gaussian(0,0) -> error
- "negative sigma": Gaussian(0,-1) -> error

```go
func TestErrorContract_Fuzzy_Gaussian(t *testing.T) {
    t.Parallel()

    t.Run("zero sigma", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Gaussian(0, 0)
        if err == nil {
            t.Fatal("error contract violated: Gaussian(0,0) should return error")
        }
    })

    t.Run("negative sigma", func(t *testing.T) {
        t.Parallel()
        _, err := fuzzy.Gaussian(0, -1)
        if err == nil {
            t.Fatal("error contract violated: Gaussian(0,-1) should return error")
        }
    })
}
```

**TestErrorContract_Fuzzy_Sample** (2 subtests):
- "lo greater than hi": Sample(fn,10,5,100) -> error (fn from valid Triangular)
- "zero samples": Sample(fn,0,10,0) -> error

```go
func TestErrorContract_Fuzzy_Sample(t *testing.T) {
    t.Parallel()

    t.Run("lo greater than hi", func(t *testing.T) {
        t.Parallel()
        fn, _ := fuzzy.Triangular(0, 5, 10)
        _, err := fuzzy.Sample(fn, 10, 5, 100)
        if err == nil {
            t.Fatal("error contract violated: Sample(fn,10,5,100) should return error")
        }
    })

    t.Run("zero samples", func(t *testing.T) {
        t.Parallel()
        fn, _ := fuzzy.Triangular(0, 5, 10)
        _, err := fuzzy.Sample(fn, 0, 10, 0)
        if err == nil {
            t.Fatal("error contract violated: Sample(fn,0,10,0) should return error")
        }
    })
}
```

---

### Section 4.2: math/stats/ -- Distribution Constructors

**TestErrorContract_Stats_Constructors** (18 subtests):

```go
func TestErrorContract_Stats_Constructors(t *testing.T) {
    t.Parallel()

    t.Run("Normal zero sigma", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewNormal(0, 0)
        if err == nil {
            t.Fatal("error contract violated: NewNormal(0,0) should return error")
        }
    })

    t.Run("Normal negative sigma", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewNormal(0, -1)
        if err == nil {
            t.Fatal("error contract violated: NewNormal(0,-1) should return error")
        }
    })

    t.Run("Normal valid", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewNormal(0, 1)
        if err != nil {
            t.Fatalf("error contract violated: valid NewNormal should not error: %v", err)
        }
    })

    t.Run("Exponential zero lambda", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewExponential(0)
        if err == nil {
            t.Fatal("error contract violated: NewExponential(0) should return error")
        }
    })

    t.Run("Exponential negative lambda", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewExponential(-1)
        if err == nil {
            t.Fatal("error contract violated: NewExponential(-1) should return error")
        }
    })

    t.Run("Beta zero alpha", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewBeta(0, 1)
        if err == nil {
            t.Fatal("error contract violated: NewBeta(0,1) should return error")
        }
    })

    t.Run("Beta negative beta", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewBeta(1, -1)
        if err == nil {
            t.Fatal("error contract violated: NewBeta(1,-1) should return error")
        }
    })

    t.Run("Binomial zero n", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewBinomial(0, 0.5)
        if err == nil {
            t.Fatal("error contract violated: NewBinomial(0,0.5) should return error")
        }
    })

    t.Run("Binomial invalid prob", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewBinomial(10, -0.1)
        if err == nil {
            t.Fatal("error contract violated: NewBinomial(10,-0.1) should return error")
        }
    })

    t.Run("Binomial prob > 1", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewBinomial(10, 1.1)
        if err == nil {
            t.Fatal("error contract violated: NewBinomial(10,1.1) should return error")
        }
    })

    t.Run("Gamma zero alpha", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewGamma(0, 1)
        if err == nil {
            t.Fatal("error contract violated: NewGamma(0,1) should return error")
        }
    })

    t.Run("FDist zero d1", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewFDist(0, 5)
        if err == nil {
            t.Fatal("error contract violated: NewFDist(0,5) should return error")
        }
    })

    t.Run("ChiSquared zero df", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewChiSquared(0)
        if err == nil {
            t.Fatal("error contract violated: NewChiSquared(0) should return error")
        }
    })

    t.Run("StudentT zero df", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewStudentT(0)
        if err == nil {
            t.Fatal("error contract violated: NewStudentT(0) should return error")
        }
    })

    t.Run("Poisson zero lambda", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewPoisson(0)
        if err == nil {
            t.Fatal("error contract violated: NewPoisson(0) should return error")
        }
    })

    t.Run("Uniform equal bounds", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewUniform(5, 5)
        if err == nil {
            t.Fatal("error contract violated: NewUniform(5,5) should return error")
        }
    })

    t.Run("Weibull zero k", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewWeibull(0, 1)
        if err == nil {
            t.Fatal("error contract violated: NewWeibull(0,1) should return error")
        }
    })

    t.Run("Lognormal zero sigma", func(t *testing.T) {
        t.Parallel()
        _, err := stats.NewLognormal(0, 0)
        if err == nil {
            t.Fatal("error contract violated: NewLognormal(0,0) should return error")
        }
    })
}
```

---

### Section 4.3: math/logic/predicate/ -- Quantifiers

**TestErrorContract_Predicate_EmptyCollection** (4 subtests):

```go
func TestErrorContract_Predicate_EmptyCollection(t *testing.T) {
    t.Parallel()

    pred := func(_ int) bool { return true }

    t.Run("ForAll empty", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.ForAll([]int{}, pred)
        if err == nil {
            t.Fatal("error contract violated: ForAll on empty collection should return error")
        }
    })

    t.Run("Exists empty", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.Exists([]int{}, pred)
        if err == nil {
            t.Fatal("error contract violated: Exists on empty collection should return error")
        }
    })

    t.Run("Count empty", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.Count([]int{}, pred)
        if err == nil {
            t.Fatal("error contract violated: Count on empty collection should return error")
        }
    })

    t.Run("Filter empty", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.Filter([]int{}, pred)
        if err == nil {
            t.Fatal("error contract violated: Filter on empty collection should return error")
        }
    })
}
```

**TestErrorContract_Predicate_NilPredicate** (2 subtests):

```go
func TestErrorContract_Predicate_NilPredicate(t *testing.T) {
    t.Parallel()

    coll := []int{1, 2, 3}

    t.Run("ForAll nil predicate", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.ForAll[int](coll, nil)
        if err == nil {
            t.Fatal("error contract violated: ForAll with nil predicate should return error")
        }
    })

    t.Run("Exists nil predicate", func(t *testing.T) {
        t.Parallel()
        _, err := predicate.Exists[int](coll, nil)
        if err == nil {
            t.Fatal("error contract violated: Exists with nil predicate should return error")
        }
    })
}
```

---

### Section 4.4: engine/mcdm/ -- AHP and TOPSIS Validation

**TestErrorContract_AHP** (4 subtests):

```go
func TestErrorContract_AHP(t *testing.T) {
    t.Parallel()

    t.Run("empty matrix", func(t *testing.T) {
        t.Parallel()
        _, err := ahp.Analyze(ahp.PairwiseMatrix{})
        if err == nil {
            t.Fatal("error contract violated: empty matrix should return error")
        }
    })

    t.Run("non-square matrix", func(t *testing.T) {
        t.Parallel()
        _, err := ahp.Analyze(ahp.PairwiseMatrix{{1, 2}})
        if err == nil {
            t.Fatal("error contract violated: non-square matrix should return error")
        }
    })

    t.Run("Rank empty weights", func(t *testing.T) {
        t.Parallel()
        _, err := ahp.Rank([]float64{}, [][]float64{})
        if err == nil {
            t.Fatal("error contract violated: Rank with empty weights should return error")
        }
    })

    t.Run("Rank dimension mismatch", func(t *testing.T) {
        t.Parallel()
        _, err := ahp.Rank([]float64{0.5, 0.5}, [][]float64{{1, 2, 3}})
        if err == nil {
            t.Fatal("error contract violated: Rank with mismatched dimensions should return error")
        }
    })
}
```

**TestErrorContract_TOPSIS** (2 subtests):

```go
func TestErrorContract_TOPSIS(t *testing.T) {
    t.Parallel()

    t.Run("empty matrix", func(t *testing.T) {
        t.Parallel()
        _, err := topsis.Rank([][]float64{}, []topsis.Criterion{})
        if err == nil {
            t.Fatal("error contract violated: empty input should return error")
        }
    })

    t.Run("dimension mismatch", func(t *testing.T) {
        t.Parallel()
        _, err := topsis.Rank(
            [][]float64{{1, 2}, {3}},
            []topsis.Criterion{{Weight: 0.5, Benefit: true}, {Weight: 0.5, Benefit: true}},
        )
        if err == nil {
            t.Fatal("error contract violated: jagged matrix should return error")
        }
    })
}
```

---

### Section 4.5: engine/bayesian/ -- Network Validation

**TestErrorContract_BayesianNetwork** (4 subtests):

```go
func TestErrorContract_BayesianNetwork(t *testing.T) {
    t.Parallel()

    t.Run("duplicate variable", func(t *testing.T) {
        t.Parallel()

        bn := network.NewNetwork()
        cpt := bayesian.NewCPT("X", nil)
        cpt.Set(stats.Assignment{}, stats.Distribution{"a": 0.5, "b": 0.5})
        err := bn.AddNode(network.Node{Variable: "X", CPT: cpt, Outcomes: []stats.Outcome{"a", "b"}})
        if err != nil {
            t.Fatalf("first AddNode should succeed: %v", err)
        }

        err = bn.AddNode(network.Node{Variable: "X", CPT: cpt, Outcomes: []stats.Outcome{"a", "b"}})
        if err == nil {
            t.Fatal("error contract violated: duplicate variable should return error")
        }
    })

    t.Run("validate detects missing parent", func(t *testing.T) {
        t.Parallel()

        bn := network.NewNetwork()
        cpt := bayesian.NewCPT("Y", []stats.Var{"Missing"})
        cpt.Set(stats.Assignment{"Missing": "a"}, stats.Distribution{"x": 0.5, "y": 0.5})
        bn.AddNode(network.Node{
            Variable: "Y", Parents: []stats.Var{"Missing"}, CPT: cpt,
            Outcomes: []stats.Outcome{"x", "y"},
        })

        err := bn.Validate()
        if err == nil {
            t.Fatal("error contract violated: network with missing parent should fail validation")
        }
    })

    t.Run("validate detects cycle", func(t *testing.T) {
        t.Parallel()

        bn := network.NewNetwork()
        cptA := bayesian.NewCPT("A", []stats.Var{"B"})
        cptA.Set(stats.Assignment{"B": "t"}, stats.Distribution{"t": 0.5, "f": 0.5})
        cptA.Set(stats.Assignment{"B": "f"}, stats.Distribution{"t": 0.5, "f": 0.5})
        bn.AddNode(network.Node{
            Variable: "A", Parents: []stats.Var{"B"}, CPT: cptA,
            Outcomes: []stats.Outcome{"t", "f"},
        })

        cptB := bayesian.NewCPT("B", []stats.Var{"A"})
        cptB.Set(stats.Assignment{"A": "t"}, stats.Distribution{"t": 0.5, "f": 0.5})
        cptB.Set(stats.Assignment{"A": "f"}, stats.Distribution{"t": 0.5, "f": 0.5})
        bn.AddNode(network.Node{
            Variable: "B", Parents: []stats.Var{"A"}, CPT: cptB,
            Outcomes: []stats.Outcome{"t", "f"},
        })

        err := bn.Validate()
        if err == nil {
            t.Fatal("error contract violated: cyclic network should fail validation")
        }
    })

    t.Run("validate detects no outcomes", func(t *testing.T) {
        t.Parallel()

        bn := network.NewNetwork()
        cpt := bayesian.NewCPT("X", nil)
        cpt.Set(stats.Assignment{}, stats.Distribution{"a": 0.5, "b": 0.5})
        bn.AddNode(network.Node{
            Variable: "X", CPT: cpt, Outcomes: []stats.Outcome{}, // empty outcomes
        })

        err := bn.Validate()
        if err == nil {
            t.Fatal("error contract violated: node with no outcomes should fail validation")
        }
    })
}
```

**TestErrorContract_BayesianCPT** (2 subtests):

```go
func TestErrorContract_BayesianCPT(t *testing.T) {
    t.Parallel()

    t.Run("lookup missing config", func(t *testing.T) {
        t.Parallel()

        cpt := bayesian.NewCPT("X", []stats.Var{"A"})
        cpt.Set(stats.Assignment{"A": "1"}, stats.Distribution{"yes": 0.5, "no": 0.5})

        _, err := cpt.Lookup(stats.Assignment{"A": "nonexistent"})
        if err == nil {
            t.Fatal("error contract violated: lookup with missing config should return error")
        }
    })

    t.Run("validate invalid distribution", func(t *testing.T) {
        t.Parallel()

        cpt := bayesian.NewCPT("X", nil)
        cpt.Set(stats.Assignment{}, stats.Distribution{"yes": 0.3, "no": 0.3}) // sums to 0.6, not 1.0

        err := cpt.Validate()
        if err == nil {
            t.Fatal("error contract violated: distribution not summing to 1.0 should fail validation")
        }
    })
}
```

---

### Section 4.6: engine/causal/ -- SCM Validation

**TestErrorContract_CausalSCM** (4 subtests):

```go
func TestErrorContract_CausalSCM(t *testing.T) {
    t.Parallel()

    t.Run("duplicate variable", func(t *testing.T) {
        t.Parallel()

        scm := model.NewSCM()
        err := scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 0 })
        if err != nil {
            t.Fatalf("first AddVariable should succeed: %v", err)
        }

        err = scm.AddVariable("X", nil, func(_ map[string]float64) float64 { return 1 })
        if err == nil {
            t.Fatal("error contract violated: duplicate variable should return error")
        }
    })

    t.Run("nil equation", func(t *testing.T) {
        t.Parallel()

        scm := model.NewSCM()
        err := scm.AddVariable("X", nil, nil)
        if err == nil {
            t.Fatal("error contract violated: nil equation should return error")
        }
    })

    t.Run("validate missing parent", func(t *testing.T) {
        t.Parallel()

        scm := model.NewSCM()
        scm.AddVariable("Y", []string{"Missing"}, func(p map[string]float64) float64 {
            return p["Missing"]
        })

        err := scm.Validate()
        if err == nil {
            t.Fatal("error contract violated: model with missing parent should fail validation")
        }
    })

    t.Run("validate cycle", func(t *testing.T) {
        t.Parallel()

        scm := model.NewSCM()
        scm.AddVariable("A", []string{"B"}, func(p map[string]float64) float64 { return p["B"] })
        scm.AddVariable("B", []string{"A"}, func(p map[string]float64) float64 { return p["A"] })

        err := scm.Validate()
        if err == nil {
            t.Fatal("error contract violated: cyclic model should fail validation")
        }
    })
}
```

---

## Appendix C: Error Contract Inventory (include as reference comment)

Total: 48 error functions covered across 6 subsections.

| Package | Function | Invalid Input | Expected Error |
|---------|----------|---------------|----------------|
| math/fuzzy | Triangular(5,3,7) | a > b | ErrInvalidRange |
| math/fuzzy | Triangular(1,5,3) | b > c | ErrInvalidRange |
| math/fuzzy | Triangular(5,5,5) | a == c | ErrInvalidRange |
| math/fuzzy | Trapezoidal(5,3,7,9) | b < a | ErrInvalidRange |
| math/fuzzy | Gaussian(0,0) | sigma=0 | ErrInvalidRange |
| math/stats | NewNormal(0,0) | sigma=0 | ErrInvalidParameter |
| ... (full inventory in ACCEPTANCE_TESTS.md Appendix C) |

## Verification
```
cd modules/compute/tests/acceptance
go vet ./...
go test -run TestErrorContract -count=1 -v ./...
```
