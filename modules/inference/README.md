# inference

Reasoning engines built on `modules/maths/`. Each sub-module implements a different
inference paradigm: given inputs (facts, evidence, observations), produce conclusions
with full traceability.

The separation is deliberate: **maths = primitives**, **inference = engines**.
Application code uses inference/ and never needs to call maths/ directly (though it can).

## Packages

### classical/

Propositional logic inference: forward and backward chaining over boolean rules
with provenance tracking.

| Package | What it provides |
|---------|-----------------|
| **explain/** | Trace types: `Origin` (Asserted/Derived), `Provenance`, `Step`, `Trace`. Value types — pure data, no interface. |
| **rules/** | `Rule` interface: `Name`, `Priority`, `Condition` (Formula), `Conclusion` (variable assignments), `Fires`, `Produces`. Options: `WithPriority`. |
| **facts/** | `FactBase` interface: `Assert`, `Derive`, `Retract`, `Get`, `Snapshot`, `Provenance`, `AllProvenance`, `Clone`. Tracks origin of every fact. |
| **engine/** | `Engine` interface: `Forward` (data-driven chaining) and `Backward` (goal-driven chaining). Options: `WithMaxIterations`, `WithStrategy` (AllMatches/FirstMatch). |
| **examples/** | Integration tests and benchmarks. |

**How it works**: define rules as `IF condition THEN conclusion`. Forward chaining starts
from known facts and fires every applicable rule until no new facts are derived. Backward
chaining starts from a goal and works backwards through rules to find a proof.

Every step is recorded in a `Trace` with full provenance: which rule fired, what facts
existed before, what was produced.

### bayesian/

Probabilistic inference over Bayesian networks: compute P(query | evidence) with
full trace of the computation.

| Package | What it provides |
|---------|-----------------|
| **explain/** | Trace types: `Phase` (Initialize/Propagate/Marginalize/Complete), `Factor`, `Step`, `Posterior`, `Trace`. Value types. |
| **network/** | `Network` interface: `AddNode`, `Node`, `Nodes`, `Parents`, `Children`, `TopologicalOrder`, `Validate`. DAG of nodes with CPTs. |
| **evidence/** | `EvidenceBase` interface: `Observe`, `Retract`, `Get`, `Observed`, `Clone`. Manages observed variable outcomes. |
| **engine/** | `Engine` interface: `Query(network, evidence, variable) Result`. Algorithms: `Enumeration` (exact, brute-force) and `VariableElimination` (exact, efficient). Options: `WithAlgorithm`, `WithEliminationOrder`. |

**How it works**: define a Bayesian network (DAG of random variables with conditional
probability tables). Set evidence (observed outcomes). Query for the posterior distribution
of any unobserved variable. The engine marginalizes over all hidden variables.

Variable elimination is typically faster than enumeration — it avoids redundant computation
by factoring the joint distribution and eliminating variables one at a time.

### fuzzy/

Fuzzy inference: map crisp inputs through linguistic rules to crisp outputs, with
Mamdani and Sugeno methods.

| Package | What it provides |
|---------|-----------------|
| **explain/** | Trace types: `Phase` (Fuzzification/RuleEvaluation/Aggregation/Defuzzification), `Membership`, `Activation`, `Step`, `Output`, `Trace`. Value types. |
| **variable/** | `Variable` interface: `Name`, `Min`, `Max`, `Terms`, `Term`, `Fuzzify`, `Resolution`. Linguistic variables with named fuzzy terms. Options: `WithResolution`. |
| **rules/** | `Rule` interface: `Name`, `Conditions`, `Operator` (And/Or), `Consequent`, `Weight`. Fuzzy IF-THEN rules. Options: `WithOperator`, `WithWeight`. |
| **engine/** | `Engine` interface: `Infer(inputs) Result`. Methods: `Mamdani` (clip+aggregate+defuzzify) and `Sugeno` (weighted average of singletons). Options: `WithMethod`, `WithTNorm`, `WithTConorm`, `WithDefuzzifyFn`, `WithSugenoOutputs`. |

**How it works**: define input/output variables with fuzzy terms (e.g., temperature:
cold/warm/hot). Define rules (IF temp is hot AND humidity is high THEN fan is fast).
Feed crisp inputs. The engine fuzzifies inputs, evaluates rules, aggregates outputs,
and defuzzifies to a crisp result.

Mamdani produces fuzzy output sets and defuzzifies (centroid, bisector, etc.). Sugeno
uses singleton output values and computes a weighted average — faster but less expressive.

## Use cases

### classical/

- **Business rule engines**: "IF client is premium AND order > $1000 THEN apply discount"
- **Compliance checking**: "IF employee has access AND no training THEN flag violation"
- **Diagnostic systems**: "IF symptom A AND symptom B THEN diagnosis C" with backward chaining to find which tests to run
- **Configuration validation**: forward-chain constraints to detect conflicts
- **Workflow automation**: derive next actions from current state and rules

### bayesian/

- **Medical diagnosis**: P(disease | symptoms, test results) using a network of conditions
- **Fraud detection**: P(fraud | transaction pattern, location, amount)
- **Risk assessment**: P(default | income, history, market) with multiple evidence sources
- **Sensor fusion**: combine uncertain readings from multiple sensors into a coherent estimate
- **Root cause analysis**: given observed failures, what is the most probable root cause?
- **Recommendation**: P(user likes X | purchase history, demographics)

### fuzzy/

- **Control systems**: temperature -> fan speed, where "somewhat hot" means "medium-high fan"
- **Credit scoring**: combine income, history, employment into a gradual risk score
- **Quality assessment**: response time + accuracy + uptime -> service quality grade
- **Dynamic pricing**: demand + inventory + competition -> price adjustment
- **HVAC/IoT**: humidity + temperature + occupancy -> climate settings
- **Any scoring** where hard thresholds (>80 = good) are too rigid and you need gradual transitions

## Common architecture

All three engines share the same structural pattern:

```
<paradigm>/
├── explain/    ← trace types (what happened during inference)
├── <domain>/   ← domain model (rules, network, variables)
├── engine/     ← inference algorithms + options
└── examples/   ← integration tests, benchmarks
```

Every engine produces a `Result` containing:
1. The computed output (derived facts, posterior distribution, crisp values)
2. A `Trace` documenting every step of the computation

This traceability is deliberate: inference engines must be **explainable**. You should
always be able to ask "why did you conclude X?" and get a step-by-step answer.

## How it connects to maths/

```
maths/logic       -->  inference/classical   (Formula, Var, Fact, Eval)
maths/probability -->  inference/bayesian    (Distribution, CPT, Factor, Var, Assignment)
maths/fuzzy       -->  inference/fuzzy       (Degree, MembershipFn, TNormFn, DefuzzifyFn)
```

maths/ provides the mathematical building blocks. inference/ provides the reasoning
algorithms. The boundary is clear: maths/ never knows about engines, traces, or
inference state. inference/ imports maths/ types and builds stateful engines around them.

## Extending this module

Potential new engines, roughly ordered by dependency chain:

| Engine | Depends on | What it does |
|--------|------------|-------------|
| `predicate/` | `maths/logic/predicate/` | First-order logic inference: forward/backward chaining with unification. One rule covers all matching objects. "For all employees with X, apply Y." |
| `temporal/` | `maths/logic/temporal/` + `maths/automata/` | Model checking: does a system satisfy temporal properties? "This workflow always terminates." "Eventually the payment is processed." |
| `statistical/` | `maths/stats/` + `maths/linalg/` | Statistical inference: regression (linear, logistic), hypothesis testing, classification. Data-driven conclusions. |
| `markov/` | `maths/markov/` + `maths/probability/` | Markov Decision Processes: policy iteration, value iteration. Optimal decisions in stochastic environments. |
| `constraint/` | `maths/sets/`, `maths/graph/` | Constraint satisfaction: arc consistency, backtracking with propagation. Scheduling, configuration, resource allocation. |
| `abductive/` | `inference/classical/` | Abductive reasoning: given observations, find the best explanation. Inverse of deduction. "The grass is wet" -> "it probably rained." |
| `causal/` | `inference/bayesian/` + `maths/graph/` | Causal inference: Pearl's do-calculus, interventions, counterfactuals. Beyond correlation to causation. |
| `belief/` | `maths/probability/` | Dempster-Shafer theory: belief functions, plausibility. Alternative to Bayes when exact probabilities are unavailable. |
| `ensemble/` | All engines | Multi-paradigm meta-engine: combine results from multiple engines with configurable strategy and weights. |

### Viability analysis

The goal is to provide **reasoning engines for small-to-medium scale problems**, not to
compete with specialized tools (SPIN for model checking, R for statistics, TensorFlow
for ML). Each engine should be self-contained, explainable, and composable.

#### predicate/ — Viability: Medium-Low. Large project, high value.

This is the jump from propositional to first-order logic. The chaining engine is
structurally similar to classical/ but every step requires **unification** — matching
`employee(X)` with `employee(juan)` to bind X=juan.

Implementation requires:
- `maths/logic/predicate/` first (terms, unification, substitution)
- Modified forward/backward chaining with variable binding at each step
- Occurs check (prevent circular bindings like X=f(X))
- Depth limits and loop detection (first-order logic is **undecidable** in general)

The value is enormous: one predicate rule replaces hundreds of propositional rules.
Business rule engines become dramatically more expressive. But the implementation is
substantial — unification alone has subtle edge cases, and the search space management
(avoiding infinite loops) needs careful design.

**Scope**: bounded fragment only. Function-free clauses (Datalog-like) cover the practical
cases (business rules, database-style queries, type checking) without hitting
undecidability. Full first-order logic with function symbols is a research project.

#### temporal/ — Viability: Medium. Useful but scope-limited.

Model checking verifies that a system (represented as a state graph) satisfies temporal
properties (expressed in LTL or CTL). The engine converts temporal formulas to automata
and checks reachability on the product of system and property automata.

Requires `maths/logic/temporal/` (temporal operators) and `maths/automata/` (DFA/NFA,
product construction, reachability).

**The bottleneck**: state space explosion. A system with 20 boolean state variables has
2^20 = ~1M states. With 30 variables, a billion. Real model checkers (SPIN, NuSMV) use
Binary Decision Diagrams (BDDs) and symbolic abstraction. Without BDDs, yarumo is
limited to small models.

**Scope**: useful for verifying simple protocols, workflows, and state machines with
tens of states. Not for verifying operating systems or hardware designs.

#### statistical/ — Viability: Medium. The great debate.

Statistical inference goes from **data to model**: given observations, fit parameters,
test hypotheses, make predictions. This is the inverse direction of probability
(model to data).

- **Linear regression**: closed-form solution (normal equations), needs `maths/linalg/`
  for matrix operations. Feasible for tens of variables.
- **Logistic regression**: no closed-form — needs iterative optimization (gradient
  descent or Newton-Raphson). Needs `maths/optim/` or `maths/autodiff/`.
- **Hypothesis testing**: mostly closed-form formulas using distribution CDFs.
- **Classification**: decision trees, naive Bayes — each is a well-defined algorithm.

**The debate**: is regression "inference" or "fitting"? Philosophically, statistical
inference means "drawing conclusions from data under uncertainty" — it fits. Practically,
it overlaps with ML territory. The boundary should be: **statistical/ handles classical
statistical methods** (regression, ANOVA, confidence intervals). If it starts needing
GPUs, it's out of scope.

**Dependency chain**: needs `maths/stats/`, `maths/linalg/`, possibly `maths/optim/`.
This is the engine with the most upstream dependencies.

#### markov/ — Viability: Medium. Clean algorithms, scale-limited.

Markov Decision Processes (MDPs) extend Markov chains with actions and rewards.
The agent observes a state, takes an action, transitions probabilistically, and
receives a reward. The goal: find the optimal policy (which action in each state
maximizes expected cumulative reward).

- **Policy evaluation**: given a policy, compute value of each state. Solve a linear
  system (connects to `maths/linalg/`).
- **Value iteration**: iteratively update state values until convergence. Simple loop,
  converges fast for small state spaces.
- **Policy iteration**: alternate between evaluating and improving the policy.

**The bottleneck**: curse of dimensionality. With S states and A actions, the transition
table has S x A x S entries. For S=1000 that's manageable. For S=1M it's not. Real RL
systems use function approximation (neural networks) to handle large state spaces —
that's deep RL, out of scope.

**Scope**: tabular MDPs for small state spaces. Foundation concepts that connect
probability with decision-making. Not a replacement for OpenAI Gym.

#### constraint/ — Viability: High. Independent and practical.

Constraint Satisfaction Problems (CSP) are a distinct paradigm: define variables with
domains and constraints between them. The engine finds assignments that satisfy all
constraints (or proves none exist).

- **Arc consistency (AC-3)**: prune domains by removing values that can't participate
  in any valid assignment. Simple queue-based algorithm.
- **Backtracking**: try values, propagate constraints, backtrack on failure.
- **Heuristics**: variable ordering (most constrained first), value ordering
  (least constraining first).

**Why it's high viability**: no heavy math dependencies. The algorithms are
well-understood, not computationally heavy per step, and immediately useful.
Scheduling, configuration, timetabling, Sudoku — these are real problems that
CSP solvers handle elegantly.

**No upstream dependencies on other maths/ packages** (beyond maybe sets/ for domain
operations). Can be implemented immediately with what exists today.

#### abductive/ — Viability: Medium. Conceptually clean, search is the challenge.

Abductive reasoning inverts deduction: instead of "premises -> conclusion", it asks
"given the conclusion (observation), what premises (explanations) would produce it?"

In medical diagnosis: deduction says "flu -> fever". Abduction says "fever is observed,
flu is a possible explanation."

Implementation approaches:
- **Set-cover**: find the minimal set of hypotheses that explains all observations.
  NP-hard in general, but feasible for small hypothesis spaces.
- **Bayesian abduction**: rank explanations by P(explanation | observation). This
  leverages the existing bayesian/ engine — abduction becomes "inference to the best
  explanation" via Bayes' theorem.
- **Logic-based**: use backward chaining from classical/ but collect candidate premises
  instead of proving them.

**The challenge**: the search space of possible explanations can be exponential. Pruning
strategies are needed. For small hypothesis spaces (tens of candidates), brute force works.

**Scope**: works well for diagnostic systems (medical, technical) where the hypothesis
space is bounded and enumerable.

#### causal/ — Viability: Low. Mathematically involved, field still evolving.

Pearl's causal inference framework goes beyond Bayesian networks: instead of just
observing correlations, it reasons about **interventions** ("what happens if I DO X?"
vs "what happens when I SEE X?") and **counterfactuals** ("what would have happened
if I had done Y instead?").

Key concepts:
- **do-calculus**: rules for computing P(Y | do(X)) from observational data.
  Requires identifying valid adjustment sets in a causal DAG.
- **Graph surgery**: computing the effect of an intervention by modifying the graph
  (removing incoming edges to the intervened variable).
- **Counterfactuals**: requires structural equations and cross-world reasoning.

**Why low viability**: do-calculus is mathematically subtle. Identifying valid adjustments
requires checking graphical criteria (back-door, front-door) that have non-trivial
algorithmic implementations. Counterfactuals add another layer of complexity with
structural equation models. The academic field itself is still producing new results.

**If implemented**: start with graph surgery (interventional queries) only. This is
feasible and immediately useful — "if we change the price, what happens to sales?"
without confounding. Counterfactuals can wait.

**Scope**: interventional queries on small causal DAGs. Full do-calculus and
counterfactuals are a research project.

#### belief/ — Viability: Medium. Well-defined math, niche use.

Dempster-Shafer theory is an alternative to Bayesian probability for reasoning under
uncertainty. Where Bayes requires assigning exact probabilities to every hypothesis,
Dempster-Shafer allows expressing **ignorance** — "I don't know" is a valid state,
distinct from "50/50".

Key concepts:
- **Mass function**: assigns belief mass to subsets of hypotheses (not just individual
  outcomes). `m({fraud}) = 0.3`, `m({legit}) = 0.5`, `m({fraud, legit}) = 0.2` means
  "20% of the evidence is ambiguous."
- **Belief and plausibility**: belief is the minimum support for a hypothesis,
  plausibility is the maximum. The gap represents ignorance.
- **Dempster's rule of combination**: merge evidence from independent sources.

**Implementation**: the math is well-defined and not computationally heavy for small
hypothesis spaces. The main operation (Dempster's combination rule) iterates over
subsets — exponential in the number of hypotheses. For 10 hypotheses, 2^10 = 1024
subsets — feasible. For 30 hypotheses, infeasible.

**Scope**: useful when evidence is imprecise or sources have varying reliability.
Sensor fusion (conflicting sensors), intelligence analysis (uncertain reports),
safety assessment (incomplete data). Niche but well-scoped.

#### ensemble/ — Viability: Premature. Interesting concept, wait for need.

A meta-engine that combines results from multiple inference paradigms:
- classical/ says: "the rule fires" (deterministic)
- bayesian/ says: "P(fraud) = 0.85" (probabilistic)
- fuzzy/ says: "risk = high with degree 0.7" (gradual)

Ensemble would aggregate these into a unified decision via configurable strategy
(weighted voting, Dempster-Shafer combination, fuzzy aggregation, etc.).

**Why premature**: for ensemble/ to make sense, you need:
1. A common interface or result type across engines
2. Multiple engines reasoning about the **same domain** simultaneously
3. A concrete use case that justifies the aggregation overhead

With 3 engines today, the concept is academically interesting but has no concrete
driver. As more engines are added (predicate/, constraint/, statistical/), the need
for multi-paradigm orchestration may emerge naturally.

**Recommendation**: do not implement until a real use case demands it. Document the
concept and revisit when the engine portfolio grows.

#### Summary

| Engine | Viability | Worth it? | Key constraint |
|--------|-----------|-----------|----------------|
| `predicate/` | Medium-Low | High value, large project | Bounded fragment only. Undecidable in general. Needs maths/logic/predicate/ first |
| `temporal/` | Medium | Depends on use case | State explosion without BDDs. Small models only |
| `statistical/` | Medium | Yes, but scope it | Classical statistics only. Not ML. Needs stats/ + linalg/ |
| `markov/` | Medium | Yes, bounded | Tabular MDPs only. Curse of dimensionality for large state spaces |
| `constraint/` | High | Yes | Most independent engine. No heavy upstream dependencies |
| `abductive/` | Medium | Yes, for diagnostics | Search space can be exponential. Small hypothesis spaces |
| `causal/` | Low | Start with interventions only | Full do-calculus is a research project. Graph surgery is feasible |
| `belief/` | Medium | Niche but well-scoped | Exponential in hypothesis count. Useful for imprecise evidence |
| `ensemble/` | Premature | Wait for need | No concrete use case yet. Revisit when engine portfolio grows |

### Guidelines for new engines

- **Same structural pattern**: explain/, domain/, engine/, examples/
- **Trace everything**: every engine must produce an explainable Trace
- **Interface-first**: public interfaces, private implementations
- **Options pattern**: configurable algorithms and parameters
- **Depend on maths/**: engines consume primitives, they don't reinvent them
- **100% test coverage**: individual test functions, `t.Parallel()`, no testify
- See [CODING_STANDARDS.md](CODING_STANDARDS.md) for full conventions
