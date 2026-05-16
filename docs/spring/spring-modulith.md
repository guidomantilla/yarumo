# Spring Modulith — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-modulith
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Modulith (v2.0.6, with 2.1.0-RC1 preview) is an opinionated toolkit for building modular monoliths with Spring Boot. Its scope splits into four largely independent concerns:

1. **Module structure verification** — defines module boundaries by Java package convention (`example.order`, with `example.order.internal` hidden) and verifies them at test time via ArchUnit (`ApplicationModules.verify()`).
2. **Event Publication Registry** (the centerpiece) — a persistent outbox that writes one row per `@TransactionalEventListener` inside the business transaction, then an async aspect transitions it through `PUBLISHED → PROCESSING → COMPLETED / FAILED / RESUBMITTED`. Backed by JPA / JDBC / MongoDB / Neo4j starters. Ships a staleness monitor, completion-mode policy (UPDATE / DELETE / ARCHIVE), and a resubmission API (`FailedEventPublications.resubmit(ResubmissionOptions)`).
3. **Event externalization** — `@Externalized("topic::#{...}")` forwards selected publications to Kafka / AMQP / JMS / Spring Messaging *after* the in-process listener completes, reusing the same registry for crash recovery.
4. **Docs + runtime support** — PlantUML / AsciiDoc generation, `/actuator/modulith` endpoint, per-module Micrometer spans, Moments API (passage-of-time events), integration test slice (`@ApplicationModuleTest`).

JVM coupling is **high** for verification (ArchUnit + Java packages + annotations) and Boot starters; **medium-low** for the conceptual outbox/event-registry pattern, which is broker- and ORM-agnostic and the only piece worth porting.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Event Publication Registry** | One row per listener written in the business tx; aspect drives it through 5 states with `getCompletionAttempts()` / `getLastResubmissionDate()`. | This is a richer transactional outbox than the textbook `pending/sent` sketch — it has the lifecycle and recovery hooks every hand-rolled outbox eventually grows. Direct fit for a new `modules/outbox/`. |
| 2 | **Staleness monitor** | Configurable durations per state (`spring.modulith.events.staleness.published / processing / resubmitted`); scheduled task marks stuck rows `FAILED`. Disabled when all values are zero. | Crash-recovery for events stuck mid-flight — most outbox implementations forget this. v1 must include it. |
| 3 | **Completion modes** (UPDATE / DELETE / ARCHIVE) | Configurable retention: keep rows for audit (UPDATE + manual purge), auto-delete on completion, or copy to archive table. | The two real production needs (audit vs. fast table) are explicit and named — saves a design discussion in every consumer. |
| 4 | **Resubmission API** | `FailedEventPublications.resubmit(ResubmissionOptions.defaults().withBatchSize(100).withMaxInFlight(10).withMinAge(5min).withFilter(...))`. Status FAILED → RESUBMITTED → PROCESSING. | Bounded, batched, age-filtered retry. The API surface is exactly what a Go port needs to copy verbatim (minus annotations). |
| 5 | **Event externalization** (`@Externalized`) | Selected events forwarded to Kafka / AMQP / JMS / Spring Messaging *after* listener completion. SpEL routing key, programmatic `EventExternalizationConfiguration.select / mapping / routeKey`. Reuses the registry for broker-failure recovery. | Cleanly separates "in-process domain event" from "cross-service published event". Maps onto `modules/messaging/events/` (§ 1.3) + `modules/outbox/` cooperation. |
| 6 | **`@ApplicationModuleListener` ergonomics** | One annotation = `@Async + @Transactional(REQUIRES_NEW) + @TransactionalEventListener`. Forces "new tx, async, post-commit" defaults. | The concept transfers without the annotation: yarumo's pub/sub default subscriber shape should be **post-commit, fresh `context.Context`, separate goroutine** — not synchronous in-tx. |
| 7 | **`AssertablePublishedEvents`** test helper | `assertThat(events).contains(OrderCompleted.class).matching(OrderCompleted::getOrderId, ref.getId())` inside `@ApplicationModuleTest`. | Worth porting as a primitive in `modules/messaging/events/` — a recorder publisher + matcher helpers. Low cost, high test ergonomics. |
| 8 | **Serialize externalization** flag | `spring.modulith.events.externalization.serialize-externalization=true` — one event in flight at a time, prevents later events overtaking earlier on resubmission spikes. | Tiny but important: an outbox without ordering control will silently reorder on retry. Adopt as an explicit `Ordered` mode. |

## 3. Long-tail features (skip)

- **`ApplicationModules.of(...).verify()` (ArchUnit boundary tests)** — `go-arch-lint` / `archtest` cover this via config. No new value.
- **Package convention for modules** (`example.order.internal` hidden, base package = API) — Go enforces this with `internal/` at compile time. Doesn't need a tool.
- **Named interfaces** (`@NamedInterface("spi")` + `allowedDependencies = "order :: spi"`) — Go's idiomatic equivalent (separate sub-packages) already covers this.
- **Nested / open application modules** — Java-package-tree workaround for what Go solves with `internal/`. Skip.
- **PlantUML / AsciiDoc module documentation generation** — yarumo's per-module `graph.go` dependency images already cover the "module map" need. Skip generating textual docs.
- **`/actuator/modulith` endpoint** — Spring-Boot-specific. Equivalent in yarumo would be a JSON dump of the dependency graph; low value, defer.
- **Micrometer per-module spans (`spring-modulith-observability`)** — `modules/telemetry/otel/` already exists; adding a `module.name` span attribute is one line in middleware, not a feature.
- **Moments (passage-of-time events API)** — niche; cron / scheduled jobs in `modules/managed/CronWorker` cover this.
- **Neo4j / JPA starters** — irrelevant; yarumo only commits to `gorm` (+ later `mongo`, `goredis`).
- **`@Externalized` SpEL routing** — yarumo will use explicit Go (`func(T) string`), not an expression DSL.
- **`@Modulithic` god-config** — no central "list all modules + shared modules + additional packages" struct.
- **`@ApplicationModuleTest` test slice** — Go test isolation is per-package; no equivalent need.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**:

- **`modules/messaging/` (§ 1.3 events/)** — yarumo's nominal-typed in-process pub/sub façade (`Subscribe[T]` over `DirectChannel`). Modulith's `@ApplicationModuleListener` ergonomics inform the **default subscriber semantics**: post-commit, fresh `context.Context` (no parent cancellation propagation), separate goroutine. `AssertablePublishedEvents` translates cleanly to a `Recorder` publisher + matcher helpers in `messaging/events/`.
- **`modules/messaging/` core (§ 1.3)** — Modulith's externalization layer (events → Kafka/AMQP) overlaps with the existing `Publisher` driver design. The new insight is the **two-stage contract**: in-process listeners fire first; only after success does the externalizer push to a broker. The cooperation point is the registry — both layers consume the same outbox rows.
- **`modules/datasource/gorm/` (§ 1.1 row-level audit hooks)** — the registry table itself lives behind the `gorm` driver; Modulith's `JdbcEventPublicationRepository` is the closest analogue. yarumo's audit hooks (`CreatedAt` / `UpdatedAt`) are a different concern from the publication's `completion_date` — distinct but adjacent.

**Gaps to fill**:

The Event Publication Registry has **no landing zone** in the current § 1 modules. `messaging/events/` is the in-process façade; `datasource/gorm/` is the driver. Neither owns the *persistent state machine for delivery*. This analysis proposes a new top-level module `modules/outbox/` (see § 6).

Specifically:

1. Outbox lifecycle states beyond `pending/sent` — needs `PUBLISHED / PROCESSING / COMPLETED / FAILED / RESUBMITTED` with `getCompletionAttempts()` from day one.
2. Staleness sweeper — crash recovery for events stuck in `PROCESSING`.
3. Completion modes — explicit `UPDATE` / `DELETE` / `ARCHIVE` config.
4. Resubmission API — `BatchSize` / `MaxInFlight` / `MinAge` / `Filter` predicate.
5. Standard test recorder + matcher for `messaging/events/`.
6. Documented "serialize externalization" mode for ordering guarantees.

**Anti-patterns to avoid**:

1. **Annotations / reflection-based listener wiring** — `@ApplicationModuleListener` magic does not translate. yarumo subscribers register explicitly: `events.Subscribe[OrderCompleted](pub, handler)`.
2. **SpEL routing keys** — externalization config takes a Go function `func(T) string`, not an expression DSL.
3. **`@Modulithic` god-config** — no central "list all modules + shared modules + additional packages" struct.
4. **ArchUnit-style runtime boundary verification baked into yarumo** — defer to `go-arch-lint`.
5. **Tying outbox storage to one ORM** — Modulith's JPA / JDBC / Mongo / Neo4j starters are useful as a list of backends to support, but the abstraction is per-driver in yarumo (`outbox/gorm/`, `outbox/mongo/`), not a Boot-style starter.
6. **Implicit audit via `UPDATE` mode** — completion mode is *event-delivery* retention, not business-event audit. Document the boundary; don't conflate.
7. **Generic event payloads stored as JSON blobs only** — Modulith's `EventSerializer` abstraction is right; yarumo should accept a pluggable serializer (default JSON, opt-in protobuf / avro) rather than hard-coding Jackson-equivalent.

## 5. Recommendation

**PARTIAL — adopt the Event Publication Registry design into a new top-level `modules/outbox/`; keep boundary verification rejected; absorb the listener semantics + test helpers into `modules/messaging/events/`.**

Modulith's primary insight — the **rich outbox lifecycle** (5 states, staleness monitor, completion modes, resubmission API, serialize-externalization flag) — is materially better than any textbook outbox sketch and is reusable without bringing in any Spring/JVM coupling. It also clarifies a missing contract between in-process pub/sub and broker forwarding: in-process listeners run post-commit *first*, externalization to a broker happens *after* listener success, both backed by the same persistent registry.

Beyond that, every other feature Modulith adds (boundary verification, package conventions, module docs, actuator endpoint, observability spans, Moments) is **already covered** by Go-native tooling (`internal/`, `go-arch-lint`, existing `graph.go`, `modules/telemetry/otel/`, `modules/managed/CronWorker`) or by `modules/messaging/events/` itself. Net new value: a new top-level `modules/outbox/` module and a small set of test helpers for `messaging/events/`.

## 6. Proposed yarumo placement

**NEW module**: `modules/outbox/` (not yet in ROADMAP).

**Why a new module, not a sub-package of `messaging/` or `datasource/`**: the registry is a state machine with its own lifecycle (`staleness` worker), its own storage contract (`Repository` SPI over multiple drivers), and its own public API (`FailedPublications.Resubmit`, `CompletedPublications.DeleteOlderThan`). It depends on both `messaging/` (to receive events) and `datasource/` (to persist them) — putting it inside either creates a circular conceptual dependency. It belongs at the same tier as those two.

**Subpackages**:

```
modules/outbox/
  publication.go        Publication{ID, EventType, Payload, Status, CompletionAttempts, LastResubmissionDate, CreatedAt, CompletedAt}
  status.go             Status enum: PUBLISHED / PROCESSING / COMPLETED / FAILED / RESUBMITTED
  completion.go         CompletionMode enum: UPDATE / DELETE / ARCHIVE
  options.go            ResubmissionOptions{BatchSize, MaxInFlight, MinAge, Filter}, WithBatchSize, WithMinAge, WithFilter
  repository.go         Repository interface (SPI): Save, MarkProcessing, MarkCompleted, MarkFailed, FindIncomplete, FindFailed, DeleteCompletedOlderThan
  registry.go           Registry — wraps Repository, exposes CompletedPublications / FailedPublications / IncompletePublications views
  serializer.go         Serializer interface (default: JSON via encoding/json)
  errors.go             ErrPublicationNotFound, ErrRepositoryNil, etc. (TypedError pattern)

  gorm/                 GORM-backed Repository (default, P1). Schema migration helper.
  mongo/                MongoDB-backed Repository (later).
  jdbc/                 database/sql-backed Repository (no ORM, later — matches Modulith's JdbcEventPublicationRepository).

  staleness/            Periodic sweeper (managed.Worker). Config: StalePublished, StaleProcessing, StaleResubmitted durations.
                        Disabled when all three are zero (matches Modulith default).

  externalize/          Externalizer driver — consumes COMPLETED in-process publications and forwards to a broker via messaging/.
                        Config: select func(T) bool, routeKey func(T) string, ordering (Parallel | Serialized).
                        On broker failure: marks the row FAILED in the registry — same recovery path as listener failures.

  testing/              Recorder publisher + AssertPublishedEvents matcher (port of AssertablePublishedEvents).
                        Lives here, not in messaging/events/, because the recorder is registry-backed.
```

**Public API shape (sketch)**:

```go
package outbox

type Status int
const (
    StatusPublished Status = iota
    StatusProcessing
    StatusCompleted
    StatusFailed
    StatusResubmitted
)

type CompletionMode int
const (
    ModeUpdate CompletionMode = iota  // default — keep rows, manual purge
    ModeDelete                         // delete on completion
    ModeArchive                        // copy to archive table, delete original
)

type ResubmissionOptions struct { /* private fields, With* funcs */ }
func NewResubmissionOptions(opts ...Option) ResubmissionOptions

type Registry struct { /* ... */ }
func (r *Registry) Save(ctx context.Context, eventType string, payload []byte) (*Publication, error)
func (r *Registry) Failed() FailedPublications
func (r *Registry) Completed() CompletedPublications
func (r *Registry) Incomplete() IncompletePublications

type FailedPublications interface {
    Resubmit(ctx context.Context, opts ResubmissionOptions) (int, error)
}
```

**Internal deps**:

- `modules/messaging/` (§ 1.3) — `events/` for the in-process listener contract; `Publisher` drivers (kafka/, rabbitmq/) for the externalizer.
- `modules/datasource/gorm/` (§ 1.1) — default storage + `WithTransaction(ctx, db, fn)` helper for atomic save-in-business-tx.
- `modules/managed/` — for `staleness/` worker lifecycle (Start / Stop / Done).
- `modules/common/resilience/` — retry budget for `Resubmit`; circuit breaker for the externalizer's broker calls.
- `modules/common/log/slog/slogctx/` — context-bound attrs (publication ID, attempt count) on every log line.

**Go libraries to wrap**:

- None mandatory — schema + worker are plain SQL/GORM. Outbox-specific Go libs in the ecosystem (`oklog/outbox`, various `transactional-outbox` ports) are immature; designing from scratch is fine and avoids inheriting their quirks.
- Optional: `riverqueue/river` if the resubmission API ever needs distributed coordination (deferred — single-node sweeper is fine for v1).

**Out of scope for v1**:

- `ARCHIVE` completion mode implementation (design the enum + interface to support it, default to `UPDATE`, ship `ARCHIVE` in v2 once a real compliance consumer asks).
- MongoDB / JDBC drivers (ship `gorm/` first).
- Distributed coordination for the staleness sweeper (single-node, leader-elected via `managed.Lifecycle` if multi-node ever shows up).
- ArchUnit-style boundary verification (`tools/modulith/` stays rejected — `go-arch-lint` covers it).
- Module documentation generation (PlantUML / AsciiDoc).
- `/actuator/modulith` JSON endpoint.
- SpEL routing keys (use `func(T) string`).
- Neo4j backend.
- Moments / passage-of-time API.

## 7. Open questions

- Should `Status = RESUBMITTED` be distinct from `PUBLISHED`, or folded into a single retry counter? Modulith keeps them separate for staleness math (different timeout per state). Same likely applies here — recommend keeping all five.
- `CompletionMode = ARCHIVE` requires a second table. Worth shipping in v1, or defer to v2 once a real compliance consumer asks? Lean: design the interface to support it, default to `UPDATE`, implement `ARCHIVE` later.
- Should `modules/messaging/events/` enforce the post-commit / new-context default, or expose both shapes (sync-in-tx vs. async-post-commit)? Modulith offers both via `@TransactionalEventListener` vs. `@ApplicationModuleListener`. Recommend: yarumo picks **async-post-commit as default**, with sync-in-tx as opt-in for cases where the caller explicitly wants atomic side effects.
- Externalization ordering — does an `Externalizer` consume from the outbox table, or hook into the in-process publisher with its own outbox row? Modulith uses the same registry for both (the externalizer is itself a transactional event listener). Simplest design, recommended.
- Should the staleness sweeper own its own goroutine, or compose into a generic `modules/managed/CronWorker`? Either works — `outbox/staleness/` registering a worker against `managed/` is cleaner separation.
- Where does the `Serializer` (JSON vs. proto vs. avro) get configured — per registry instance, per event type, or per externalizer driver? Modulith abstracts it via `EventSerializer`. Recommend: per registry instance, default JSON, override via `Option`.

## 8. ROADMAP delta proposed (NOT applied)

- **Add `modules/outbox/` as a new top-level module under § 1** with `gorm/`, `mongo/` (later), `jdbc/` (later), `staleness/`, `externalize/`, `testing/` subpackages. Internal deps: `messaging/`, `datasource/gorm/`, `managed/`, `common/resilience/`, `common/log/slog/slogctx/`. Priority: **Medium** — gates any consumer that needs at-least-once delivery across service boundaries; can ship after `messaging/` (§ 1.3) and `datasource/gorm/` (§ 1.1) are in place.
- **Extend § 1.3 (`modules/messaging/events/`)** to document the standard listener contract — post-commit, fresh `context.Context` (caller cancellation does not propagate), separate goroutine, error → outbox `FAILED`. Add a `Recorder` test publisher (port of `AssertablePublishedEvents`) — though the registry-backed version lives in `outbox/testing/`.
- **Note `modulith verify`** (ArchUnit-style architectural constraint checking) — **REJECT**, `go-arch-lint` covers it. No new tool.
- **Note PlantUML / AsciiDoc module docs** — **REJECT**, per-module `graph.go` covers the "module map" need.
- **Note `/actuator/modulith` endpoint** — **DEFER**, low value; if needed, a one-off JSON dump of the dependency graph is trivial.
- **Note Moments API (passage-of-time events)** — **REJECT**, `modules/managed/CronWorker` covers scheduled triggers.
