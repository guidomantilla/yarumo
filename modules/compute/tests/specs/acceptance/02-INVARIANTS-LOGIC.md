# Acceptance Tests — Prompt 02: invariants_logic_test.go

## Context

Read `00-context.md` for project structure, imports, and coding standards.

## Role

You are a Go testing engineer. Generate the file `invariants_logic_test.go`.

## Output

Generate exactly ONE file: `invariants_logic_test.go`

Place it at: `modules/compute/tests/acceptance/invariants_logic_test.go`

## Constraints

- Package: `package acceptance_test`
- No testify — use `t.Fatal`/`t.Fatalf`
- No table-driven tests
- `t.Parallel()` on every test function and every subtest
- Only import public APIs (see 00-context.md for import aliases)
- No inline assignments (`if err := fn(); err != nil` forbidden)

## Helper References

This file uses the following helpers defined in `helpers_test.go`:

- `generateFormulas(depth int, vars []logic.Var) []logic.Formula` — builds all formulas up to depth with given variables
- `isCNF(f logic.Formula) bool` — checks that a formula is in Conjunctive Normal Form
- `isDNF(f logic.Formula) bool` — checks that a formula is in Disjunctive Normal Form
- `isClause(f logic.Formula) bool` — checks that a formula is a disjunction of literals
- `isCNFConjunct(f logic.Formula) bool` — checks CNF conjunct structure
- `isDNFConjunct(f logic.Formula) bool` — checks DNF conjunct structure
- `isDNFDisjunct(f logic.Formula) bool` — checks DNF disjunct structure

## Required Imports

```go
import (
    "fmt"
    "testing"
    "time"

    "github.com/guidomantilla/yarumo/compute/math/logic"
    "github.com/guidomantilla/yarumo/compute/math/logic/entailment"
    "github.com/guidomantilla/yarumo/compute/math/logic/predicate"
    "github.com/guidomantilla/yarumo/compute/math/logic/sat"
    "github.com/guidomantilla/yarumo/compute/math/logic/temporal"
)
```

## Tests

The file must contain exactly 19 test functions organized in 5 sections.

---

### Section 1.1: logic/ — Transformations (7 tests)

#### TestAcceptance_NNF_preserves_equivalence

// Strategy: Exhaustive within bounds
// Reference: Mendelson "Introduction to Mathematical Logic" §1.4
// Verifications: ~1000 (all formulas in corpus)

```go
func TestAcceptance_NNF_preserves_equivalence(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            nnf := logic.NNF(f)
            if !logic.Equivalent(f, nnf) {
                t.Fatalf("formula %d: NNF not equivalent: %s vs %s", i, f, nnf)
            }
        }
    })
}
```

#### TestAcceptance_CNF_preserves_equivalence

// Strategy: Exhaustive within bounds
// Reference: Mendelson "Introduction to Mathematical Logic" §1.4
// Verifications: ~1000

```go
func TestAcceptance_CNF_preserves_equivalence(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            cnf := logic.CNF(f)
            if !logic.Equivalent(f, cnf) {
                t.Fatalf("formula %d: CNF not equivalent: %s vs %s", i, f, cnf)
            }
        }
    })
}
```

#### TestAcceptance_DNF_preserves_equivalence

// Strategy: Exhaustive within bounds
// Reference: Mendelson "Introduction to Mathematical Logic" §1.4
// Verifications: ~1000

```go
func TestAcceptance_DNF_preserves_equivalence(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            dnf := logic.DNF(f)
            if !logic.Equivalent(f, dnf) {
                t.Fatalf("formula %d: DNF not equivalent: %s vs %s", i, f, dnf)
            }
        }
    })
}
```

#### TestAcceptance_CNF_structural_form

// Strategy: Exhaustive within bounds
// Reference: Mendelson "Introduction to Mathematical Logic" §1.4
// Verifications: ~1000

```go
func TestAcceptance_CNF_structural_form(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            cnf := logic.CNF(f)
            if !isCNF(cnf) {
                t.Fatalf("formula %d: CNF result is not in CNF form: %s", i, cnf)
            }
        }
    })
}
```

#### TestAcceptance_DNF_structural_form

// Strategy: Exhaustive within bounds
// Reference: Mendelson "Introduction to Mathematical Logic" §1.4
// Verifications: ~1000

```go
func TestAcceptance_DNF_structural_form(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            dnf := logic.DNF(f)
            if !isDNF(dnf) {
                t.Fatalf("formula %d: DNF result is not in DNF form: %s", i, dnf)
            }
        }
    })
}
```

#### TestAcceptance_Simplify_idempotence

// Strategy: Exhaustive within bounds
// Reference: Simplify(Simplify(f)).Equals(Simplify(f)) for all f
// Verifications: ~1000

```go
func TestAcceptance_Simplify_idempotence(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            s1 := logic.Simplify(f)
            s2 := logic.Simplify(s1)
            if !s1.Equals(s2) {
                t.Fatalf("formula %d: Simplify not idempotent: %s != %s", i, s1, s2)
            }
        }
    })
}
```

#### TestAcceptance_Simplify_preserves_equivalence

// Strategy: Exhaustive within bounds
// Reference: Equivalent(f, Simplify(f)) for all f
// Verifications: ~1000

```go
func TestAcceptance_Simplify_preserves_equivalence(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            simplified := logic.Simplify(f)
            if !logic.Equivalent(f, simplified) {
                t.Fatalf("formula %d: Simplify not equivalent: %s vs %s", i, f, simplified)
            }
        }
    })
}
```

---

### Section 1.2: logic/sat/ — DPLL (6 tests)

#### TestAcceptance_SAT_soundness_exhaustive

// Strategy: Exhaustive within bounds
// Reference: DPLL algorithm soundness — if SAT returns a model, the model satisfies the formula
// Verifications: ~1000

```go
func TestAcceptance_SAT_soundness_exhaustive(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            cnf := sat.FromFormula(logic.CNF(f))
            satisfiable, model := sat.Solve(cnf)
            if satisfiable && !f.Eval(model) {
                t.Fatalf("formula %d: SAT returned model that does not satisfy formula: %s, model=%v", i, f, model)
            }
        }
    })
}
```

#### TestAcceptance_SAT_completeness_exhaustive

// Strategy: Exhaustive within bounds
// Reference: DPLL algorithm completeness — if UNSAT, no truth table row satisfies
// Verifications: ~1000

```go
func TestAcceptance_SAT_completeness_exhaustive(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            if logic.IsSatisfiable(f) {
                continue
            }
            rows := logic.TruthTable(f)
            for _, row := range rows {
                if row.Result {
                    t.Fatalf("formula %d: DPLL says UNSAT but truth table has satisfying row: %s, assignment=%v", i, f, row.Assignment)
                }
            }
        }
    })
}
```

#### TestAcceptance_CNF_preserves_satisfiability

// Strategy: Exhaustive within bounds
// Reference: CNF conversion preserves satisfiability
// Verifications: ~1000

```go
func TestAcceptance_CNF_preserves_satisfiability(t *testing.T) {
    t.Parallel()
    vars := []logic.Var{"P", "Q", "R"}
    corpus := generateFormulas(2, vars)
    t.Run("all formulas", func(t *testing.T) {
        t.Parallel()
        for i, f := range corpus {
            original := logic.IsSatisfiable(f)
            converted := logic.IsSatisfiable(logic.CNF(f))
            if original != converted {
                t.Fatalf("formula %d: satisfiability changed by CNF: original=%v, converted=%v, formula=%s", i, original, converted, f)
            }
        }
    })
}
```

#### TestAcceptance_SAT_adversarial_XOR_chain

// Strategy: Adversarial
// Reference: XOR is hard for resolution-based solvers (Tseitin 1968)
// Verifications: 1

```go
func TestAcceptance_SAT_adversarial_XOR_chain(t *testing.T) {
    t.Parallel()

    t.Run("XOR chain is satisfiable", func(t *testing.T) {
        t.Parallel()

        // P XOR Q = (P | Q) & (!P | !Q)
        // Chain: P XOR Q XOR R XOR S
        p, q, r, s := logic.Var("P"), logic.Var("Q"), logic.Var("R"), logic.Var("S")
        xorPQ := logic.AndF{
            L: logic.OrF{L: p, R: q},
            R: logic.OrF{L: logic.NotF{F: p}, R: logic.NotF{F: q}},
        }
        xorPQR := logic.AndF{
            L: logic.OrF{L: xorPQ, R: r},
            R: logic.OrF{L: logic.NotF{F: xorPQ}, R: logic.NotF{F: r}},
        }
        xorAll := logic.AndF{
            L: logic.OrF{L: xorPQR, R: s},
            R: logic.OrF{L: logic.NotF{F: xorPQR}, R: logic.NotF{F: s}},
        }

        if !logic.IsSatisfiable(xorAll) {
            t.Fatal("XOR chain should be satisfiable")
        }
    })
}
```

#### TestAcceptance_SAT_adversarial_pigeonhole

// Strategy: Adversarial
// Reference: Haken 1985 — PHP requires exponential proofs in resolution
// Verifications: 1

```go
func TestAcceptance_SAT_adversarial_pigeonhole(t *testing.T) {
    t.Parallel()

    t.Run("PHP 3,2 is unsatisfiable", func(t *testing.T) {
        t.Parallel()

        // p_i_j = pigeon i is in hole j
        // Each pigeon must be in at least one hole: (p_i_1 | p_i_2) for i=1,2,3
        // Each hole has at most one pigeon: !(p_i_j & p_k_j) for i!=k
        p := func(i, j int) logic.Var {
            return logic.Var(fmt.Sprintf("p%d_%d", i, j))
        }

        // At least one hole per pigeon
        atLeast1 := logic.AndF{
            L: logic.OrF{L: p(1, 1), R: p(1, 2)},
            R: logic.AndF{
                L: logic.OrF{L: p(2, 1), R: p(2, 2)},
                R: logic.OrF{L: p(3, 1), R: p(3, 2)},
            },
        }

        // At most one pigeon per hole
        atMost1 := logic.AndF{
            L: logic.OrF{L: logic.NotF{F: p(1, 1)}, R: logic.NotF{F: p(2, 1)}},
            R: logic.AndF{
                L: logic.OrF{L: logic.NotF{F: p(1, 1)}, R: logic.NotF{F: p(3, 1)}},
                R: logic.AndF{
                    L: logic.OrF{L: logic.NotF{F: p(2, 1)}, R: logic.NotF{F: p(3, 1)}},
                    R: logic.AndF{
                        L: logic.OrF{L: logic.NotF{F: p(1, 2)}, R: logic.NotF{F: p(2, 2)}},
                        R: logic.AndF{
                            L: logic.OrF{L: logic.NotF{F: p(1, 2)}, R: logic.NotF{F: p(3, 2)}},
                            R: logic.OrF{L: logic.NotF{F: p(2, 2)}, R: logic.NotF{F: p(3, 2)}},
                        },
                    },
                },
            },
        }

        php := logic.AndF{L: atLeast1, R: atMost1}

        if logic.IsSatisfiable(php) {
            t.Fatal("PHP(3,2) should be unsatisfiable")
        }
    })
}
```

#### TestAcceptance_SAT_adversarial_tautology

// Strategy: Adversarial
// Reference: (P v ~P) is a tautology for any P
// Verifications: 1

```go
func TestAcceptance_SAT_adversarial_tautology(t *testing.T) {
    t.Parallel()

    t.Run("conjunction of tautological clauses", func(t *testing.T) {
        t.Parallel()

        // (P1 | !P1) & (P2 | !P2) & ... & (P10 | !P10)
        vars := make([]logic.Var, 10)
        for i := range vars {
            vars[i] = logic.Var(fmt.Sprintf("V%d", i))
        }

        var f logic.Formula = logic.OrF{L: vars[0], R: logic.NotF{F: vars[0]}}
        for i := 1; i < len(vars); i++ {
            clause := logic.OrF{L: vars[i], R: logic.NotF{F: vars[i]}}
            f = logic.AndF{L: f, R: clause}
        }

        if !logic.IsSatisfiable(f) {
            t.Fatal("tautological conjunction should be satisfiable")
        }
    })
}
```

---

### Section 1.3: logic/entailment/ (4 tests)

#### TestAcceptance_Entailment_exhaustive_crosscheck

// Strategy: Exhaustive within bounds
// Reference: Semantic definition: A |= B iff every model of A is a model of B
// Verifications: ~500 (all pairs of depth<=1 formulas)

```go
func TestAcceptance_Entailment_exhaustive_crosscheck(t *testing.T) {
    t.Parallel()

    vars := []logic.Var{"P", "Q"}
    corpus := generateFormulas(1, vars)

    t.Run("all pairs", func(t *testing.T) {
        t.Parallel()

        for i, f1 := range corpus {
            for j, f2 := range corpus {
                entails := entailment.Entails([]logic.Formula{f1}, f2)

                // Semantic check: for every assignment where f1 is true, f2 must be true
                semantic := true
                rows := logic.TruthTable(f1)
                for _, row := range rows {
                    if row.Result && !f2.Eval(row.Assignment) {
                        semantic = false
                        break
                    }
                }

                if entails != semantic {
                    t.Fatalf("pair (%d,%d): Entails=%v but semantic=%v, f1=%s, f2=%s", i, j, entails, semantic, f1, f2)
                }
            }
        }
    })
}
```

#### TestAcceptance_Countermodel_validation_exhaustive

// Strategy: Exhaustive within bounds
// Reference: Definition of countermodel
// Verifications: ~250 (non-entailment pairs)

```go
func TestAcceptance_Countermodel_validation_exhaustive(t *testing.T) {
    t.Parallel()

    vars := []logic.Var{"P", "Q"}
    corpus := generateFormulas(1, vars)

    t.Run("all non-entailment pairs", func(t *testing.T) {
        t.Parallel()

        for i, f1 := range corpus {
            for j, f2 := range corpus {
                holds, counter := entailment.EntailsWithCounterModel([]logic.Formula{f1}, f2)
                if holds {
                    continue
                }

                if !f1.Eval(counter) {
                    t.Fatalf("pair (%d,%d): countermodel does not satisfy premise: f1=%s, model=%v", i, j, f1, counter)
                }

                if f2.Eval(counter) {
                    t.Fatalf("pair (%d,%d): countermodel satisfies conclusion: f2=%s, model=%v", i, j, f2, counter)
                }
            }
        }
    })
}
```

#### TestAcceptance_Entailment_known_answer

// Strategy: Known-answer
// Reference: Enderton "Mathematical Introduction to Logic" §1
// Verifications: 5 (modus ponens, modus tollens, hypothetical syllogism, disjunctive syllogism, affirming consequent fallacy)

```go
func TestAcceptance_Entailment_known_answer(t *testing.T) {
    t.Parallel()

    p, q, r := logic.Var("P"), logic.Var("Q"), logic.Var("R")

    t.Run("modus ponens", func(t *testing.T) {
        t.Parallel()

        // {P, P=>Q} |= Q
        if !entailment.Entails([]logic.Formula{p, logic.ImplF{L: p, R: q}}, q) {
            t.Fatal("modus ponens should hold")
        }
    })

    t.Run("modus tollens", func(t *testing.T) {
        t.Parallel()

        // {!Q, P=>Q} |= !P
        if !entailment.Entails([]logic.Formula{logic.NotF{F: q}, logic.ImplF{L: p, R: q}}, logic.NotF{F: p}) {
            t.Fatal("modus tollens should hold")
        }
    })

    t.Run("hypothetical syllogism", func(t *testing.T) {
        t.Parallel()

        // {P=>Q, Q=>R} |= P=>R
        if !entailment.Entails([]logic.Formula{logic.ImplF{L: p, R: q}, logic.ImplF{L: q, R: r}}, logic.ImplF{L: p, R: r}) {
            t.Fatal("hypothetical syllogism should hold")
        }
    })

    t.Run("disjunctive syllogism", func(t *testing.T) {
        t.Parallel()

        // {P|Q, !P} |= Q
        if !entailment.Entails([]logic.Formula{logic.OrF{L: p, R: q}, logic.NotF{F: p}}, q) {
            t.Fatal("disjunctive syllogism should hold")
        }
    })

    t.Run("affirming consequent is fallacy", func(t *testing.T) {
        t.Parallel()

        // {Q, P=>Q} |/= P
        if entailment.Entails([]logic.Formula{q, logic.ImplF{L: p, R: q}}, p) {
            t.Fatal("affirming the consequent should NOT hold")
        }
    })
}
```

---

### Section 1.4: logic/predicate/ (1 test)

#### TestAcceptance_Predicate_boundaries

// Strategy: Boundary cases
// Reference: FOL restricted to finite domains
// Verifications: 5 subtests (singleton ForAll, singleton Exists, empty domain error, always false, always true)

```go
func TestAcceptance_Predicate_boundaries(t *testing.T) {
    t.Parallel()

    alwaysTrue := func(_ int) bool { return true }
    alwaysFalse := func(_ int) bool { return false }
    isEven := func(x int) bool { return x%2 == 0 }

    t.Run("singleton domain ForAll equals single check", func(t *testing.T) {
        t.Parallel()

        result, err := predicate.ForAll([]int{42}, isEven)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if !result {
            t.Fatal("ForAll([42], isEven) should be true")
        }
    })

    t.Run("singleton domain Exists equals single check", func(t *testing.T) {
        t.Parallel()

        result, err := predicate.Exists([]int{42}, isEven)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if !result {
            t.Fatal("Exists([42], isEven) should be true")
        }
    })

    t.Run("empty domain returns error", func(t *testing.T) {
        t.Parallel()

        _, err := predicate.ForAll([]int{}, alwaysTrue)
        if err == nil {
            t.Fatal("error contract violated: ForAll on empty domain should return error")
        }

        _, err = predicate.Exists([]int{}, alwaysFalse)
        if err == nil {
            t.Fatal("error contract violated: Exists on empty domain should return error")
        }
    })

    t.Run("always false predicate", func(t *testing.T) {
        t.Parallel()

        domain := []int{1, 2, 3, 4, 5}

        result, err := predicate.ForAll(domain, alwaysFalse)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if result {
            t.Fatal("ForAll(domain, alwaysFalse) should be false")
        }

        count, err := predicate.Count(domain, alwaysFalse)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if count != 0 {
            t.Fatalf("Count(domain, alwaysFalse) should be 0, got %d", count)
        }
    })

    t.Run("always true predicate", func(t *testing.T) {
        t.Parallel()

        domain := []int{1, 2, 3, 4, 5}

        result, err := predicate.ForAll(domain, alwaysTrue)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if !result {
            t.Fatal("ForAll(domain, alwaysTrue) should be true")
        }

        count, err := predicate.Count(domain, alwaysTrue)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }

        if count != 5 {
            t.Fatalf("Count(domain, alwaysTrue) should be 5, got %d", count)
        }
    })
}
```

---

### Section 1.5: logic/temporal/ (1 test)

#### TestAcceptance_Temporal_boundaries

// Strategy: Boundary exact
// Reference: Bounded model checking (Biere et al. 2003)
// Verifications: 5 subtests (ResponseWithin at deadline, one ns after, Sequence with duplicates, Before simultaneous, FrequencyWithin exact)

```go
func TestAcceptance_Temporal_boundaries(t *testing.T) {
    t.Parallel()

    t.Run("ResponseWithin exactly at deadline passes", func(t *testing.T) {
        t.Parallel()

        maxDur := 100 * time.Millisecond
        trace := temporal.Trace{
            {Name: "trigger", Time: time.Unix(0, 0)},
            {Name: "response", Time: time.Unix(0, int64(maxDur))},
        }

        result := temporal.ResponseWithin(trace, "trigger", "response", maxDur)
        if !result {
            t.Fatal("response at exactly the deadline should pass")
        }
    })

    t.Run("ResponseWithin one nanosecond after deadline fails", func(t *testing.T) {
        t.Parallel()

        maxDur := 100 * time.Millisecond
        trace := temporal.Trace{
            {Name: "trigger", Time: time.Unix(0, 0)},
            {Name: "response", Time: time.Unix(0, int64(maxDur)+1)},
        }

        result := temporal.ResponseWithin(trace, "trigger", "response", maxDur)
        if result {
            t.Fatal("response one nanosecond after deadline should fail")
        }
    })

    t.Run("Sequence with duplicates finds first match", func(t *testing.T) {
        t.Parallel()

        trace := temporal.Trace{
            {Name: "A", Time: time.Unix(0, 0)},
            {Name: "B", Time: time.Unix(0, 1)},
            {Name: "A", Time: time.Unix(0, 2)},
            {Name: "B", Time: time.Unix(0, 3)},
            {Name: "C", Time: time.Unix(0, 4)},
        }

        result := temporal.Sequence(trace, []string{"A", "B", "C"})
        if !result {
            t.Fatal("Sequence [A,B,C] should be found in [A,B,A,B,C]")
        }
    })

    t.Run("Before with simultaneous events fails", func(t *testing.T) {
        t.Parallel()

        sameTime := time.Unix(1, 0)
        trace := temporal.Trace{
            {Name: "a", Time: sameTime},
            {Name: "b", Time: sameTime},
        }

        result := temporal.Before(trace, "a", "b")
        if result {
            t.Fatal("Before(a, b) with simultaneous events should fail")
        }
    })

    t.Run("FrequencyWithin exact threshold passes", func(t *testing.T) {
        t.Parallel()

        window := 10 * time.Second
        trace := temporal.Trace{
            {Name: "event", Time: time.Unix(1, 0)},
            {Name: "event", Time: time.Unix(3, 0)},
            {Name: "event", Time: time.Unix(5, 0)},
        }

        result := temporal.FrequencyWithin(trace, "event", 3, window)
        if !result {
            t.Fatal("exactly minCount events in window should pass")
        }

        result = temporal.FrequencyWithin(trace, "event", 4, window)
        if result {
            t.Fatal("minCount-1 events should fail")
        }
    })
}
```

---

## Verification

After generating, run:
```
cd modules/compute/tests/acceptance
go vet ./...
go test -run TestAcceptance -count=1 -v ./...
```
