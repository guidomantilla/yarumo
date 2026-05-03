# Testing Taxonomy

> A complete classification of software testing types, from lowest to highest level,
> plus orthogonal dimensions that apply across all levels.
>
> Date: 2026-03-02 | Version: 1

---

## The testing pyramid

```
                        /\
                       /  \
                      / AC \        Acceptance Tests
                     /______\       "The analyst creates a ruleset and executes a decision"
                    /        \
                   /   E2E    \     End-to-End Tests
                  /____________\    "HTTP request -> API -> SDK -> engine -> DB -> response"
                 /              \
                /    Contract    \   Contract Tests
               /__________________\  "Repository fulfills the interface regardless of Postgres or Memory"
              /                    \
             /     Component        \  Component Tests
            /________________________\ "The inference module produces correct results given a ruleset"
           /                          \
          /       Integration          \  Integration Tests
         /______________________________\ "Parser + Evaluator work together correctly"
        /                                \
       /            Unit                  \  Unit Tests
      /____________________________________\ "The Simplify function reduces A and A to A"
```

**Key insight**: the higher up the pyramid, the fewer tests you need, the slower they run,
and the more behavior they protect. The lower, the more tests you have, the faster they run,
and the more specific they are.

---

## Level 1 — Unit Tests

**What**: a single function, a single method, in total isolation. Dependencies are mocked
or absent.

**Scope**: inside ONE package. No external dependencies.

**What breaks them**: internal changes to the package.

**Characteristics**:
- Fastest to run (milliseconds)
- Highest quantity (hundreds or thousands)
- Test implementation details — they WILL break on refactors, and that's expected
- Every branch, every edge case, every error path

**Example**: the `Simplify` function reduces `A and A` to `A`. The `Parse` function with
input `A and B` returns an `And` node with children `A`, `B`.

---

## Level 2 — Integration Tests

**What**: two or more packages collaborating. They verify that pieces fit together.

**Scope**: crosses package boundaries WITHIN a module.

**What breaks them**: changes in interfaces between packages, changes in internal formats.

**Characteristics**:
- Test the seams between components
- More realistic than unit tests but still fast
- Don't need external infrastructure (no DB, no HTTP)
- Catch "works in isolation but breaks when connected" bugs

**Example**: the parser parses a formula -> the evaluator evaluates it -> the result is
correct. Or: the engine uses the logic module internally -> conclusions are correct.

---

## Level 3 — Component Tests

**What**: an entire module as a black box. Data goes in, results come out. Internal
structure doesn't matter.

**Scope**: one complete module.

**What breaks them**: changes in the external behavior of the module. Internal refactors
do NOT break them.

**Characteristics**:
- The module is tested through its public boundary only
- Input and output are expressed as plain data (not internal types when possible)
- If you rewrite the module from scratch but preserve behavior, these tests still pass
- Slower than integration tests but much more resilient to refactors

**Example**: the inference module receives rules + facts as data -> responds with
conclusions. You don't know or care whether it uses forward chaining, RETE, or magic
internally.

---

## Level 4 — Contract Tests

**What**: verify that implementations of an interface fulfill the contract. If you have
`Repository` as an interface, every implementation (Memory, Postgres, Filesystem) must
pass THE SAME tests.

**Scope**: the interface and all its implementations.

**What breaks them**: implementations that don't fulfill the complete interface, or that
have divergent behavior.

**Characteristics**:
- One test suite, parameterized over N implementations
- Every implementation runs the exact same scenarios
- Ensure substitutability (Liskov Substitution Principle)
- Particularly valuable when implementations are added over time

**Example**: a parameterized test suite runs against `MemoryRepository`,
`PostgresRepository`, `FilesystemRepository` — all three must pass identically.
Save -> Get returns what was saved. Delete -> Get returns not found. List returns
all saved items.

---

## Level 5 — End-to-End Tests (E2E)

**What**: the complete system running as a real user would use it, including real
infrastructure (HTTP, database, message queues, etc.).

**Scope**: the entire application, deployed and running.

**What breaks them**: any change in any layer — they are the most fragile but the most
realistic.

**Characteristics**:
- Slowest to run (seconds to minutes)
- Require real infrastructure (or testcontainers)
- Fewest in quantity
- Catch deployment, configuration, and wiring bugs
- Fragile — break for many reasons, not all meaningful

**Example**: an HTTP client sends `POST /api/v1/decisions` with JSON -> the server
processes it -> returns 200 with result -> the result is verified in the audit database.

---

## Level 6 — Acceptance Tests

**What**: declarative business scenarios. They don't say how, they say what. They express
eternal domain truths that survive any implementation change.

**Scope**: the entire business domain.

**What breaks them**: only changes that alter the functional behavior of the system.

**Characteristics**:
- Written in business language, not code language
- Describe WHAT must happen, not HOW
- Survive complete rewrites — if the behavior is preserved, they pass
- Act as living documentation of system requirements
- Fewest in quantity but highest in business value

**Example**: "A client with good credit and sufficient collateral is approved for a loan".
"A ruleset with contradictory rules is flagged during validation".

---

## Orthogonal dimensions

These are not "levels" but "dimensions" — they apply at any level of the pyramid.

### Property-Based Tests

**What**: instead of fixed inputs, generate random inputs and verify that a PROPERTY always
holds.

**Characteristics**:
- Find edge cases that manual tests can't imagine
- Express invariants: "for ALL inputs, this property holds"
- Require thinking about properties rather than examples
- Complement example-based tests, don't replace them

**Example**: "For ANY random logical formula, `Simplify(f)` is equivalent to `f`".
Generate 10,000 random formulas and verify. "For ANY valid probability distribution,
the sum is 1.0".

### Regression Tests

**What**: a test born from a bug found in production or during development. "This broke
once, never again".

**Characteristics**:
- Institutional memory of bugs
- Each one tells a story: "on date X, we found that Y broke because Z"
- Prevent re-introduction of fixed bugs
- Accumulate over time as a safety net

**Example**: "The parser used to fail with `not not not A` — now there's a test that
verifies triple negation parses correctly".

### Benchmark / Performance Tests

**What**: they don't verify correctness but speed, throughput, memory usage.

**Characteristics**:
- Detect performance regressions
- Establish baselines: "this operation takes N ms with M inputs"
- Should run in CI but with tolerance for environment variance
- Separate from correctness tests — a correct but slow system still fails these

**Example**: "Forward chaining with 100 rules and 50 facts completes in under 10ms".
"1000 concurrent decision executions don't degrade correctness".
"The API responds to `/decisions` in under 50ms at p99".

### Smoke Tests

**What**: minimal verification that the system starts and responds. They don't verify
deep correctness.

**Characteristics**:
- First filter in CI/CD — if this fails, don't run anything else
- Extremely fast (seconds)
- Test basic wiring: "does the server start?", "does the DB connect?", "does a trivial
  request succeed?"
- High signal-to-noise ratio

**Example**: "The server starts, responds 200 on `/health`, and can execute a trivial
decision".

### Mutation Tests

**What**: automatically modify your code (mutate a condition, change a `+` to `-`) and
verify that your tests detect it. If a mutant survives, your tests have a gap.

**Characteristics**:
- Measure the QUALITY of your tests, not of your code
- Expensive to run (re-run full test suite per mutation)
- Most valuable for critical code where correctness is non-negotiable
- Complement code coverage — 100% coverage doesn't mean 100% mutation kill rate

**Example**: in a math library, change `>=` to `>` in a boundary check. If no test
fails, there's a missing edge case test.

---

## How the levels relate

```
Acceptance    "A client with good credit is approved"
    |          needs
E2E           "POST /decisions returns 200 with approved=true"
    |          needs
Contract      "PostgresRepository.Save -> Get round-trips correctly"
    |          needs
Component     "inference module: rules + facts -> correct conclusions"
    |          needs
Integration   "parser + evaluator work together"
    |          needs
Unit          "Simplify(A and A) returns A"
```

Each level ASSUMES the levels below it work. If unit tests fail, don't bother running
integration tests. If component tests fail, don't bother running E2E.

---

## Anti-patterns

| Anti-pattern | Problem | Fix |
|---|---|---|
| **Ice cream cone** (many E2E, few units) | Slow CI, flaky tests, hard to debug failures | Invert: more units, fewer E2E |
| **Testing implementation** at high levels | Component test that checks internal data structures | Test through public boundary only |
| **Testing behavior** at low levels | Unit test that sets up 5 collaborators to verify a scenario | Push up to integration or component level |
| **No contract tests** with multiple implementations | Postgres works, Memory doesn't, discovered in production | One suite, parameterized over all implementations |
| **Acceptance tests as E2E** | Business scenarios tied to HTTP endpoints and DB schemas | Keep acceptance declarative; E2E is a separate concern |
| **No property-based** for math/logic code | Edge cases discovered by users, not by tests | Add property tests for invariants |

---

## Decision guide: which type do I need?

| Question | Answer |
|---|---|
| "Does this function work?" | Unit test |
| "Do these two packages work together?" | Integration test |
| "Does this module behave correctly as a whole?" | Component test |
| "Do all implementations of this interface behave the same?" | Contract test |
| "Does the deployed system work end-to-end?" | E2E test |
| "Does the system do what the business needs?" | Acceptance test |
| "Does this property hold for ALL inputs?" | Property-based test |
| "Will this old bug come back?" | Regression test |
| "Is it fast enough?" | Benchmark test |
| "Does it even start?" | Smoke test |
| "Are my tests good enough?" | Mutation test |
