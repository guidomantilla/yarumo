# Acceptance Test Specifications — modules/compute/

> **Code generation**: To generate Go test code from this spec, execute the prompts
> in `acceptance/00-CONTEXT.md` through `acceptance/07-PERFORMANCE.md`.
> See `README.md` in this directory for the full flow.

## Metadata

- Generated: 2026-03-09
- Based on: modules/compute/CORRECTNESS.md (rounds 1, 2 & 3, 2026-03-09)
- Existing tests reviewed:
  - math/logic/examples/properties_test.go (22 property tests)
  - math/logic/examples/examples_test.go (10 example tests)
  - math/logic/examples/benchmark_test.go
  - math/fuzzy/examples/properties_test.go (22 property tests)
  - math/fuzzy/examples/examples_test.go (20 example tests)
  - math/fuzzy/examples/benchmark_test.go
  - math/stats/examples/properties_test.go (12 property tests)
  - math/stats/examples/examples_test.go (8 example tests)
  - math/stats/examples/benchmark_test.go
  - engine/deductive/examples/examples_test.go (13 tests)
  - engine/deductive/examples/benchmark_test.go
  - engine/bayesian/examples/examples_test.go (14 tests)
  - engine/bayesian/examples/benchmark_test.go
  - engine/fuzzy/examples/examples_test.go (15 tests)
  - engine/fuzzy/examples/benchmark_test.go
  - engine/causal/examples/examples_test.go (10 tests)
  - engine/causal/examples/benchmark_test.go
  - engine/mcdm/examples/examples_test.go (22 tests)
  - engine/mcdm/examples/benchmark_test.go
  - NO examples/ for: math/graph/, math/fsm/, math/markov/
- Strategies: Exhaustive-within-bounds + Adversarial + Known-answer + Golden Files + Error Contracts + Performance Baselines

## Coverage Summary

| Area | Existing tests | New tests | Total verifications |
|------|---------------|-----------|---------------------|
| 1.1 logic/ transformations | 32 | 7 | ~2870 |
| 1.2 logic/sat/ DPLL | 32 | 6 | ~1237 |
| 1.3 logic/entailment/ | 32 | 7 | ~177 |
| 1.4 logic/predicate/ | unit tests | 4 | 16 |
| 1.5 logic/temporal/ | unit tests | 17 | 17 |
| 1.6 math/sets/ | 0 | 14 | ~2684 |
| 1.7 fuzzy/ axioms | 42 | 20 | ~87329 |
| 1.8 stats/ distributions | 20 | 29 | ~189 |
| 1.9 math/graph/ | 0 | 22 | ~4705 |
| 1.10 math/fsm/ | 0 | 8 | ~45 |
| 1.11 math/markov/ | 0 | 10 | ~23 |
| 1.12 engine/deductive/ | 13 | 4 | ~16 |
| 1.13 engine/bayesian/ | 14 | 4 | ~31 |
| 1.14 engine/fuzzy/ | 15 | 3 | ~648 |
| 1.15 engine/causal/ | 10 | 3 | ~10 |
| 1.16 engine/mcdm/ | 22 | 3 | ~16 |
| 2. Golden scenario | 0 | 3 | ~30 |
| 3. Golden files | 0 | 5 | ~15 |
| 4. Error contracts | 0 | 9 | ~70 |
| 5. Performance baselines | 0 | 10 | ~17 |
| **Total** | **264** | **179** | **~99,945** |

## Required Helpers

The implementation will need the following helpers (described here, not as code):

- **generateFormulas(depth, vars)**: Generates all propositional formulas up to a given depth using the given variable names. depth=0 produces each variable, True, False. depth=1 adds Not(f) for each depth=0 formula. depth=2 adds BinOp(f1,f2) for each pair from depth<=1 across {And, Or, Impl, Iff}, plus Not(f) for depth=1 formulas. If depth=3 exceeds 5000 formulas, cap at depth=2. Corpus is generated once and shared across all tests in section 1.1 and 1.2.

- **fuzzyGrid(step)**: Generates a grid of floating-point values from 0.0 to 1.0 with the given step. For step=0.05, this produces 21 points: [0.0, 0.05, 0.10, ..., 1.0]. Used for exhaustive verification of fuzzy axioms.

- **assertFloat(name, got, want, tolerance)**: Compares two floating-point values with a given tolerance. Reports failure with context including the name, expected and actual values, and tolerance used.

- **makeRainNetwork()**: Constructs the standard Rain-Sprinkler-WetGrass Bayesian network with the CPTs specified in section 1.10.

- **makeLoanNetwork()**: Constructs the CreditHistory-IncomeLevel-Default Bayesian network with CPTs for the golden scenario.

- **makeTippingEngine()**: Constructs the standard Mamdani tipping engine with food/service inputs and tip output.

- **Tolerances**:
  - floatTolerance = 1e-9 (general floating-point arithmetic)
  - probTolerance = 1e-6 (probabilities with marginalization accumulation)
  - defuzzTolerance = 0.5 (defuzzification discretization error)
  - goldenBayesian = 1e-4 (Bayesian golden file comparisons)
  - goldenFuzzy = 1e-2 (fuzzy golden file comparisons)

---

## Section 1: Mathematical Invariants

### 1.1 logic/ — Transformations (Strategy A: exhaustive)

**Formula corpus construction**: generateFormulas(depth, vars) builds all syntactically distinct formulas up to a given depth using variables {P, Q, R}.

- depth=0 produces 5 atomic formulas: P, Q, R, True, False
- depth=1 adds Not(f) for each depth=0 formula, producing 5 additional formulas (10 total)
- depth=2 adds BinOp(f1, f2) for every ordered pair of depth<=1 formulas across all four binary operators {And, Or, Impl, Iff}, plus Not(f) for each depth=1 formula. This yields approximately 400 formulas at depth=2, plus all formulas from lower depths, for a total corpus of roughly 410 formulas.
- If depth=3 would exceed 5000 formulas, the corpus is capped at depth=2.

The corpus is generated once and shared across all tests in this section. Every test iterates over every formula in the corpus.

#### Test: NNF preserves equivalence (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, NNF(f) is logically equivalent to f.
- **Reference**: Mendelson "Introduction to Mathematical Logic" §1.4; Enderton "A Mathematical Introduction to Logic" §1.5
- **Verifications**: ~410 (one per formula in corpus)
- **Prerequisite**: math/logic/examples/properties_test.go covers NNF on selected formulas
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute nf = NNF(f)
  - Call Equivalent(f, nf)
  - Expected: true
- **Failure criterion**: Equivalent(f, NNF(f)) returns false for any formula in the corpus.
- **Expected values**: true for every formula. No exceptions.

#### Test: CNF preserves equivalence (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, CNF(f) is logically equivalent to f.
- **Reference**: Mendelson §1.4; Enderton §1.5
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go covers CNF on selected formulas
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute cf = CNF(f)
  - Call Equivalent(f, cf)
  - Expected: true
- **Failure criterion**: Equivalent(f, CNF(f)) returns false for any formula.
- **Expected values**: true for every formula. No exceptions.

#### Test: DNF preserves equivalence (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, DNF(f) is logically equivalent to f.
- **Reference**: Mendelson §1.4; Enderton §1.5
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go covers DNF on selected formulas
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute df = DNF(f)
  - Call Equivalent(f, df)
  - Expected: true
- **Failure criterion**: Equivalent(f, DNF(f)) returns false for any formula.
- **Expected values**: true for every formula. No exceptions.

#### Test: CNF structural form (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, CNF(f) is either an atomic formula, a literal, or has the structural form And(Or(...), Or(...), ...) where each clause is a disjunction of literals.
- **Reference**: Enderton §1.5
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute cf = CNF(f)
  - Walk the AST of cf and verify:
    - The top-level node is And, Or, Not(variable), a variable, True, or False
    - If the top-level is And, every child is Or, Not(variable), a variable, True, or False
    - If a child is Or, every grandchild is Not(variable), a variable, True, or False
    - No nested And appears inside an Or; no nested Or appears inside another Or
- **Failure criterion**: Any CNF(f) contains a nested And inside an Or clause, or an Or inside another Or, or any non-literal leaf inside a clause.
- **Expected values**: Every formula passes the structural check.

#### Test: DNF structural form (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, DNF(f) is either an atomic formula, a literal, or has the structural form Or(And(...), And(...), ...) where each term is a conjunction of literals.
- **Reference**: Enderton §1.5
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute df = DNF(f)
  - Walk the AST and verify the dual structural form of CNF
- **Failure criterion**: Any DNF(f) contains a nested Or inside an And term, or an And inside another And.
- **Expected values**: Every formula passes the structural check.

#### Test: Simplify idempotence (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, Simplify(Simplify(f)).Equals(Simplify(f)) is true. A single pass must reach the fixed point.
- **Reference**: Standard boolean algebra; convergent rewriting system.
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute s1 = Simplify(f)
  - Compute s2 = Simplify(s1)
  - Call s2.Equals(s1)
  - Expected: true
- **Failure criterion**: Simplify(Simplify(f)).Equals(Simplify(f)) returns false for any formula.
- **Expected values**: true for every formula.

#### Test: Simplify preserves equivalence (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every propositional formula f, Simplify(f) is logically equivalent to f.
- **Reference**: Mendelson §1.4; standard boolean algebra identities.
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - Compute sf = Simplify(f)
  - Call Equivalent(f, sf)
  - Expected: true
- **Failure criterion**: Equivalent(f, Simplify(f)) returns false for any formula.
- **Expected values**: true for every formula.

---

### 1.2 logic/sat/ — DPLL (Strategy A + B)

#### Test: SAT soundness (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: If DPLL reports satisfiable, the returned model actually satisfies the formula.
- **Reference**: Davis, Putnam, Logemann, Loveland (1962). Soundness: no false positives.
- **Verifications**: ~350 (satisfiable formulas in corpus)
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f in generateFormulas(2, {P, Q, R}):
  - If IsSatisfiable(f) == true, obtain model
  - Evaluate f.Eval(model)
  - Expected: true
- **Failure criterion**: Any satisfiable formula evaluates to false under the returned model.
- **Expected values**: f.Eval(model) == true for every satisfiable formula.

#### Test: SAT completeness via truth table (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: If DPLL reports unsatisfiable, no truth assignment makes the formula true.
- **Reference**: Davis et al. (1962). Completeness: no false negatives.
- **Verifications**: ~60 (unsatisfiable formulas in corpus)
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f where IsSatisfiable(f) == false:
  - Enumerate all 2^|vars| truth assignments
  - For each assignment, evaluate f
  - Expected: every evaluation returns false
- **Failure criterion**: Any unsatisfiable formula evaluates to true under any truth assignment.
- **Expected values**: false for all assignments of all unsatisfiable formulas.

#### Test: CNF preserves satisfiability (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every formula f, f and CNF(f) have the same satisfiability status.
- **Reference**: Enderton §1.5
- **Verifications**: ~410
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each formula f:
  - sat1 = IsSatisfiable(f)
  - sat2 = IsSatisfiable(CNF(f))
  - Expected: sat1 == sat2
- **Failure criterion**: Satisfiability status differs for any formula.
- **Expected values**: Match for every formula.

#### Test: XOR chain deep nesting (adversarial)
- **Strategy**: Adversarial
- **Invariant**: DPLL correctly handles deeply nested XOR structures. XOR(a,b) = Not(Iff(a,b)).
- **Reference**: XOR chains are a standard adversarial SAT benchmark. P XOR Q XOR R XOR S has exactly 8 satisfying assignments out of 16.
- **Verifications**: 3
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**:
  - Construct XOR chain with 4 variables {P, Q, R, S}
  - IsSatisfiable must return true
  - Model must satisfy the formula
  - Truth table count must be exactly 8
- **Failure criterion**: IsSatisfiable returns false, model fails, or count != 8.
- **Expected values**: sat=true, 8 satisfying assignments.

#### Test: Pigeon hole PHP(3,2) (adversarial)
- **Strategy**: Adversarial
- **Invariant**: PHP(3,2) is unsatisfiable. 3 pigeons, 2 holes, no two pigeons share a hole.
- **Reference**: Haken (1985) proved PHP(n,n-1) requires exponential resolution proofs.
- **Verifications**: 2
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**:
  - Construct PHP(3,2) as CNF: 6 variables p_ij, 3 at-least-one clauses, 6 at-most-one clauses
  - IsSatisfiable must return false
  - Truth table confirms 0 satisfying assignments out of 64
- **Failure criterion**: IsSatisfiable returns true or any assignment satisfies all clauses.
- **Expected values**: UNSAT, 0 satisfying assignments.

#### Test: All-true tautology (adversarial)
- **Strategy**: Adversarial
- **Invariant**: (P1 OR NOT P1) AND ... AND (P10 OR NOT P10) is satisfiable (it is a tautology).
- **Reference**: Law of excluded middle; standard propositional logic.
- **Verifications**: 2
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**:
  - Construct the formula with N=10 distinct variables
  - IsSatisfiable must return true
  - Model must satisfy the formula
- **Failure criterion**: IsSatisfiable returns false or model does not satisfy the formula.
- **Expected values**: sat=true.

---

### 1.3 logic/entailment/ (Strategy A + B)

#### Test: Entailment cross-check (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every pair (f1, f2) of depth<=1 formulas, Entails([f1], f2) agrees with semantic verification via truth table.
- **Reference**: Enderton "A Mathematical Introduction to Logic" §1. Semantic entailment: Gamma |= phi.
- **Verifications**: 100 (10 x 10 pairs)
- **Prerequisite**: math/logic/examples/properties_test.go and examples_test.go
- **Subtests**: For each ordered pair (f1, f2) from 10 depth<=1 formulas:
  - Call Entails([f1], f2)
  - Compute semantic check: enumerate all truth assignments, check that every assignment satisfying f1 also satisfies f2
  - Expected: agreement
- **Failure criterion**: Entails returns different boolean than semantic check for any pair.
- **Expected values**: Agreement for all 100 pairs.

#### Test: Countermodel validation (exhaustive)
- **Strategy**: Exhaustive
- **Invariant**: For every non-entailing pair, the countermodel satisfies f1 and falsifies f2.
- **Reference**: Enderton §1.
- **Verifications**: ~70
- **Prerequisite**: math/logic/examples/properties_test.go
- **Subtests**: For each pair where Entails returns false:
  - Obtain countermodel via EntailsWithCounterModel
  - Verify f1.Eval(countermodel) == true
  - Verify f2.Eval(countermodel) == false
- **Failure criterion**: Any countermodel fails to satisfy f1 or fails to falsify f2.
- **Expected values**: Every countermodel is valid.

#### Test: Modus ponens (known-answer)
- **Strategy**: Known-answer
- **Invariant**: {P, P => Q} |= Q.
- **Reference**: Enderton §1. Most fundamental rule of inference.
- **Verifications**: 1
- **Subtests**: premises=[Var("P"), Impl(Var("P"), Var("Q"))], conclusion=Var("Q"). Expected: true.
- **Failure criterion**: Entails returns false.
- **Expected values**: true.

#### Test: Modus tollens (known-answer)
- **Strategy**: Known-answer
- **Invariant**: {NOT Q, P => Q} |= NOT P.
- **Reference**: Enderton §1.
- **Verifications**: 1
- **Subtests**: Expected: true.
- **Failure criterion**: Entails returns false.

#### Test: Hypothetical syllogism (known-answer)
- **Strategy**: Known-answer
- **Invariant**: {P => Q, Q => R} |= P => R. Implication is transitive.
- **Reference**: Enderton §1.
- **Verifications**: 1
- **Expected values**: true.

#### Test: Disjunctive syllogism (known-answer)
- **Strategy**: Known-answer
- **Invariant**: {P OR Q, NOT P} |= Q.
- **Reference**: Enderton §1.
- **Verifications**: 1
- **Expected values**: true.

#### Test: Affirming the consequent is a fallacy (known-answer)
- **Strategy**: Known-answer
- **Invariant**: {Q, P => Q} does NOT entail P. This is a formal fallacy.
- **Reference**: Enderton §1. Countermodel: P=false, Q=true.
- **Verifications**: 2
- **Subtests**:
  - Entails must return false
  - Countermodel must be {P: false, Q: true}
- **Failure criterion**: Entails returns true or countermodel is invalid.
- **Expected values**: false. Countermodel: {P: false, Q: true}.

---

### 1.4 logic/predicate/ (Strategy B: boundary)

#### Test: Singleton domain (boundary)
- **Strategy**: Adversarial (boundary)
- **Invariant**: On a singleton domain, ForAll and Exists collapse to direct predicate application.
- **Reference**: Standard first-order logic.
- **Verifications**: 4
- **Subtests**:
  - ForAll([42], trueAt42) -> true
  - ForAll([42], falseAt42) -> false
  - Exists([42], trueAt42) -> true
  - Exists([42], falseAt42) -> false
- **Failure criterion**: Any quantifier inconsistent with direct application.
- **Expected values**: true, false, true, false.

#### Test: Empty domain — standard FOL (boundary)
- **Strategy**: Adversarial (boundary)
- **Invariant**: ForAll([], P) = true (vacuous truth), Exists([], P) = false, Count([], P) = 0, Filter([], P) = [].
- **Reference**: Standard first-order logic with empty domains.
- **Verifications**: 4
- **Subtests**: domain=[], anyPredicate (irrelevant).
- **Failure criterion**: ForAll returns false, Exists returns true, Count non-zero, or Filter non-empty.
- **Expected values**: true, false, 0, [].

#### Test: Always-false predicate (boundary)
- **Strategy**: Adversarial (boundary)
- **Invariant**: For |domain|>=1 with alwaysFalse: ForAll=false, Exists=false, Count=0, Filter=[].
- **Reference**: Standard first-order logic.
- **Verifications**: 4
- **Subtests**: domain=[1,2,3,4,5], alwaysFalse.
- **Expected values**: false, false, 0, [].

#### Test: Always-true predicate (boundary)
- **Strategy**: Adversarial (boundary)
- **Invariant**: For alwaysTrue: ForAll=true, Exists=true, Count=len(domain), Filter=domain.
- **Reference**: Standard first-order logic.
- **Verifications**: 4
- **Subtests**: domain=[1,2,3,4,5], alwaysTrue.
- **Expected values**: true, true, 5, [1,2,3,4,5].

---

### 1.5 logic/temporal/ (Strategy B: precise boundary)

#### Test: ResponseWithin exactly at deadline (boundary)
- **Strategy**: Adversarial (precise boundary)
- **Invariant**: ResponseWithin accepts response at exactly maxDuration (inclusive).
- **Reference**: Manna & Pnueli.
- **Verifications**: 1
- **Subtests**: trigger at t=0, response at t=maxDuration. Expected: pass.
- **Failure criterion**: Rejects response at exact deadline.

#### Test: ResponseWithin one nanosecond after deadline (boundary)
- **Strategy**: Adversarial (precise boundary)
- **Invariant**: ResponseWithin rejects response at maxDuration + 1ns.
- **Reference**: Manna & Pnueli.
- **Verifications**: 1
- **Subtests**: trigger at t=0, response at t=maxDuration + 1ns. Expected: fail.
- **Failure criterion**: Accepts response after deadline.

#### Test: FrequencyWithin exactly at threshold (boundary)
- **Strategy**: Adversarial (precise boundary)
- **Invariant**: Accepts exactly minCount events, rejects minCount-1.
- **Reference**: Manna & Pnueli.
- **Verifications**: 2
- **Subtests**: minCount=5. 5 events: pass. 4 events: fail.
- **Failure criterion**: Rejects minCount or accepts fewer.

#### Test: Sequence with duplicates (boundary)
- **Strategy**: Adversarial (boundary)
- **Invariant**: Greedy subsequence matching handles repeated events. [A,B,A,B,C] matches [A,B,C].
- **Reference**: Standard subsequence matching.
- **Verifications**: 1
- **Expected values**: pass (A[0], B[1], C[4]).

#### Test: Before simultaneous events (boundary)
- **Strategy**: Adversarial (precise boundary)
- **Invariant**: Before(a, b) returns false when a and b occur at the same time (strict ordering).
- **Reference**: Standard strict ordering semantics.
- **Verifications**: 1
- **Expected values**: false.

#### Test: Always — all hold (LTL)
- **Strategy**: Known-answer
- **Invariant**: Always(phi) is true when phi holds at every position.
- **Reference**: Pnueli (1977).
- **Verifications**: 1
- **Subtests**: trace=[true,true,true,true,true]. Always(phi) at position 0.
- **Expected values**: true.

#### Test: Always — one violation (LTL)
- **Strategy**: Known-answer
- **Invariant**: Always(phi) is false when phi fails at any single position.
- **Reference**: Pnueli (1977).
- **Verifications**: 1
- **Subtests**: trace=[true,true,false,true,true]. Always(phi) at position 0.
- **Expected values**: false.

#### Test: Always — empty trace (LTL)
- **Strategy**: Known-answer
- **Invariant**: Always(phi) on empty trace is vacuously true.
- **Reference**: Pnueli (1977); Manna & Pnueli.
- **Verifications**: 1
- **Expected values**: true.

#### Test: Next — normal (LTL)
- **Strategy**: Known-answer
- **Invariant**: Next(phi) at position i is true iff phi holds at position i+1.
- **Reference**: Pnueli (1977).
- **Verifications**: 2
- **Subtests**: trace=[a,b,c]. Next(phi_b) at 0: true. Next(phi_a) at 0: false.
- **Expected values**: true, false.

#### Test: Next — last position (LTL)
- **Strategy**: Known-answer
- **Invariant**: Next at last position returns false (no next state).
- **Reference**: Pnueli (1977); finite-trace semantics.
- **Verifications**: 1
- **Expected values**: false.

#### Test: Until — classic (LTL)
- **Strategy**: Known-answer
- **Invariant**: (phi U psi) is true when phi holds until psi eventually holds.
- **Reference**: Pnueli (1977).
- **Verifications**: 1
- **Subtests**: trace=[¬ψ,¬ψ,ψ] with phi at positions 0,1. Until(phi,psi) at 0.
- **Expected values**: true.

#### Test: Until — psi never holds (LTL)
- **Strategy**: Known-answer
- **Invariant**: (phi U psi) is false if psi never holds. Until requires psi to eventually hold.
- **Reference**: Pnueli (1977). Strong until.
- **Verifications**: 1
- **Expected values**: false.

#### Test: Until — psi holds immediately (LTL)
- **Strategy**: Known-answer
- **Invariant**: (phi U psi) is true at i if psi holds at i (phi not required).
- **Reference**: Pnueli (1977). Vacuous satisfaction when j=i.
- **Verifications**: 1
- **Expected values**: true.

#### Test: Release — dual of Until (LTL)
- **Strategy**: Known-answer
- **Invariant**: (phi R psi) ≡ ¬(¬phi U ¬psi). Duality must hold on several traces.
- **Reference**: Pnueli (1977); Manna & Pnueli.
- **Verifications**: 4
- **Subtests**:
  - Trace 1: [phi=F psi=T, phi=F psi=T, phi=T psi=T]. Both true.
  - Trace 2: [phi=F psi=T, phi=F psi=F]. Both false.
  - Trace 3: [phi=T psi=T, phi=F psi=F]. Both true.
  - Trace 4: [phi=F psi=F]. Both false.
- **Failure criterion**: Release and NOT(Until(NOT phi, NOT psi)) disagree on any trace.

#### Test: Release — psi holds forever (LTL)
- **Strategy**: Known-answer
- **Invariant**: (phi R psi) is true when psi holds at every position and phi never holds.
- **Reference**: Pnueli (1977); Manna & Pnueli.
- **Verifications**: 1
- **Expected values**: true.

#### Test: Since — past-time (LTL)
- **Strategy**: Known-answer
- **Invariant**: Since(phi, psi) at position i is true if psi held at some past j and phi held at all positions since.
- **Reference**: Manna & Pnueli.
- **Verifications**: 1
- **Subtests**: trace=[psi,phi,phi,phi]. Since(phi,psi) at position 3.
- **Expected values**: true.

#### Test: Since — psi never held (LTL)
- **Strategy**: Known-answer
- **Invariant**: Since(phi, psi) is false if psi never held in the past.
- **Reference**: Manna & Pnueli. Strong since.
- **Verifications**: 1
- **Expected values**: false.

---

### 1.6 math/sets/ — Set operations (Strategy A: exhaustive + B: known-answer)

#### Test: SetUnionCommutativity
- **Strategy**: Exhaustive
- **Invariant**: Union(A, B) == Union(B, A) for all pairs of subsets
- **Reference**: Halmos "Naive Set Theory", Chapter 3
- **Verifications**: 256 (16 x 16 subsets of {1,2,3,4})
- **Subtests**: All 256 ordered pairs (A, B). Verify Equals(Union(A, B), Union(B, A)).
- **Failure criterion**: Any pair where equality fails.

#### Test: SetIntersectCommutativity
- **Strategy**: Exhaustive
- **Invariant**: Intersect(A, B) == Intersect(B, A)
- **Reference**: Halmos, Chapter 4
- **Verifications**: 256
- **Failure criterion**: Any pair where equality fails.

#### Test: SetUnionAssociativity
- **Strategy**: Exhaustive
- **Invariant**: Union(Union(A, B), C) == Union(A, Union(B, C))
- **Reference**: Halmos, Chapter 3
- **Verifications**: 512 (8 x 8 x 8 subsets of {1,2,3})
- **Failure criterion**: Any triple where equality fails.

#### Test: SetIntersectAssociativity
- **Strategy**: Exhaustive
- **Invariant**: Intersect(Intersect(A, B), C) == Intersect(A, Intersect(B, C))
- **Reference**: Halmos, Chapter 4
- **Verifications**: 512

#### Test: SetDifferenceCorrectness
- **Strategy**: Exhaustive
- **Invariant**: Difference(A, B) contains exactly elements in A not in B.
- **Reference**: Halmos, Chapter 5
- **Verifications**: 256
- **Subtests**: All 256 pairs. Verify element-by-element.
- **Failure criterion**: Any element of result is in B, or any element of A not in B is missing.

#### Test: SetSymmetricDifferenceCorrectness
- **Strategy**: Exhaustive
- **Invariant**: SymmetricDifference(A, B) == Union(Difference(A, B), Difference(B, A))
- **Reference**: Halmos, Chapter 5
- **Verifications**: 256

#### Test: SetSubsetReflexivity
- **Strategy**: Exhaustive
- **Invariant**: Subset(A, A) == true for every set
- **Reference**: Halmos, Chapter 1
- **Verifications**: 16

#### Test: SetSubsetAntisymmetry
- **Strategy**: Exhaustive
- **Invariant**: If Subset(A, B) && Subset(B, A) then Equals(A, B)
- **Reference**: Halmos, Chapter 1
- **Verifications**: 256 (only diagonal pairs satisfy both conditions)

#### Test: SetUnionIdentity
- **Strategy**: Exhaustive
- **Invariant**: Union(A, ∅) == A for all A
- **Reference**: Halmos, Chapter 3
- **Verifications**: 16

#### Test: SetIntersectAnnihilation
- **Strategy**: Exhaustive
- **Invariant**: Intersect(A, ∅) == ∅ for all A
- **Reference**: Halmos, Chapter 4
- **Verifications**: 16

#### Test: SetDifferenceWithEmpty
- **Strategy**: Exhaustive
- **Invariant**: Difference(A, ∅) == A and Difference(∅, A) == ∅
- **Reference**: Halmos, Chapter 5
- **Verifications**: 32

#### Test: SetContainsCorrectness
- **Strategy**: Known-answer
- **Invariant**: Contains returns true iff element is in the set
- **Reference**: Halmos, Chapter 1
- **Verifications**: 3
- **Expected values**: Contains({1,2,3}, 2)=true, Contains({1,2,3}, 4)=false, Contains(∅, 1)=false.

#### Test: SetEqualsReflexivity
- **Strategy**: Exhaustive
- **Invariant**: Equals(A, A) == true for every set
- **Reference**: Halmos, Chapter 1
- **Verifications**: 16

#### Test: SetEqualsSymmetry
- **Strategy**: Exhaustive
- **Invariant**: Equals(A, B) == Equals(B, A)
- **Reference**: Halmos, Chapter 1
- **Verifications**: 256

---

### 1.7 fuzzy/ — Axioms (Strategy A: exhaustive grid)

Grid: step=0.05, 21 points [0.0, 0.05, ..., 1.0]. Pairs: 441. Triples: 9261.

#### Test: FuzzyTNormCommutativity
- **Strategy**: Exhaustive grid
- **Invariant**: |t(a, b) - t(b, a)| < 1e-15 for all t-norms and all grid pairs
- **Reference**: Klement, Mesiar & Pap "Triangular Norms" (2000), Definition 1.1
- **Verifications**: 1323 (441 pairs x 3 t-norms: Min, Product, Lukasiewicz)
- **Prerequisite**: math/fuzzy/examples/properties_test.go
- **Failure criterion**: Any pair where |t(a, b) - t(b, a)| >= 1e-15.

#### Test: FuzzyTNormAssociativity
- **Strategy**: Exhaustive grid
- **Invariant**: |t(t(a, b), c) - t(a, t(b, c))| < 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.1
- **Verifications**: 27783 (9261 triples x 3 t-norms)
- **Failure criterion**: Any triple where difference >= 1e-15.

#### Test: FuzzyTNormMonotonicity
- **Strategy**: Exhaustive grid
- **Invariant**: For a1 <= a2, t(a1, b) <= t(a2, b) + 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.1
- **Verifications**: ~13800 (~4600 ordered pairs per t-norm x 3)
- **Failure criterion**: Any case where t(a1, b) > t(a2, b) + 1e-15.

#### Test: FuzzyTNormIdentity
- **Strategy**: Exhaustive grid
- **Invariant**: |t(a, 1.0) - a| < 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.1
- **Verifications**: 63 (21 x 3)

#### Test: FuzzyTConormCommutativity
- **Strategy**: Exhaustive grid
- **Invariant**: |s(a, b) - s(b, a)| < 1e-15 for all t-conorms (Max, ProbabilisticSum, BoundedSum)
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.13
- **Verifications**: 1323

#### Test: FuzzyTConormAssociativity
- **Strategy**: Exhaustive grid
- **Invariant**: |s(s(a, b), c) - s(a, s(b, c))| < 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.13
- **Verifications**: 27783

#### Test: FuzzyTConormMonotonicity
- **Strategy**: Exhaustive grid
- **Invariant**: For a1 <= a2, s(a1, b) <= s(a2, b) + 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.13
- **Verifications**: ~13800

#### Test: FuzzyTConormIdentity
- **Strategy**: Exhaustive grid
- **Invariant**: |s(a, 0.0) - a| < 1e-15
- **Reference**: Klement, Mesiar & Pap (2000), Definition 1.13
- **Verifications**: 63

#### Test: FuzzyLukasiewiczBoundaryAtZero
- **Strategy**: Known-answer
- **Invariant**: Luk(0.3, 0.6) = max(0.3+0.6-1, 0) = 0.0 exact.
- **Reference**: Klement, Mesiar & Pap (2000), Example 1.7
- **Verifications**: 1
- **Expected values**: 0.0.

#### Test: FuzzyLukasiewiczBoundaryAtTransition
- **Strategy**: Known-answer
- **Invariant**: Transition from 0 to positive at a+b=1.
- **Reference**: Klement, Mesiar & Pap (2000), Example 1.7
- **Verifications**: 2
- **Expected values**: Luk(0.5, 0.5)=0.0, Luk(0.5, 0.51)=0.01.

#### Test: FuzzyProductSubnormalStability
- **Strategy**: Adversarial
- **Invariant**: Product(1e-300, 1e-300) >= 0 and not NaN.
- **Reference**: IEEE 754-2019, §7.5
- **Verifications**: 1
- **Expected values**: >= 0 (likely 0.0 due to underflow).

#### Test: FuzzyTriangularPeak
- **Strategy**: Known-answer
- **Invariant**: Triangular(a, b, c) at x=b equals 1.0.
- **Reference**: Zadeh (1965); Klir & Yuan Ch. 2
- **Verifications**: 3
- **Expected values**: mu(b)=1.0 for Triangular(1,5,9), Triangular(0,0,1), Triangular(0,1,1).

#### Test: FuzzyTriangularZeros
- **Strategy**: Known-answer
- **Invariant**: mu(a)=0.0, mu(c)=0.0.
- **Reference**: Zadeh (1965)
- **Verifications**: 2
- **Expected values**: 0.0 at both endpoints.

#### Test: FuzzyTriangularOutOfRange
- **Strategy**: Known-answer
- **Invariant**: mu(a-1)=0.0, mu(c+1)=0.0.
- **Reference**: Zadeh (1965)
- **Verifications**: 2
- **Expected values**: 0.0 outside support.

#### Test: FuzzyTrapezoidalPlateau
- **Strategy**: Known-answer
- **Invariant**: Trapezoidal(a,b,c,d) at b, c, and (b+c)/2 equals 1.0.
- **Reference**: Klir & Yuan Ch. 2
- **Verifications**: 3
- **Expected values**: 1.0 at all three points.

#### Test: FuzzyGaussianAtCenter
- **Strategy**: Known-answer
- **Invariant**: Gaussian(center, sigma) at center equals 1.0.
- **Reference**: Klir & Yuan Ch. 2
- **Verifications**: 1
- **Expected values**: 1.0 exact (exp(0)=1).

#### Test: FuzzyGaussianSymmetry
- **Strategy**: Exhaustive grid
- **Invariant**: |mu(center+x) - mu(center-x)| < 1e-15 for Gaussian.
- **Reference**: Klir & Yuan Ch. 2
- **Verifications**: 21

#### Test: FuzzyDefuzzificationOutputInRange
- **Strategy**: Exhaustive
- **Invariant**: For each defuzzification method, output falls within [xMin, xMax].
- **Reference**: Klir & Yuan Ch. 11
- **Verifications**: 20 (5 methods x 4 distributions: uniform, left-skewed, right-skewed, bimodal)
- **Subtests**: Methods: Centroid, Bisector, MOM, LOM, SOM. x in [0, 10]. Result must be in [0, 10].
- **Failure criterion**: Any result outside [0, 10].

#### Test: FuzzyFuzzifyCorrectness
- **Strategy**: Known-answer
- **Invariant**: Fuzzify(fn, x) == fn(x) for all built-in membership functions.
- **Reference**: Klir & Yuan Ch. 2; trivially correct — direct delegation to MembershipFn.
- **Verifications**: 16 (4 functions × 4 x values)
- **Subtests**: Triangular(1,5,9), Trapezoidal(1,3,7,9), Gaussian(5,2), Sigmoid(5,1) each at x = {0, 3, 5, 10}.
- **Failure criterion**: |Fuzzify(fn, x) - fn(x)| >= 1e-15.
- **Expected values**: Fuzzify(fn, x) == fn(x) exactly for each pair.

#### Test: FuzzyClipCorrectness
- **Strategy**: Known-answer
- **Invariant**: Clip(fn, level)(x) == min(fn(x), level).
- **Reference**: Mamdani alpha-cut (standard implication method); Klir & Yuan Ch. 11.
- **Verifications**: 16 (4 x values × 4 levels)
- **Subtests**: fn = Triangular(0,5,10). Levels [0.0, 0.3, 0.7, 1.0]. x values [0, 3, 5, 8].
- **Failure criterion**: |Clip(fn, level)(x) - min(fn(x), level)| >= 1e-15.

#### Test: FuzzyScaleCorrectness
- **Strategy**: Known-answer
- **Invariant**: Scale(fn, level)(x) == fn(x) * level.
- **Reference**: Larsen (1980) product implication; Klir & Yuan Ch. 11.
- **Verifications**: 18 (4 x values × 4 levels + 2 identity/annihilation checks)
- **Subtests**: fn = Triangular(0,5,10). Levels [0.0, 0.3, 0.7, 1.0]. x values [0, 3, 5, 8]. Additionally: Scale(fn, 1)(x) == fn(x) (identity), Scale(fn, 0)(x) == 0 (annihilation).
- **Failure criterion**: |Scale(fn, level)(x) - fn(x)*level| >= 1e-15.

#### Test: FuzzyClipVsScaleShapePreservation
- **Strategy**: Known-answer
- **Invariant**: Clip(fn, level)(x) >= Scale(fn, level)(x) for all x where fn(x) > 0, because min(a, b) >= a*b for a,b ∈ [0,1].
- **Reference**: Klir & Yuan Ch. 11 (Mamdani vs Larsen implication comparison).
- **Verifications**: 12 (3 levels × 4 x values with fn(x) > 0)
- **Subtests**: fn = Triangular(0,5,10). Levels [0.3, 0.5, 0.7]. x values [2, 4, 5, 8].
- **Failure criterion**: Clip(fn, level)(x) < Scale(fn, level)(x) for any x where fn(x) > 0.

#### Test: FuzzyAggregateMaxCorrectness
- **Strategy**: Known-answer
- **Invariant**: AggregateMax(f1, f2, ...)(x) == max(f1(x), f2(x), ...).
- **Reference**: Klir & Yuan Ch. 3 (fuzzy set union via max).
- **Verifications**: 10 (5 x values × 2 configurations: 2 functions and 3 functions)
- **Subtests**: Config 1: f1=Triangular(0,3,6), f2=Triangular(4,7,10) at x={0,3,5,7,10}. Config 2: f1=Triangular(0,3,6), f2=Triangular(3,5,7), f3=Triangular(6,8,10) at x={0,3,5,8,10}.
- **Failure criterion**: |AggregateMax(fns...)(x) - max(fn_i(x))| >= 1e-15.

#### Test: FuzzySampleCorrectness
- **Strategy**: Known-answer
- **Invariant**: Sample(fn, lo, hi, n) produces n evenly-spaced (x, y) points.
- **Reference**: Klir & Yuan Ch. 11 (discretization for defuzzification).
- **Verifications**: 8
- **Subtests**: fn=Triangular(0,5,10). Sample(fn, 0, 10, 11): first=(0, 0.0), last=(10, 0.0), step=1.0, count=11. Sample(fn, 0, 10, 1): returns [(0, fn(0))]. Sample(fn, 2, 8, 4): step=2.0, points at x={2,4,6,8}.
- **Failure criterion**: Point count differs from n, first point x != lo, last point x != hi, step != (hi-lo)/(n-1) for n>=2.

---

### 1.8 stats/ — Distributions (Strategy B: known-answer + adversarial)

Tolerance: 1e-9 for PDF/CDF, 1e-12 for Mean/Variance (exact values).

#### Test: StatsNormalDistribution
- **Strategy**: Known-answer
- **Invariant**: Normal(0, 1) matches reference values.
- **Reference**: NIST; Casella & Berger Table A.
- **Verifications**: 5
- **Expected values**:
  - PDF(0) = 0.398942280401433 (1e-9)
  - CDF(0) = 0.5 (1e-9)
  - CDF(1.96) = 0.975002104852 (1e-9)
  - Mean = 0.0 (1e-12)
  - Variance = 1.0 (1e-12)

#### Test: StatsNormalCDFSymmetry
- **Strategy**: Known-answer
- **Invariant**: CDF(-x) = 1 - CDF(x) for standard normal.
- **Reference**: NIST; normal distribution symmetry.
- **Verifications**: 1
- **Expected values**: CDF(-1.96) = 0.024997895148 (1e-9). |CDF(-1.96) - (1 - CDF(1.96))| < 1e-12.

#### Test: StatsExponentialDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger §3.3.
- **Verifications**: 4
- **Expected values**: PDF(0)=2.0, CDF(1)=0.864664716763387, Mean=0.5, Variance=0.25.

#### Test: StatsBetaDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger §3.3; NIST.
- **Verifications**: 3
- **Expected values**: Mean=2/7≈0.285714285714, Variance=10/392≈0.025510204082, CDF(0.5)=0.890625.

#### Test: StatsStudentTDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger Table A; NIST.
- **Verifications**: 5
- **Expected values**: PDF(0)=0.379609319956, CDF(0)=0.5, CDF(2.571)≈0.975 (1e-3), Mean=0.0, Variance=5/3≈1.666666666667.

#### Test: StatsChiSquaredDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger §5.3; NIST.
- **Verifications**: 3
- **Expected values**: Mean=3.0, Variance=6.0, CDF(7.815)≈0.95 (1e-3).

#### Test: StatsPoissonDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger §3.2.
- **Verifications**: 4
- **Expected values**: PMF(0)=0.049787068368, PMF(3)=0.224041807660, Mean=3.0, Variance=3.0.

#### Test: StatsBinomialDistribution
- **Strategy**: Known-answer
- **Reference**: Casella & Berger §3.2.
- **Verifications**: 3
- **Expected values**: PMF(3)=0.266827932, Mean=3.0, Variance=2.1.

#### Test: StatsUniformDistribution
- **Strategy**: Known-answer
- **Invariant**: Uniform(2, 8) matches reference values.
- **Reference**: Casella & Berger §3.3.
- **Verifications**: 4
- **Expected values**: PDF(5)=0.166666666667 (1e-9), CDF(5)=0.5, Mean=5.0, Variance=3.0.

#### Test: StatsGammaDistribution
- **Strategy**: Known-answer
- **Invariant**: Gamma(3, 2) (shape=3, rate=2) matches reference values.
- **Reference**: Casella & Berger §3.3; NIST.
- **Verifications**: 3
- **Expected values**: Mean=1.5, Variance=0.75, CDF(1) verify via regularized incomplete gamma.

#### Test: StatsLognormalDistribution
- **Strategy**: Known-answer
- **Invariant**: Lognormal(0, 1) matches reference values.
- **Reference**: Casella & Berger §3.3.
- **Verifications**: 4
- **Expected values**: Mean=1.648721270700 (1e-9), Variance=4.670774270471 (1e-9), CDF(1)=0.5, PDF(1) verify.

#### Test: StatsWeibullDistribution
- **Strategy**: Known-answer
- **Invariant**: Weibull(2, 1) (k=2, λ=1) matches reference values.
- **Reference**: Casella & Berger §3.3; NIST.
- **Verifications**: 5
- **Expected values**: PDF(1)=0.735758882343 (1e-9), CDF(1)=0.632120558829 (1e-9), Mean=0.886226925453 (1e-9), Variance=0.214601836603 (1e-9).

#### Test: StatsFDistribution
- **Strategy**: Known-answer
- **Invariant**: F(5, 10) matches reference values.
- **Reference**: Casella & Berger §5.3; NIST.
- **Verifications**: 2
- **Expected values**: Mean=1.25 (1e-12), Variance=1.354166666667 (1e-9).

#### Test: StatsGumbelDistribution
- **Strategy**: Known-answer
- **Invariant**: Gumbel(0, 1) matches reference values.
- **Reference**: NIST; Kotz & Nadarajah (2000).
- **Verifications**: 3
- **Expected values**: Mean=0.577215664902 (1e-9), Variance=1.644934066848 (1e-9), CDF(0)=0.367879441171 (1e-9).

#### Test: StatsParetoDistribution
- **Strategy**: Known-answer
- **Invariant**: Pareto(1, 3) (xm=1, α=3) matches reference values.
- **Reference**: Casella & Berger §3.3.
- **Verifications**: 3
- **Expected values**: CDF(2)=0.875, Mean=1.5, Variance=0.75.

#### Test: StatsGeometricDistribution
- **Strategy**: Known-answer
- **Invariant**: Geometric(0.3) (failures-before-success) matches reference values.
- **Reference**: Casella & Berger §3.2.
- **Verifications**: 4
- **Expected values**: PMF(0)=0.3, PMF(2)=0.147, Mean=2.333333333333 (1e-9), Variance=7.777777777778 (1e-9).

#### Test: StatsHypergeometricDistribution
- **Strategy**: Known-answer
- **Invariant**: Hypergeometric(50, 10, 5) (N=50, K=10, n=5) matches reference values.
- **Reference**: Casella & Berger §3.2.
- **Verifications**: 2
- **Expected values**: Mean=1.0, Variance=0.734693877551 (1e-9).

#### Test: StatsNegativeBinomialDistribution
- **Strategy**: Known-answer
- **Invariant**: NegativeBinomial(5, 0.4) (r=5, p=0.4) matches reference values.
- **Reference**: Casella & Berger §3.2.
- **Verifications**: 2
- **Expected values**: Mean=7.5, Variance=18.75.

#### Test: StatsWelfordCatastrophicCancellation
- **Strategy**: Adversarial
- **Invariant**: Welford's algorithm resists catastrophic cancellation.
- **Reference**: Welford (1962); Knuth TAOCP Vol 2 §4.2.2.
- **Verifications**: 1
- **Subtests**: Data: [1e8+1, 1e8+2, 1e8+3]. Population variance = 2/3 ≈ 0.666666666667.
- **Failure criterion**: |computed - 0.666667| >= 1e-9.
- **Expected values**: 0.666666666667.

#### Test: StatsWelfordIdenticalData
- **Strategy**: Adversarial
- **Invariant**: Variance of identical values is exactly zero.
- **Verifications**: 1
- **Expected values**: Data [5,5,5,5,5]. Variance = 0.0 exact.

#### Test: StatsWelfordSingleDatum
- **Strategy**: Adversarial
- **Invariant**: Population variance of single observation is zero.
- **Verifications**: 1
- **Expected values**: Data [42]. Variance = 0.0 exact.

#### Test: StatsWindowedStatsConsistency
- **Strategy**: Adversarial
- **Invariant**: WindowedStats matches full recomputation after each insertion.
- **Reference**: Sliding window correctness contract.
- **Verifications**: 100
- **Subtests**: 100 values (fixed seed), window=10. Compare Mean and Variance against recomputed values.
- **Failure criterion**: |WindowedStats.Mean() - recomputed| >= 1e-9 or |WindowedStats.Variance() - recomputed| >= 1e-9.

#### Test: StatsHypothesisTTestKnownOutcome
- **Strategy**: Known-answer
- **Invariant**: T-test correctly does not reject H0.
- **Reference**: Casella & Berger §8.3.
- **Verifications**: 4
- **Subtests**: Data [2.1, 2.3, 1.9, 2.0, 2.2], H0: mu=2.0.
- **Expected values**: mean=2.1, s≈0.158113883008, t≈1.414213562373, p≈0.230 (> 0.05, do not reject).

#### Test: StatsHypothesisTTestRejection
- **Strategy**: Known-answer
- **Invariant**: T-test correctly rejects H0 when true mean is far from hypothesized.
- **Reference**: Casella & Berger §8.3.
- **Verifications**: 2
- **Subtests**: Data [10.1, 10.3, 9.9, 10.2, 10.4, 10.0, 10.5], H0: mu=9.0.
- **Expected values**: p < 0.001. Reject H0.

#### Test: StatsHypothesisWelchTTest
- **Strategy**: Known-answer
- **Invariant**: Welch's two-sample t-test correctly rejects H0 of equal means.
- **Reference**: Casella & Berger §8.3 (Welch-Satterthwaite degrees of freedom).
- **Verifications**: 2
- **Subtests**: Sample1 [5.1, 5.3, 4.9], Sample2 [3.1, 3.2, 2.9]. Different means, similar variance.
- **Failure criterion**: p >= 0.01 (should reject H0).
- **Expected values**: p < 0.01. Reject H0 (equal means).

#### Test: StatsHypothesisChiSquaredGoodnessOfFit
- **Strategy**: Known-answer
- **Invariant**: Chi-squared goodness-of-fit test detects non-uniform distribution.
- **Reference**: Casella & Berger §10.3.
- **Verifications**: 3
- **Subtests**: Observed [50, 30, 20], Expected [33.3, 33.3, 33.4]. Statistic = Σ(O-E)²/E, df=2.
- **Failure criterion**: Statistic computation wrong or p >= 0.05.
- **Expected values**: Statistic ≈ 13.51 (1e-1), p < 0.05. Reject H0 (uniform).

#### Test: StatsHypothesisKSTest
- **Strategy**: Known-answer
- **Invariant**: KS test correctly accepts matching distribution and rejects mismatched distribution.
- **Reference**: NIST; Conover (1999).
- **Verifications**: 2
- **Subtests**: Sample from standard normal (fixed seed, n=100). KS against Normal(0,1): do NOT reject (p > 0.05). KS against Normal(5,1): reject (p < 0.001).
- **Failure criterion**: False rejection of matching distribution, or false acceptance of mismatched distribution.
- **Expected values**: Same-distribution p > 0.05. Different-distribution p < 0.001.

#### Test: StatsHypothesisANOVA
- **Strategy**: Known-answer
- **Invariant**: ANOVA correctly identifies identical vs. different group means.
- **Reference**: Casella & Berger §11.2.
- **Verifications**: 4
- **Subtests**: Identical groups [5,6,7], [5,6,7], [5,6,7]: F=0, p=1.0, do not reject. Different groups [1,2,3], [10,11,12], [20,21,22]: F very large, p < 0.001, reject.
- **Failure criterion**: Wrong F-statistic or wrong rejection decision.
- **Expected values**: Identical: F=0.0, p=1.0. Different: p < 0.001. Reject H0.

#### Test: StatsHypothesisFisherExact
- **Strategy**: Known-answer
- **Invariant**: Fisher exact test on 2×2 table yields correct p-value.
- **Reference**: Fisher (1922); Agresti (2002) §3.5.
- **Verifications**: 2
- **Subtests**: Table [[8,2],[1,9]]. Two-tailed p-value via hypergeometric calculation.
- **Failure criterion**: p-value deviates from reference by > 1e-6.
- **Expected values**: Verify against exact hypergeometric computation. p ≈ 0.0011 (1e-4).

#### Test: StatsBayesPrecision
- **Strategy**: Known-answer
- **Invariant**: Bayes' theorem diagnostic testing yields correct posterior.
- **Reference**: Kahneman & Tversky; Casella & Berger §1.3.
- **Verifications**: 2
- **Subtests**: P(Disease)=0.001, P(Pos|Disease)=0.99, P(Pos|NoDisease)=0.05.
- **Expected values**: P(Positive)=0.05094 (1e-9), P(Disease|Positive)=0.019434041226 (1e-5).

---

### 1.9 math/graph/ (Strategy A: exhaustive + B: known-answer)

No existing examples/ tests. All tests are new.

#### Test: Directed adjacency exhaustive
- **Strategy**: Exhaustive
- **Invariant**: Neighbors(v) returns exactly the outgoing targets for all directed graphs on 3 vertices.
- **Reference**: CLRS Ch. 22; Diestel "Graph Theory".
- **Verifications**: 192 (64 graphs x 3 vertices)
- **Subtests**: All 2^6=64 directed graphs on {A, B, C}. For each, verify Neighbors(v) matches expected set.
- **Failure criterion**: Neighbors returns wrong set for any graph/vertex.

#### Test: Undirected symmetry exhaustive
- **Strategy**: Exhaustive
- **Invariant**: Adjacency is symmetric for all undirected graphs on 3 vertices.
- **Reference**: Diestel Definition 1.1.1.
- **Verifications**: 48 (8 graphs x 6 pair checks)
- **Subtests**: All 2^3=8 undirected graphs on {A, B, C}. Verify u in Neighbors(v) iff v in Neighbors(u).

#### Test: DAG topological sort exhaustive
- **Strategy**: Exhaustive
- **Invariant**: TopologicalSort produces valid ordering for all DAGs on 4 vertices.
- **Reference**: CLRS Ch. 22.4.
- **Verifications**: All acyclic orientations among 4096 possible directed graphs on 4 vertices.
- **Subtests**: For each DAG, verify for every edge (u,v): index(u) < index(v).

#### Test: DAG cycle rejection exhaustive
- **Strategy**: Exhaustive
- **Invariant**: All cyclic directed graphs on 4 vertices are rejected by DAG construction.
- **Reference**: CLRS Ch. 22.4.
- **Verifications**: All non-DAG directed graphs among 4096.

#### Test: BFS shortest path known
- **Strategy**: Known-answer
- **Invariant**: Dijkstra distances match CLRS values.
- **Reference**: CLRS Ch. 24.
- **Verifications**: 5
- **Subtests**: 5-vertex graph {S,A,B,C,D}. All weights=1.0. dist(S,S)=0, dist(S,A)=1, dist(S,B)=1, dist(S,C)=2, dist(S,D)=2.

#### Test: Centrality star graph
- **Strategy**: Known-answer
- **Invariant**: Centrality measures for star graph produce known values.
- **Reference**: Freeman (1979).
- **Verifications**: 6
- **Subtests**: Undirected star: center C, 4 leaves, weights=1.0.
  - DegreeCentrality(C) = 4/4 = 1.0
  - DegreeCentrality(leaf) = 1/4 = 0.25
  - BetweennessCentrality(C) = 6.0 (raw Brandes, undirected halving)
  - BetweennessCentrality(leaf) = 0.0
  - ClosenessCentrality(C) = 4/4 = 1.0
  - ClosenessCentrality(leaf) = 4/7

#### Test: Bipartite verification
- **Strategy**: Known-answer
- **Invariant**: K(2,3) is bipartite; K(3) is not.
- **Reference**: Diestel Proposition 1.6.1.
- **Verifications**: 2
- **Expected values**: IsBipartite(K23)=true, IsBipartite(K3)=false.

#### Test: MST known weight
- **Strategy**: Known-answer
- **Invariant**: MST total weight matches known value.
- **Reference**: CLRS Ch. 23.
- **Verifications**: 3
- **Subtests**: 4 vertices {A,B,C,D}. Edges: A-B(1), A-C(4), B-C(2), B-D(6), C-D(3). MST: A-B(1)+B-C(2)+C-D(3)=6.
- **Expected values**: edgeCount=3, totalWeight=6.0.

#### Test: Graph coloring validity
- **Strategy**: Known-answer
- **Invariant**: Greedy coloring assigns no same color to adjacent vertices.
- **Reference**: Diestel Ch. 5.
- **Verifications**: 3
- **Subtests**: K3 (ChromaticNumber=3), path P4 (ChromaticNumber=2). Verify no adjacent pair shares color.

#### Test: Matching known cardinality
- **Strategy**: Known-answer
- **Invariant**: Maximum bipartite matching has expected cardinality.
- **Reference**: Hopcroft & Karp (1973).
- **Verifications**: 2
- **Expected values**: |matching(K22)|=2, |matching(K23)|=2.

#### Test: Tree properties
- **Strategy**: Known-answer
- **Invariant**: Tree on n vertices has n-1 edges, is connected, is acyclic.
- **Reference**: Diestel Proposition 1.5.1.
- **Verifications**: 4
- **Expected values**: 5-node tree: EdgeCount=4, acyclic, all reachable from root, correct parent/leaf.

#### Test: BFS level order
- **Strategy**: Known-answer
- **Invariant**: BFS visits in non-decreasing distance order.
- **Reference**: CLRS Theorem 22.5.
- **Verifications**: 1
- **Subtests**: 6-vertex directed graph. BFS from source. Record visit order. Verify dist(visited[i]) <= dist(visited[i+1]) for all consecutive pairs.
- **Failure criterion**: Any vertex visited before a closer vertex.
- **Expected values**: Visit order respects non-decreasing distance from source.

#### Test: DFS all reachable
- **Strategy**: Known-answer
- **Invariant**: DFS visits all vertices in a connected component.
- **Reference**: CLRS Theorem 22.10.
- **Verifications**: 2
- **Subtests**: Connected graph K4: DFS from vertex 0 visits all 4 vertices. Disconnected graph {0-1-2} + {3-4}: DFS from 0 visits {0,1,2} only.
- **Failure criterion**: DFS misses a reachable vertex or includes an unreachable vertex.
- **Expected values**: Connected: visited count = 4. Disconnected from 0: visited = {0,1,2}, count = 3.

#### Test: SCC known decomposition
- **Strategy**: Known-answer
- **Invariant**: Tarjan SCC produces correct communicating classes.
- **Reference**: Tarjan (1972); CLRS Ch. 22.5.
- **Verifications**: 1
- **Subtests**: 8-vertex graph with edges forming 5 SCCs: {A,B,C} (cycle A→B→C→A), {D,E} (cycle D→E→D), {F}, {G}, {H} (singletons). Edges C→D, E→F, F→G, G→H connect components.
- **Failure criterion**: SCC count differs from 5, or any vertex assigned to wrong component.
- **Expected values**: 5 SCCs with compositions as listed.

#### Test: Bridges known
- **Strategy**: Known-answer
- **Invariant**: Bridge detection identifies edges whose removal disconnects the graph.
- **Reference**: Tarjan (1974).
- **Verifications**: 1
- **Subtests**: Graph with two triangles {A-B-C} and {D-E-F} connected by single edge C-D.
- **Failure criterion**: Bridge set differs from {C-D}.
- **Expected values**: Bridges = {C-D}. Count = 1.

#### Test: Articulation points known
- **Strategy**: Known-answer
- **Invariant**: AP detection identifies vertices whose removal disconnects the graph.
- **Reference**: Tarjan (1974).
- **Verifications**: 1
- **Subtests**: Same two-triangle graph connected by C-D.
- **Failure criterion**: Articulation point set differs from {C, D}.
- **Expected values**: APs = {C, D}. Count = 2.

#### Test: PageRank convergence
- **Strategy**: Known-answer
- **Invariant**: PageRank values sum to 1.0 and match hand-computed reference.
- **Reference**: Brin & Page (1998); Langville & Meyer (2004).
- **Verifications**: 5
- **Subtests**: 4-node directed graph: A→B, A→C, B→C, C→A, D→C. Damping factor d=0.85. Hand-computed PageRank: solve PR = (1-d)/n + d*M^T*PR where M is the column-stochastic link matrix. Node D has no outgoing links to other nodes in the graph (dangling node handled by uniform distribution).
- **Failure criterion**: |sum(PageRank) - 1.0| >= 1e-9 or any individual PageRank deviates from reference by > 1e-3.
- **Expected values**: sum(PR) = 1.0. PR(A) ≈ 0.372, PR(B) ≈ 0.196, PR(C) ≈ 0.394, PR(D) ≈ 0.038. Tolerance 1e-3.

#### Test: Max flow known
- **Strategy**: Known-answer
- **Invariant**: Edmonds-Karp returns correct max flow.
- **Reference**: CLRS Ch. 26.
- **Verifications**: 1
- **Subtests**: Network S→A(3), S→B(2), A→T(2), B→T(3), A→B(1). Derivation: path S→A→T bottleneck=2, path S→B→T bottleneck=2, total=4. Or: path S→A→B→T bottleneck=1, path S→A→T bottleneck=2, path S→B→T bottleneck=1, total varies by path selection. Max flow = 4.
- **Failure criterion**: Max flow differs from 4.0.
- **Expected values**: MaxFlow = 4.0. Tolerance 1e-9.

#### Test: Floyd-Warshall known
- **Strategy**: Known-answer
- **Invariant**: All-pairs shortest paths match hand-computed values.
- **Reference**: CLRS Ch. 25.2.
- **Verifications**: 9
- **Subtests**: 3-vertex directed graph. A->B(1), B->C(2), A->C(10). dist[A][C]=3 (via B, not direct 10).

#### Test: Dijkstra weighted shortest path known-answer
- **Strategy**: Known-answer
- **Invariant**: Dijkstra computes correct shortest weighted paths on a graph with varying positive edge weights.
- **Reference**: CLRS 4th ed., Ch. 22.3; Dijkstra, E.W. (1959). "A note on two problems in connexion with graphs."
- **Verifications**: 5
- **Prerequisite**: BFS shortest path test covers unweighted case
- **Subtests**:
  - 5-vertex directed graph {A,B,C,D,E}. Edges: A→B(4), A→C(2), C→B(1), B→D(3), B→E(1), C→D(5), D→E(2).
  - dist(A,A)=0, dist(A,C)=2 (direct), dist(A,B)=3 (A→C→B), dist(A,E)=4 (A→C→B→E), dist(A,D)=6 (A→C→B→D).
  - Derivation: Process A: tentative B=4, C=2. Process C(2): B=min(4, 2+1)=3, D=min(∞, 2+5)=7. Process B(3): D=min(7, 3+3)=6, E=min(∞, 3+1)=4. Process E(4): no improvement. Process D(6): E=min(4, 6+2)=4.
- **Failure criterion**: Any distance differs from reference value.
- **Expected values**: [A=0, B=3, C=2, D=6, E=4]. Tolerance 1e-9.

#### Test: Bellman-Ford negative edges known-answer
- **Strategy**: Known-answer
- **Invariant**: Bellman-Ford computes shortest paths in graphs with negative edge weights (but no negative cycles) and detects negative cycles when present.
- **Reference**: CLRS 4th ed., Ch. 22.1; Bellman, R. (1958). "On a routing problem."
- **Verifications**: 6
- **Prerequisite**: Dijkstra test covers non-negative weights
- **Subtests**:
  - 4-vertex directed graph {A,B,C,D}. Edges: A→B(1), A→C(4), B→C(-2), B→D(3), C→D(1).
  - dist(A,A)=0, dist(A,B)=1, dist(A,C)=-1 (A→B→C via negative edge), dist(A,D)=0 (A→B→C→D = 1+(-2)+1).
  - Negative cycle detection: graph with edges A→B(1), B→C(-3), C→A(1). Cycle weight = 1+(-3)+1 = -1 < 0. Must return error or indicate negative cycle.
- **Failure criterion**: Any distance differs from reference, or negative cycle not detected.
- **Expected values**: [A=0, B=1, C=-1, D=0]. Negative cycle: error returned. Tolerance 1e-9.

#### Test: AllPaths between two vertices known-answer
- **Strategy**: Known-answer
- **Invariant**: AllPaths returns all distinct simple paths between source and target vertex.
- **Reference**: Diestel "Graph Theory" 5th ed., Ch. 1 — path enumeration.
- **Verifications**: 3
- **Prerequisite**: BFS/DFS tests cover reachability
- **Subtests**:
  - Diamond directed graph: A→B, A→C, B→D, C→D. AllPaths(A,D) returns 2 paths: [A,B,D] and [A,C,D].
  - Linear path: A→B→C. AllPaths(A,C) returns 1 path: [A,B,C].
  - Disconnected: A→B, C→D (no path A to D). AllPaths(A,D) returns 0 paths.
- **Failure criterion**: Path count differs from expected, or any returned path is not a valid simple path from source to target.
- **Expected values**: Diamond: 2 paths. Linear: 1 path. Disconnected: 0 paths.

---

### 1.10 math/fsm/ (Strategy A: exhaustive + B: known-answer)

No existing examples/ tests. All tests are new.

#### Test: DFA acceptance exhaustive
- **Strategy**: Exhaustive
- **Invariant**: DFA accepts iff string ends in "1" for all binary strings up to length 4.
- **Reference**: Hopcroft, Motwani & Ullman Ch. 2.
- **Verifications**: 31 (all binary strings of length 0-4)
- **Subtests**: DFA {Q0(initial), Q1}. Q0 on "0"->Q0, Q0 on "1"->Q1, Q1 on "0"->Q0, Q1 on "1"->Q1. 15 accepted (end in "1"), 16 rejected.

#### Test: Determinism check
- **Strategy**: Known-answer
- **Invariant**: Deterministic() distinguishes DFA from NFA.
- **Reference**: Hopcroft et al. Definition 2.1 vs 2.3.
- **Verifications**: 2
- **Subtests**: Complete DFA with one transition per (state, symbol): Deterministic()=true. NFA with two transitions from state A on symbol "x": Deterministic()=false.
- **Failure criterion**: DFA classified as non-deterministic, or NFA classified as deterministic.
- **Expected values**: DFA: true. NFA: false.

#### Test: Reachability disconnected
- **Strategy**: Known-answer
- **Invariant**: Reachable excludes structurally unreachable states.
- **Reference**: Hopcroft et al. Ch. 2.
- **Verifications**: 2
- **Subtests**: States {A,B,C,D}, transitions A→B→C, D has no incoming from reachable states. Reachable from A = {A,B,C}. D is unreachable.
- **Failure criterion**: Unreachable state included in reachable set, or reachable state excluded.
- **Expected values**: Reachable(A) = {A,B,C}. IsReachable(A,D) = false.

#### Test: Binary divisible by 3
- **Strategy**: Known-answer
- **Invariant**: 3-state DFA correctly recognizes binary numbers divisible by 3.
- **Reference**: Hopcroft et al. Ch. 2.
- **Verifications**: 4
- **Expected values**: "110"(=6) accepted, "111"(=7) rejected, "0" accepted, "" accepted.

#### Test: Even number of 'a'
- **Strategy**: Known-answer
- **Invariant**: 2-state DFA accepts strings with even count of 'a'.
- **Reference**: Hopcroft et al. Ch. 2.
- **Verifications**: 7
- **Expected values**: "" accepted, "b" accepted, "aa" accepted, "ab" rejected, "ba" rejected, "aab" accepted, "aba" accepted.

#### Test: Transition function determinism
- **Strategy**: Known-answer
- **Verifications**: 3
- **Subtests**: Single transition: true. Complete DFA: true. Duplicate transition: false.

#### Test: Dead states
- **Strategy**: Known-answer
- **Invariant**: DeadStates returns non-accepting states from which no accepting state is reachable.
- **Reference**: Hopcroft et al. Ch. 2.
- **Verifications**: 2
- **Expected values**: {A,B,C,Dead} where Dead self-loops and is non-accepting: DeadStates=[Dead]. All-reachable-to-accepting: DeadStates=[].

#### Test: Completeness check
- **Strategy**: Known-answer
- **Invariant**: IsComplete returns true iff every (state, event) pair has a transition.
- **Reference**: Hopcroft et al. Ch. 2.
- **Verifications**: 2
- **Subtests**: DFA with 2 states, alphabet {0,1}, missing transition (B,1): IsComplete()=false. Same DFA with all 4 transitions defined: IsComplete()=true.
- **Failure criterion**: Incomplete DFA reported as complete, or vice versa.
- **Expected values**: Incomplete: false. Complete: true.

---

### 1.11 math/markov/ (Strategy B: known-answer + adversarial)

No existing examples/ tests. All tests are new.

#### Test: Weather chain steady state
- **Strategy**: Known-answer
- **Invariant**: Stationary distribution matches balance equations.
- **Reference**: Norris "Markov Chains" Ch. 1.
- **Verifications**: 2
- **Subtests**: States [Sunny, Rainy]. P=[[0.8, 0.2], [0.4, 0.6]].
  - Derivation: 0.2·π_S = 0.4·π_R, π_S + π_R = 1. π_S = 2/3, π_R = 1/3.
- **Expected values**: π_Sunny = 2/3, π_Rainy = 1/3. Tolerance 1e-9.

#### Test: N-step transition
- **Strategy**: Known-answer
- **Invariant**: P^2 computed correctly by matrix-vector multiplication.
- **Reference**: Norris Ch. 1.
- **Verifications**: 4
- **Expected values**:
  - StepN(Sunny, 2): Sunny=0.72, Rainy=0.28.
  - StepN(Rainy, 2): Sunny=0.56, Rainy=0.44.

#### Test: Gambler's ruin absorption
- **Strategy**: Known-answer
- **Invariant**: Absorption probabilities and mean times match closed-form solutions.
- **Reference**: Kemeny & Snell Ch. 3.
- **Verifications**: 6
- **Subtests**: States [S0,S1,S2,S3]. S0, S3 absorbing. P(1→0)=P(1→2)=P(2→1)=P(2→3)=0.5.
  - P(S1→S0)=2/3, P(S1→S3)=1/3, P(S2→S0)=1/3, P(S2→S3)=2/3.
  - MeanAbsorption(S1)=2.0, MeanAbsorption(S2)=2.0.

#### Test: Classification communicating classes
- **Strategy**: Known-answer
- **Invariant**: Classify correctly identifies transient/recurrent/absorbing via SCC.
- **Reference**: Norris §1.3-1.5; Kemeny & Snell Ch. 2-3.
- **Verifications**: 5
- **Subtests**: 5-state chain. {A,B} transient, {C,D} transient, {E} absorbing.

#### Test: Period
- **Strategy**: Known-answer
- **Invariant**: Period via BFS-discrepancy for periodic chain.
- **Reference**: Norris Definition 1.2.1.
- **Verifications**: 2
- **Subtests**: 2-state oscillating A↔B. Period(A)=2. IsErgodic()=false.

#### Test: Mean first passage
- **Strategy**: Known-answer
- **Invariant**: Mean first passage time matches closed-form solution.
- **Reference**: Kemeny & Snell §4.4.
- **Verifications**: 3
- **Expected values**: MFP(Sunny→Rainy)=5.0, MFP(Rainy→Sunny)=2.5, MFP(Sunny→Sunny)=0.0.

#### Test: Simulate stochastic consistency
- **Strategy**: Known-answer + statistical
- **Invariant**: Empirical stationary distribution from simulation approximates theoretical π.
- **Reference**: Norris §1.1 (Ergodic theorem for Markov chains).
- **Verifications**: 3
- **Subtests**: Weather chain P=[[0.8, 0.2], [0.4, 0.6]]. Theoretical steady state: π_Sunny=2/3, π_Rainy=1/3. Run 10000 simulations (fixed seed) of 100 steps each. Empirical frequency of each state at step 100.
- **Failure criterion**: |empirical_freq(state) - π(state)| >= 0.05 for any state (statistical tolerance for 10000 samples).
- **Expected values**: Empirical π_Sunny ≈ 0.667 ± 0.05, π_Rainy ≈ 0.333 ± 0.05.

#### Test: Periodic chain steady state (adversarial)
- **Strategy**: Adversarial
- **Invariant**: Steady state exists for periodic chain despite P^n not converging.
- **Reference**: Norris Theorem 1.7.2.
- **Verifications**: 2
- **Expected values**: 2-state oscillating. π_A=0.5, π_B=0.5. Tolerance 1e-9.

#### Test: Nearly-absorbing convergence (adversarial)
- **Strategy**: Adversarial
- **Invariant**: Steady state converges with near-degenerate probabilities.
- **Reference**: Norris Ch. 1.
- **Verifications**: 2
- **Subtests**: P=[[0.9999, 0.0001], [0.0001, 0.9999]]. By symmetry: π_A=π_B=0.5.
- **Failure criterion**: Steady() errors or deviates from 0.5 by > 1e-9.

#### Test: Invalid row rejection (adversarial)
- **Strategy**: Adversarial
- **Invariant**: NewChain rejects matrix where rows don't sum to 1.
- **Reference**: Norris Definition 1.1.1.
- **Verifications**: 1
- **Subtests**: [[0.5, 0.3], [0.4, 0.6]]. Row 0 sums to 0.8. Must return error.

---

### 1.12 engine/deductive/ (Strategy B: adversarial)

#### Test: Fixed point with cycles
- **Strategy**: Adversarial
- **Invariant**: Forward chaining terminates on cyclic rules due to monotonicity.
- **Reference**: Russell & Norvig "AIMA" Ch. 7-9, monotonic fixpoint theorem.
- **Verifications**: 4
- **Prerequisite**: engine/deductive/examples/examples_test.go
- **Subtests**:
  - "cyclic A-B-C": Rules A->B, B->C, C->A. Initial {A: true}. Steps <= 3. All 3 facts derived.
  - "self-referencing": Rule A->A. Initial {A: true}. Steps == 0, no new facts.
  - "mutual cycle": A->B, B->A. Initial {A: true}. Steps == 1, B derived.
  - "4-node cycle": A->B, B->C, C->D, D->A. Steps == 3.
- **Failure criterion**: Hangs past maxIterations, or steps exceed derivable facts count, or expected facts missing.

#### Test: Deep rule chain
- **Strategy**: Adversarial
- **Invariant**: Forward chaining propagates through 25-rule chain. Backward chaining respects depth limits.
- **Reference**: Russell & Norvig "AIMA" Ch. 7-9.
- **Verifications**: 4
- **Subtests**:
  - Forward: 25 rules V0->V1->...->V25. Derives V25 in 25 steps.
  - Backward depth=30: proves V25.
  - Backward depth=10: does NOT prove V25.
  - Backward depth=25: proves V25 (exact boundary).
- **Failure criterion**: Forward does not derive V25, or backward with sufficient depth fails, or backward with insufficient depth succeeds.

#### Test: Clone-on-attempt isolation
- **Strategy**: Adversarial
- **Invariant**: Failed rule attempts do not pollute the factbase.
- **Reference**: Russell & Norvig "AIMA" Ch. 9.
- **Verifications**: 3
- **Subtests**:
  - "failed AND rule": Condition (A AND B), A=true, B=false. After attempt, snapshot == initial.
  - "backward failed attempt": Rule chain fails to prove, snapshot unchanged.
  - "mixed success/failure": One rule fails, another succeeds. Only successful derivation persists.
- **Failure criterion**: Snapshot contains facts from failed attempts.

#### Test: Conflict resolution determinism
- **Strategy**: Adversarial
- **Invariant**: PriorityOrder fires all applicable rules. FirstMatch fires one per pass.
- **Reference**: Russell & Norvig "AIMA" Ch. 7.
- **Verifications**: 5
- **Subtests**:
  - PriorityOrder fires both same-priority rules.
  - FirstMatch eventually derives all facts.
  - Priority ordering: lower priority number fires first in trace.
  - Three rules same priority all fire.
  - FirstMatch with three rules takes three passes.

---

### 1.13 engine/bayesian/ (Strategy A: exhaustive + B: known-answer)

Rain network CPTs:
- P(Rain=t)=0.2, P(Sprinkler=t|Rain=t)=0.01, P(Sprinkler=t|Rain=f)=0.4
- P(WG=t|S=t,R=t)=0.99, P(WG=t|S=t,R=f)=0.9, P(WG=t|S=f,R=t)=0.8, P(WG=t|S=f,R=f)=0.0

#### Test: VE equals Enumeration for all queries
- **Strategy**: Exhaustive
- **Invariant**: VE and Enumeration produce identical posteriors for all queries and evidence subsets.
- **Reference**: Koller & Friedman "PGM" Ch. 9.
- **Verifications**: ~24 (3 query variables x ~8 evidence combinations)
- **Subtests**: For each query variable, for each evidence subset, verify |VE - Enum| <= 1e-9.
- **Failure criterion**: Any divergence > 1e-9.

#### Test: Elimination order invariance
- **Strategy**: Exhaustive
- **Invariant**: VE result does not depend on elimination order.
- **Reference**: Koller & Friedman "PGM" Ch. 9.
- **Verifications**: 2
- **Subtests**: Try all permutations of hidden variable elimination. Verify identical results.

#### Test: Hand-calculated Rain posterior
- **Strategy**: Known-answer
- **Invariant**: P(Rain=true|WetGrass=true) = 0.35770.
- **Reference**: Koller & Friedman; manual derivation.
- **Verifications**: 3
- **Subtests**:
  - P(Rain=t|WG=t) = 0.16038/0.44838 = 0.35770. Tolerance: goldenBayesian (1e-4).
  - P(Rain=t|Rain=t) = 1.0 (evidence clamping).
  - For all variables without evidence: sum(posterior) == 1.0 (1e-9).
- **Expected values**: 0.35770 ± 1e-4, 1.0 exact, sums = 1.0.

#### Test: Explaining away (D-separation)
- **Strategy**: Known-answer
- **Invariant**: P(Rain|WG=t, Spr=t) < P(Rain|WG=t). Observing sprinkler reduces rain probability.
- **Reference**: Koller & Friedman "PGM" Ch. 3; Pearl.
- **Verifications**: 3
- **Subtests**:
  - P(Rain=t|WG=t) = 0.35770 ± 1e-4.
  - P(Rain=t|WG=t, Spr=t) < P(Rain=t|WG=t) (strict inequality).
  - Both posteriors verified against both VE and Enumeration.

---

### 1.14 engine/fuzzy/ (Strategy A: exhaustive monotonicity + B: known-answer)

#### Test: Tipping monotonicity
- **Strategy**: Exhaustive
- **Invariant**: Tip is non-decreasing when one input increases with the other fixed.
- **Reference**: Mamdani & Assilian (1975).
- **Verifications**: 42 (21 food values + 21 service values)
- **Subtests**:
  - Fix service=5.0, vary food 0-10 step 0.5. Tip sequence non-decreasing within defuzzTolerance.
  - Fix food=5.0, vary service 0-10 step 0.5. Same.
- **Failure criterion**: Any consecutive decrease > defuzzTolerance (0.5).

#### Test: Single-rule Sugeno
- **Strategy**: Known-answer
- **Invariant**: Full activation Sugeno produces exact singleton output.
- **Reference**: Sugeno (1985).
- **Verifications**: 1
- **Expected values**: activation=1.0, output = 1.0*25.0/1.0 = 25.0. Tolerance 1e-9.

#### Test: Output bounds all methods
- **Strategy**: Exhaustive (boundary)
- **Invariant**: For every defuzzification method and input combination, tip is in [0, 30].
- **Reference**: Mamdani & Assilian (1975); Sugeno (1985).
- **Verifications**: 605 (5 methods x 11 food x 11 service values)
- **Failure criterion**: Any tip outside [0, 30].

---

### 1.15 engine/causal/ (Strategy B: adversarial + known-answer)

#### Test: Confounder SCM
- **Strategy**: Adversarial
- **Invariant**: do-calculus distinguishes causation from correlation.
- **Reference**: Pearl "Causality" Ch. 3, 7.
- **Verifications**: 4
- **Subtests**:
  - "confounder do does not change Y": SCM U→X(=U*2), U→Y(=U*3+X*0). do(X=5) does NOT change Y.
  - "real causal effect": SCM U→X(=U*2), X→Y(=X+10). do(X=5): Y=15.
  - "do differs from observe": Results differ with confounder.
  - "intervention idempotence": SCM X→Y(=X+3). Propagate X=5 gives Y=8. Intervene do(X=5) gives Y=8.

#### Test: Counterfactual preservation
- **Strategy**: Known-answer
- **Invariant**: Non-descendant variables preserve factual values in counterfactuals.
- **Reference**: Pearl "Causality" Ch. 7.
- **Verifications**: 4
- **Subtests**:
  - SCM X→Z(=X*2)→Y(=Z+3). Factual X=5. do(Z=7): X=5 (preserved), Z=7, Y=10.
  - Diamond SCM U→X, U→Z, X→Y, Z→Y (Y=X+Z). do(X=10): Z keeps factual value.
  - do(X=10) given factual X=5: Z=20, Y=23.
  - Causal effect measurement: deltaY/deltaX matches coefficient.
- **Expected values**: All exact within 1e-9.

#### Test: Intervention idempotence
- **Strategy**: Known-answer
- **Invariant**: Without confounders, do(X=v) == observe(X=v).
- **Reference**: Pearl "Causality" Ch. 3.
- **Verifications**: 2
- **Expected values**: Simple chain Y=8 for both. Longer chain Z=10, Y=13 for both.

---

### 1.16 engine/mcdm/ (Strategy B: known-answer + boundary)

#### Test: AHP known answer
- **Strategy**: Known-answer
- **Invariant**: Perfectly consistent matrix produces exact weights with CR=0.
- **Reference**: Saaty (1980).
- **Verifications**: 5
- **Subtests**:
  - Matrix [[1,2,6],[1/2,1,3],[1/6,1/3,1]]. Weights=[0.6, 0.3, 0.1] (1e-6). CR < 1e-10.
  - Weights sum to 1.0 (1e-9).
  - Saaty textbook 3x3: weights≈[0.633, 0.260, 0.106] (0.01). CR≈0.033.
- **Expected values**: As specified.

#### Test: AHP CR boundary
- **Strategy**: Boundary
- **Invariant**: CR < 0.10 threshold.
- **Reference**: Saaty (1980).
- **Verifications**: 2
- **Subtests**:
  - Consistent matrix: Consistent=true, CR < 0.10.
  - Highly inconsistent circular preferences: Consistent=false, CR >= 0.10.

#### Test: TOPSIS known answer
- **Strategy**: Known-answer
- **Invariant**: TOPSIS scores and rankings for known cases.
- **Reference**: Hwang & Yoon (1981).
- **Verifications**: 5
- **Subtests**:
  - Strict dominance: A=[10,10,10], B=[1,1,1]. Score(A)=1.0, Score(B)=0.0.
  - Ideal alternative: Score = 1.0.
  - Anti-ideal: Score = 0.0.
  - Benefit vs cost mixed: [10,1] dominates [1,10].
  - Weight sensitivity: weights [0.9,0.1] vs [0.1,0.9] produce opposite rankings.

---

## Section 2: Golden Scenario — Cross-Paradigm Loan Approval Pipeline

### Principle

One end-to-end scenario exercises all 5 engine paradigms in sequence. Every numerical value is hand-calculated. Each phase feeds into the next.

### Phase 1: Deductive Eligibility

Rules:
- R1 "eligibility": age_ok AND no_fraud → eligible
- R2 "can-apply": eligible AND has_income → can_apply

### Phase 2: Bayesian Default Probability

Network: CreditHistory → Default, IncomeLevel → Default.
- P(CH=good)=0.7, P(CH=bad)=0.3, P(IL=high)=0.6, P(IL=low)=0.4
- P(D=yes|CH=good,IL=high)=0.02, P(D=yes|CH=good,IL=low)=0.10
- P(D=yes|CH=bad,IL=high)=0.15, P(D=yes|CH=bad,IL=low)=0.40

Marginal: P(D=yes) = 0.02×0.7×0.6 + 0.10×0.7×0.4 + 0.15×0.3×0.6 + 0.40×0.3×0.4
= 0.0084 + 0.028 + 0.027 + 0.048 = 0.1114

### Phase 3: Fuzzy Risk Assessment

Input: debt_ratio (0-100). Output: risk (0-100).
Terms: low=Trapezoidal(0,0,20,40), medium=Triangular(30,50,70), high=Trapezoidal(60,80,100,100).
Rules: debt_ratio low→risk low, debt_ratio medium→risk medium, debt_ratio high→risk high.

### Phase 4: Causal What-If

SCM: Income → DebtRatio (=100-Income×0.8) → RiskScore (=DebtRatio×0.5).
- Factual Income=50: DebtRatio=60, RiskScore=30.
- do(Income=80): DebtRatio=36, RiskScore=18.

### Phase 5: MCDM Loan Option Ranking

Options: A(rate=5,term=30,amount=200), B(rate=4,term=15,amount=150), C(rate=6,term=20,amount=250).
Criteria: rate(cost,0.5), term(benefit,0.3), amount(benefit,0.2).

TOPSIS hand calculation:
- Vector norms: rate=sqrt(77)=8.775, term=sqrt(1525)=39.051, amount=sqrt(125000)=353.553
- Ideal: rate=0.228(min/B), term=0.230(max/A), amount=0.141(max/C)
- S(A)=0.674, S(B)=0.470, S(C)=0.332.
- Ranking: A > B > C.

### Variant 1: Clear Approval

#### Test: TestAcceptance_LoanApproval_approved
- Phase 1: age_ok=true, no_fraud=true, has_income=true → eligible=true, can_apply=true. Steps=2.
- Phase 2: Evidence CH=good, IL=high. P(D=yes)=0.02. Tolerance: probTolerance (1e-6).
- Phase 3: debt_ratio=60. Risk in [30, 70].
- Phase 4: Factual Income=50 → RiskScore=30.0. do(Income=80) → RiskScore=18.0. Tolerance: 1e-9.
- Phase 5: 3 scores in [0,1]. Option A highest. Tolerance: 1e-9.
- Trace: each phase returns non-nil result.

### Variant 2: Clear Rejection

#### Test: TestAcceptance_LoanApproval_rejected
- Phase 1: age_ok=true, no_fraud=false → eligible NOT derived. Short circuit.

### Variant 3: Borderline

#### Test: TestAcceptance_LoanApproval_borderline
- Phase 1: can_apply=true. Steps=2.
- Phase 2: Evidence CH=good, IL=low. P(D=yes)=0.10. Tolerance: 1e-6.
- Phase 3: debt_ratio=50. medium(50)=1.0. Risk in [30, 70].
- Phase 4: Factual Income=62.5 → DebtRatio=50, RiskScore=25. do(Income=75) → DebtRatio=40, RiskScore=20.
- Phase 5: All scores in [0,1]. Option A highest.

### Trace Verification

Each paradigm produces:
- Deductive: result.Trace non-nil, len matches Steps
- Bayesian: result.Posterior with ≥2 outcomes
- Fuzzy: result.Outputs with ≥1 key
- Causal: result.Values with all SCM variables
- MCDM: result.Scores with length = number of alternatives

---

## Section 3: Golden Files — Behavioral Snapshots

### Principle

Freeze concrete outputs to detect behavioral drift. Failure message: "behavioral regression".

### 3.1 Golden Bayesian — Rain Network

- **P(Rain=t|WG=t)**: 0.35770. Tolerance: goldenBayesian (1e-4).
  - Derivation: P(WG=t)=0.44838, P(Rain=t,WG=t)=0.16038, ratio=0.35770.
- **P(Rain=t|WG=t, Spr=t)**: [TO_VALIDATE against engine output]. Tolerance: 1e-4.
- **P(Spr=t|WG=t)**: 0.64680. Tolerance: 1e-4.
  - Derivation: P(S=t,WG=t)=0.28998, ratio=0.28998/0.44838=0.64680.
- **Medical network P(D=present|T=positive)**: 0.01943. Tolerance: 1e-4.
  - Derivation: P(T+)=0.05094, P(D,T+)=0.00099, ratio=0.01943.

### 3.2 Golden Deductive — Business Scenario

Rules: high_spend AND loyal → premium → discount → notify.
- Initial: {high_spend: true, loyal: true}.
- Derived: {premium, discount, notify} = all true.
- Steps: 3.
- Trace: length 3, rule names in firing order.
- Provenance: premium/discount/notify = Derived. high_spend/loyal = Asserted.

### 3.3 Golden Fuzzy — Canonical Tipping

| food | service | method | expected tip |
|------|---------|--------|-------------|
| 1.0 | 1.0 | Mamdani/Centroid | [TO_FREEZE] |
| 5.0 | 5.0 | Mamdani/Centroid | [TO_FREEZE] |
| 9.0 | 9.0 | Mamdani/Centroid | [TO_FREEZE] |
| 5.0 | 5.0 | Mamdani/Bisector | [TO_FREEZE] |
| 5.0 | 5.0 | Mamdani/MeanOfMax | [TO_FREEZE] |
| 9.0 | 9.0 | Sugeno | [TO_FREEZE] |

Procedure: run current engine, capture values to 6 decimal places, replace [TO_FREEZE] markers.
Tolerance: goldenFuzzy (1e-2). Until frozen, validate only that tip is in [0, 25].

### 3.4 Golden Causal — Linear SCM (Z=2X, Y=Z+3)

| Operation | Input | X | Z | Y |
|-----------|-------|-----|------|------|
| Propagate | X=5 | 5.0 | 10.0 | 13.0 |
| Intervene | do(Z=7) | 0.0 | 7.0 | 10.0 |
| Counterfactual | factual X=5, do(X=10) | 10.0 | 20.0 | 23.0 |
| Counterfactual | factual X=5, do(Z=7) | 5.0 | 7.0 | 10.0 |

Tolerance: floatTolerance (1e-9).

### 3.5 Golden MCDM

**AHP**: Consistent 3x3 matrix [[1,2,6],[1/2,1,3],[1/6,1/3,1]].
- Weights: [0.6, 0.3, 0.1]. Tolerance: 1e-6.
- CR: 0.0. Tolerance: 1e-9.

**TOPSIS**: A=[10,10,10], B=[1,1,1], all benefit, equal weights.
- Score(A) = 1.0, Score(B) = 0.0. Tolerance: 1e-9.

### Tolerance Summary

- Exact: integers, booleans, strings
- floatTolerance (1e-9): causal, TOPSIS dominant, AHP CR
- goldenBayesian (1e-4): Bayesian posteriors
- goldenFuzzy (1e-2): fuzzy defuzzified outputs
- 1e-6: AHP weights

---

## Section 4: Error Contracts

### Principle

Every public constructor/validator must reject invalid input with the expected error.
Naming: TestErrorContract_[Package]_[Function]. Failure message: "error contract violated".

### 4.1 math/fuzzy/ — Membership Constructors

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| Triangular(5, 3, 7) | a > b | ErrInvalidRange |
| Triangular(1, 5, 3) | b > c | ErrInvalidRange |
| Triangular(5, 5, 5) | a == c | ErrInvalidRange |
| Trapezoidal(5, 3, 7, 9) | b < a | ErrInvalidRange |
| Trapezoidal(1, 2, 2, 1) | d < a | ErrInvalidRange |
| Gaussian(0, 0) | sigma == 0 | ErrInvalidRange |
| Gaussian(0, -1) | sigma < 0 | ErrInvalidRange |
| Sample(fn, 10, 5, 100) | lo > hi | ErrInvalidRange |
| Sample(fn, 0, 10, 0) | n == 0 | ErrEmptySamples |

Valid counterpart: Triangular(1,5,9), Trapezoidal(0,2,8,10) → no error.

### 4.2 math/stats/ — Distribution Constructors

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewNormal(0, 0) | sigma == 0 | ErrInvalidParameter |
| NewNormal(0, -1) | sigma < 0 | ErrInvalidParameter |
| NewExponential(0) | lambda == 0 | ErrInvalidParameter |
| NewExponential(-1) | lambda < 0 | ErrInvalidParameter |
| NewBeta(0, 1) | alpha == 0 | ErrInvalidParameter |
| NewBeta(1, -1) | beta < 0 | ErrInvalidParameter |
| NewBinomial(0, 0.5) | n == 0 | ErrInvalidParameter |
| NewBinomial(10, -0.1) | p < 0 | ErrInvalidProb |
| NewBinomial(10, 1.1) | p > 1 | ErrInvalidProb |
| NewGamma(0, 1) | alpha == 0 | ErrInvalidParameter |
| NewChiSquared(0) | k == 0 | ErrInvalidDegreesOfFreedom |
| NewStudentT(0) | nu == 0 | ErrInvalidDegreesOfFreedom |
| NewFDist(0, 5) | d1 == 0 | ErrInvalidDegreesOfFreedom |
| NewPoisson(0) | lambda == 0 | ErrInvalidParameter |
| NewUniform(5, 5) | min == max | ErrInvalidParameter |
| NewWeibull(0, 1) | k == 0 | ErrInvalidParameter |
| NewLognormal(0, 0) | sigma == 0 | ErrInvalidParameter |

### 4.3 math/logic/predicate/ — Quantifiers

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| ForAll(coll, nil) | nil predicate | ErrNilPredicate |
| Exists(coll, nil) | nil predicate | ErrNilPredicate |

Valid empty-domain behavior (tested in section 1.4, NOT error contracts):
- ForAll([], pred) → true, Exists([], pred) → false, Count([], pred) → 0, Filter([], pred) → []

### 4.4 engine/bayesian/network/ — Network Construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddNode (duplicate) | Same variable twice | error ("duplicate") |
| Validate (cycle) | A→B→A | ErrCyclicNetwork |
| Validate (missing parent) | Unregistered parent | error ("not in network") |
| Validate (no outcomes) | Empty Outcomes | error ("no outcomes") |

### 4.5 engine/causal/model/ — SCM Construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddVariable (duplicate) | Duplicate variable | ErrDuplicateVariable |
| AddVariable (nil eq) | Nil equation | ErrNilEquation |
| Validate (cycle) | Cyclic SCM | ErrCyclicModel |
| Validate (missing parent) | Unregistered parent | ErrParentNotFound |

### 4.6 engine/mcdm/ — Validation

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| ahp.Analyze([]) | Empty matrix | ErrEmptyMatrix |
| ahp.Analyze(non-square) | Non-square matrix | ErrNotSquareMatrix |
| ahp.Rank([], evals) | Empty weights | ErrEmptyMatrix |
| ahp.Rank(w, mismatched) | Dimension mismatch | ErrDimensionMismatch |
| topsis.Rank([], criteria) | Empty matrix | ErrEmptyInput |
| topsis.Rank(matrix, mismatched) | Dimension mismatch | ErrDimensionMismatch |

### 4.7 math/graph/ — Graph Construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddNode (duplicate) | Same ID twice | ErrDuplicateNode |
| AddEdge (missing from) | Non-existent From | ErrNodeNotFound |
| AddEdge (missing to) | Non-existent To | ErrNodeNotFound |
| DAG AddEdge (cycle) | Creates cycle | ErrCycleDetected |
| DAG AddEdge (self-loop) | From == To | ErrSelfLoop |
| Bipartite AddEdge (same partition) | Both left | ErrNotBipartite |
| ShortestPath (disconnected) | No path exists | ErrNoPath |
| ShortestPath (missing node) | Non-existent source | ErrNodeNotFound |
| TopologicalSort (cyclic) | Cyclic directed graph | ErrNotDAG |
| BFS (missing start) | Non-existent start | ErrNodeNotFound |
| DFS (missing start) | Non-existent start | ErrNodeNotFound |
| Tree RemoveNode (root) | Removing root node | ErrInvalidEdge |
| Tree AddEdge (multiple parents) | Node already has parent | ErrMultipleParents |
| Tree NewTreeFrom (multiple roots) | DAG with >1 root | ErrNotTree |
| Coloring (empty graph) | Graph with no vertices | error or valid empty coloring |

Valid counterparts: AddNode with unique ID, AddEdge between existing nodes preserving DAG/tree constraints, Tree RemoveNode on leaf, Tree AddEdge to parentless node, Coloring on non-empty graph → no error.

### 4.8 math/fsm/ — FSM Construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewMachine (unknown To state) | To state not in states | ErrInvalidTransition |
| NewMachine (no initial) | Initial not in states | ErrNoInitialState |
| NewMachine (duplicate state) | Duplicate state ID | ErrDuplicateState |
| NewMachine (empty event) | Empty event string | ErrInvalidEvent |
| Send (unknown event) | Event with no transition | ErrTransitionNotFound |

### 4.9 math/markov/ — Chain Construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewChain (row sum != 1) | Row sums to 0.8 | ErrInvalidRow |
| NewChain (negative prob) | P[i][j] = -0.1 | ErrInvalidProbability |
| NewChain (non-square) | 1 row, 2 states | ErrInvalidMatrix |
| NewChain (empty) | 0 states | ErrEmptyChain |
| NewChain (duplicate state) | Same state name twice | ErrDuplicateState |
| Steady (reducible) | Not irreducible | ErrNotIrreducible |
| Absorption (no absorbing) | No absorbing states | ErrNoAbsorbingStates |

---

## Section 5: Performance Baselines

### Principle

No absolute times. Three strategies:
1. **Termination**: algorithm must complete within 5s timeout.
2. **Relative scaling**: T(large)/T(small) must be below a threshold.
3. **Algorithm comparison**: neither algorithm more than 10x slower.

Failure message: "performance regression".

### 5.1 Termination Under Pressure

| Test | Input | Timeout |
|------|-------|---------|
| DPLL 15 vars | 15-variable cyclic implication chain, ~100 clauses | 5s |
| Forward chaining 200 rules | 200 sequential rules V0→V1→...→V200 | 5s |
| Forward chaining cyclic | 50 cyclic rules V0→V1→...→V49→V0 | 5s |
| VE 6-variable network | 6-variable Bayesian network, evidence on last, query first | 5s |
| BFS 1000 vertices | Sparse directed graph, each node → next 3 | 5s |
| DFS 1000 vertices | Same graph | 5s |
| FSM reachability 500 | 500 states, each → next 5 (modular) | 5s |

### 5.2 Relative Scaling

| Test | Small | Large | Expected ratio |
|------|-------|-------|----------------|
| Forward chaining | 10 sequential rules | 100 sequential rules | < 150 |
| DPLL | 5-variable chain | 10-variable chain | < 100 |
| VE | 3-variable chain | 5-variable chain | < 100 |
| Shortest path | 100 vertices sparse | 1000 vertices sparse | < 150 |
| Topological sort | 100-node linear DAG | 1000-node linear DAG | < 20 |
| FSM acceptance | 100 states linear | 1000 states linear | < 15 |
| Markov steady state | 10-state irreducible | 50-state irreducible | < 500 |
| Markov absorption | 10 transient + 2 absorbing | 48 transient + 2 absorbing | < 2000 |

### 5.3 Algorithm Comparison

| Test | Algorithms | Input | Constraint |
|------|-----------|-------|------------|
| VE vs Enumeration | VariableElimination, Enumeration | Rain network, WG=true, query Rain | Same result (1e-6), ratio 0.1-10 |
| Mamdani vs Sugeno | Mamdani/Centroid, Sugeno | 2-rule tipping, food=7.0 | Ratio 0.1-10 |

### 5.4 Graph Algorithms

Covered by 5.1 (BFS/DFS termination) and 5.2 (shortest path, topological sort scaling).

### 5.5 FSM Operations

Covered by 5.1 (reachability termination) and 5.2 (acceptance scaling).

### 5.6 Markov Chain

Covered by 5.2 (steady state and absorption scaling).

---

## Appendix A: Tests That Already Exist and Are NOT Duplicated

The following existing test files were reviewed. Their coverage is not duplicated in this specification; new tests complement (not repeat) them.

| Package | File | Tests | Justification |
|---------|------|-------|---------------|
| math/logic/ | examples/properties_test.go | 22 property tests | Cover NNF/CNF/DNF/Simplify on selected formulas. New tests extend to exhaustive corpus (~410 formulas). |
| math/logic/ | examples/examples_test.go | 10 example tests | Cover specific scenarios. New tests add adversarial and cross-check strategies. |
| math/fuzzy/ | examples/properties_test.go | 22 property tests | Cover t-norm/t-conorm properties on sampled values. New tests use exhaustive grid (441 pairs, 9261 triples). |
| math/fuzzy/ | examples/examples_test.go | 20 example tests | Cover membership functions and defuzzification on selected inputs. New tests add boundary and degenerate cases. |
| math/stats/ | examples/properties_test.go | 12 property tests | Cover distribution properties. New tests add NIST reference values and adversarial numerical stability. |
| math/stats/ | examples/examples_test.go | 8 example tests | Cover specific distribution calculations. New tests add Welford cancellation, WindowedStats consistency. |
| engine/deductive/ | examples/examples_test.go | 13 tests | Cover basic forward/backward scenarios. New tests add cyclic rules, deep chains, clone-on-attempt isolation. |
| engine/bayesian/ | examples/examples_test.go | 14 tests | Cover basic VE/Enumeration. New tests add exhaustive evidence-sweep, elimination order invariance, explaining away. |
| engine/fuzzy/ | examples/examples_test.go | 15 tests | Cover Mamdani/Sugeno scenarios. New tests add exhaustive monotonicity, output bounds across all methods. |
| engine/causal/ | examples/examples_test.go | 10 tests | Cover propagation/intervention/counterfactual. New tests add confounder SCM, diamond graph, idempotence. |
| engine/mcdm/ | examples/examples_test.go | 22 tests | Cover AHP/TOPSIS scenarios. New tests add CR boundary, ideal/anti-ideal, weight sensitivity. |
| math/graph/ | (no examples/) | 0 | All graph acceptance tests are new. |
| math/fsm/ | (no examples/) | 0 | All FSM acceptance tests are new. |
| math/markov/ | (no examples/) | 0 | All Markov acceptance tests are new. |

Each existing benchmark_test.go file is also reviewed but not listed (benchmarks are not functional tests).

---

## Appendix B: Manual Derivations of Expected Values

- **B.1**: P(Rain=t|WG=t) = 0.16038/0.44838 = 0.35770
- **B.2**: P(Default=yes) = 0.0084 + 0.028 + 0.027 + 0.048 = 0.1114
- **B.3**: Causal: Propagate X=5→Z=10→Y=13; Intervene do(Z=7)→Y=10; CF do(X=10)→Y=23; CF do(Z=7)→X=5,Z=7,Y=10
- **B.4**: AHP weights [0.6, 0.3, 0.1], CR=0.0 for consistent matrix
- **B.5**: P(Spr=t|WG=t) = 0.28998/0.44838 = 0.64680
- **B.6**: TOPSIS dominance: A=[10,10,10]→S=1.0, B=[1,1,1]→S=0.0
- **B.7**: TOPSIS loan: S(A)=0.674, S(B)=0.470, S(C)=0.332; A > B > C
- **B.8**: Gambler's ruin: P(S1→S0)=2/3, P(S1→S3)=1/3, MFP(Sunny→Rainy)=5.0
- **B.9**: Weather steady state: π_Sunny=2/3, π_Rainy=1/3
- **B.10**: P^2[S→S]=0.72, P^2[S→R]=0.28

---

## Appendix C: Error Contract Inventory

Complete inventory of public functions that return errors, grouped by package.

| # | Package | Function | Invalid input | Expected error |
|---|---------|----------|--------------|----------------|
| 1 | math/fuzzy | Triangular(5,3,7) | a > b | ErrInvalidRange |
| 2 | math/fuzzy | Triangular(1,5,3) | b > c | ErrInvalidRange |
| 3 | math/fuzzy | Triangular(5,5,5) | a == c | ErrInvalidRange |
| 4 | math/fuzzy | Trapezoidal(5,3,7,9) | b < a | ErrInvalidRange |
| 5 | math/fuzzy | Trapezoidal(1,2,2,1) | d < a | ErrInvalidRange |
| 6 | math/fuzzy | Gaussian(0,0) | sigma == 0 | ErrInvalidRange |
| 7 | math/fuzzy | Gaussian(0,-1) | sigma < 0 | ErrInvalidRange |
| 8 | math/fuzzy | Sample(fn,10,5,100) | lo > hi | ErrInvalidRange |
| 9 | math/fuzzy | Sample(fn,0,10,0) | n == 0 | ErrEmptySamples |
| 10 | math/stats | NewNormal(0,0) | sigma == 0 | ErrInvalidParameter |
| 11 | math/stats | NewNormal(0,-1) | sigma < 0 | ErrInvalidParameter |
| 12 | math/stats | NewExponential(0) | lambda == 0 | ErrInvalidParameter |
| 13 | math/stats | NewExponential(-1) | lambda < 0 | ErrInvalidParameter |
| 14 | math/stats | NewBeta(0,1) | alpha == 0 | ErrInvalidParameter |
| 15 | math/stats | NewBeta(1,-1) | beta < 0 | ErrInvalidParameter |
| 16 | math/stats | NewBinomial(0,0.5) | n == 0 | ErrInvalidParameter |
| 17 | math/stats | NewBinomial(10,-0.1) | p < 0 | ErrInvalidProb |
| 18 | math/stats | NewBinomial(10,1.1) | p > 1 | ErrInvalidProb |
| 19 | math/stats | NewGamma(0,1) | alpha == 0 | ErrInvalidParameter |
| 20 | math/stats | NewChiSquared(0) | k == 0 | ErrInvalidDegreesOfFreedom |
| 21 | math/stats | NewStudentT(0) | nu == 0 | ErrInvalidDegreesOfFreedom |
| 22 | math/stats | NewFDist(0,5) | d1 == 0 | ErrInvalidDegreesOfFreedom |
| 23 | math/stats | NewPoisson(0) | lambda == 0 | ErrInvalidParameter |
| 24 | math/stats | NewUniform(5,5) | min == max | ErrInvalidParameter |
| 25 | math/stats | NewWeibull(0,1) | k == 0 | ErrInvalidParameter |
| 26 | math/stats | NewLognormal(0,0) | sigma == 0 | ErrInvalidParameter |
| 27 | logic/predicate | ForAll(coll,nil) | nil predicate | ErrNilPredicate |
| 28 | logic/predicate | Exists(coll,nil) | nil predicate | ErrNilPredicate |
| 29 | engine/bayesian | AddNode(duplicate) | duplicate variable | error ("duplicate") |
| 30 | engine/bayesian | Validate(cycle) | A→B→A | ErrCyclicNetwork |
| 31 | engine/bayesian | Validate(missing parent) | unregistered parent | error ("not in network") |
| 32 | engine/bayesian | Validate(no outcomes) | empty Outcomes | error ("no outcomes") |
| 33 | engine/causal | AddVariable(duplicate) | duplicate variable | ErrDuplicateVariable |
| 34 | engine/causal | AddVariable(nil eq) | nil equation | ErrNilEquation |
| 35 | engine/causal | Validate(cycle) | cyclic SCM | ErrCyclicModel |
| 36 | engine/causal | Validate(missing parent) | unregistered parent | ErrParentNotFound |
| 37 | engine/mcdm | ahp.Analyze([]) | empty matrix | ErrEmptyMatrix |
| 38 | engine/mcdm | ahp.Analyze(non-square) | non-square matrix | ErrNotSquareMatrix |
| 39 | engine/mcdm | ahp.Rank([],evals) | empty weights | ErrEmptyMatrix |
| 40 | engine/mcdm | ahp.Rank(w,mismatched) | dimension mismatch | ErrDimensionMismatch |
| 41 | engine/mcdm | topsis.Rank([],criteria) | empty matrix | ErrEmptyInput |
| 42 | engine/mcdm | topsis.Rank(matrix,mismatched) | dimension mismatch | ErrDimensionMismatch |
| 43 | math/graph | AddNode(duplicate) | same ID twice | ErrDuplicateNode |
| 44 | math/graph | AddEdge(missing from) | non-existent From | ErrNodeNotFound |
| 45 | math/graph | AddEdge(missing to) | non-existent To | ErrNodeNotFound |
| 46 | math/graph | DAG AddEdge(cycle) | creates cycle | ErrCycleDetected |
| 47 | math/graph | DAG AddEdge(self-loop) | From == To | ErrSelfLoop |
| 48 | math/graph | Bipartite AddEdge(same partition) | both left | ErrNotBipartite |
| 49 | math/graph | ShortestPath(disconnected) | no path exists | ErrNoPath |
| 50 | math/graph | ShortestPath(missing node) | non-existent source | ErrNodeNotFound |
| 51 | math/graph | TopologicalSort(cyclic) | cyclic directed graph | ErrNotDAG |
| 52 | math/graph | BFS(missing start) | non-existent start | ErrNodeNotFound |
| 53 | math/graph | DFS(missing start) | non-existent start | ErrNodeNotFound |
| 54 | math/graph | Tree RemoveNode(root) | removing root | ErrInvalidEdge |
| 55 | math/graph | Tree AddEdge(multiple parents) | node already has parent | ErrMultipleParents |
| 56 | math/graph | Tree NewTreeFrom(multiple roots) | DAG with >1 root | ErrNotTree |
| 57 | math/graph | Coloring(empty graph) | no vertices | error or empty coloring |
| 58 | math/fsm | NewMachine(unknown To) | To state not in states | ErrInvalidTransition |
| 59 | math/fsm | NewMachine(no initial) | initial not in states | ErrNoInitialState |
| 60 | math/fsm | NewMachine(duplicate state) | duplicate state ID | ErrDuplicateState |
| 61 | math/fsm | NewMachine(empty event) | empty event string | ErrInvalidEvent |
| 62 | math/fsm | Send(unknown event) | no transition | ErrTransitionNotFound |
| 63 | math/markov | NewChain(row sum != 1) | row sums to 0.8 | ErrInvalidRow |
| 64 | math/markov | NewChain(negative prob) | P[i][j] = -0.1 | ErrInvalidProbability |
| 65 | math/markov | NewChain(non-square) | 1 row, 2 states | ErrInvalidMatrix |
| 66 | math/markov | NewChain(empty) | 0 states | ErrEmptyChain |
| 67 | math/markov | NewChain(duplicate state) | same state twice | ErrDuplicateState |
| 68 | math/markov | Steady(reducible) | not irreducible | ErrNotIrreducible |
| 69 | math/markov | Absorption(no absorbing) | no absorbing states | ErrNoAbsorbingStates |

**Total: 69 error contract entries** (9 fuzzy + 17 stats + 2 predicate + 4 bayesian + 4 causal + 6 mcdm + 15 graph + 5 fsm + 7 markov).
