# maths

Mathematical primitives for reasoning systems. Pure functions, no side effects,
no external dependencies.

This module provides the building blocks that `modules/inference` uses to build
applied reasoning engines. The separation is deliberate: **maths = primitives**,
**inference = engines**.

## Packages

### logic/

Propositional logic: formulas as data structures that you can build, parse, transform,
simplify, and analyze.

| Area | Functions | What it does |
|------|-----------|--------------|
| **Nodes** | `Var`, `TrueF`, `FalseF`, `NotF`, `AndF`, `OrF`, `ImplF`, `IffF` | Formula building blocks. Implement the `Formula` interface (String, Eval, Vars). |
| **Parser** | `Parse`, `ParseWith` | Text to formula: `"A & (B | C)"` becomes an AST. Supports Unicode operators, keywords, and programmer syntax. |
| **Eval** | `Formula.Eval(facts)` | Evaluate a formula against a truth assignment. |
| **Transform** | `ToNNF`, `ToCNF`, `ToDNF` | Normal form conversions (Negation, Conjunctive, Disjunctive). |
| **Simplify** | `Simplify` | 18 algebraic simplification rules applied until fixpoint. |
| **Analysis** | `TruthTable`, `Equivalent`, `FailCases` | Truth tables, equivalence checking, counterexample generation. |
| **Satisfiability** | `IsSatisfiable`, `IsContradiction`, `IsTautology` | Brute-force or pluggable SAT solver via `RegisterSATSolver`. |
| **SAT** | `sat.Solve`, `sat.FromFormula` | DPLL algorithm on CNF formulas. ~10x faster than brute-force. |
| **Entailment** | `Entails`, `EntailsWithCounterModel` | Does a conclusion follow from premises? Returns countermodel on failure. |
| **Format** | `Format` | Unicode rendering: `(A ∧ B) → ¬C`. |

### probability/

Discrete probability primitives. Currently all code lives in `probability/` directly.
When continuous probability is added, the package will be restructured into:

```
probability/           <-- shared types (Var, Outcome, Prob, Assignment)
probability/discrete/  <-- current content (Distribution, CPT, Factor, Bayes)
probability/continuous/ <-- future (Normal, Exponential, Beta, Gamma, PDF/CDF)
```

| Area | Functions | What it does |
|------|-----------|--------------|
| **Types** | `Var`, `Outcome`, `Prob`, `Distribution`, `Assignment` | Random variables, outcomes, and probability maps. |
| **Distribution** | `IsValid`, `Normalize`, `Complement`, `Entropy` | Validate, rescale, complement, and measure information content. |
| **Bayes** | `Bayes`, `TotalProbability`, `ChainRule`, `Independent` | Bayes' theorem P(H\|E), total probability, chain rule, independence. |
| **CPT** | `NewCPT`, `Set`, `Lookup`, `Validate` | Conditional probability tables: P(child \| parents). |
| **Factor** | `NewFactor`, `Multiply`, `SumOut`, `Restrict`, `NormalizeFactor` | Factor algebra for variable elimination algorithms. |

**Discrete vs continuous**: this package covers **discrete** probability — variables
with countable outcomes (yes/no, red/green/blue, 1..6). Continuous probability
(Normal, Exponential, Gamma) operates on density functions and integrals over
ranges, and is not yet implemented.

### fuzzy/

Fuzzy logic: membership functions, fuzzy operators, and defuzzification methods.

| Area | Functions | What it does |
|------|-----------|--------------|
| **Types** | `Degree`, `MembershipFn`, `TNormFn`, `TConormFn`, `DefuzzifyFn`, `Set`, `Point` | Fuzzy truth values in [0,1], function types, named fuzzy sets. |
| **Membership** | `Triangular`, `Trapezoidal`, `Gaussian`, `Sigmoid`, `Constant` | Standard membership function shapes. Map crisp values to degrees. |
| **T-norms** | `Min`, `Product`, `Lukasiewicz` | Fuzzy AND operators (intersection). |
| **T-conorms** | `Max`, `ProbabilisticSum`, `BoundedSum` | Fuzzy OR operators (union). |
| **Operations** | `Fuzzify`, `Clip`, `Scale`, `AggregateMax`, `Sample` | Evaluate, clip/scale by alpha-cut, aggregate, and discretize. |
| **Complement** | `Complement` | Fuzzy NOT: 1 - degree. |
| **Defuzzify** | `Centroid`, `Bisector`, `MeanOfMax`, `LargestOfMax`, `SmallestOfMax` | Convert fuzzy output back to a crisp value. |

## Use cases

### logic/

- **Rule engines**: express business rules as formulas, evaluate them against facts
- **Validation**: check if a set of constraints is satisfiable before deploying a configuration
- **Compliance**: given premises (regulations), does a conclusion (action) follow logically?
- **Circuit design**: model digital circuits as boolean formulas, verify properties
- **Puzzles/games**: solve logic puzzles (Sudoku constraints as SAT problems)

### probability/

This is for **reasoning under uncertainty**, not for surveys or descriptive statistics.

- **Medical diagnosis**: P(disease | symptoms) using Bayesian networks
- **Fraud detection**: P(fraud | transaction patterns)
- **Spam filtering**: P(spam | words in email)
- **Sensor fusion**: combine uncertain readings from multiple sensors
- **Risk assessment**: P(default | client profile, market conditions)
- **Recommendation**: P(user likes X | purchase history)
- **Decision support**: given incomplete observations, what is the most probable state?

The CPT and Factor types are specifically designed to feed into Bayesian network
inference algorithms (see `modules/inference/bayesian/`).

### fuzzy/

This is for domains where categories are **gradual**, not binary.

- **Control systems**: "temperature is *somewhat high*" -> adjust valve *a little*
- **Credit scoring**: "income is *medium*" + "history is *good*" -> risk is *low-medium*
- **Quality assessment**: "response time is *fast*" + "accuracy is *high*" -> priority *low*
- **Dynamic pricing**: "demand is *high*" + "inventory is *low*" -> price *high*
- **HVAC/IoT**: "humidity is *medium*" + "temperature is *hot*" -> fan speed *high*
- **Any scoring system** where hard thresholds (>80 = good) are too rigid

The membership functions and operators feed into fuzzy inference engines
(see `modules/inference/fuzzy/` for Mamdani and Sugeno methods).

## How it connects to inference/

```
maths/logic       -->  inference/classical   (forward/backward chaining)
maths/probability -->  inference/bayesian    (enumeration, variable elimination)
maths/fuzzy       -->  inference/fuzzy       (Mamdani, Sugeno)
```

maths/ provides the math. inference/ provides the reasoning. Application code uses
inference/ and never needs to call maths/ directly (though it can).

## Extending this module

Potential new packages, roughly ordered by utility:

| Package | What it would provide | Enables |
|---------|-----------------------|---------|
| `stats/` | Descriptive statistics (mean, variance, percentiles), hypothesis testing, confidence intervals | Data analysis, A/B testing, survey analysis |
| `probability/continuous/` | Continuous distributions (Normal, Exponential, Beta, Gamma), PDF/CDF. Triggers refactor: current code moves to `probability/discrete/`, shared types stay in `probability/` | Gaussian Mixture Models, Kalman filters, continuous Bayesian inference |
| `linalg/` | Vectors, matrices, decompositions (LU, QR, SVD, Cholesky), eigenvalues | ML foundations, PCA, regression, neural networks |
| `autodiff/` | Automatic differentiation (forward/reverse mode), symbolic derivatives, chain rule as computation | Exact gradients for backpropagation, Jacobians, Hessians without numerical approximation |
| `numerical/` | Root finding (bisection, Newton-Raphson, secante), interpolation (Lagrange, splines), numerical integration (trapezoid, Simpson), iterative solvers (Jacobi, Gauss-Seidel) | Solving f(x)=0, curve fitting, approximating integrals, large sparse systems |
| `optim/` | Gradient descent, Newton's method, line search, constrained optimization. Depends on `autodiff/` for exact gradients, `linalg/` for matrix operations | Training ML models, parameter fitting |
| `graph/` | Graph structures, BFS/DFS, shortest path, topological sort | Network analysis, dependency resolution, planning |
| `logic/temporal/` | LTL/CTL operators (Always, Eventually, Until, Next), model checking | Verification of systems, protocol validation, workflow correctness |
| `logic/predicate/` | First-order logic: variables, constants, predicates, quantifiers (forall/exists), unification, resolution | General rules ("all employees with X get Y"), Prolog-like reasoning |
| `sets/` | Generic `Set[T]` with union, intersection, difference, complement, power set, product | Replaces ad-hoc `map[T]struct{}` patterns used across packages |
| `combinatorics/` | Permutations, combinations, binomial coefficient, inclusion-exclusion, factorial | Probability calculations, enumeration, complexity analysis |
| `markov/` | Markov chains: transition matrices, stationary distribution, n-step probabilities, absorbing chains | Prediction, PageRank, weather models, queueing theory |
| `automata/` | Deterministic/non-deterministic finite state machines (DFA/NFA), state transitions, reachability | Model checking for temporal/, workflow verification, protocol validation |
| `fuzzy/` (type-2) | Interval membership functions (uncertainty over uncertainty), type-reduction | Environments with noisy or imprecise membership definitions |

### Viability analysis

The goal is **not** to compete with gonum, NumPy, or PyTorch. These packages provide
primitives for yarumo's inference engines at small-to-medium scale. For large-scale
ML or heavy numerical computation, use specialized tools.

#### stats/ — Viability: High

Simple arithmetic: mean is a for loop, median is sort + pick middle, variance is
another for loop. Even hypothesis testing uses closed-form formulas (not iterative).
The most complex piece would be the inverse CDF of the t-Student distribution for
confidence intervals, but well-documented numerical approximations exist.

gonum/stat already exists and is mature. Reimplementing basic descriptive statistics
is trivial and keeps yarumo dependency-free. Advanced statistical tests (ANOVA,
non-parametric tests) would be reinventing the wheel — only implement what the
inference engines need.

#### probability/continuous/ — Viability: High

PDF and CDF of a Normal distribution are closed-form formulas (`math.Exp`, `math.Erfc`).
Go stdlib already has `math.Erfc` which is the key piece. Exponential and Uniform are
trivial. Beta and Gamma need the incomplete Gamma function — more complex but Go has
`math.Gamma` and `math.Lgamma`.

Sampling (generating random numbers following a distribution) uses `math/rand/v2` plus
known methods (Box-Muller for Normal, inverse CDF for Exponential).

No heavy iteration. Just evaluating mathematical functions.

#### linalg/ — Viability: Medium. Scope must be bounded.

Two worlds:

**The simple part** (vectors, matrix multiplication, transpose, inverse of small matrices,
determinant): feasible in pure Go, not computationally heavy. A 100x100 matrix
multiplication runs in milliseconds.

**The heavy part** (SVD, eigenvalues, factorizations of large matrices): serious
implementations (LAPACK, BLAS) are written in Fortran/C optimized at the CPU instruction
level (SIMD, cache blocking). Reimplementing SVD in pure Go will be 10-100x slower than
LAPACK for large matrices.

**The reality**: yarumo's use case is not "train a neural network with millions of
parameters" — that's Python/PyTorch with GPUs. The use case is "solve a system of 50
equations for an inference model" or "PCA over 20 variables". For that, pure Go is
more than enough.

**Scope**: implement basics (vectors, matrices, LU, QR, eigenvalues for small symmetric
matrices). SVD can be a simplified version. If someone needs 10,000x10,000 matrices,
they should use gonum or Python.

#### autodiff/ — Viability: Medium. Only autodiff, not a CAS.

This is **not** about symbolic differentiation ("give me the derivative of x^2+3x as
2x+3") or symbolic integration. That would be a Computer Algebra System (CAS) like
Mathematica/SymPy — a massive project with no place in yarumo.

What it actually provides is **automatic differentiation (autodiff)**:

- **Forward mode**: propagate derivatives alongside values. Each operation (`+`, `*`,
  `sin`) knows its derivative and propagates it. Not iterative, not heavy — O(n) where
  n is the number of operations. Very feasible in Go.
- **Reverse mode**: this is backpropagation. Build a computation graph, then traverse it
  backwards. More complex to implement, but feasible for small graphs.

**Scope**: start with forward mode (for gradients and Jacobians). Reverse mode is optional
and only needed if yarumo ever supports training neural-network-like models.

#### numerical/ — Viability: High. Not as heavy as it seems.

**Root finding** (Newton-Raphson, bisection): Newton-Raphson converges in 5-10 iterations
typically (quadratic convergence). Each iteration is ONE evaluation of f(x) and ONE of
f'(x). Not heavy at all. Bisection is even simpler — just evaluates f(x) once per
iteration.

**Interpolation** (Lagrange, splines): NOT iterative. Direct formula. Lagrange is O(n^2),
splines require solving a tridiagonal system in O(n). Trivial.

**Numerical integration** (trapezoid, Simpson): evaluates the function n times, but n=100
already gives good precision for smooth functions. No differentiation needed — just
evaluate f(x) at evenly-spaced points and sum.

**What IS heavy**: integration in many dimensions (triple, quadruple integrals). Cost
grows exponentially. That is Monte Carlo territory and out of scope.

**Iterative solvers** (Jacobi, Gauss-Seidel): converge fast for well-conditioned matrices.
These overlap with linalg/ and would live better there.

#### optim/ — Viability: Medium. Scope matters.

**Basic gradient descent**: trivial. `x = x - lr * gradient` in a loop. 10 lines of code.

**SGD, momentum, Adam**: variations of the same loop. Feasible, not heavy.

**Newton's method (multivariable)**: needs to compute and invert the Hessian (matrix of
second derivatives). For 10 variables that's a 10x10 matrix — trivial. For 10,000
variables it's a 10,000x10,000 matrix — infeasible. That's why real ML uses L-BFGS
(approximates the Hessian without computing it).

**The real bottleneck**: optim/ itself is not heavy. The heavy part is the function being
optimized. Optimizing f(x) = x^2 + 3x runs in microseconds. Optimizing a neural network
with millions of parameters — the bottleneck is the network, not the optimizer.

**Scope**: viable for small models (the kind yarumo's inference engines use). For deep
learning, nobody would use this — they'd use PyTorch with GPUs.

#### graph/ — Viability: High

Graphs are the most natural data structure for Go. BFS/DFS are simple recursive functions.
Dijkstra is a heap + a loop. Topological sort is DFS with a stack.

Nothing computationally questionable. Algorithms are O(V+E) or O(V log V + E) worst case.

Bayesian networks are already directed acyclic graphs. Today inference/bayesian/ defines
its own network structure internally. With graph/, that structure generalizes.

#### logic/temporal/ — Viability: Medium

LTL/CTL over small models (tens of states) is feasible. Model checking is essentially
exploring a state graph — BFS/DFS.

**The problem**: state space explosion. A system with 20 boolean variables has 2^20 = ~1M
states. With 30 variables, a billion. This is the same problem as SAT (already solved with
DPLL). Real model checkers (SPIN, NuSMV) use BDDs (Binary Decision Diagrams) and symbolic
abstraction to handle this.

**Scope**: useful for small models (simple protocols, workflows). For real large-system
verification, BDDs would be needed — a complex project in itself.

#### fuzzy/ type-2 — Viability: High

Type-2 is type-1 with an interval instead of a point. Where you had `Degree = 0.7`, now
you have `Interval = [0.6, 0.8]`. Operations (t-norms, defuzzify) apply to the interval
endpoints.

The extra step is **type-reduction** (Karnik-Mendel): an iterative algorithm that converges
in ~5-10 iterations operating on small vectors. Not heavy.

Natural extension of existing fuzzy/ package.

#### logic/predicate/ — Viability: Medium-Low. Large project, high value.

First-order logic extends propositional logic from "A AND B" to "for all X: employee(X)
AND active(X) implies has_salary(X)". One rule covers all objects instead of one rule
per object.

Implementation requires:

- **Terms**: variables (`X`), constants (`juan`), functions (`parent(X)`)
- **Unification**: can `employee(X)` match `employee(juan)`? Yes, X=juan. This is the
  core algorithm — pattern matching with variable substitution.
- **Resolution**: inference algorithm (like forward/backward chaining but with variable
  substitution at each step)
- **Quantifiers**: forall and exists, with scoping rules

This is the heart of Prolog. The algorithms (unification, resolution) are well-understood
and not computationally heavy per step. The complexity comes from the search space —
first-order logic is **undecidable** in the general case (unlike propositional which is
always decidable via SAT). Practical implementations use depth limits and loop detection.

**Scope**: implement a bounded fragment (no infinite domains, function-free clauses).
This covers the practical cases (business rules, database queries, type checking)
without hitting undecidability.

Would split into `maths/logic/predicate/` (terms, unification, substitution) and
`inference/predicate/` (resolution engine, query answering).

#### sets/ — Viability: High. Small and useful.

Generic `Set[T]` type with standard operations: union, intersection, difference,
complement, subset check, power set, cartesian product.

Today every package that needs set operations uses `map[T]struct{}` and writes the
logic ad-hoc. `mergeVars` in probability/factor.go is literally set union. `Vars()` in
logic/ deduplicates with a map — that's set insertion.

Implementation is trivial — thin wrapper over `map[T]struct{}` with methods. Maybe
30-40 functions total. The question is whether it lives in `maths/sets/` (mathematical
primitive) or `common/sets/` (general utility). Since it has no math-specific semantics,
`common/sets/` might be the better home.

#### combinatorics/ — Viability: High. Very small.

Factorial, permutations, combinations, binomial coefficient — each is 5-15 lines.
Inclusion-exclusion is a loop applying the formula. No iteration, no convergence,
just arithmetic.

The real question is whether this justifies a standalone package. The total API might
be ~10 functions. Could live as helpers within packages that need them. But having them
in one place avoids reimplementation and makes them discoverable.

**Risk**: almost none. The only subtlety is overflow — factorial(21) exceeds int64.
Use `math/big` or document the limit.

#### markov/ — Viability: High. Connects probability/ with temporal processes.

A Markov chain is an automaton where transitions have probabilities. The core data
structure is a transition matrix (each row is a probability distribution — exactly
what `probability.Distribution` already is).

Key operations:

- **n-step probability**: multiply the transition matrix n times. For small matrices
  this is fast. For large ones, use eigenvalue decomposition (connects to linalg/).
- **Stationary distribution**: the steady state the chain converges to. Solve a linear
  system (connects to linalg/) or iterate until convergence.
- **Absorbing chains**: which states are final? What's the expected time to reach them?
- **Classification**: transient vs recurrent states, periodicity.

All operations are either matrix arithmetic or simple graph traversal. No heavy
computation for the scale yarumo targets.

**Dependency**: benefits from linalg/ (matrix multiplication, eigenvalues) but can
start with a simple implementation using slices of slices.

#### automata/ — Viability: High. Prerequisite for temporal/.

DFA (deterministic) and NFA (non-deterministic) finite state machines. A state, a set
of transitions (state + event -> next state), an initial state, accepting states.

Key operations:

- **Reachability**: can state B be reached from state A? (BFS/DFS — trivial)
- **NFA to DFA conversion**: subset construction algorithm. Well-known, O(2^n) worst
  case but practical cases are much smaller.
- **Minimization**: merge equivalent states. Uses partition refinement.
- **Product construction**: combine two automata (for checking "does system X satisfy
  property Y expressed as an automaton?").

This is the foundation for `logic/temporal/` — temporal formulas get converted to
automata, and model checking becomes reachability analysis on the product automaton.

Without `automata/`, `logic/temporal/` would have to implement its own state machine
representation internally. With it, temporal/ can focus on the logic and delegate
the state machine mechanics.

**Scope**: DFA and NFA with basic operations. Pushdown automata and Turing machines
are out of scope (academic interest only for yarumo's purposes).

#### Summary

| Package | Viability | Worth it? | Key constraint |
|---------|-----------|-----------|----------------|
| `stats/` | High | Yes | Basic descriptive + what inference engines need |
| `probability/continuous/` | High | Yes | Closed-form formulas, stdlib helps |
| `linalg/` | Medium | Yes, bounded | Small-medium matrices only. Do not compete with LAPACK |
| `autodiff/` | Medium | Yes, autodiff only | Forward mode first. This is NOT a CAS |
| `numerical/` | High | Yes | Not as heavy as it seems. Newton converges in 5-10 iterations |
| `optim/` | Medium | Yes, bounded | For inference models, not for deep learning |
| `graph/` | High | Yes | Cleanest package to implement |
| `logic/temporal/` | Medium | Depends | Useful for small models, state explosion on large ones |
| `logic/predicate/` | Medium-Low | High value, large project | Bounded fragment only. Undecidable in general |
| `sets/` | High | Yes | Trivial implementation. Maybe better in common/sets/ |
| `combinatorics/` | High | Yes, but small | ~10 functions. Watch for int64 overflow on factorial |
| `markov/` | High | Yes | Connects probability/ with temporal processes. Benefits from linalg/ |
| `automata/` | High | Yes, prerequisite for temporal/ | DFA/NFA only. No pushdown or Turing machines |
| `fuzzy/` type-2 | High | Yes | Natural extension of existing code |

### Open debate: Operations Research

Operations Research (OR) is **applied optimization for business decisions**: "what is
the best possible decision given these constraints?" Its main areas are:

| Area | What it solves | Example |
|------|----------------|---------|
| Linear programming | Optimize a linear function subject to linear constraints | Maximize profit producing A and B with limited resources |
| Integer programming | Same but variables must be integers | How many trucks of each type to buy? |
| Queueing theory | Model waiting systems | How many checkout lines to keep wait < 5 min? |
| Dynamic programming | Optimal sequential decisions | Cheapest route visiting these cities in order |
| Game theory | Optimal decisions when another rational player is also deciding | What price to set when a competitor is also pricing? |
| Monte Carlo simulation | Model complex systems via random sampling | What happens to delivery times if demand rises 20%? |
| Transport/assignment | Move things from origins to destinations at minimum cost | Which warehouses supply which stores? |
| Scheduling | Assign tasks to resources over time | In what order to process these jobs on these machines? |

**The key insight**: OR is not a mathematical primitive — it is an **application** of
several primitives. It is closer to inference/ than to maths/.

How OR maps to planned packages:

| OR area | Covered by | Status |
|---------|------------|--------|
| Linear/integer programming | `optim/` (simplex method, branch and bound) + `linalg/` | Planned |
| Queueing theory | `probability/continuous/` (Exponential, Poisson) + `markov/` | Planned |
| Dynamic programming | Algorithm pattern, not a library. Implemented case by case | N/A |
| Game theory | `linalg/` (payoff matrices) + `optim/` (Nash equilibria) | Planned |
| Monte Carlo simulation | `probability/` + random sampling | Partially covered |
| Transport | Special case of linear programming | Covered by optim/ |
| Scheduling | `graph/` (dependencies) + constrained optimization | Planned |
| Markov decision processes | `markov/` + `optim/` | Planned |

**Conclusion**: OR does not need its own package. When linalg/ + optim/ + graph/ +
probability/continuous/ + markov/ exist, the building blocks for OR problems are already
in place. What would sit on top is an `inference/or/` or application-level layer that
assembles these primitives for specific OR problem types (e.g., a simplex solver, a
transportation problem solver). That decision can wait until there is a concrete use case.

### Guidelines for new packages

- **Pure primitives**: no side effects, no I/O, no external dependencies
- **Value types**: public structs with public fields, construct directly
- **Simple errors**: `errors.New` sentinels, no `TypedError` pattern
- **100% test coverage**: individual test functions, `t.Parallel()`, no testify
- See [CODING_STANDARDS.md](CODING_STANDARDS.md) for full conventions
