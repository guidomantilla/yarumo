## ROLE

You are a testing engineer specialized in formal verification. You are performing
a SURGICAL PATCH on an existing acceptance test specification.

## TASK

The file `modules/compute/tests/specs/ACCEPTANCE_TESTS.md` is 1843 lines and mostly
correct. It is MISSING exactly 5 blocks of content that were recently added to the
generation prompt. Your job is to generate ONLY the missing blocks, formatted identically
to the existing spec, ready to be inserted at the specified locations.

**Output format**: For each missing block, output:

```
=== INSERT AFTER LINE: <line number> ===
=== CONTEXT: "<text of the line after which to insert>" ===

<content to insert>

=== END INSERT ===
```

## ABSOLUTE CONSTRAINTS

- Generate ONLY the 5 missing blocks listed below. Nothing else.
- Match the EXACT style of the existing spec: same heading levels, same field names
  (Strategy, Invariant, Reference, Verifications, Subtests, Failure criterion, Expected values).
- Do NOT regenerate anything that already exists in the spec.
- Do NOT change section numbers — the spec uses §1.7 for fuzzy, §1.8 for stats,
  §1.9 for graph, §1.11 for markov. Keep those numbers.
- Include the Coverage Summary table update as the last block.

## SOURCE MATERIAL

All content below comes from `modules/compute/prompts/ACCEPTANCE_TESTS.md` (the generation
prompt) which is already aligned with `modules/compute/CORRECTNESS.md`.

## MISSING BLOCK 1: Fuzzy Operations (6 tests)

**Insert after**: Section 1.7, after the last existing test "FuzzyDefuzzificationOutputInRange"
(line 747, text: "- **Failure criterion**: Any result outside [0, 10].")

**Source from prompt §1.7 "Operations"**:

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

Generate 6 test specifications in the same format as the existing fuzzy tests (####-level
headings, Strategy, Invariant, Reference, Verifications, Subtests, Expected values, etc.).

## MISSING BLOCK 2: 10 Distribution Known-Answer Tests

**Insert after**: Section 1.8, after the last existing distribution test "StatsBinomialDistribution"
(line 807, text: "- **Expected values**: PMF(3)=0.266827932, Mean=3.0, Variance=2.1.")

The existing spec has: Normal, Exponential, Beta, StudentT, ChiSquared, Poisson, Binomial.
Missing 10 distributions with these reference values:

- **Uniform(2, 8)**: PDF(5)=1/6≈0.166666666667, CDF(5)=0.5, Mean=5.0, Variance=3.0
- **Gamma(3, 2)** (shape=3, rate=2): Mean=1.5, Variance=0.75, CDF(1) verify via regularized incomplete gamma
- **Lognormal(0, 1)**: Mean=exp(0.5)≈1.648721270700, Variance=(exp(1)-1)*exp(1)≈4.670774270471, CDF(1)=0.5
- **Weibull(2, 1)** (k=2, λ=1): PDF(1)=2*exp(-1)≈0.735758882343, CDF(1)=1-exp(-1)≈0.632120558829, Mean=√π/2≈0.886226925453, Variance=1-π/4≈0.214601836603
- **FDist(5, 10)**: Mean=10/8=1.25, Variance=2*100*13/(5*64*6)≈1.354166666667
- **Gumbel(0, 1)**: Mean=γ≈0.577215664902, Variance=π²/6≈1.644934066848, CDF(0)=exp(-1)≈0.367879441171
- **Pareto(1, 3)** (xm=1, α=3): CDF(2)=1-(1/2)³=0.875, Mean=1.5, Variance=0.75
- **Geometric(0.3)** (failures-before-success): PMF(0)=0.3, PMF(2)=0.7²×0.3=0.147, Mean=7/3≈2.333333333333, Variance=70/9≈7.777777777778
- **Hypergeometric(50, 10, 5)** (N=50, K=10, n=5): Mean=1.0, Variance≈0.734693877551
- **NegativeBinomial(5, 0.4)** (r=5, p=0.4): Mean=7.5, Variance=18.75

Tolerance: 1e-9 for PDF/CDF, 1e-12 for Mean/Variance (exact values).

Generate 10 test specifications (one per distribution), same format as StatsNormalDistribution.

## MISSING BLOCK 3: 5 Hypothesis Tests

**Insert after**: Section 1.8, after the existing "StatsHypothesisTTestRejection" test
(line 852, text: "- **Expected values**: p < 0.001. Reject H0.")

The existing spec has: TTest known outcome + TTest rejection.
Missing 5 additional hypothesis tests:

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

- **Fisher exact 2×2**: Table [[8,2],[1,9]]. Two-tailed p-value.
  Verify against known hypergeometric calculation.

Generate 5 test specifications, same format as StatsHypothesisTTestKnownOutcome.

## MISSING BLOCK 4: PageRank Convergence Test

**Insert after**: Section 1.9, after the existing "Articulation points known" test
(line 995, text: "- **Expected values**: APs = {C, D}. Count = 2.")

The spec already has SCC, Bridges, Articulation Points, Max flow, Dijkstra, Bellman-Ford,
Floyd-Warshall, AllPaths. Missing only PageRank.

- **PageRank convergence**: Small known graph (e.g., 4-node with damping=0.85).
  Verify that sum(PageRank) ≈ 1.0 and values match reference computation.
  Reference: Brin & Page (1998).

Generate 1 test specification with full detail (graph structure, edge list,
hand-computed PageRank values, tolerance).

## MISSING BLOCK 5: Markov Simulate Stochastic Consistency

**Insert after**: Section 1.11, after the existing "Mean first passage" test
(line 1170, text: "- **Expected values**: MFP(Sunny→Rainy)=5.0, MFP(Rainy→Sunny)=2.5, MFP(Sunny→Sunny)=0.0.")

The spec has: steady state, N-step, absorption, classification, period, MFP, periodic
adversarial, nearly-absorbing adversarial, invalid row. Missing only Simulate.

- **Simulate stochastic consistency**: Run 10000 simulations (fixed seed) of the weather chain
  for 100 steps. Empirical stationary distribution should approximate theoretical π
  within statistical tolerance (~0.05 for 10000 samples).
  Reference: Norris §1.1.

Generate 1 test specification with full detail (chain, parameters, expected distribution,
tolerance justification).

## MISSING BLOCK 6: Coverage Summary Table Update

**Replace**: Lines 36-58 of the spec (the Coverage Summary table).

Current values that need updating:
- §1.7 fuzzy/ axioms: change "14" new tests → "20" (added 6 Operations tests)
- §1.8 stats/ distributions: change "14" new tests → "29" (added 10 distributions + 5 hypothesis)
- §1.9 math/graph/: change "21" new tests → "22" (added PageRank)
- §1.11 math/markov/: change "9" new tests → "10" (added Simulate)
- **Total** new tests: change "162" → "179"
- Update total verifications as needed

Generate the complete replacement table.

## FINAL INSTRUCTIONS

- Output EXACTLY 6 insert/replace blocks, nothing more.
- Each block must be self-contained and ready to copy-paste.
- Maintain the `---` horizontal rule conventions of the existing spec.
- Every test must have: #### heading, Strategy, Invariant, Reference, Verifications,
  Subtests (if applicable), Failure criterion, Expected values.
