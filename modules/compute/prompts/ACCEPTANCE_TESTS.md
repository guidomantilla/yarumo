## ROLE

You are a testing engineer specialized in formal verification and mathematical
testing. You know property-based testing and acceptance testing for scientific
software. Your output is a SPECIFICATION reviewable by a human expert,
not code.

## TASK

Generate an ACCEPTANCE TEST SPECIFICATION (pure text, no code)
for the inference library modules/compute/ (math/ + engine/).

Output file must be saved at: modules/compute/tests/specs/ACCEPTANCE_TESTS.md

Based on:

1. The mathematical correctness analysis: modules/compute/CORRECTNESS.md
2. The existing code in modules/compute/math/ and modules/compute/engine/
3. The existing tests in the examples/ directories of each package

## ABSOLUTE CONSTRAINTS

- Only generate test SPECIFICATIONS, never code.
- DO NOT duplicate tests that already exist in the examples/ directories of each package.
  First READ the existing tests to know what is already covered.
- Tests must be REVIEWABLE SPECIFICATIONS: describe what is tested, why, with what
  inputs, and what values are expected. DO NOT include code in any language — the
  implementation is done by other prompts (.claude/prompts/acceptance/01-07.md).
- Each test must have a THEORETICAL REFERENCE that justifies the expected value.
- Test names in English, descriptive.

## RIGOR METHODOLOGY

Each invariant is tested with TWO complementary strategies:

### Strategy A: Exhaustive-within-bounds (for enumerable properties)

Enumerate ALL cases up to a finite limit and verify the property on EACH one.
This turns a "property test" into a complete verification within the bound.

- **Logical formulas**: Generate all propositional formulas up to depth=3
  with variables {P, Q, R} using the constructors {Var, Not, And, Or, Impl, Iff, True, False}.
  This produces a finite corpus.
- **Fuzzy values**: Grid over [0, 1] with step 0.05 = 21 points. For pairs: 21x21 = 441 combinations.
  For triples: 21x21x21 = 9261 combinations.
- **Bayesian variables**: For each variable in a network, for each possible
  evidence combination (2^observed), verify the property.

### Strategy B: Adversarial + known-answer (for numerical properties)

Pathological inputs designed to break, plus reference values from textbooks.

- **Catastrophic cancellation**: Values designed to exploit precision loss
  (e.g., [1e8+1, 1e8+2, 1e8+3] for Welford).
- **Boundary values**: 0, 1, epsilon, 1-epsilon, +Inf, -Inf, NaN for numerical functions.
- **Textbook known-answer**: Values with published and verified answers
  (Koller & Friedman for Bayesian, Pearl for causal, Saaty for AHP, NIST for stats).
- **Near-degenerate**: AHP matrices with CR just below 0.10, networks with
  extreme probabilities (0.001, 0.999), SCMs with long paths.

## PROJECT CONTEXT

### Library structure

The inference library has two modules:

**math/** — Pure mathematical foundations (15 packages):
- logic/ — propositional logic (NNF, CNF, DNF, simplify, parser)
- logic/sat/ — SAT solver (DPLL)
- logic/entailment/ — semantic entailment with countermodels
- logic/predicate/ — quantifiers over finite collections (ForAll, Exists, Count, Filter)
- logic/temporal/ — bounded temporal assertions (ResponseWithin, FrequencyWithin, Eventually, Before, Elapsed, Sequence) + LTL primitives (Always, Next, Until, Release, Since)
- fuzzy/ — membership functions (triangular, trapezoidal, gaussian, sigmoid, constant), t-norms/t-conorms, defuzzification (Centroid, Bisector, MeanOfMax, LargestOfMax, SmallestOfMax), Fuzzify, Clip, Scale, AggregateMax, Sample
- sets/ — set operations (Union, Intersection, Difference, SymmetricDifference, IsSubset, IsSuperset, Contains, Equal)
- stats/ — 17 distributions (Normal, Exponential, Uniform, Beta, Gamma, ChiSquared, StudentT, Lognormal, Weibull, FDist, Gumbel, Pareto, Poisson, Binomial, Geometric, Hypergeometric, NegativeBinomial), hypothesis testing, RunningStats (Welford), WindowedStats, descriptive functions (Mean, Variance, Median, Percentile, Correlation, LinearRegression, RSquared, Skewness, Kurtosis, WeightedMean, GeometricMean, HarmonicMean, MAD, Mode, etc.), Bayes theorem (BayesTheorem, TotalProbability, ChainRule, Independent), probability functions (IsValid, Normalize, Complement, Entropy)
- graph/ — directed, undirected, DAG, bipartite graphs. Traversal (BFS, DFS), paths (Dijkstra, Bellman-Ford, Floyd-Warshall, AllPaths), centrality (Degree, Betweenness, Closeness, PageRank), structure (TopologicalSort, ConnectedComponents, SCC, Bridges, ArticulationPoints, Diameter), optimization (MST, MaxFlow, Matching), matrix (ToMatrix, MatrixMultiply tropical, MatrixPower, TransitiveClosure), substructures (Bron-Kerbosch cliques, Eulerian path, coloring), properties (IsDAG, IsTree, IsBipartite, InDegree, OutDegree), operations (Subgraph, Union, Intersection, Complement, Reverse, CartesianProduct), multigraph, tree
- fsm/ — finite state machines: transitions, guards, determinism, reachability, dead states, completeness
- markov/ — Markov chains: transition, steady state, classification, absorption, mean first passage, step-n, simulate

**engine/** — 5 inference paradigms:
- deductive/ — forward + backward chaining, rules, facts, provenance
- bayesian/ — variable elimination + enumeration, Bayesian networks with CPTs
- fuzzy/ — Mamdani + Sugeno, fuzzy rules, linguistic variables
- causal/ — Pearl levels 1-2 (propagation, do-operator, counterfactuals)
- mcdm/ — AHP (pairwise comparison + consistency) + TOPSIS (ideal solution distance)

Each paradigm has: inference engine, explainability (traces), and auxiliary packages.

### Existing tests (READ to avoid duplication)

Each package has an examples/ directory with existing tests:
- math/logic/examples/, math/fuzzy/examples/, math/stats/examples/
- engine/bayesian/examples/, engine/causal/examples/, engine/deductive/examples/
- engine/fuzzy/examples/, engine/mcdm/examples/

READ these files before specifying tests to avoid duplication.

### Exclusions

- math/logic/parser/ — string-to-formula parser (software engineering, not mathematics).
  Its correctness is covered by unit tests, not formal analysis or acceptance tests.

### Architectural principle

The acceptance tests are an EXTERNAL CONSUMER of the public API.
Not a friend of the implementation.

- They only use the public API — never internals
- If something cannot be tested without accessing internals, that reveals a gap in the API
- If a refactor changes internals but maintains the API, these tests DO NOT break
- If a refactor changes the API, these tests DO break — and that is correct

### Numerical tolerances

- floatTolerance: 1e-9 for general floating-point arithmetic
- probTolerance: 1e-6 for probabilities (accumulate error in marginalization)
- defuzzTolerance: 0.5 for defuzzification (discretization introduces error)
- goldenBayesian: 1e-4 for Bayesian golden files
- goldenFuzzy: 1e-2 for fuzzy golden files

## SECTION 1: MATHEMATICAL INVARIANTS

For EACH section of CORRECTNESS.md, generate tests that protect the invariants
declared as GUARANTEED. The test must FAIL if someone breaks the invariant.

DO NOT generate tests that already exist. READ each existing test file first
and document in the output what is already covered.

### 1.1 math/logic/ — Transformations (Strategy A: exhaustive)

#### Formula corpus
Define helper generateFormulas(depth, vars) that generates all formulas up to a given depth:
- depth=0: each variable, True, False
- depth=1: Not(f) for each f of depth=0
- depth=2: BinOp(f1, f2) for each f1,f2 of depth<=1, for each BinOp in {And, Or, Impl, Iff}; plus Not(f) for each f of depth=1
- depth=3: same, combining depth<=2
- Estimate size: with 3 vars, depth=2 generates ~200 formulas. depth=3 may generate thousands; if it exceeds 5000, limit to depth=2.

#### Tests

- **NNF preserves equivalence (exhaustive)**: For EACH formula in the corpus,
  verify Equivalent(f, NNF(f)) == true.
  Failure criterion: ANY formula where NNF is not equivalent.

- **CNF preserves equivalence (exhaustive)**: Same with CNF(f).

- **DNF preserves equivalence (exhaustive)**: Same with DNF(f).

- **CNF structural form**: For EACH formula in the corpus, verify that
  CNF(f) has the form And(Or(...), Or(...), ...) or is atomic/literal.
  Recursively verify that the structure is correct.

- **DNF structural form**: Same, verifying Or(And(...), And(...), ...).

- **Simplify idempotence (exhaustive)**: For EACH formula in the corpus,
  verify that Simplify(Simplify(f)).Equals(Simplify(f)).
  If it fails, Simplify did not reach the fixed point in one pass.

- **Simplify preserves equivalence (exhaustive)**: For EACH formula in the corpus,
  verify Equivalent(f, Simplify(f)) == true.

- **FailCases correctness (exhaustive)**: For EACH formula in the corpus,
  verify that FailCases(f) returns exactly the valuations where f evaluates to false.
  Cross-check: for each valuation in FailCases(f), verify f.Eval(valuation) == false.
  Also verify completeness: the count |FailCases(f)| + |satisfying rows from TruthTable(f)| == 2^|vars|.

- **IsTautology cross-check (exhaustive)**: For EACH formula in the corpus,
  verify IsTautology(f) == (len(FailCases(f)) == 0).
  Also verify IsTautology(f) == !IsSatisfiable(Not(f)).

- **IsContradiction cross-check (exhaustive)**: For EACH formula in the corpus,
  verify IsContradiction(f) == !IsSatisfiable(f).
  Also verify IsContradiction(f) == IsTautology(Not(f)).

### 1.2 math/logic/sat/ — DPLL (Strategy A: exhaustive + Strategy B: adversarial)

#### Exhaustive

- **SAT soundness (exhaustive)**: For EACH formula in the corpus (section 1.1):
  If IsSatisfiable(f) == true, obtain model := SatisfyingAssignment(f),
  verify f.Eval(model) == true.
  Failure criterion: there exists a formula where the returned model does not satisfy it.

- **SAT completeness via truth table (exhaustive)**: For EACH formula in the corpus:
  If IsSatisfiable(f) == false, verify via TruthTable(f) that
  NO row is true (all valuations fail).
  Failure criterion: DPLL says UNSAT but a satisfying valuation exists.

- **CNF preserves satisfiability (exhaustive)**: For EACH formula in the corpus:
  IsSatisfiable(f) == IsSatisfiable(CNF(f)).

#### Adversarial

- **XOR chain (deep nesting)**: Build f = P XOR Q XOR R XOR S (as a chain
  of negated Iffs). High depth, many CNF clauses. Verify that DPLL
  finds a correct model.

- **Pigeon hole PHP(3,2)**: 3 pigeons, 2 holes. Formulate as CNF.
  Known unsatisfiable. Verify DPLL returns UNSAT.
  Reference: Haken 1985 — PHP requires exponential proofs in resolution.

- **All-true tautology**: (P OR NOT P) AND (Q OR NOT Q) AND ... for N=10 variables.
  Verify satisfiable (trivially: all true).

### 1.3 math/logic/entailment/ — (Strategy A: exhaustive + B: known-answer)

#### Exhaustive

- **Entailment cross-check (exhaustive)**: For EACH pair (f1, f2) of formulas
  in the corpus with depth<=1 (limit combinatorics):
  Verify that Entails([f1], f2) matches the semantic verification:
  "for every valuation, if f1 is true then f2 is true".
  Implement the semantic verification via TruthTable.

- **Countermodel validation (exhaustive)**: For EACH pair where Entails==false:
  Obtain countermodel via EntailsWithCounterModel.
  Verify that the countermodel satisfies f1 AND negates f2.

#### Known-answer (classical logic)

- **Modus ponens**: {P, P=>Q} |= Q. True.
- **Modus tollens**: {NOT Q, P=>Q} |= NOT P. True.
- **Hypothetical syllogism**: {P=>Q, Q=>R} |= P=>R. True.
- **Disjunctive syllogism**: {P OR Q, NOT P} |= Q. True.
- **Affirming consequent (fallacy)**: {Q, P=>Q} |= P. FALSE — must return false.
  Reference: Enderton "Mathematical Introduction to Logic" §1.

### 1.4 math/logic/predicate/ — (Strategy B: boundary cases)

NOTE: The basic cases (ForAll true/false, Exists true/false, Count, Filter)
are probably already covered in the existing tests. READ first.

Only generate NEW tests that cover:

- **Singleton domain**: ForAll([x], P) == P(x). Exists([x], P) == P(x).
  Only case where ForAll and Exists are equivalent.
- **Empty domain — standard FOL semantics**: ForAll([], P) returns true (vacuous truth).
  Exists([], P) returns false (vacuous falsity). Count([], P) returns 0. Filter([], P) returns [].
  No errors — follows standard first-order logic conventions.
- **Always-false predicate**: ForAll(domain, alwaysFalse) == false for |domain| >= 1.
  Count(domain, alwaysFalse) == 0. Filter(domain, alwaysFalse) == [].
- **Always-true predicate**: ForAll(domain, alwaysTrue) == true.
  Count(domain, alwaysTrue) == len(domain). Filter(domain, alwaysTrue) == domain.

### 1.5 math/logic/temporal/ — (Strategy B: precise boundary cases)

NOTE: READ the existing tests in temporal/ first. Probably has coverage.
Only generate EXACT BOUNDARY tests:

#### Bounded assertions

- **ResponseWithin exactly at deadline**: Trigger at t=0, response at t=maxDuration exactly.
  Must pass (deadline is inclusive: candidate.Time <= deadline).
- **ResponseWithin one nanosecond after**: Trigger at t=0, response at t=maxDuration+1ns.
  Must fail.
- **FrequencyWithin exactly at threshold**: Exactly minCount events in the window.
  Must pass. minCount-1 events: must fail.
- **Sequence with duplicates**: Trace [A, B, A, B, C] searching for [A, B, C].
  Must pass (greedy finds A[0], B[1], C[4]).
- **Before simultaneous**: a and b occur at the same time. Before(a, b) must fail
  (ev.Time >= firstB.Time rejects simultaneous).
- **Eventually — event present**: Trace [A, B, C]. Eventually(trace, "B") must return true.
- **Eventually — event absent**: Trace [A, C]. Eventually(trace, "B") must return false.
- **Eventually — empty trace**: Eventually([], "A") must return false.
- **Elapsed — normal**: Trace [A at t=0, B at t=5s]. Elapsed(trace, "A", "B") must return 5s.
- **Elapsed — missing event**: Trace [A]. Elapsed(trace, "A", "B") must return error
  (event not found in trace).
- **Elapsed — same event**: Elapsed(trace, "A", "A") — verify behavior
  (0 duration or error, depending on implementation).

#### LTL primitives

- **Always — all hold**: Trace where φ holds at every position. Always(φ) must return true.
- **Always — one violation**: Trace where φ holds everywhere except one position.
  Always(φ) must return false.
- **Always — empty trace**: Always(φ) on empty trace. Vacuously true.
- **Next — normal**: Trace [a, b, c]. Next(φ_b) at position 0 must return true
  (φ_b holds at position 1). Next(φ_a) at position 0 must return false.
- **Next — last position**: Next at the last position of the trace.
  Must return false (no next position exists).
- **Until — classic**: Trace [¬ψ, ¬ψ, ψ] with φ holding at positions 0 and 1.
  (φ U ψ) at position 0 must return true.
- **Until — ψ never holds**: Trace where ψ never holds.
  (φ U ψ) must return false (ψ must eventually hold).
- **Until — ψ holds immediately**: Trace [ψ, ...]. (φ U ψ) at position 0 must return true
  (ψ holds immediately, φ not required).
- **Release — dual of Until**: Verify that (φ R ψ) ≡ ¬(¬φ U ¬ψ) on several traces.
- **Release — ψ holds forever**: Trace where ψ holds at all positions and φ never holds.
  (φ R ψ) must return true (ψ holds forever is sufficient).
- **Since — past-time**: Trace [ψ, φ, φ, φ]. Since(φ, ψ) at position 3 must return true
  (ψ held at position 0, φ held at all positions since).
- **Since — ψ never held**: Trace where ψ never held. Since(φ, ψ) must return false.

### 1.6 math/sets/ — Set operations (Strategy A: exhaustive on small sets + B: known-answer)

READ the existing tests in sets/ first. Generate only what is NOT covered.

#### Exhaustive on small sets

- **Union commutativity**: For ALL pairs of subsets of {1,2,3,4} (2^4 × 2^4 = 256 pairs),
  verify Union(A, B) == Union(B, A).

- **Intersection commutativity**: Same 256 pairs, verify Intersection(A, B) == Intersection(B, A).

- **Union associativity**: For ALL triples of subsets of {1,2,3} (2^3 × 2^3 × 2^3 = 512 triples),
  verify Union(Union(A, B), C) == Union(A, Union(B, C)).

- **Intersection associativity**: Same 512 triples.

- **Difference correctness**: For ALL pairs of subsets of {1,2,3,4},
  verify Difference(A, B) contains exactly the elements in A that are not in B.

- **SymmetricDifference correctness**: For ALL pairs of subsets of {1,2,3,4},
  verify SymmetricDifference(A, B) == Union(Difference(A, B), Difference(B, A)).

- **IsSubset reflexivity**: For ALL subsets A of {1,2,3,4}, IsSubset(A, A) == true.

- **IsSubset antisymmetry**: For ALL pairs where IsSubset(A, B) && IsSubset(B, A),
  verify Equal(A, B) == true.

#### Identity and empty set laws

- **Union identity**: Union(A, ∅) == A for all A.
- **Intersection annihilation**: Intersection(A, ∅) == ∅ for all A.
- **Difference with empty**: Difference(A, ∅) == A. Difference(∅, A) == ∅.
- **Contains correctness**: Contains({1,2,3}, 2) == true. Contains({1,2,3}, 4) == false.
  Contains(∅, 1) == false.
- **Equal reflexivity**: Equal(A, A) == true for all A.
- **Equal symmetry**: Equal(A, B) == Equal(B, A) for all pairs.
- **IsSuperset dual of IsSubset**: For ALL pairs of subsets of {1,2,3,4},
  verify IsSuperset(A, B) == IsSubset(B, A).

### 1.7 math/fuzzy/ — Fuzzy Logic (Strategy A: exhaustive grid)

#### Value grid
Define helper fuzzyGrid(step) that generates [0.0, step, 2*step, ..., 1.0].
Use step=0.05 (21 points, 441 pairs, 9261 triples).

#### Tests per t-norm (Min, Product, Lukasiewicz)

For EACH t-norm, for EACH pair/triple in the grid:

- **Commutativity**: |t(a,b) - t(b,a)| < 1e-15.
  441 verifications per t-norm. 3 t-norms = 1323 total.

- **Associativity**: |t(t(a,b),c) - t(a,t(b,c))| < 1e-15.
  9261 verifications per t-norm. 3 t-norms = 27783 total.

- **Monotonicity**: For each a1 <= a2 in the grid and each b:
  t(a1, b) <= t(a2, b) + 1e-15.
  ~4600 verifications per t-norm.

- **Identity**: |t(a, 1.0) - a| < 1e-15. 21 verifications per t-norm.

#### Tests per t-conorm (Max, ProbabilisticSum, BoundedSum)

Same, with identity t-conorm(a, 0.0) == a.

#### Boundary tests for t-norms

- **Lukasiewicz at 0**: Luk(0.3, 0.6) = max(0.3+0.6-1, 0) = 0. Verify exact.
- **Lukasiewicz at transition**: Luk(0.5, 0.5) = max(0, 0) = 0.
  Luk(0.5, 0.51) = max(0.01, 0) = 0.01.
- **Product with epsilon**: Product(1e-300, 1e-300) — must not be exactly 0
  (underflow is platform-dependent, but the result must be >= 0).

#### Membership functions

- **Triangular peak**: For Triangular(a, b, c), mu(b) == 1.0 exact.
- **Triangular zeros**: mu(a) == 0.0, mu(c) == 0.0.
- **Triangular out of range**: mu(a-1) == 0.0, mu(c+1) == 0.0.
- **Trapezoidal plateau**: For Trapezoidal(a, b, c, d),
  mu(b) == 1.0, mu(c) == 1.0, mu((b+c)/2) == 1.0.
- **Gaussian at center**: Gaussian(center, sigma), mu(center) == 1.0.
- **Gaussian symmetry**: |mu(center+x) - mu(center-x)| < 1e-15.
- **Sigmoid monotonicity**: For Sigmoid(center, slope) with positive slope,
  mu is monotonically increasing. Verify mu(center-2) < mu(center) < mu(center+2).
  mu(center) == 0.5 (inflection point). At extremes: mu(center-10*slope) ≈ 0, mu(center+10*slope) ≈ 1.
- **Sigmoid negative slope**: Sigmoid(center, -slope) is monotonically decreasing.
  Verify mu(center-2) > mu(center) > mu(center+2).
- **Constant value**: Constant(0.7)(x) == 0.7 for all x (any x value).
  Constant(0.0)(x) == 0.0. Constant(1.0)(x) == 1.0.

#### Complement (Zadeh)

- **Complement involution (exhaustive)**: For EACH value d in the fuzzy grid,
  verify Complement(Complement(d)) == d to 1e-15.
  Reference: Zadeh complement C(d) = 1-d, involution property C(C(x)) = x.
- **Complement boundary**: Complement(0.0) == 1.0 exact. Complement(1.0) == 0.0 exact.
- **Complement mid-point**: Complement(0.5) == 0.5 exact (fixed point).

#### Defuzzification bounds

- **Output in range**: For EACH defuzzification method (Centroid, Bisector, MOM, LOM, SOM),
  given a set of pairs (x, mu(x)) where x in [0, 10]:
  The result must be in [0, 10].
  Test with distributions: uniform, left-skewed, right-skewed, bimodal.

#### Operations (Fuzzify, Clip, Scale, AggregateMax, Sample)

- **Fuzzify correctness**: Fuzzify(fn, x) == fn(x) for all built-in membership functions.
  Verify with Triangular, Trapezoidal, Gaussian, Sigmoid at several x values.
  Trivially correct — direct delegation to MembershipFn.

- **Clip correctness (alpha-cut)**: Clip(fn, level)(x) == min(fn(x), level).
  Grid of x values and levels [0.0, 0.3, 0.7, 1.0].
  Reference: Mamdani alpha-cut (standard implication method).

- **Scale correctness (Larsen)**: Scale(fn, level)(x) == fn(x) * level.
  Same grid as Clip. Reference: Larsen (1980) product implication.
  Verify identity: Scale(fn, 1)(x) == fn(x).
  Verify annihilation: Scale(fn, 0)(x) == 0.

- **Clip vs Scale shape preservation**: For the same fn and level,
  Clip truncates (flat top at level), Scale compresses (preserves shape).
  Verify that Clip(fn, level)(x) >= Scale(fn, level)(x) for all x where fn(x) > 0
  (Clip ≥ Scale because min(a, b) ≥ a*b for a,b ∈ [0,1]).

- **AggregateMax correctness**: AggregateMax(f1, f2, ...)(x) == max(f1(x), f2(x), ...).
  Test with 2 and 3 overlapping membership functions at multiple x values.

- **Sample correctness**: Sample(fn, lo, hi, n) produces n evenly-spaced points.
  Verify: first point x=lo, last point x=hi, step=(hi-lo)/(n-1) for n≥2.
  Special case: Sample(fn, lo, hi, 1) returns [(lo, fn(lo))].

### 1.8 math/stats/ — Distributions (Strategy B: known-answer + adversarial)

#### Known-answer — reference values

For each distribution, verify against values computed with arbitrary precision
(Wolfram Alpha, NIST, or scipy with float128).

- **Normal(0,1)**:
  - PDF(0) = 1/sqrt(2*pi) ≈ 0.398942280401433
  - CDF(0) = 0.5
  - CDF(1.96) ≈ 0.975002104852
  - CDF(-1.96) ≈ 0.024997895148
  - Mean = 0, Variance = 1

- **Exponential(lambda=2)**:
  - PDF(0) = 2.0
  - CDF(1) = 1 - e^(-2) ≈ 0.864664716763387
  - Mean = 0.5, Variance = 0.25

- **Beta(2, 5)**:
  - Mean = 2/7 ≈ 0.285714285714
  - Variance = 10/392 ≈ 0.025510204082
  - CDF(0.5) ≈ 0.890625 (I_0.5(2,5), exactly computable)

- **StudentT(5)**:
  - PDF(0) = Gamma(3)/(sqrt(5*pi)*Gamma(2.5)) ≈ 0.379609319956
  - CDF(0) = 0.5 (symmetry)
  - CDF(2.571) ≈ 0.975 (critical t for alpha=0.05, two-tailed, df=5)
  - Mean = 0, Variance = 5/3

- **ChiSquared(3)**:
  - Mean = 3, Variance = 6
  - CDF(7.815) ≈ 0.95 (critical value for alpha=0.05)

- **Poisson(3)**:
  - PMF(0) = e^(-3) ≈ 0.049787068368
  - PMF(3) = e^(-3) * 3^3 / 6 ≈ 0.224041807660
  - Mean = 3, Variance = 3

- **Binomial(10, 0.3)**:
  - PMF(3) = C(10,3) * 0.3^3 * 0.7^7 ≈ 0.266827932
  - Mean = 3, Variance = 2.1

- **Uniform(2, 8)**:
  - PDF(5) = 1/6 ≈ 0.166666666667
  - CDF(5) = 0.5
  - Mean = 5.0, Variance = 3.0

- **Gamma(3, 2)** (shape=3, rate=2):
  - Mean = 1.5, Variance = 0.75
  - CDF(1) — verify via regularized incomplete gamma

- **Lognormal(0, 1)**:
  - Mean = exp(0.5) ≈ 1.648721270700
  - Variance = (exp(1)-1)*exp(1) ≈ 4.670774270471
  - CDF(1) = 0.5 (symmetry of underlying normal)

- **Weibull(2, 1)** (k=2, lambda=1):
  - PDF(1) = 2*exp(-1) ≈ 0.735758882343
  - CDF(1) = 1-exp(-1) ≈ 0.632120558829
  - Mean = Γ(3/2) = sqrt(π)/2 ≈ 0.886226925453
  - Variance = 1 - π/4 ≈ 0.214601836603

- **FDist(5, 10)**:
  - Mean = 10/8 = 1.25
  - Variance = 2*100*13/(5*64*6) ≈ 1.354166666667

- **Gumbel(0, 1)**:
  - Mean = γ (Euler-Mascheroni) ≈ 0.577215664902
  - Variance = π²/6 ≈ 1.644934066848
  - CDF(0) = exp(-1) ≈ 0.367879441171

- **Pareto(1, 3)** (xm=1, alpha=3):
  - CDF(2) = 1 - (1/2)³ = 0.875
  - Mean = 1.5, Variance = 0.75

- **Geometric(0.3)** (failures-before-success):
  - PMF(0) = 0.3
  - PMF(2) = 0.7² × 0.3 = 0.147
  - Mean = 7/3 ≈ 2.333333333333
  - Variance = 70/9 ≈ 7.777777777778

- **Hypergeometric(50, 10, 5)** (N=50, K=10, n=5):
  - Mean = 1.0
  - Variance = 5×10×40×45/(2500×49) ≈ 0.734693877551

- **NegativeBinomial(5, 0.4)** (r=5, p=0.4):
  - Mean = 7.5
  - Variance = 18.75

Tolerance: 1e-9 for PDF/CDF, 1e-12 for Mean/Variance (exact values).

#### Adversarial — numerical stability

- **Welford catastrophic cancellation**: Data [1e8+1, 1e8+2, 1e8+3].
  Variance (population) = 2/3 ≈ 0.666667.
  The naive algorithm Σ(x²) - (Σx)²/n produces garbage due to cancellation.
  Welford must give the correct value with tolerance 1e-9.

- **Welford with identical data**: [5, 5, 5, 5, 5]. Variance = 0 exact.

- **Welford with a single datum**: [42]. Population variance = 0. Sample variance = 0 (n=1).

- **WindowedStats vs full recompute**: Insert 100 random values (fixed seed)
  into WindowedStats with window=10. After each insertion, compute Mean and
  Variance directly over the last min(i+1, 10) values. Compare.
  Tolerance: 1e-9.

- **Hypothesis t-test known outcome**: Sample [2.1, 2.3, 1.9, 2.0, 2.2],
  H0: mu=2.0. Manual: mean=2.1, s=0.1581, t=2.0/0.0707≈1.414, df=4.
  p-value ≈ 0.230 (two-tailed). Do not reject at alpha=0.05.
  Reference: Casella & Berger example 8.3.

- **Hypothesis t-test rejection**: Sample [10.1, 10.3, 9.9, 10.2, 10.4, 10.0, 10.5],
  H0: mu=9.0. The t-value must be large, p-value < 0.001. Reject.

- **Welch's two-sample t-test**: Sample1 [5.1, 5.3, 4.9], Sample2 [3.1, 3.2, 2.9].
  Different means, similar variance. p-value < 0.01. Reject H0 (equal means).
  Reference: Casella & Berger §8.3 (Welch-Satterthwaite degrees of freedom).

- **Chi-squared goodness-of-fit**: Observed [50, 30, 20], Expected [33.3, 33.3, 33.4].
  Statistic = Σ(O-E)²/E, df=2. Verify statistic and that p < 0.05.
  Reference: Casella & Berger §10.3.

- **KS test known outcome**: Sample from standard normal (fixed seed, n=100).
  KS test against Normal(0,1) should NOT reject (p > 0.05).
  KS test against Normal(5,1) should reject (p < 0.001).

- **ANOVA known outcome**: 3 groups: [5,6,7], [5,6,7], [5,6,7] (identical).
  F-statistic = 0, p = 1.0. Do not reject.
  3 groups: [1,2,3], [10,11,12], [20,21,22]. F very large, p < 0.001. Reject.
  Reference: Casella & Berger §11.2.

- **Fisher exact 2x2**: Table [[8,2],[1,9]]. Two-tailed p-value.
  Verify against known hypergeometric calculation.

- **Mann-Whitney U test**: Sample1 [1, 3, 5, 7], Sample2 [2, 4, 6, 8, 10].
  Different distributions. Compute U statistic manually:
  Rank all values combined, sum ranks of each group.
  U1 = n1*n2 + n1*(n1+1)/2 - R1. Verify U statistic and p-value.
  Reference: Mann & Whitney (1947).

- **Mann-Whitney U — identical samples**: Sample1 = Sample2 = [1, 2, 3].
  U should indicate no difference (p ≈ 1.0 or large p). Do not reject.

- **Bayes with precision**: P(Disease|Positive) with prior=0.001, sensitivity=0.99,
  false_positive_rate=0.05. P(Positive) = 0.001*0.99 + 0.999*0.05 = 0.05094.
  P(Disease|Positive) = 0.001*0.99/0.05094 ≈ 0.01943. Verify to 1e-5.
  Reference: Base rate fallacy (Kahneman & Tversky).

### 1.9 engine/deductive/ — (Strategy B: adversarial)

READ the existing tests in deductive/ first. Generate only what is NOT covered.

- **Fixed point with cycles**: Rules A→B, B→C, C→A (cycle).
  Forward chaining must terminate (monotonicity: no new facts after 1 iteration
  if A is already true). Verify that steps <= 3 and all 3 facts are derived.

- **Deep rule chain**: A→B, B→C, ..., Y→Z (25 rules in chain).
  Forward: must derive Z in 25 steps. Backward with depth=30: must derive Z.
  Backward with depth=10: must NOT derive Z (completeness sacrificed).

- **Clone-on-attempt isolation**: Rule with condition (A AND B), where A=true but B=false.
  The rule attempts to fire but fails on B. Verify that the factbase after the attempt
  is IDENTICAL to the factbase before the attempt (no side effects).
  Implement: compare Snapshot() before and after.

- **Conflict resolution determinism**: Two rules with the same priority that
  derive different facts. Verify that BOTH fire eventually (forward
  is not first-match by default, but PriorityOrder that fires all applicable rules).

- **Forward chaining — oscillation detection (contradictory rules)**: Two rules that
  overwrite each other's conclusions: R1: A → B, R2: B → NOT_A (or equivalent).
  The oscillation detection must compare state at end-of-iteration with start-of-iteration;
  if equal (net effect is zero), the loop terminates.
  Verify: terminates (does not loop forever) and steps are bounded.
  Reference: CORRECTNESS.md — "oscillation detection compares the state at
  end-of-iteration with start-of-iteration."

- **Backward chaining — CWA with negated conditions**: Rule with condition NOT(fraud).
  Fact "fraud" is NOT in the factbase. Under CWA, unprovable variables default to false,
  so NOT(fraud) evaluates to true and the rule fires.
  Verify: backward chaining derives the conclusion when the negated variable is absent.
  Reference: CORRECTNESS.md — "rules with negated conditions fire correctly
  when the negated variable is absent or unprovable."

### 1.10 engine/bayesian/ — (Strategy A: exhaustive over the network + B: known-answer)

READ the existing tests in bayesian/ first. Generate only what is NOT covered.

#### Exhaustive over rain network

- **VE == Enumeration for ALL variables**: For each variable {Rain, Sprinkler, WetGrass},
  for each possible evidence subset over the other variables
  (e.g., Rain: {}, {Sprinkler=true}, {Sprinkler=false}, {WetGrass=true}, ...):
  Verify that VE and Enumeration produce the same posterior.
  Total: 3 variables × ~8 evidence combinations = ~24 queries.
  Tolerance: 1e-9.

- **Elimination order invariance**: For Rain with evidence WetGrass=true,
  try ALL permutations of elimination order of the hidden variables
  (Sprinkler). For the medical network: try all permutations of
  {TestResult, Symptom} when querying Disease. Verify identical results.

#### Known-answer (Koller & Friedman)

- **Rain network hand-calculated**:
  P(Rain=true | WetGrass=true):
  P(WG=t) = P(WG=t|R=t,S=t)P(S=t|R=t)P(R=t) + P(WG=t|R=t,S=f)P(S=f|R=t)P(R=t) +
             P(WG=t|R=f,S=t)P(S=t|R=f)P(R=f) + P(WG=t|R=f,S=f)P(S=f|R=f)P(R=f)
           = 0.99*0.01*0.2 + 0.8*0.99*0.2 + 0.9*0.4*0.8 + 0.0*0.6*0.8
           = 0.00198 + 0.1584 + 0.288 + 0.0
           = 0.44838
  P(Rain=t, WG=t) = 0.00198 + 0.1584 = 0.16038
  P(Rain=true | WetGrass=true) = 0.16038/0.44838 ≈ 0.35770
  Verify to tolerance 1e-4.

- **Prior consistency**: For EACH variable without evidence: sum(posterior) == 1.0 to 1e-9.

- **Evidence clamping**: Observe Rain=true, query Rain.
  P(Rain=true | Rain=true) == 1.0 exact.

- **D-separation**: In the rain network, Rain and Sprinkler are independent a priori
  (not d-connected without observing WetGrass). Verify that P(Sprinkler|Rain=true) gives
  a different result from P(Sprinkler|Rain=true, WetGrass=true) — WetGrass is a
  collider that activates the dependency (explaining away).

- **CPT missing rows — silent probability 0**: Build a network where a node's CPT
  is missing a row for a specific parent configuration. Query with that configuration.
  The missing row contributes probability 0 during factor construction.
  Verify the result is consistent (no error, probability reflects the 0 contribution).
  Reference: CORRECTNESS.md — "CPT validation does not check completeness
  (all parent configurations present). Missing rows silently contribute probability 0."

### 1.11 engine/fuzzy/ — (Strategy A: exhaustive monotonicity + B: known-answer)

READ the existing tests in fuzzy engine/ first. Generate only what is NOT covered.

#### Exhaustive monotonicity

- **Tipping monotonicity**: Fix service=5.0. Vary food from 0 to 10 with step=0.5 (21 values).
  Compute tip for each value. Verify that the tip sequence is NON-DECREASING.
  Repeat fixing food=5.0 and varying service.
  Note: monotonicity depends on the rules. If the rules are monotone (better input →
  better output), the system must be too. Document the conditions.

#### Known-answer

- **Single-rule Sugeno**: A single input "x" with a single term "high" (membership=1 at x=10),
  a single output "y" with singleton=25. For input x=10: activation=1.0,
  output = 1.0*25 / 1.0 = 25.0 exact.

- **Output bounds**: For EACH defuzzification method, with tipping rules and
  extreme inputs (food=0,service=0), (food=10,service=10), (food=5,service=5):
  output must be in [0, 30] (range of the tip variable).

- **No rules fire — output 0**: Build a fuzzy engine where the input values
  produce zero membership in all rule antecedents (e.g., input far outside
  all membership function ranges). All rule strengths are 0.
  Mamdani: output == 0.0. Sugeno: output == 0.0.
  Verify no panic and output is exactly 0.
  Reference: CORRECTNESS.md — "No rules fire -> output 0.0" and
  "Rules with zero strength correctly skipped."

- **Missing input variable — closed-world degree 0**: Engine with input variables
  {temperature, humidity}. Call Infer with only {"humidity": 50} (temperature missing).
  The missing variable is skipped during fuzzification (degree 0 for all its terms).
  Rules referencing temperature get activation 0. Output reflects only humidity-based rules.
  Verify: no error, no panic. Output is 0 if all rules require temperature.
  Reference: CORRECTNESS.md — "Missing input variable → degree 0 for all terms
  (closed-world assumption)."

### 1.12 engine/causal/ — (Strategy B: adversarial + known-answer)

READ the existing tests in causal/ first. Generate only what is NOT covered.

#### Adversarial: confounder SCM

Build SCM with confounder:
```
U (exogenous, value=1)
U → X (X = U * 2)
U → Y (Y = U * 3 + X * 0)
```
Without intervention (observational): if we observe X=2, then U=1, Y=3.
With do(X=5): U does not change (still 1), Y = U*3 + X*0 = 3.
do(X=5) does NOT change Y because X has no direct edge to Y.

Build ANOTHER SCM with real causal effect:
```
U → X (X = U * 2)
X → Y (Y = X + 10)
```
Observational: X=2, Y=12.
do(X=5): Y = 5+10 = 15. do DOES change Y.

- **do != observe with confounder**: Compare Propagate(observations) vs Intervene(interventions)
  and verify that they give DIFFERENT results in the presence of a confounder.

#### Known-answer: Pearl examples

- **Counterfactual preservation**: SCM X→Z→Y, factual X=5, hypothetical do(Z=7).
  X is not a descendant of Z, so X keeps its factual value (5).
  Y is a descendant of Z, so Y is recomputed: Y = 7+3 = 10.
  Verify: result.Values["X"] == 5 (preserved), result.Values["Y"] == 10 (recomputed).

- **Intervention idempotence**: SCM X→Y (Y=X+3). Propagate with X=5 gives Y=8.
  Intervene with do(X=5) must give EXACTLY the same result: Y=8.
  Difference from Propagate: none (because the intervention equals the observation).

- **Diamond graph**: SCM U→X, U→Z, X→Y, Z→Y (Y = X + Z).
  Factual: U=1 → X=f1(U), Z=f2(U), Y=X+Z.
  do(X=10): only X changes, Z keeps its factual value (because U→Z does not go through X).
  Y is recomputed with X=10 and Z=factual_Z.

### 1.13 engine/mcdm/ — (Strategy B: known-answer + boundary)

#### Known-answer AHP (Saaty 1980)

- **Perfectly consistent 3x3 matrix**:
  If the real weights are [0.6, 0.3, 0.1], the comparison matrix is:
  | 1   | 2   | 6   |
  | 1/2 | 1   | 3   |
  | 1/6 | 1/3 | 1   |
  CR must be exactly 0.0 (or < 1e-10).
  Weights must be [0.6, 0.3, 0.1] to tolerance 1e-9.

- **Saaty textbook example**: Use the published example in Saaty (1980) table 3.1
  (criteria selection for school). Verify that the weights and CR match the
  published values.

- **CR boundary**: Build a matrix with CR = 0.099 (just below the threshold).
  Must pass validation. Slightly perturb to CR = 0.101.
  Must fail validation.

- **Weights sum to 1**: For ANY valid matrix, sum(weights) == 1.0 to 1e-9.

#### Known-answer TOPSIS (Hwang & Yoon 1981)

- **Strict dominance**: Alternative A = [10, 10, 10], B = [1, 1, 1].
  All benefit criteria. A dominates B. Score(A) > Score(B).

- **Ideal alternative**: If A == exact positive ideal solution:
  Score(A) must be 1.0 (D+ = 0, D- > 0, C = D-/(0+D-) = 1).

- **Anti-ideal alternative**: If A == negative ideal solution:
  Score(A) must be 0.0 (D- = 0, C = 0/(D++0) = 0).

- **Benefit vs cost**: With mixed criteria (column 1 = benefit, column 2 = cost):
  Alternative [10, 1] (high benefit, low cost) must dominate [1, 10].

- **Weight sensitivity**: Change weights dramatically and verify that the ranking changes.
  E.g., weights [0.9, 0.1] vs [0.1, 0.9] with alternatives that favor different criteria.

### 1.14 math/graph/ — (Strategy A: exhaustive on small graphs + B: known-answer)

READ the existing tests in graph/ first. Generate only what is NOT covered.

#### Exhaustive on small graphs

- **Directed graph — adjacency correctness**: Build all directed graphs on 3 vertices
  (2^(3*2)=64 possible edge sets). For each graph, verify that Neighbors(v) returns
  exactly the vertices reachable by a single edge from v.

- **Undirected graph — symmetry**: Build all undirected graphs on 3 vertices
  (2^3=8 possible edge sets). For each graph, verify that if (u,v) is an edge
  then (v,u) is also an edge (adjacency is symmetric).

- **DAG — topological sort correctness**: Build all DAGs on 4 vertices
  (enumerate acyclic orientations). For each DAG, verify that TopologicalSort
  returns an ordering where for every edge (u,v), u appears before v.

- **DAG — cycle rejection**: For each non-DAG (has a cycle), verify that
  the DAG constructor or validation returns an error.

#### Known-answer

- **BFS shortest path**: Known graph (Dijkstra's example from CLRS Fig. 24.6):
  vertices {s, t, x, y, z}, edges with weights. Verify shortest distances match
  published values.

- **Centrality — star graph**: Star graph with center c and 4 leaves.
  Degree centrality of c = 4/4 = 1.0. Betweenness of c = 1.0 (all paths go through c).
  Closeness of c = 4/4 = 1.0.
  Reference: Freeman (1979) centrality measures.

- **Bipartite verification**: K(2,3) (complete bipartite) must be bipartite.
  K(3) (complete graph on 3 vertices, triangle) must NOT be bipartite.

- **MST — known weight**: Graph with known MST weight from CLRS.
  Verify that the total weight of the MST matches the expected value.

- **Graph coloring — validity**: For any coloring produced by greedy coloring,
  verify that no two adjacent vertices share the same color.

- **Matching — known cardinality**: Bipartite graph with known maximum matching size.
  Verify the matching size is correct.

- **Tree properties**: A tree on n vertices must have exactly n-1 edges,
  be connected, and be acyclic.

- **Tree RemoveEdge cascade**: Tree with root A, edges A→B, B→C, B→D, A→E.
  RemoveEdge(A→B) must cascade-remove the disconnected subtree {B, C, D}.
  After removal: tree contains only {A, E} with edge A→E.
  Verify tree invariants hold (single root, n-1 edges, connected, acyclic).
  Reference: CORRECTNESS.md edge case — "Tree.RemoveEdge cascade-removes the
  disconnected subtree (edge.To + all descendants)."

- **Tree RemoveNode cascade**: Same tree. RemoveNode(B) must cascade-remove
  the entire subtree rooted at B: {B, C, D}.
  After removal: tree contains only {A, E} with edge A→E.
  Reference: CORRECTNESS.md edge case — "Tree.RemoveNode cascade-removes the
  entire subtree rooted at the removed node."

#### Multigraph

- **DirectedMultigraph parallel edges**: Add two edges A→B with different IDs and weights.
  Verify both edges are stored. Neighbors(A) returns B (once). EdgesBetween(A,B) returns 2 edges.
- **UndirectedMultigraph parallel edges**: Add two edges A-B with different IDs.
  Verify both edges are stored. Adjacency is symmetric.
- **Multigraph IsDirected**: DirectedMultigraph.IsDirected() == true.
  UndirectedMultigraph.IsDirected() == false.

#### Traversal invariants

- **BFS level order**: From a source vertex, BFS must visit vertices in
  non-decreasing distance order. Verify that dist(visited[i]) <= dist(visited[i+1]).

- **DFS — all reachable visited**: For each connected component, verify that
  DFS from any vertex in the component visits all vertices in that component.

#### Shortest path algorithms

- **Dijkstra weighted shortest path**: Known graph with varying positive weights.
  Verify shortest distances match hand-computed values.
  Reference: CLRS Ch. 24; Dijkstra (1959).

- **Bellman-Ford with negative edges**: Graph with negative edge weights (no negative cycles).
  Verify distances match hand-computed values.
  Also: graph WITH a negative cycle — must return error or indicate negative cycle.
  Reference: CLRS Ch. 24; Bellman (1958).

- **Floyd-Warshall all-pairs**: 3-vertex directed graph with known all-pairs distances.
  Verify that shorter indirect paths are preferred over direct edges.
  Reference: CLRS Ch. 25.

- **AllPaths known**: Diamond graph A→B, A→C, B→D, C→D.
  AllPaths(A,D) returns 2 paths. Disconnected: 0 paths.
  Reference: Diestel "Graph Theory" Ch. 1.

#### Structure algorithms

- **SCC (Tarjan) known decomposition**: Directed graph with known SCCs.
  Verify SCC count and compositions match expected values.
  Reference: Tarjan (1972); CLRS Ch. 22.5.

- **Bridges known**: Two triangles connected by a single edge.
  Verify the bridge set contains exactly the connecting edge.
  Reference: Tarjan (1974).

- **Articulation points known**: Same two-triangle graph.
  Verify the articulation point set matches expected vertices.
  Reference: Tarjan (1974).

#### Optimization algorithms

- **Max flow known (Edmonds-Karp)**: Network with known max flow value.
  Verify max flow matches hand-computed value.
  Reference: CLRS Ch. 26.

- **MinCut — max-flow min-cut theorem**: Same network as max flow test.
  Verify that the min cut capacity equals the max flow value.
  Verify that the cut partitions the graph into two sets S (containing source)
  and T (containing sink), with no residual capacity from S to T.
  Reference: Max-flow min-cut theorem (Ford & Fulkerson 1956); CLRS Ch. 26.

#### Centrality — PageRank

- **PageRank convergence**: Small known graph (e.g., 4-node with damping=0.85).
  Verify that sum(PageRank) ≈ 1.0 and values match reference computation.
  Reference: Brin & Page (1998).

#### Documented limitations (informational tests)

These tests document known limitations from CORRECTNESS.md. They do not protect
invariants — they verify that the limitations ARE as documented.

- **Dijkstra with negative weights — incorrect result**: Graph A→B(1), B→C(-5), A→C(0).
  Dijkstra from A to C returns 0 (direct), but true shortest is -4 (via B).
  Verify Dijkstra returns the WRONG answer (0, not -4).
  Verify Bellman-Ford returns the CORRECT answer (-4).
  Reference: CORRECTNESS.md — "Dijkstra does not validate for negative weights
  (silently gives wrong answers)."

- **Floyd-Warshall with negative cycle — no detection**: Graph with A→B(1), B→C(-3), C→A(1).
  Negative cycle weight = -1. Floyd-Warshall completes without error but produces
  incorrect distances (negative diagonal or arbitrarily negative values).
  Verify it does NOT return an error (unlike Bellman-Ford which would detect it).
  Reference: CORRECTNESS.md — "Floyd-Warshall does not detect negative cycles."

- **Betweenness ignores edge weights**: Graph A→B(1), A→C(100), B→C(1).
  Betweenness treats all edges as unweighted. The shortest weighted path A→B→C
  is ignored; betweenness treats A→C as a direct 1-hop path.
  Verify that Betweenness(B) == 0 (no shortest unweighted paths through B, since
  A→C is direct). In weighted shortest paths, B WOULD be on the shortest path.
  Reference: CORRECTNESS.md — "Betweenness uses unweighted BFS (ignores edge weights)."

- **Closeness centrality — disconnected graph (Wasserman & Faust)**: Graph with two
  components: {A-B-C} and {D-E}. Closeness(A) must use the Wasserman & Faust
  generalization: closeness = reachable/(n-1) × reachable/totalDist.
  Verify Closeness(A) > 0 (not infinity or zero) despite disconnected graph.
  Reference: CORRECTNESS.md — "For disconnected graphs, uses Wasserman & Faust
  generalization with reachable node count." Freeman (1979).

- **Undirected self-loop degree convention**: Undirected graph with vertex A having
  a self-loop. Degree(A) must count the self-loop as 2 (standard convention).
  If A also has an edge to B: Degree(A) == 3 (2 for self-loop + 1 for A-B).
  Reference: CORRECTNESS.md — "Self-loops stored once, degree counts self-loops as 2."

### 1.15 math/fsm/ — (Strategy A: exhaustive on small automata + B: known-answer)

READ the existing tests in fsm/ first. Generate only what is NOT covered.

#### Exhaustive on small automata

- **DFA acceptance exhaustive**: Build a DFA over alphabet {0, 1} with 2 states
  that accepts strings ending in "1". Enumerate ALL strings up to length 4
  (2^0 + 2^1 + 2^2 + 2^3 + 2^4 = 31 strings). For each string, verify that
  the DFA accepts iff the string ends in "1".

- **Determinism check**: For a DFA, verify that for EVERY (state, symbol) pair
  there is EXACTLY one transition. For an NFA, verify that there exists at least
  one (state, symbol) pair with 0 or >1 transitions.

#### Known-answer

- **Reachability — disconnected states**: FSM with states {A, B, C, D} where
  D has no incoming transitions from any reachable state.
  Reachable({A}) must return {A, B, C} (or whatever is reachable) and exclude D.
  Reference: Hopcroft et al. Ch. 2.

- **Acceptance — classic examples**:
  - Binary string divisible by 3: DFA with 3 states. Verify that "110" (=6) is accepted,
    "111" (=7) is rejected, "0" (=0) is accepted, "" (empty) is accepted (0 is divisible by 3).
  - Even number of 'a's: DFA with 2 states over {a, b}. Verify acceptance on
    "", "b", "aa", "ab", "ba", "aab", "aba".

- **Transition function determinism**: After adding transitions, verify that
  IsDeterministic() correctly identifies DFA vs NFA.

- **Dead states**: States with no outgoing transitions.
  Build FSM with states {A, B, C, Dead} where Dead has self-loop and is non-accepting.
  Verify DeadStates returns [Dead]. Build FSM where all states reach accepting: DeadStates=[].
  Reference: Hopcroft et al. Ch. 2.

- **Completeness check**: IsComplete returns true iff every (state, event) pair has a transition.
  Build incomplete DFA (missing one transition): IsComplete=false.
  Build complete DFA (all transitions defined): IsComplete=true.
  Reference: Hopcroft et al. Ch. 2.

- **IsComplete — empty alphabet (vacuous)**: FSM with states but no events/transitions.
  IsComplete must return true (vacuously — no (state, event) pairs to check).
  Reference: CORRECTNESS.md — "Empty alphabet returns isComplete=true (vacuously)."

- **Guard evaluation order — insertion order**: FSM with state A, event "go",
  two guarded transitions: T1 (guard: x>5, target: B) added first,
  T2 (guard: x>3, target: C) added second. With x=6 (both guards true):
  first-match semantics → T1 fires (B is reached, not C).
  With x=4 (only T2 guard true): T2 fires (C is reached).
  Reference: CORRECTNESS.md — "Guard evaluation order is insertion order."

### 1.16 math/markov/ — (Strategy B: known-answer + adversarial)

READ the existing tests in markov/ first. Generate only what is NOT covered.

#### Known-answer

- **2-state weather chain**: States {Sunny, Rainy}.
  Transition matrix: P(Sunny→Sunny)=0.8, P(Sunny→Rainy)=0.2,
  P(Rainy→Sunny)=0.4, P(Rainy→Rainy)=0.6.
  Steady state: π·P = π and π_S + π_R = 1.
  Solving: 0.8·π_S + 0.4·π_R = π_S → 0.4·π_R = 0.2·π_S → π_S = 2·π_R.
  With π_S + π_R = 1: π_S = 2/3, π_R = 1/3.
  Verify to tolerance 1e-9.
  Reference: Norris "Markov Chains" Ch. 1.

- **StepN — 2-step transition**: Same weather chain. StepN(2) gives P^2:
  P^2[Sunny→Sunny] = 0.8*0.8 + 0.2*0.4 = 0.72.
  P^2[Sunny→Rainy] = 0.8*0.2 + 0.2*0.6 = 0.28.
  Verify exact values to tolerance 1e-9.

- **StepN — n=0 edge case**: StepN(0) from state Sunny must return identity:
  P(Sunny→Sunny) = 1.0, P(Sunny→Rainy) = 0.0.
  Reference: CORRECTNESS.md — "StepN with n=0 returns initial state with probability 1."

- **Absorbing chain — gambler's ruin**: States {0, 1, 2, 3} where 0 and 3 are absorbing.
  P(1→0)=0.5, P(1→2)=0.5, P(2→1)=0.5, P(2→3)=0.5.
  Absorption probabilities from state 1: P(absorb at 0) = 2/3, P(absorb at 3) = 1/3.
  From state 2: P(absorb at 0) = 1/3, P(absorb at 3) = 2/3.
  Expected absorption time from state 1: 2 steps. From state 2: 2 steps.
  Reference: Kemeny & Snell "Finite Markov Chains" Ch. 3.

- **Classification — communicating classes**: Chain with states {A, B, C, D, E}.
  A↔B (communicate), C↔D (communicate), A→C (one-way), E is absorbing.
  Communicating classes: {A, B}, {C, D}, {E}.
  {A, B} is transient (can leave to C). {C, D} is transient (if can reach E) or
  recurrent (if cannot leave). {E} is recurrent (absorbing).
  Verify classification matches expected.

#### Adversarial

- **Periodic chain**: States {A, B} with P(A→B)=1, P(B→A)=1.
  Period = 2. Steady state still exists: π_A = π_B = 0.5.
  But P^n does NOT converge (oscillates). Verify that SteadyState
  returns the correct stationary distribution despite periodicity.

- **Nearly-absorbing chain**: Transition matrix with one state having
  P(self)=0.9999 and P(other)=0.0001. SteadyState must still converge
  (not get stuck in numerical issues). Verify that the steady state
  assigns almost all probability to the near-absorbing state.

- **Absorption with non-absorbing recurrent class**: Chain with states {T1, T2, A, R1, R2}.
  A is absorbing. R1↔R2 form a non-absorbing recurrent class. T1→A(0.5), T1→R1(0.5). T2→T1(1.0).
  Absorption probabilities from T1: P(absorb at A) = 0.5, P(enter {R1,R2}) = 0.5.
  Probabilities sum to 1.0 but across absorbing AND non-absorbing recurrent classes.
  Reference: CORRECTNESS.md — "absorption probabilities for transient states may sum to
  less than 1 (complement is probability of entering non-absorbing recurrent class)."

- **Mean first passage**: Weather chain. MFP(Sunny→Rainy) and MFP(Rainy→Sunny).
  Verify against closed-form solution via (I-Q)h = 1.
  MFP(from==to) = 0 (convention: already there).
  Reference: Kemeny & Snell §4.4.

- **Period**: 2-state oscillating chain A↔B (P(A→B)=1, P(B→A)=1).
  Period(A)=2. IsErgodic=false (period > 1).
  Reference: Norris Definition 1.2.1.

- **Simulate stochastic consistency**: Run 10000 simulations (fixed seed) of the weather chain
  for 100 steps. Empirical stationary distribution should approximate theoretical π
  within statistical tolerance (~0.05 for 10000 samples).
  Reference: Norris §1.1.

- **Row validation**: A matrix with rows not summing to 1.0 must be rejected
  as invalid. E.g., [[0.5, 0.3], [0.4, 0.6]] — row 0 sums to 0.8.

### 1.17 math/stats/ — Descriptive Functions (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for 20 functions. READ existing tests first.

#### Known-answer — reference values (NIST, Wolfram Alpha)

- **Mean**: [2, 4, 6, 8] → 5.0 exact.
- **Variance (population)**: [2, 4, 6, 8] → 5.0 exact.
- **SampleVariance (Bessel's)**: [2, 4, 6, 8] → 20/3 ≈ 6.666666666667.
- **StdDev**: [2, 4, 6, 8] → sqrt(5) ≈ 2.236067977500.
- **Median (odd)**: [1, 3, 5, 7, 9] → 5.0 exact.
- **Median (even)**: [1, 3, 5, 7] → 4.0 exact (average of 3 and 5).
- **Percentile**: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10], p=25 → 3.25 (linear interpolation).
  p=50 → 5.5. p=75 → 7.75. p=0 → 1. p=100 → 10.
- **Covariance (population)**: X=[1,2,3,4,5], Y=[2,4,6,8,10] → 4.0 exact (perfect linear).
- **SampleCovariance**: Same X,Y → 5.0 exact (n-1 denominator).
- **Correlation (Pearson)**: X=[1,2,3,4,5], Y=[2,4,6,8,10] → 1.0 exact (perfect positive).
  X=[1,2,3,4,5], Y=[10,8,6,4,2] → -1.0 exact (perfect negative).
  X=[1,2,3,4,5], Y=[5,5,5,5,5] → 0.0 (no variance in Y — undefined, but typically 0).
- **LinearRegression (OLS)**: X=[1,2,3,4,5], Y=[2,4,6,8,10] →
  slope=2.0, intercept=0.0 exact.
  Reference: OLS: slope=Cov(X,Y)/Var(X), intercept=ȳ-slope·x̄.
- **RSquared**: Same data → 1.0 exact (perfect fit).
- **Skewness (Fisher-Pearson)**: [1, 2, 2, 3, 3, 3, 4, 4, 5] — verify against
  reference calculation. Symmetric data → 0.0.
- **Kurtosis (excess)**: Normal-like data should have kurtosis ≈ 0.
  Uniform [1,2,3,4,5,6,7,8,9,10] → -1.2 (platykurtic). Verify against Wolfram Alpha.
- **WeightedMean**: values=[2,4,6], weights=[1,2,3] → (2+8+18)/6 = 28/6 ≈ 4.666667.
- **WeightedVariance**: Same values/weights → verify against formula Σw(x-μ̄)²/Σw.
- **GeometricMean**: [1, 2, 4, 8] → (1×2×4×8)^(1/4) = 64^(1/4) ≈ 2.828427124746.
- **HarmonicMean**: [1, 2, 4] → 3/(1 + 0.5 + 0.25) = 3/1.75 ≈ 1.714285714286.
- **MAD (Median Absolute Deviation)**: [1, 2, 3, 4, 5] → median=3,
  deviations=[2,1,0,1,2], MAD=median([0,1,1,2,2])=1.0.
- **Mode**: [1, 2, 2, 3, 3, 3] → 3. [1, 1, 2, 2] → either 1 or 2 (multi-modal).

Tolerance: 1e-9 for all computed values.

### 1.18 math/stats/ — Probability Functions (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for IsValid, Normalize, Complement, Entropy.

- **IsValid — valid distribution**: [0.3, 0.3, 0.4] → true (sums to 1.0, all in [0,1]).
- **IsValid — negative probability**: [0.5, -0.1, 0.6] → false.
- **IsValid — sum != 1**: [0.3, 0.3, 0.3] → false (sums to 0.9).
- **Normalize**: [2, 3, 5] → [0.2, 0.3, 0.5] exact.
- **Normalize — already normalized**: [0.25, 0.75] → [0.25, 0.75] unchanged.
- **Complement**: 0.3 → 0.7 exact. 0.0 → 1.0. 1.0 → 0.0.
- **Entropy — uniform**: [0.25, 0.25, 0.25, 0.25] → log₂(4) = 2.0 bits exact.
  Maximum entropy for n outcomes.
- **Entropy — certain**: [1.0, 0.0, 0.0] → 0.0 bits exact (using 0·log(0)=0 convention).
- **Entropy — binary**: [0.5, 0.5] → 1.0 bit exact.
- **Entropy — skewed**: [0.9, 0.1] → -0.9·log₂(0.9) - 0.1·log₂(0.1) ≈ 0.468996.
  Reference: Shannon (1948).

Tolerance: 1e-9 for all values.

### 1.19 math/stats/ — Bayes Theorem (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for BayesTheorem, TotalProbability, ChainRule, Independent.
Section 1.8 already has one adversarial Bayes test. These cover the remaining functions.

- **BayesTheorem — textbook**: P(H)=0.01, P(E|H)=0.95, P(E)=0.059.
  P(H|E) = 0.01×0.95/0.059 ≈ 0.161017. Verify to 1e-6.
  Reference: Bayes (1763), standard formulation.

- **TotalProbability**: Hypotheses H1, H2, H3 with P(H1)=0.2, P(H2)=0.5, P(H3)=0.3.
  P(E|H1)=0.1, P(E|H2)=0.4, P(E|H3)=0.8.
  P(E) = 0.2×0.1 + 0.5×0.4 + 0.3×0.8 = 0.02 + 0.20 + 0.24 = 0.46.
  Verify exact to 1e-9.

- **ChainRule**: P(A)=0.5, P(B|A)=0.3, P(C|A,B)=0.8.
  P(A,B,C) = P(A)·P(B|A)·P(C|A,B) = 0.5×0.3×0.8 = 0.12 exact.
  Reference: Product rule of probability.

- **Independent**: P(A)=0.4, P(B)=0.6. If independent: P(A∩B)=0.24.
  Independent(0.4, 0.6, 0.24) → true. Independent(0.4, 0.6, 0.30) → false.

- **TotalProbability — partition of one**: Single hypothesis P(H)=1.0, P(E|H)=0.7.
  P(E) = 0.7. Degenerate case.

Tolerance: 1e-9 for exact products, 1e-6 for divisions.

### 1.20 math/graph/ — Matrix Operations (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for ToMatrix, MatrixMultiply (tropical), MatrixPower, TransitiveClosure.

- **ToMatrix — directed 3-vertex graph**: Vertices {A, B, C}, edges A→B(3), B→C(2), A→C(7).
  Distance matrix:
  | | A | B | C |
  |---|---|---|---|
  | A | 0 | 3 | 7 |
  | B | Inf | 0 | 2 |
  | C | Inf | Inf | 0 |
  Verify exact values (0 diagonal, Inf for non-adjacent, min weight for multi-edges).
  Reference: Standard adjacency/distance matrix (CLRS Ch. 25).

- **ToMatrix — undirected**: Same graph but undirected. Matrix must be symmetric:
  M[A][B] == M[B][A] == 3.

- **MatrixMultiply (tropical min-plus)**: Given two 2×2 matrices:
  A = [[0, 3], [Inf, 0]], B = [[0, Inf], [2, 0]].
  C[i][j] = min_k(A[i][k] + B[k][j]).
  C[0][0] = min(0+0, 3+2) = 0. C[0][1] = min(0+Inf, 3+0) = 3.
  C[1][0] = min(Inf+0, 0+2) = 2. C[1][1] = min(Inf+Inf, 0+0) = 0.
  C = [[0, 3], [2, 0]]. Verify exact.
  Reference: Tropical semiring (Maclagan & Sturmfels).

- **MatrixPower — 2-step distances**: For the 3-vertex graph above,
  M² (tropical) gives 2-step shortest distances.
  M²[A][C] = min(M[A][A]+M[A][C], M[A][B]+M[B][C], M[A][C]+M[C][C])
           = min(0+7, 3+2, 7+0) = 5.
  Verify that M²[A][C] == 5 (shorter via B than direct).

- **TransitiveClosure (Warshall)**: Graph A→B, B→C.
  Transitive closure must include A→C (reachable via B).
  TC[A][C] == true. TC[C][A] == false.
  Reference: Warshall (1962).

- **TransitiveClosure — complete graph**: K₃. All pairs reachable. TC is all true.

Tolerance: exact comparison for integer weights.

### 1.21 math/graph/ — Substructures, Properties & Operations (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for Bron-Kerbosch, Eulerian Path, Coloring, graph operations, property queries. Section 1.14 already tests Coloring. These cover the rest.

#### Cliques (Bron-Kerbosch)

- **K₄ cliques**: Complete graph K₄ (4 vertices, all edges).
  Only maximal clique is {A, B, C, D} (the whole graph).
  Reference: Bron & Kerbosch (1973).

- **Triangle + pendant**: Graph with triangle {A,B,C} plus vertex D connected only to A.
  Maximal cliques: {A,B,C} and {A,D}. Verify count == 2.

- **Independent set (no edges)**: 3 vertices, no edges.
  Each vertex is its own maximal clique. 3 cliques of size 1.

#### ChromaticNumber

- **ChromaticNumber — greedy upper bound**: ChromaticNumber returns the number
  of colors used by the greedy coloring, which is an UPPER BOUND on the true
  chromatic number. For K₃ (triangle): ChromaticNumber == 3 (exact, since χ(K₃)=3).
  For a bipartite graph: ChromaticNumber <= 3 (true χ=2, greedy may use 2 or 3).
  For an independent set (no edges): ChromaticNumber == 1 (exact).
  NOTE: This is a documented limitation — greedy coloring does NOT guarantee
  the true chromatic number (NP-hard).
  Reference: CORRECTNESS.md formal gap; Diestel "Graph Theory" Ch. 5.

#### Eulerian Path (Hierholzer)

- **Eulerian circuit — K₃ cycle**: Triangle A-B-C-A (undirected).
  All vertices have even degree (2). Eulerian circuit exists.
  Verify that the returned path visits all 3 edges exactly once and returns to start.
  Reference: Euler (1736), Hierholzer (1873).

- **Eulerian path — bridge graph**: Path A-B-C (undirected).
  A and C have odd degree (1), B has even degree (2).
  Exactly 2 odd-degree vertices → Eulerian path exists (A to C or C to A).
  Verify path visits all 2 edges exactly once.

- **No Eulerian path**: Graph with 4 odd-degree vertices.
  Must return error or indicate no Eulerian path exists.

#### Graph Operations

- **Subgraph**: Graph {A,B,C,D} with edges. Subgraph({A,B,C}) must contain
  only vertices A,B,C and edges between them. Edges to D excluded.

- **Union**: G1 = {A→B}, G2 = {B→C}. Union = {A→B, B→C}.
  Vertices: {A,B,C}. All edges from both graphs present.

- **Intersection**: G1 = {A→B, B→C}, G2 = {A→B, C→D}.
  Intersection: only {A→B} (common edge). Vertices: {A, B}.

- **Complement**: Graph {A,B,C} with edges {A→B, B→C}.
  Complement contains all edges NOT in original: {A→C, B→A, C→A, C→B}.
  Reference: Diestel "Graph Theory" Ch. 1.

- **Reverse**: Directed graph {A→B, B→C}.
  Reverse: {B→A, C→B}. All edge directions flipped.

- **Cartesian Product**: P₂ □ P₂ (path of 2 × path of 2) = C₄ (4-cycle).
  Verify vertex count = 4, edge count = 4.
  Reference: Imrich & Klavžar "Product Graphs".

#### Property Queries

- **IsDAG**: DAG → true. Graph with cycle → false. Undirected graph → false.
- **IsTree**: Tree → true. Graph with cycle → false. Disconnected forest → false.
- **IsBipartite**: K(2,3) → true. K₃ (triangle) → false.
- **InDegree / OutDegree**: Directed graph A→B, A→C, B→C.
  OutDegree(A)=2, InDegree(A)=0, OutDegree(B)=1, InDegree(B)=1, InDegree(C)=2, OutDegree(C)=0.

### 1.22 math/graph/ — Connected Components & Diameter (Strategy B: known-answer)

CORRECTNESS.md guarantees soundness for Connected Components and Diameter in the Structure section.
Section 1.14 already tests SCC, Bridges, Articulation Points, Topological Sort. These cover the rest.

#### Connected Components

- **Single component**: Connected undirected graph on 5 vertices.
  ConnectedComponents returns 1 component with all 5 vertices.

- **Three components**: Undirected graph with {A-B}, {C-D-E}, {F}.
  ConnectedComponents returns 3 components: {A,B}, {C,D,E}, {F}.

- **Single vertex**: One vertex, no edges. 1 component of size 1.

- **Complete graph**: K₅. 1 component with all 5 vertices.
  Reference: Diestel "Graph Theory" Ch. 1.

#### Diameter

- **Path graph P₅**: Vertices 1-2-3-4-5 (linear path).
  Diameter = 4 (distance from 1 to 5).
  Reference: Diestel "Graph Theory" Ch. 1.

- **Complete graph K₄**: Diameter = 1 (every pair is adjacent).

- **Cycle C₆**: 6-vertex cycle. Diameter = 3 (opposite vertices).

- **Star graph S₅**: Center + 4 leaves. Diameter = 2 (leaf to leaf via center).

- **Disconnected graph**: Two components. Diameter = -1 (convention for disconnected).
  Reference: CORRECTNESS.md states "Returns -1 for disconnected graphs."

#### Reachable, Ancestors, Descendants

- **Reachable from source**: Directed graph A→B→C, A→D. Reachable(A) must return {A, B, C, D}.
  Reachable(C) must return {C} (no outgoing edges from C).
- **Reachable — disconnected**: Vertex E with no edges. Reachable(E) == {E}.
- **Ancestors**: DAG A→B→C, D→B. Ancestors(C) must return {A, B, D} (all nodes with paths to C).
  Ancestors(A) must return {} (root node).
- **Descendants**: Same DAG. Descendants(A) must return {B, C}.
  Descendants(C) must return {} (leaf node).
  Reference: CORRECTNESS.md states "Reachable, Ancestors, Descendants: BFS/DFS-based, all correct."

## SECTION 2: GOLDEN SCENARIO CROSS-PARADIGM

Design ONE end-to-end "Loan Approval" scenario that demonstrates the integration of
multiple paradigms. The scenario must have HAND-CALCULATED VALUES.

### Scenario: Loan Approval Pipeline

A system that processes a loan application in 5 phases:

#### Phase 1 — Deductive: Eligibility

Rules:
- age >= 18 AND no_fraud → eligible
- eligible AND has_income → can_apply

Factual input: age_ok=true, no_fraud=true, has_income=true.
Expected: eligible=true, can_apply=true in 2 steps.

Rejection input: age_ok=true, no_fraud=false.
Expected: eligible=false, can_apply not derived.

#### Phase 2 — Bayesian: Default probability

Network: CreditHistory → Default, IncomeLevel → Default.
Concrete CPTs (invent reasonable values and calculate by hand):

P(CreditHistory=good) = 0.7, P(CreditHistory=bad) = 0.3
P(IncomeLevel=high) = 0.6, P(IncomeLevel=low) = 0.4
P(Default=yes | CreditHistory=good, IncomeLevel=high) = 0.02
P(Default=yes | CreditHistory=good, IncomeLevel=low) = 0.10
P(Default=yes | CreditHistory=bad, IncomeLevel=high) = 0.15
P(Default=yes | CreditHistory=bad, IncomeLevel=low) = 0.40

Query: P(Default=yes | CreditHistory=good, IncomeLevel=high) = 0.02.
Calculate by hand: with direct evidence, the posterior equals the CPT.

Query without evidence: P(Default=yes) = marginalize.
Calculate the exact value by hand and verify.

#### Phase 3 — Fuzzy: Risk Assessment

Variables:
- Input: income_ratio (0-100), debt_ratio (0-100)
- Output: risk_level (0-100)

Membership functions and concrete rules. For each variant,
calculate the Mamdani result by hand (or at least verify
that it is in the expected range with precise bounds).

#### Phase 4 — Causal: What-if

SCM: Income → DebtRatio → RiskScore
With concrete equations: DebtRatio = 100 - Income*0.8, RiskScore = DebtRatio * 0.5.

Factual: Income=50 → DebtRatio=60 → RiskScore=30.
Counterfactual do(Income=80): DebtRatio=36 → RiskScore=18.
Verify exact values.

#### Phase 5 — MCDM: Option ranking

3 loan options: {rate, term, amount}.
Criteria: rate (cost, minimize), term (benefit, maximize), amount (benefit, maximize).
Weights: [0.5, 0.3, 0.2].

Calculate TOPSIS scores by hand and verify the ranking.

### Variants with exact expected values

Each variant must have ALL intermediate values calculated.
Do not accept "the result should be reasonable" — specify the number.

1. **Clear approval**: Ideal applicant. Deductive: eligible. Bayesian: P(default) < 0.05.
   Fuzzy: risk < 30. Causal: what-if shows improvement. MCDM: option A wins.
   ALL values calculated by hand.

2. **Clear rejection**: Fraud detected. Deductive: NOT eligible (short circuit).
   The other paradigms are NOT executed (or are executed and confirm rejection).

3. **Borderline case**: Deductive: eligible. Bayesian: P(default) ≈ 0.15.
   Fuzzy: risk ≈ 50. Causal: do(income+20%) lowers risk to ≈ 35.
   MCDM: tied options.
   ALL values calculated by hand.

### Trace verification

For EACH variant, verify that EACH paradigm produces a non-empty Trace
with the correct phases:
- Deductive: steps with rule names
- Bayesian: Initialize, Propagate/Marginalize, Complete
- Fuzzy: Fuzzification, RuleEvaluation, Defuzzification
- Causal: Propagation/Intervention/Counterfactual, Complete
- MCDM: (no trace — verify result directly)

## SECTION 3: GOLDEN FILES (Behavioral Snapshots)

Mathematical invariants protect PROPERTIES. Golden files protect
CONCRETE BEHAVIORS: exact outputs that must NOT change between refactors.

If a refactor changes an output, the test fails. That FORCES investigation of whether the
change is intentional or a regression. Without golden files, a subtle drift
(tip=16.234 → tip=16.189) goes unnoticed because both are "in range".

### Principle

For each paradigm, define ONE canonical scenario with fixed inputs and outputs
frozen to the maximum precision the algorithm supports. The test compares against
the frozen value. If it changes, the test fails.

### 3.1 Golden: Bayesian — Rain network

Canonical scenario: P(Rain=true | WetGrass=true) with the standard rain network.
Value calculated by hand (see section 1.10): 0.35770.
This value must NOT change between refactors.
STRICT tolerance: 1e-4 — if it changes in the 4th decimal place, something broke.

Also freeze:
- P(Rain=true | WetGrass=true, Sprinkler=true) — explaining away
- P(Sprinkler=true | WetGrass=true) — marginal query
- Medical network: P(Disease=present | TestResult=positive)
- Each value with manual derivation in Appendix B.

### 3.2 Golden: Deductive — Business scenario

Freeze the exact result of the existing business scenario:
- Derived facts: {high_spend, loyal, premium, discount, notify}
- Steps: exact number of steps
- Trace: rule names in exact firing order
- Provenance of each fact (Asserted vs Derived, rule name)

### 3.3 Golden: Fuzzy — Canonical tipping

Freeze the outputs of the tipping problem with representative inputs:

| food | service | method | expected tip (frozen) |
|------|---------|--------|----------------------|
| 1.0 | 1.0 | Mamdani/Centroid | [compute and freeze] |
| 5.0 | 5.0 | Mamdani/Centroid | [compute and freeze] |
| 9.0 | 9.0 | Mamdani/Centroid | [compute and freeze] |
| 5.0 | 5.0 | Mamdani/Bisector | [compute and freeze] |
| 5.0 | 5.0 | Mamdani/MeanOfMax | [compute and freeze] |
| 9.0 | 9.0 | Sugeno | [compute and freeze] |

Procedure: run the current code, capture the values, and freeze them.
If it is not possible to run (the prompt generates specs, not executes), mark as
`[TO_FREEZE: run and capture]` and document the procedure.

### 3.4 Golden: Causal — Linear SCM

Freeze results of the SCM X→Z→Y (Z=2X, Y=Z+3):

| operation | inputs | X | Z | Y |
|-----------|--------|---|---|---|
| Propagate | X=5 | 5.0 | 10.0 | 13.0 |
| Intervene | do(Z=7) | 0.0 | 7.0 | 10.0 |
| Counterfactual | factual X=5, do(X=10) | 10.0 | 20.0 | 23.0 |
| Counterfactual | factual X=5, do(Z=7) | 5.0 | 7.0 | 10.0 |

These values may exist in existing unit tests, but without the "golden" label.
The acceptance test REPEATS them as explicit golden files with
"behavioral regression" failure message — their purpose is to detect drift between refactors.

### 3.5 Golden: MCDM — AHP + TOPSIS

Freeze:
- AHP: weights and CR for the consistent 3x3 matrix
- TOPSIS: scores and ranking for the dominance scenario

### Tolerance criteria for golden files

- Exact values (integers, booleans, strings): exact comparison
- Floating point from simple operations (causal): 1e-9
- Floating point from marginalization (bayesian): 1e-4
- Floating point from defuzzification (fuzzy): 1e-2 (discretization)

The failure message must ALWAYS say "behavioral regression" to distinguish it
from a mathematical invariant failure.

## SECTION 4: ERROR CONTRACTS

A mega-refactor can change error handling without noticing.
Error contracts verify that invalid inputs STILL produce the
same errors after the refactor.

### Principle

For each public function that returns an error, verify:
1. Invalid input → expected error (sentinel or type)
2. Valid input → NO error

### 4.1 math/fuzzy/ — Membership constructors

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| Triangular(5, 3, 7) | a=5 > b=3 | ErrInvalidRange |
| Triangular(1, 5, 3) | b=5 > c=3 | ErrInvalidRange |
| Triangular(5, 5, 5) | a == c | ErrInvalidRange |
| Trapezoidal(5, 3, 7, 9) | b < a | ErrInvalidRange |
| Trapezoidal(1, 2, 2, 1) | d < a | ErrInvalidRange |
| Gaussian(0, 0) | sigma == 0 | ErrInvalidRange |
| Gaussian(0, -1) | sigma < 0 | ErrInvalidRange |
| Sample(fn, 10, 5, 100) | lo > hi | ErrInvalidRange |
| Sample(fn, 0, 10, 0) | n == 0 | ErrEmptySamples |

For each row: verify that the returned error matches the sentinel
(using error matching or direct comparison per the module's pattern).
And verify that the valid counterpart does NOT return an error.

### 4.2 math/stats/ — Distribution constructors

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
| NewGumbel(0, 0) | beta == 0 | ErrInvalidParameter |
| NewGumbel(0, -1) | beta < 0 | ErrInvalidParameter |
| NewPareto(0, 3) | xm == 0 | ErrInvalidParameter |
| NewPareto(1, 0) | alpha == 0 | ErrInvalidParameter |
| NewGeometric(0) | p == 0 | ErrInvalidProb |
| NewGeometric(1.1) | p > 1 | ErrInvalidProb |
| NewHypergeometric(0, 5, 3) | N == 0 | ErrInvalidParameter |
| NewHypergeometric(50, -1, 5) | K < 0 | ErrInvalidParameter |
| NewHypergeometric(50, 60, 5) | K > N | ErrInvalidParameter |
| NewHypergeometric(50, 10, 0) | n == 0 | ErrInvalidParameter |
| NewHypergeometric(50, 10, 60) | n > N | ErrInvalidParameter |
| NewNegativeBinomial(0, 0.5) | r == 0 | ErrInvalidParameter |
| NewNegativeBinomial(5, 0) | p == 0 | ErrInvalidProb |
| NewNegativeBinomial(5, 1.1) | p > 1 | ErrInvalidProb |

Generate ONE test per constructor that verifies:
- Invalid input → correct error
- Valid input → nil error

### 4.3 math/logic/predicate/ — Quantifiers

Empty collections follow standard FOL semantics (vacuous truth/falsity),
so they are NOT error conditions. Only nil predicates are errors.

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| ForAll(coll, nil) | nil predicate | ErrNilPredicate |
| Exists(coll, nil) | nil predicate | ErrNilPredicate |

Valid empty-domain behavior (tested in section 1.4, NOT error contracts):
- ForAll([], pred) → true (vacuous truth)
- Exists([], pred) → false (vacuous falsity)
- Count([], pred) → 0
- Filter([], pred) → []

### 4.4 engine/bayesian/network/ — Network construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddNode (duplicate) | Node with already existing variable | error (contains "duplicate") |
| Validate (cycle) | Network with A→B→A | ErrCyclicNetwork |
| Validate (missing parent) | Node with unregistered parent | error (contains "not in network") |
| Validate (no outcomes) | Node without outcomes | error (contains "no outcomes") |

### 4.5 engine/causal/model/ — SCM construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddVariable (duplicate) | Duplicate variable | ErrDuplicateVariable (via ErrCausal) |
| AddVariable (nil eq) | Nil equation | ErrNilEquation (via ErrCausal) |
| Validate (cycle) | Cyclic SCM | ErrCyclicModel (via ErrCausal) |
| Validate (missing parent) | Unregistered parent | ErrParentNotFound (via ErrCausal) |

### 4.6 engine/mcdm/ — Validation

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| ahp.Analyze([]) | Empty matrix | ErrEmptyMatrix |
| ahp.Analyze(non-square) | Rows of different length | ErrNotSquareMatrix |
| ahp.Rank([], evals) | Empty weights | ErrEmptyMatrix |
| ahp.Rank(w, mismatched) | Inconsistent dims | ErrDimensionMismatch |
| topsis.Rank([], criteria) | Empty matrix | ErrEmptyInput |
| topsis.Rank(matrix, mismatched) | Inconsistent dims | ErrDimensionMismatch |

### 4.7 math/graph/ — Graph construction and operations

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| AddNode (duplicate) | Same ID twice | ErrDuplicateNode |
| AddEdge (missing from) | Non-existent From vertex | ErrNodeNotFound |
| AddEdge (missing to) | Non-existent To vertex | ErrNodeNotFound |
| DAG AddEdge (cycle) | Edge that creates a cycle | ErrCycleDetected |
| DAG AddEdge (self-loop) | From == To | ErrSelfLoop |
| Bipartite AddEdge (same partition) | Both vertices in same partition | ErrNotBipartite |
| ShortestPath (disconnected) | No path between vertices | ErrNoPath |
| ShortestPath (missing node) | Non-existent source vertex | ErrNodeNotFound |
| TopologicalSort (cyclic) | Cyclic directed graph | ErrNotDAG |
| BFS (missing start) | Non-existent start vertex | ErrNodeNotFound |
| DFS (missing start) | Non-existent start vertex | ErrNodeNotFound |
| Tree RemoveNode (root) | Removing root node | ErrInvalidEdge |
| Tree AddEdge (multiple parents) | Node already has parent | ErrMultipleParents |
| Tree NewTreeFrom (multiple roots) | DAG with >1 root | ErrNotTree |
| Coloring (empty graph) | Graph with no vertices | error or valid empty coloring |

### 4.8 math/fsm/ — FSM construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewMachine (unknown To state) | To state not in states | ErrInvalidTransition |
| NewMachine (no initial) | Initial state not in states | ErrNoInitialState |
| NewMachine (duplicate state) | Duplicate state ID | ErrDuplicateState |
| NewMachine (empty event) | Empty event string | ErrInvalidEvent |
| Send (unknown event) | Event with no transition from current state | ErrTransitionNotFound |

### 4.9 math/markov/ — Markov chain construction

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewChain (row sum != 1) | Row sums to 0.8 | ErrInvalidRow |
| NewChain (negative prob) | P[i][j] = -0.1 | ErrInvalidProbability |
| NewChain (non-square) | 1 row, 2 states | ErrInvalidMatrix |
| NewChain (empty) | 0 states | ErrEmptyChain |
| NewChain (duplicate state) | Same state name twice | ErrDuplicateState |
| SteadyState (reducible) | Not irreducible chain | ErrNotIrreducible |
| Absorption (no absorbing) | No absorbing states | ErrNoAbsorbingStates |
| MeanFirstPassage (unreachable) | Target unreachable from source | ErrSingularMatrix |

### 4.10 engine/deductive/ — Rule engine construction

READ the code to identify exact error sentinels. Likely contracts:

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewEngine (nil rules) | nil or empty rule set | error (no rules) |
| AddRule (nil condition) | Rule with nil condition | error (nil condition) |
| AddRule (nil action) | Rule with nil action/conclusion | error (nil action) |
| AddRule (empty name) | Rule with empty name | error (empty name) |

Verify valid counterparts do NOT return errors.

### 4.11 engine/fuzzy/ — Fuzzy engine construction

READ the code to identify exact error sentinels. Likely contracts:

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| NewEngine (no inputs) | Empty input variables | panic via cassert (inputVars empty) |
| NewEngine (no outputs) | Empty output variables | panic via cassert (outputVars empty) |
| NewEngine (no rules) | No rules defined | error or panic (verify) |

NOTE: Missing input at Infer time is NOT an error — the engine skips the missing
variable during fuzzification (closed-world: degree 0 for all terms). See section 1.11.
This is verified in the existing test TestEngine_Infer_missingInput.

### 4.12 math/stats/ — Descriptive functions and Bayes

| Function | Invalid input | Expected error |
|----------|--------------|----------------|
| Mean([]) | Empty slice | error (empty data) |
| Variance([]) | Empty slice | error (empty data) |
| Median([]) | Empty slice | error (empty data) |
| Percentile(data, -1) | p < 0 | error (invalid percentile) |
| Percentile(data, 101) | p > 100 | error (invalid percentile) |
| Correlation([], []) | Empty slices | error (empty data) |
| Correlation(x, y) (len mismatch) | Different lengths | error (dimension mismatch) |
| LinearRegression(x, y) (len mismatch) | Different lengths | error (dimension mismatch) |
| GeometricMean (negative) | Slice with negative value | error (non-positive value) |
| HarmonicMean (zero) | Slice with zero | error (zero value) |
| BayesTheorem(prior, likelihood, 0) | P(E) == 0 | error (division by zero) |

READ the code to confirm exact error sentinels. Verify valid counterparts do NOT return errors.

### Naming convention

All error contract tests are named TestErrorContract_[Package]_[Function].
The failure message says "error contract violated" to distinguish it from other failures.

## SECTION 5: PERFORMANCE BASELINES

A refactor that makes an algorithm 100x slower is a regression, even if it passes
all functional tests. Performance baselines detect degradation.

### Principle

DO NOT use absolute times (they depend on the machine). Instead:

1. **Guaranteed termination**: the test fails by timeout if the algorithm does not terminate
   within a reasonable limit (e.g., 5 seconds).
2. **Relative scaling**: measure T(N) and T(10N), verify that the ratio does not exceed
   the expected complexity (e.g., O(n^2) → ratio < 110).
3. **Algorithm vs algorithm**: VE must not be more than 10x slower than Enumeration
   on the same network (for small networks, both are similar).

### 5.1 Termination under pressure

These tests verify that algorithms DO NOT enter infinite loops with
inputs designed to stress:

- **DPLL with 15 variables**: Formula with 15 variables, ~100 clauses.
  Must terminate. If it does not terminate in 5 seconds, something is broken.

- **Forward chaining with 200 rules**: 200 rules in chain A→B→C→...
  Must reach fixed point. If it does not terminate in 5 seconds, something is broken.

- **Forward chaining with cyclic rules**: 50 rules forming cycles.
  Must terminate (monotonicity). If it does not terminate in 5 seconds, something is broken.

- **Variable elimination with 6 variables**: Complete Bayesian network with 6 vars.
  Must terminate. If it does not terminate in 5 seconds, something is broken.

Pattern: run the algorithm with a 5-second timeout. If it does not terminate, it is a regression.

### 5.2 Relative scaling

Measure how much execution time grows when input grows:

- **Forward chaining scaling**: Measure T(10 rules) and T(100 rules) in chain.
  Ratio must be < 150 (expected complexity ~ O(R*F) where R=rules, F=facts).
  If ratio > 150, there is a quadratic or worse regression.

- **DPLL scaling**: Measure T(5 vars) and T(10 vars) with random formulas (fixed seed).
  DPLL is exponential in worst case, but for random formulas the ratio must be
  manageable (< 100). If ratio > 1000, something is wrong.

- **VE scaling**: Measure T(3 vars) and T(5 vars).
  VE is exponential in treewidth, but for linear networks it must be polynomial.

Pattern: measure execution time with N and with 10N, compute ratio, verify against threshold.

### 5.3 Algorithm vs algorithm

- **VE vs Enumeration on rain network**: Both must produce the same result
  (already covered in section 1.10), AND the slower one must not be >10x the faster.
  For 3 variables, both are ~O(1). A ratio >10 would indicate unexpected overhead.

- **Mamdani vs Sugeno**: Sugeno should be faster (does not require defuzzification).
  Verify that Sugeno is not more than 2x slower than Mamdani (would be unexpected).

### 5.4 Graph algorithms

- **BFS/DFS termination with 1000 vertices**: Complete graph with 1000 vertices.
  BFS and DFS must terminate. If they do not terminate in 5 seconds, something is broken.

- **Shortest path scaling**: Measure T(100 vertices) and T(1000 vertices) on
  sparse random graphs (fixed seed). Expected complexity for Dijkstra: O((V+E)logV).
  Ratio must be < 150 for 10x vertex increase on sparse graphs.

- **Topological sort scaling**: DAG with N vertices in chain.
  Measure T(100) and T(1000). Expected O(V+E). Ratio must be < 20.

### 5.5 FSM operations

- **Acceptance scaling**: DFA with N states in a chain, input of length N.
  Measure T(100) and T(1000). Expected O(N). Ratio must be < 15.

- **Reachability termination**: FSM with 500 states and dense transitions.
  Must terminate. If it does not terminate in 5 seconds, something is broken.

### 5.6 Markov chain

- **Steady state convergence**: Chain with N states.
  Measure T(10 states) and T(50 states). Iterative method expected O(N² × iterations).
  Ratio must be < 500 (25x states increase → ~625x if quadratic, with some tolerance).

- **Absorption computation**: Absorbing chain with N transient states.
  Measure T(10) and T(50). Involves matrix inversion O(N³).
  Ratio must be < 2000.

### Convention

- Termination tests: verify that the algorithm terminates within the timeout
- Scaling tests: measure growth ratio T(10N)/T(N)
- Comparison tests: verify that an algorithm is not disproportionately slower than another
- Failure criterion: "performance regression" to distinguish from functional failures

## REQUIRED OUTPUT

Generate ONE markdown document of PURE SPECIFICATION — no code in any language.
This document is for a mathematical testing expert to review and say
"these tests are correct and complete" or "X is missing". The implementation in
code is done by other prompts (.claude/prompts/acceptance/01-07.md).

CRITICAL RULE: DO NOT include code blocks with test functions.
Only text, tables, mathematical derivations, and descriptions of what is tested.
The document must be independent of the implementation language.

Exact structure:

```
# Acceptance Test Specifications — modules/compute/

## Metadata
- Generated: [date]
- Based on: CORRECTNESS.md
- Existing tests reviewed: [list of files read]
- Strategies: Exhaustive-within-bounds + Adversarial + Known-answer

## Coverage summary

| Area | Existing tests | New tests | Total verifications |
|------|---------------|-----------|---------------------|
| ... | N | M | [number of individual verifications] |

## Required helpers

Describe the helpers the implementation will need (no code):
- generateFormulas(depth, vars): generates all formulas up to depth with the given vars
- fuzzyGrid(step): generates grid [0.0, step, ..., 1.0]
- assertFloat(name, got, want, tolerance): float comparison with tolerance
- Tolerances: floatTolerance=1e-9, probTolerance=1e-6, defuzzTolerance=0.5,
  goldenBayesian=1e-4, goldenFuzzy=1e-2

## Section 1: Mathematical invariants

### 1.N [Area]

#### Test: [descriptive name]
- **Strategy**: Exhaustive / Adversarial / Known-answer
- **Invariant**: [what mathematical property it protects]
- **Reference**: [theoretical source with page/section]
- **Verifications**: [number of individual verifications it runs]
- **Prerequisite**: [what existing tests already cover part of this]
- **Subtests**:
  - "subtest name": description, inputs, expected output
  - "subtest name": description, inputs, expected output
- **Failure criterion**: [when exactly should this test fail]
- **Expected values**: [list of concrete values and their derivation]

## Section 2: Golden Scenario

### Scenario: Loan Approval
- **Paradigms**: deductive + bayesian + fuzzy + causal + mcdm
- **Variants**: approved, rejected, borderline
- **Manual calculations**:
  For each variant, each phase: input → step-by-step calculation → expected output
  [tables with all intermediate values]

## Section 3: Golden Files

### 3.N [Paradigm]
- **Canonical scenario**: [description]
- **Frozen values**: [table of inputs → exact outputs with derivation]
- **Tolerance**: [value and justification]
- **Failure message**: "behavioral regression: ..."

## Section 4: Error Contracts

### 4.N [Package]
- **Functions covered**: [list]
- **Contract table**:

| Function | Invalid input | Reason | Expected error |
|----------|--------------|--------|----------------|
| ... | ... | ... | ... |

## Section 5: Performance Baselines

### 5.N [Algorithm]
- **Type**: Termination / Scaling / Comparison
- **Description**: what is measured, with what inputs, what threshold
- **Failure criterion**: "performance regression: ..."

## Appendix A: Tests that already exist and are NOT duplicated
[list with justification]

## Appendix B: Manual derivations of expected values
[step-by-step calculation of each known-answer value]

## Appendix C: Error contract inventory
[complete table: function → invalid input → expected error → sentinel]
```

## VALIDATION

Before finalizing, verify:
1. No test duplicates an existing one — justify for each.
2. Each invariant from CORRECTNESS.md marked as GUARANTEED has at least one test.
3. Exhaustive tests specify how many verifications they run.
4. Known-answer tests include the derivation of the expected value.
5. Adversarial tests explain WHY the input is pathological.
6. The golden scenario has ALL intermediate values calculated by hand.
7. Each test has a theoretical reference with section/page.
8. Golden files cover all 5 paradigms and have concrete frozen values.
9. Error contracts cover ALL public functions that return errors
    (see inventory in Appendix C: 17 stats constructors + 9 fuzzy + 2 predicate +
    4 bayesian + 4 causal + 6 mcdm + 15 graph + 5 fsm + 7 markov +
    ~4 deductive + ~3 fuzzy-engine + ~11 stats-descriptive/bayes = ~83 functions).
10. Performance baselines include at least: 7 termination, 8 scaling, 2 comparison.
11. Sections 1.17-1.22 cover all CORRECTNESS.md subsections that were not in 1.1-1.16:
    descriptive functions, probability functions, Bayes theorem, graph matrix operations,
    graph substructures/properties/operations, and connected components/diameter.
