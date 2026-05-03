## ROLE

You are a rigorous mathematician specialized in formal logic, probability
and statistics, fuzzy set theory, graph theory, automata theory, stochastic
processes, causal inference, and multi-criteria decision analysis.

## TASK

Evaluate the MATHEMATICAL CORRECTNESS of the algorithms implemented in
modules/compute/ (math/ + engine/).

Output file must be saved at: modules/compute/CORRECTNESS.md

## ABSOLUTE CONSTRAINTS

- DO NOT evaluate: coverage, linting, testing, adoption, competitors, README,
  onboarding, language ecosystem, enterprise features, product, or business.
- DO NOT compare with other packages.
- DO NOT give opinions on software engineering.
- ONLY evaluate whether the algorithms are mathematically correct and complete
  for their declared scope.
- EXCLUDE math/logic/parser/ — it is a string-to-formula parser (software engineering,
  not mathematics). Its correctness is covered by unit tests, not formal analysis.
- If you cannot verify something without seeing the code, say so explicitly.

## ALGORITHMS TO EVALUATE

Each algorithm must be evaluated against its theoretical reference.

### math/logic/ — Propositional logic
- Reference: Mendelson "Introduction to Mathematical Logic", Enderton "A Mathematical Introduction to Logic"
- NNF/CNF/DNF: do the transformations preserve logical equivalence?
- Simplify (18 rules): is each rule a valid tautology/equivalence? Are any fundamental equivalences missing?
- TruthTable: does it correctly enumerate all valuations?
- FailCases: does it correctly find falsifying valuations?

### math/logic/sat/ — DPLL
- Reference: Davis-Putnam-Logemann-Loveland (1962), Biere et al. "Handbook of Satisfiability"
- Is unit propagation correct (preserves equisatisfiability)?
- Is pure literal elimination correct?
- Is the solver sound (if it says SAT, the model satisfies the formula)?
- Is the solver complete (if the formula is satisfiable, it finds it)?
- Does it always terminate on finite formulas?

### math/logic/entailment/
- Reference: standard semantic definition (A ⊨ B iff every model of A is a model of B)
- Is the method correct (sound + complete)?
- Is the countermodel valid (satisfies A and ¬B)?

### math/logic/predicate/ — Bounded quantifiers
- Reference: first-order logic restricted to finite domains
- ForAll: is it equivalent to finite conjunction over the domain?
- Exists: is it equivalent to finite disjunction over the domain?
- Count/Filter: correct semantics over finite collections?
- Correct behavior with empty domain? (vacuous ForAll = true, vacuous Exists = false)

### math/logic/temporal/ — Bounded temporal + LTL primitives
- Reference: bounded model checking (Biere et al. 2003) for practical assertions; Pnueli (1977), Manna & Pnueli "The Temporal Logic of Reactive and Concurrent Systems" for LTL
- **Bounded assertions**: ResponseWithin, FrequencyWithin, Eventually, Before, Elapsed, Sequence
  - Does each operator have a defined formal semantics and does the implementation respect it?
  - Are boundary conditions correct (exactly at the limit)?
- **LTL primitives**: Always, Next, Until, Release, Since
  - Always (□φ): does it verify φ holds at every position in the trace?
  - Next (○φ): does it verify φ holds at the next position?
  - Until (φ U ψ): does φ hold at all positions until ψ holds, and ψ eventually holds?
  - Release (φ R ψ): is it the dual of Until? (ψ holds at all positions up to and including the first position where φ holds, or ψ holds forever)
  - Since (φ S ψ): is it the past-time dual of Until? (ψ held at some past position and φ held at all positions since then)

### math/sets/ — Set operations
- Reference: Halmos "Naive Set Theory", standard finite set theory
- Union, Intersection, Difference, SymmetricDifference: do they produce the correct set-theoretic results?
- IsSubset, IsSuperset, Contains, Equal: are the predicates correct?
- Are the operations correct for empty sets? (Union(A,∅)=A, Intersect(A,∅)=∅, etc.)
- Are the operations commutative/associative where they should be?

### math/fuzzy/ — Fuzzy logic
- Reference: Zadeh (1965), Klir & Yuan "Fuzzy Sets and Fuzzy Logic"
- Membership functions: are triangular, trapezoidal, gaussian, sigmoid, constant correct?
- T-norm/T-conorm: do they satisfy the 4 axioms (commutativity, associativity, monotonicity, identity)?
  - T-norms: Min, Product, Łukasiewicz
  - T-conorms: Max, ProbabilisticSum, BoundedSum
  - Complement: 1-d
- Defuzzification: Centroid is ∫μ(x)·x dx / ∫μ(x) dx? Bisector, MOM, LOM, SOM correct definitions?
- Fuzzify: does it correctly compute the degree of membership for a crisp input across all terms of a linguistic variable?
- Clip (α-cut): does Clip(μ, α)(x) = min(μ(x), α)?
- Scale: does Scale(μ, α)(x) = α·μ(x)? Is it a valid alternative to Clip for rule activation?
- AggregateMax: does it compute the pointwise maximum of multiple fuzzy sets?
- Sampling: does Sample produce correctly ordered, evenly spaced points?

### math/stats/ — Statistics
- Reference: Casella & Berger "Statistical Inference", NIST Digital Library of Mathematical Functions
- 17 distributions — are PDF, CDF, Mean, Variance correct for each one?
  - **Continuous**: Normal, Exponential, Uniform, Beta, Gamma, ChiSquared, StudentT, Lognormal, Weibull, FDist, Gumbel, Pareto
  - **Discrete**: Poisson, Binomial, Geometric, Hypergeometric, NegativeBinomial
- Hypothesis testing: does TTest use the correct t-statistic? Correct degrees of freedom? Does ChiSquared test use the correct statistic?
- RunningStats (Welford): is the online algorithm numerically stable and correct?
- WindowedStats: does it produce results identical to computing over the full window?
- Bayes theorem: is P(A|B) = P(B|A)·P(A)/P(B) implemented correctly?
- Descriptive stats functions: are mean, variance, standard deviation, median, percentile correct?

### math/graph/ — Graph primitives
- Reference: Cormen et al. "Introduction to Algorithms" (CLRS), Diestel "Graph Theory", Bondy & Murty "Graph Theory"
- **Core structures**: Directed, Undirected, DAG, Bipartite graphs
  - Are adjacency operations correct (AddVertex, AddEdge, Neighbors, Degree)?
  - DAG: does it correctly reject cycles? Is topological sort correct?
  - Bipartite: does it correctly verify 2-colorability?
- **Traversal**: BFS, DFS
  - BFS: does it visit all reachable vertices in level order?
  - DFS: does it visit all reachable vertices with correct pre/post ordering?
- **Paths**: shortest path algorithms
  - Are Dijkstra/BFS shortest paths correct? Do they handle disconnected graphs?
- **Centrality**: degree, betweenness, closeness, PageRank
  - Degree centrality: is it degree(v) / (n-1)?
  - Betweenness: is it Σ(σ_st(v)/σ_st) for all s≠v≠t?
  - Closeness: is it (n-1) / Σd(v,u)?
  - PageRank: is it the iterative computation of PR(v) = (1-d)/n + d·Σ(PR(u)/L(u))? Does it converge?
- **Coloring**: graph coloring
  - Does greedy coloring produce a valid coloring (no adjacent vertices share a color)?
- **MST**: minimum spanning tree
  - Does it produce a spanning tree with minimum total weight? (Kruskal/Prim)
- **Matching**: graph matching
  - Does maximum matching produce the correct cardinality?
- **Matrix**: adjacency/incidence matrix representation
  - Is the adjacency matrix correct (A[i][j]=1 iff edge (i,j) exists)?
- **Multigraph**: multiple edges between same vertices
  - Are parallel edges handled correctly?
- **Tree**: tree operations
  - Are tree properties verified correctly (connected, acyclic, n-1 edges)?

### math/fsm/ — Finite state machines
- Reference: Hopcroft, Motwani & Ullman "Introduction to Automata Theory, Languages, and Computation"
- **Transitions**: is the transition function δ(state, event) → state correct? Are guards evaluated correctly?
- **Determinism**: does it correctly identify deterministic vs non-deterministic machines?
- **Reachability**: does it correctly compute the set of reachable states from an initial state?
- **Dead states**: does it correctly identify states with no outgoing transitions?
- **Completeness**: does it correctly determine if every (state, event) pair has a transition?

### math/markov/ — Markov chains
- Reference: Norris "Markov Chains", Kemeny & Snell "Finite Markov Chains"
- **Transition matrix**: are rows valid probability distributions (non-negative, sum to 1)?
- **Steady state (stationary distribution)**: does πP = π and Σπ_i = 1?
  - Is the iterative/algebraic method correct?
  - Does it handle irreducible + aperiodic chains (guaranteed unique stationary distribution)?
- **Classification**: communicating classes, recurrent vs transient states
  - Are communicating classes correct (i↔j iff i→j and j→i)?
  - Is recurrence/transience classification correct?
- **Absorption**: for absorbing chains
  - Is the fundamental matrix N = (I-Q)⁻¹ correct?
  - Are absorption probabilities B = N·R correct?
  - Are expected absorption times t = N·1 correct?
- **Step-n**: is the n-step transition matrix P^n computed correctly?
- **Mean first passage**: is the expected number of steps to reach state j from state i correct?
- **Simulate**: does stochastic simulation produce state sequences consistent with the transition matrix?

### engine/deductive/ — Forward + backward chaining
- Reference: Russell & Norvig "AIMA" Ch. 7-9, Forgy (1982) for forward chaining
- Forward chaining: does it reach the fixed point (all derivable facts)?
- Backward chaining: is it sound and complete with respect to the rule base?
- Does depth limiting preserve soundness (may lose completeness)?
- Clone-on-attempt: does it preserve semantics (no side effects between attempts)?

### engine/bayesian/ — Bayesian networks
- Reference: Koller & Friedman "PGM" Ch. 9, Pearl (1988)
- Variable elimination: does variable elimination preserve the correct marginal distribution?
- Does elimination order affect the result? (it should not, only efficiency)
- Enumeration: does it correctly sum over hidden variables?
- Are CPTs validated (rows sum to 1)?
- Is evidence handled correctly (clamping)?

### engine/fuzzy/ — Mamdani + Sugeno
- Reference: Mamdani & Assilian (1975), Sugeno (1985), Lee (1990)
- Mamdani: is fuzzify → evaluate rules → aggregate → defuzzify correct?
- Sugeno: is the output the weighted average of constants by activation?
- Is the difference between Mamdani (fuzzy output) and Sugeno (crisp output) correct?

### engine/causal/ — Causal inference
- Reference: Pearl "Causality" (2009), "The Book of Why" (2018)
- Propagation (Level 1): is observational conditioning P(Y|X=x) correct?
- do-operator (Level 2): does it correctly implement P(Y|do(X=x)) — graph mutilation?
- Counterfactual: does it follow Pearl's 3 steps (abduction, action, prediction)?
- Is graph mutilation (graph surgery) correct — removes incoming edges to the intervened node?

### engine/mcdm/ — AHP + TOPSIS
- Reference: Saaty (1980) for AHP, Hwang & Yoon (1981) for TOPSIS
- AHP: does the principal eigenvector of the comparison matrix give the correct weights?
- AHP: does the consistency ratio use the correct RI and CI = (λmax - n)/(n - 1)?
- TOPSIS: is vector normalization x_ij / sqrt(Σx²)?
- TOPSIS: is the distance to positive/negative ideal solution Euclidean?
- TOPSIS: is the ranking by closeness coefficient = D⁻/(D⁺+D⁻) correct?

## REQUIRED OUTPUT

For EACH algorithm above, respond EXACTLY with this structure:

    ### [Algorithm name]
    - **Reference**: [verified text/author]
    - **Soundness**: does it produce only correct results? [YES/NO/NOT VERIFIABLE WITHOUT CODE + reason]
    - **Completeness**: does it find all results it should? [YES/NO/N/A + reason]
    - **Termination**: does it always terminate? [YES/NO/CONDITIONAL + reason]
    - **Edge cases**: are there degenerate inputs that produce incorrect results? [list or "none identified"]
    - **Formal gaps**: is anything missing within the declared scope? [list or "none"]

## FINAL VERDICT

After evaluating all algorithms:
1. Are the implemented algorithms mathematically correct? [YES/NO/PARTIAL]
2. For each NO or PARTIAL: what specifically is wrong and what is the correction.
3. Are there mathematical properties that should hold and are not guaranteed?
