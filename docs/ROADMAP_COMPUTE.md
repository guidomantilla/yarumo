# Roadmap — modules/compute

Forward-looking work for `modules/compute/`. The module's **scope and current state** live in [`modules/compute/README.md`](../modules/compute/README.md). The formal correctness analysis lives in [`modules/compute/CORRECTNESS.md`](../modules/compute/CORRECTNESS.md).

> **Status as of 2026-05-12**: the correctness chain (prompts → CORRECTNESS.md → ACCEPTANCE_TESTS.md spec → codegen prompts → `*_test.go`) is **largely built**. All 11 math packages and 5 engine paradigms are covered by formal analysis. Remaining gaps (one missing test file, one pending spec patch, two open proposals) are tracked in [milestone Phase 9](https://github.com/guidomantilla/yarumo/milestones/10) via [YA-0085](https://github.com/guidomantilla/yarumo/issues/85) … [YA-0088](https://github.com/guidomantilla/yarumo/issues/88).

---

## 1. Engines nuevos — process dimension

The mathematical and inference foundations are in place. The next layer is **process** engines that compose math/ primitives into runtime workflow logic.

### Dependency graph

```
engine/states      → math/fsm → math/graph         (✓ math/ ready)
engine/montecarlo  → math/markov → math/graph + math/stats   (✓ math/ ready)
engine/mining      → math/graph + math/markov + math/stats   (✓ math/ ready)
engine/steps       → common/   (no math/ dep)
```

### Packages

| Engine | Description | math/ deps |
|---|---|---|
| `engine/steps/` | Short in-process workflows: retry, timeout, compensation | None (common/ only) |
| `engine/states/` | FSM engine with business context | `math/fsm` |
| `engine/montecarlo/` | Stochastic simulation | `math/markov`, `math/stats` |
| `engine/mining/` | Process mining | `math/graph`, `math/markov`, `math/stats` |

**Status**: Planned. Math dependencies (`graph/`, `fsm/`, `markov/`) all implemented and formally verified.

**Priority**: Medium. Targets are concrete enough to ticket when work begins; until a real consumer needs them, they stay as design surface here.

---

## 2. What does NOT belong in compute/

Explicit exclusions to keep the module clean:

- I/O, networking, storage (that is SDK / app scope).
- Heavy ML runtimes (TensorFlow / PyTorch equivalents).
- Persistent state or lifecycle (that is `managed/` scope).
- Distributed computation (Spark / Flink). Out of scope; compute/ is in-process.

---

## 3. Brainstorm — math/ extensions

Ideas that would fit `math/` if a real use case appears. **Not actionable.** Filed here so we can quickly grade demand when something concrete shows up.

| Package | Scope sketch |
|---|---|
| `math/optimization/` | Linear programming, simplex, constraint satisfaction (CSP), genetic algorithms, simulated annealing |
| `math/geometry/` | Spatial indexing, distances, convex hull, Voronoi — useful for clustering, embeddings |
| `math/numeric/` | Interpolation, root-finding, numerical integration, automatic differentiation |
| `math/information/` | Entropy, mutual information, KL divergence, coding theory |
| `math/signal/` | FFT, convolution, filtering — only if there is a time-series consumer |

---

## 4. Brainstorm — engine/ extensions

Ideas that would fit `engine/` if a real use case appears.

| Engine | Scope sketch |
|---|---|
| `engine/search/` | A*, beam search, minimax / alpha-beta, MCTS |
| `engine/planning/` | STRIPS, HTN planning, PDDL-like planners |
| `engine/constraint/` | Constraint propagation, arc consistency, backtracking with heuristics |
| `engine/abductive/` | Abductive reasoning (best explanation given evidence) |

---

## 5. Brainstorm — new top-level areas in compute/

Categories not yet present in the module.

| Area | Scope sketch |
|---|---|
| `compute/learn/` | Lightweight learning without heavy ML frameworks: kNN, Naive Bayes, decision stumps, logistic regression. Tabular reinforcement learning (Q-learning, SARSA). Online learning (bandits, UCB) |
| `compute/nlp/` | Lightweight text primitives: tokenization, stemming, TF-IDF, similarity. Grammar parsing (CFG, PEG) |
| `compute/schedule/` | Scheduling with constraints, resource allocation, job-shop |
| `compute/aggregate/` | Advanced aggregation operators (OWA, Choquet integral, composite scoring) — complements MCDM |

---

## 6. Under deeper review

Items worth a focused investigation before classifying as accept / reject.

- **`compute/math/argumentation/`** — argumentation frameworks (Dung, ASPIC+), attack/defense between arguments. Candidate for `math/` at the same tier as `graph/`. Use case: formal conflict resolution after inference when rules contradict in complex systems. Decision pending until a concrete need surfaces in DaaS or Aluna.

---

## Related docs

- [`modules/compute/README.md`](../modules/compute/README.md) — scope, package index, design decisions, limitations.
- [`modules/compute/CORRECTNESS.md`](../modules/compute/CORRECTNESS.md) — formal correctness analysis (2026-03-09 round 3).
- [`modules/compute/tests/specs/README.md`](../modules/compute/tests/specs/README.md) — correctness chain documentation.
- `modules/compute/` engineering work in flight — tracked in [milestone Phase 9](https://github.com/guidomantilla/yarumo/milestones/10) ([YA-0085](https://github.com/guidomantilla/yarumo/issues/85) … [YA-0088](https://github.com/guidomantilla/yarumo/issues/88)).
