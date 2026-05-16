# Spring StateMachine — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-statemachine
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: REJECT (a top-level `modules/statemachine/` is not on the canonical roadmap; one narrow piece becomes a follow-up note for `compute/engine/states/`)

## 1. Project summary

Spring StateMachine (SSM) 4.0.1 on JDK 8+, Spring Framework 6.2.9. Implements UML statechart semantics (hierarchical states, regions, pseudo-states, history, fork/join, choice/junction) on top of Spring's reactive (Mono/Flux) and DI machinery. Scope is huge: runtime engine, configurer DSL, listener model, security integration, persistence (JPA / Redis / MongoDB), distributed coordination via ZooKeeper, Spring Boot auto-config, UML import (Eclipse Papyrus), Kryo serialization, JMX/Actuator monitoring. JVM-coupled at every layer — Reactor types in the public API, Spring Beans for wiring, `@EnableStateMachine` / `@WithStateMachine` annotations as the canonical config path.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | Statechart semantics (hierarchical states, regions) | UML-style nested states + orthogonal regions running in parallel | Long-running workflows (orders, KYC, document review) outgrow flat FSMs. Hierarchy lets shared exit/entry logic live on the parent state. |
| 2 | Pseudo-states: choice / junction / fork / join | Conditional branching + parallel split/merge baked into the model | Decision points without escaping the FSM into ad-hoc code. Fork/join is the only sane way to model parallel sub-flows that must converge. |
| 3 | History states (shallow / deep) | Resume a composite state at the last-active substate | Pause-and-resume workflows (e.g. human approval in the middle of automation). |
| 4 | Guards + actions on transitions | Pure predicate + side-effecting hook, both on transitions and state entry/exit/do | Already in `compute/math/fsm/` for guards; SSM adds typed actions with error handlers. |
| 5 | Extended state variables | Arbitrary `Map<String,Object>` attached to the machine, persisted alongside the current state | Avoids state-space explosion (you don't model "retry count" as N states). Composable with persistence. |
| 6 | Event deferral | A state can defer an event for processing in a later state | Workflow inboxes — receive-now-handle-later semantics without an external queue. |
| 7 | `StateMachineContext` snapshot | Self-describing serializable snapshot of current state(s) + extended variables + regions + history | The unit of persistence. The right abstraction even outside Spring. |
| 8 | `StateMachinePersister` interface | Save/restore by machine ID; pluggable backend | JPA / Redis / MongoDB repos provided. Mirrors the contract Go workflows reinvent. |
| 9 | Distributed ensemble (ZooKeeper) | Multi-node coordination: leader election, state replication, eventual consistency over a ZK ensemble | Active-active or active-passive failover for one logical machine across replicas. |
| 10 | Listeners + `@OnStateChanged` hooks | `stateChanged` / `stateEntered` / `transitionEnded` / `eventNotAccepted` plus annotation-driven beans | Observability + side effects without polluting the model. Equivalent in Go: callbacks list on the engine. |
| 11 | Error handling on actions | `errorAction` paired with every action; `StateMachineInterceptor.stateMachineError` for global trap | Necessary the moment actions do I/O. |
| 12 | UML config import (Papyrus) | Read a `.uml` model file and build the machine at runtime | Lets BAs draw the workflow. Niche but real for regulated industries. |
| 13 | Reactive-first API (Mono/Flux) | Non-blocking send/start/stop since 3.0 | Maps directly to Go's `context.Context` + goroutines; the *design* (non-blocking event processing) transfers, the Reactor types do not. |

## 3. Long-tail features (skip)

- **Spring Security integration** (`@EnableStateMachineSecurity`, transition permissions). Authorization belongs in `modules/auth/` (§ 1.2), not inside the FSM engine.
- **`@EnableStateMachine` / `@WithStateMachine` annotation-driven configuration**. Annotation magic is explicitly anti-pattern for yarumo.
- **Spring Boot auto-configuration** (`spring-statemachine-starter`). Bootstrap belongs in `modules/boot/`.
- **JMX / Actuator monitoring endpoints**. Replaced by `modules/telemetry/otel/` + standard health endpoints (§ 1.4).
- **Kryo serialization** (`spring-statemachine-kryo`). JVM-only format. Go equivalent: `encoding/gob` or `json` per consumer choice.
- **SpEL expressions for guards/actions**. yarumo already has `common/expressions/`; if needed, wire that — but most guards in Go are just `func(any) bool`.
- **Eclipse Papyrus UML editor coupling**. Editor tooling is out of scope for a Go library.
- **`StateMachineFactory` + per-key cached machines**. Useful concept but trivial to implement above the engine; no need to wrap.
- **`UsernamePasswordAuthenticationToken` / RememberMe / etc. in security DSL**. Same — auth concern.
- **Session-scoped state machine beans**. Spring scope artifact; irrelevant in Go.
- **ZooKeeper ensemble**. The mechanism is sound but the substrate is wrong: nobody in the Go microservice landscape standardizes on ZK; etcd, Consul, or Postgres advisory locks are more idiomatic. If yarumo ever needs distributed FSM coordination, it is a fresh design, not a port.
- **`spring-statemachine-data-jpa/redis/mongodb`** as separate modules. In yarumo this collapses to: `compute/engine/states/persister` interface + per-backend impl wired by the consumer via `modules/datasource/*` (§ 1.1).

## 4. Mapping to Yarumo

**Existing/planned modules with overlap**:

- `modules/compute/math/fsm/` (**exists**) — provides primitives: `State`, `Transition`, `Guard`, `Machine` backed by `graph.Directed`. Flat FSM only; no hierarchy, no regions, no extended state, no persistence, no actions, no listeners.
- `modules/compute/engine/states/` (**planned** per MEMORY.md / ROADMAP_COMPUTE.md) — the FSM **engine** on top of `math/fsm`. Will own runtime semantics: event dispatch, action invocation, lifecycle, listener fan-out. No design doc yet.
- `modules/compute/math/markov/` — Markov chains. Orthogonal; not a substitute.
- `sdks/processes/` (planned) — `local/` and `temporal/` workflow engines. The proper home for **durable, distributed** workflow orchestration. Temporal already solves distributed coordination, persistence, history, fork/join, and retries — better than ZK-based SSM ever did.

**Gaps this could fill** (and where they actually belong):

| SSM feature | Could it fill a yarumo gap? | Where it belongs (if anywhere) |
|---|---|---|
| Hierarchical states | Yes — `math/fsm/` is flat | `math/fsm/` extension or new `math/hsm/` subpackage. Filed as a math-layer addition, not an engine-layer one. |
| Regions (orthogonal parallel) | Yes — yarumo has no equivalent | Same as above, math-level. |
| Pseudo-states (choice/junction/fork/join/history) | Yes — `math/fsm/` has none | Math layer first (types + semantics), engine consumes. |
| Extended state variables | Engine-level | `compute/engine/states/` — straightforward `map[string]any` + getter/setter on the runtime. |
| `StateMachineContext` snapshot + `StateMachinePersister` | Engine-level | `compute/engine/states/persister/` interface; backends in `sdks/processes/local/` or per-consumer code wiring `modules/datasource/*`. |
| Event deferral | Engine-level | `compute/engine/states/` runtime feature. |
| Distributed (ZooKeeper) | **No** — wrong substrate | If ever needed, `sdks/processes/temporal/` (which is durable + distributed by construction). Do not port ZK. |
| Action error handlers | Engine-level | `compute/engine/states/` — wrap action invocations. |
| Listeners / `@OnStateChanged` | Engine-level (no annotations) | `compute/engine/states/` — function-typed listener slice. |
| UML import | Out of scope | Tools/. Brainstorm only. |

**Anti-patterns to avoid**:

1. **Annotation-driven configuration** (`@EnableStateMachine`, `@WithStateMachine`, `@OnStateChanged`). yarumo wires everything in code.
2. **Spring Bean coupling**. `StateMachineFactory` as a bean, autowired persisters, session-scoped machines — none of this exists in a Go library.
3. **Reactor in the public API**. The non-blocking *behavior* transfers; `Mono<Void>` does not. yarumo uses `context.Context` + plain return values.
4. **God DSL builder**. Spring's `StateMachineBuilder.builder().configureStates()...` chain conflates schema and runtime. yarumo separates schema (declarative `Definition`) from runtime (engine `Run`).
5. **Conflating engine and persistence**. SSM ships `data-jpa/redis/mongodb` modules inside the statemachine project. In yarumo, persistence interfaces live in `compute/engine/states/`; backends are wired by the consumer using `modules/datasource/*`.
6. **Pretending ZooKeeper is the right distributed primitive in 2026**. It isn't, for Go services.
7. **Wrapping UML editors**. Niche enterprise concern; not a library responsibility.

## 5. Recommendation

**REJECT** — a top-level `modules/statemachine/` stays off the canonical roadmap. Spring StateMachine confirms the call rather than overturning it.

Rationale:

- **Engine duplication.** `compute/math/fsm/` exists; `compute/engine/states/` is committed in MEMORY.md and ROADMAP_COMPUTE.md. Creating `modules/statemachine/` (or wrapping SSM) would shadow the planned engine without adding novel semantics that the engine cannot pick up natively at the math + engine layer.
- **The "extra surface" SSM has over `math/fsm/` is not a module — it's features for the existing layers.** Hierarchical states, regions, pseudo-states → grow `math/fsm/` (or a sibling `math/hsm/`). Extended state, persister interface, listeners, error handlers, event deferral → grow the planned `compute/engine/states/`. None of this needs a new top-level module.
- **Persistence isn't a state-machine concern, it's a datasource concern.** SSM bundles `data-jpa/redis/mongodb` only because of Spring's repository-bean pattern. In yarumo the contract is one `Persister` interface in the engine; the backend is wired by the consumer using `modules/datasource/gorm` / `goredis` / `mongo`.
- **Distributed FSM is the wrong abstraction.** What SSM solves with ZooKeeper is what Temporal solves better with durable execution. `sdks/processes/temporal/` is already on the roadmap for durable workflows; that is the right home for "this state machine survives node failure". No ZK port.
- **Java coupling.** Spring Beans + Reactor + annotations + SpEL + Papyrus UML. There is no Go library to wrap here — porting would be a re-implementation, which is exactly the kind of "framework-port-without-a-consumer" yarumo does not pursue.

**One forward-looking note** (does not change the rejection): when `compute/engine/states/` lands, the SSM feature list above is the right design checklist for it — hierarchy, regions, pseudo-states, history, extended state, persister, listeners, error handlers, event deferral. Track these as engine-design tickets under Phase 4 (milestone #10) or under the (future) `compute/engine/states/` work item, **not** as a new module.

## 6. Proposed yarumo placement

Not applicable — REJECT.

For traceability, the design notes that **would** apply to `compute/engine/states/` if/when it is picked up:

**Module**: `modules/compute/engine/states/` (already planned).

**Subpackages** (engine-level features inspired by SSM, none of them new modules):

- `compute/engine/states/` — runtime: dispatch, listeners, lifecycle.
- `compute/engine/states/persister/` — `Persister` interface + `Snapshot` type analogous to `StateMachineContext`.
- `compute/engine/states/listener/` — listener interface + slice fan-out (no annotations).
- `compute/math/hsm/` (**new math subpackage, not a new module**) — hierarchical / regional / pseudo-state primitives, sibling to `math/fsm/`. Engine consumes both.

**Internal deps**: `modules/common/` (errs, assert, log, expressions optional for guards); `modules/compute/math/fsm/` and (future) `compute/math/hsm/`.

**Go libraries to wrap**: none worth wrapping. `looplab/fsm` and `qmuntal/stateless` are flat FSMs (same level as `math/fsm`). For durable distributed workflows, **Temporal** already wins, and it belongs in `sdks/processes/temporal/`, not here.

**Out of scope for v1**: annotations, UML import, ZooKeeper, security DSL, Reactor types, JMX/Actuator, Spring Boot auto-config.

## 7. Open questions

1. Should hierarchical / regional / pseudo-state primitives live in `compute/math/fsm/` (one package) or a sibling `compute/math/hsm/` (cleaner separation, but more imports)? Defer until `compute/engine/states/` is picked up.
2. Does the engine need a built-in `Persister`, or is "the consumer wires its own using `modules/datasource/*`" enough? Lean toward the latter — provide only the `Snapshot` type and a `Persister` interface.
3. For "this workflow survives node failure", do we want a built-in distributed mode in `compute/engine/states/`, or do we punt entirely to `sdks/processes/temporal/`? Punting is consistent with the "library, not framework" stance.
4. Is there a real DaaS / Aluna use case for hierarchical states today? If not, keep `math/fsm/` flat until one appears. Don't pre-build statechart semantics speculatively.
5. UML import (`.uml` → machine definition) — is there demand for "BA draws the workflow, dev consumes"? Probably no; revisit only if a regulated-industry consumer asks.

## 8. ROADMAP delta proposed (NOT applied)

Nothing for ROADMAP_NEW_MODULES.md. The verdict is REJECT for a standalone module. Engine-level features inspired by SSM (hierarchical states, regions, pseudo-states, history, extended state, persister, listeners, error handlers, event deferral) belong inside `compute/engine/states/` — already planned under Phase 4 (milestone #10) / ROADMAP_COMPUTE.md, not under ROADMAP_NEW_MODULES.md. If/when those features are filed, they show up as tickets against `compute/engine/states/` and (if needed) a sibling `compute/math/hsm/` math subpackage — no new top-level module, no annex entry.
