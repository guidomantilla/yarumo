# Roadmap — New Modules & Tools

Engineering work planned for modules that do not yet exist and for new tools. Each item has a placement decision and is annotated with status. Bigger workspace context — placement principle, status legend, related docs — at the bottom of this file.

> **Implementation work is tracked via GitHub milestones**, not in this doc — see [open milestones](https://github.com/guidomantilla/yarumo/milestones). This file scopes **new** modules / tools that no milestone covers yet; once a track absorbs an item it leaves this doc.

## Current state (as of 2026-05-15)

| Milestone | State | Issues | Scope |
|---|---|---|---|
| [Phase 0 — Standards hygiene](https://github.com/guidomantilla/yarumo/milestone/1) | Drained | 5 / 5 | workspace-wide standards cleanup |
| [Phase 1 — Modules: Crypto](https://github.com/guidomantilla/yarumo/milestone/11) | **Closed** | 30 / 30 | `modules/common/crypto/*` (consolidated) |
| [Phase 2 — Modules: Common (non-crypto)](https://github.com/guidomantilla/yarumo/milestone/8) | Drained | 28 / 28 | rest of `modules/common/` |
| [Phase 3 — Modules: Config / Managed / Telemetry](https://github.com/guidomantilla/yarumo/milestone/9) | **Active** | 0 / 20 | `modules/config/`, `managed/`, `telemetry/otel/` |
| [Phase 4 — Modules: Compute](https://github.com/guidomantilla/yarumo/milestone/10) | Queued | 0 / 8 | `modules/compute/` correctness chain |

Items in [§ 1](#1-new-modules) and [§ 2](#2-new-tools) are **not yet ticketed** — they get filed when a track picks them up. Phase-2 follow-ups born from reviews (YA-0042..0044, YA-0154..0162, plus the unmiliestoned refactor tickets #164–#166) live in the open-issues backlog and will be folded into the appropriate track when work resumes.

---

## Placement principle

| Bucket | Where it lives | Examples |
|---|---|---|
| **Pure library, no lifecycle, no external SDK deps** | `modules/common/` | `assert`, `errs`, `crypto`, `expressions` |
| **Bootstrap (one-shot)** | `modules/config/` | viper-driven config loading |
| **Lifecycle components (Start/Stop/Done)** | `modules/managed/` | `BaseWorker`, `HttpServer`, `GrpcServer`, `CronWorker` |
| **Observability** | `modules/telemetry/otel/` | OTel SDK setup |
| **Application wiring (DI / orchestration)** | `modules/boot/` | Container, BeanFn, Run() |
| **Domain modules** | `modules/<name>/` | `auth`, `datasource`, `messaging`, `health` |
| **Standalone tooling** | `tools/<name>/` | code generators, build helpers |

A library belongs in `common/` only if it has **no lifecycle** and depends on **no external SDK**. Anything else gets its own module.

## Status legend

| Status | Meaning |
|---|---|
| **Issue YA-NNNN** | Filed and tracked on GitHub |
| **Planned** | Designed in this doc; ready to be filed when work starts |
| **Brainstorm** | Speculative; needs a real use case before filing |

---

# 1. New modules

## 1.1. `modules/datasource/` — DB and cache adapters

**Status**: Planned
**Why a new module**: connection pools have lifecycle (Start = open, Stop = close). External SDKs.
**Granularity**: one subpackage per driver. Each is independently shippable.

Migrated from go-feather-lib:

| Subpackage | Backend | Status |
|---|---|---|
| `modules/datasource/gorm/` | SQL via GORM | Planned (high prio) |
| `modules/datasource/mongo/` | MongoDB | Planned |
| `modules/datasource/goredis/` | Redis | Planned |
| `modules/datasource/gocql/` | Cassandra | Planned (low prio) |
| `modules/datasource/ldap/` | LDAP directory (read users / groups / orgs) | Planned — pairs with `auth/` (§ 1.2). Driver, not full LDAP framework. |
| `modules/datasource/vector/` | Vector DBs (pgvector, Pinecone, Weaviate, Qdrant) | Planned — unblocks RAG in `sdks/decisions/companions/embeddings/`. Mirrors Spring AI vector stores. |

Common pattern per driver: `Context` (url/server/credentials), `Connection` (open/close), `TransactionHandler` (callback-based).

**Cross-driver features planned at the `datasource/` core level**:

- **`WithTransaction(ctx, db, fn)` helper** — Go has no `@Transactional` AOP; a single function-shaped helper that runs `fn` inside a tx and rolls back on error covers 90% of the use case. Lives in the core, used by every driver.
- **Row-level audit hooks** — `CreatedBy` / `LastModifiedBy` / `CreatedAt` / `UpdatedAt` columns auto-populated via `BeforeSave` / `BeforeUpdate` hooks (Spring Data Auditing equivalent). Implemented in `modules/datasource/gorm/`. Different layer from the **event-level** audit trail in `modules/audit/` (§ 3.2) — the two cooperate.
- **Data-lifecycle policies** — retention, archival, soft / hard delete (GDPR-aware). Per-table policy declared as struct tags or programmatic config; background sweeper enforces. Cross-references `best-practices/.../data-design/lifecycle.md`. Decision pending: implement in the gorm driver only, or factor into a `datasource/lifecycle/` sub-package.

## 1.2. `modules/auth/` — Authentication, Authorization, Principals

**Status**: Planned
**Why a new module**: stateful (token validators, user loaders), depends on `datasource` for persistence, ships HTTP/gRPC middleware with lifecycle.

Migrated from go-feather-lib's `security/`:

- **`AuthenticationService`** — `Authenticate(credentials) → Principal`, `Validate(token) → Principal`.
- **`AuthorizationFilter`** — HTTP middleware that extracts the token, authorizes, injects `Principal` into context.
- **`PrincipalManager`** — Principal CRUD (base + GORM implementations).

**Internal deps**: `modules/common/crypto/passwords`, `modules/common/crypto/tokens`, `modules/datasource/` (for persistence).

> The `DelegatingPasswordEncoder` initially planned here lives in `common/crypto/passwords` instead — see [YA-0020](https://github.com/guidomantilla/yarumo/issues/20).

**Sub-modules planned within `auth/`**:

- **`auth/oauth2/`** — OAuth2 client (consume Google / GitHub / Microsoft / etc.) + resource server (validate JWTs issued by an external IdP). Spring splits this into `oauth2-client` and `oauth2-resource-server` libraries; in yarumo both share one sub-module. Reuses `modules/common/crypto/tokens` for JWT validation.
- **LDAP authentication provider** lives here. The LDAP-as-directory data source (read users/groups/orgs) lives in `modules/datasource/ldap/` — see § 1.1.

Reference: [Annex B](#annex-b--spring-security-feature-catalog) at the bottom of this doc captures the full Spring Security feature space that informed the scope.

## 1.3. `modules/messaging/` — EIP layer + brokers

**Status**: Planned
**Why a new module**: connections + consumers/producers with lifecycle. External SDKs per broker.

Two layers consolidated under a single module. The full design rationale and Spring Messaging / Spring Integration feature reference are in [Annex A](#annex-a--spring-messaging--integration-reference).

**Proposed layout**:

```
modules/messaging/
  message.go            Message[T], Headers, Builder
  channel.go            Channel interface (Send/Receive)
  channels/
    direct.go           DirectChannel (synchronous)
    queue.go            QueueChannel (buffered)
    pubsub.go           PubSubChannel (broadcast)
    priority.go         PriorityChannel
    rendezvous.go       RendezvousChannel (zero-capacity handoff)
  handler.go            MessageHandler = func(Message) error
  endpoints/            EIP components (Transformer, Filter, Router, Splitter, Aggregator, Activator)
  schema/               Schema Registry client (see below)
  events/               nominal-typed pub/sub façade (see below)
  rabbitmq/
    amqp/               driver implementing Channel interfaces
    streams/            driver implementing Channel interfaces
  kafka/
    cdc/                Debezium CDC event parsing (see below)
  nats/                 future
```

Everything explicit — no DI container, no annotations, no auto-wiring. Channels are created manually, endpoints are connected with code.

**Sub-modules planned within `messaging/`**:

- **`messaging/schema/`** — Schema Registry client. Spring Kafka integrates with Confluent Schema Registry; `data-engineering/contract-design/` lists schema contracts as a central category. Avro/Protobuf payloads need pre-publish and post-consume validation. `Registry` interface with impls `confluent/`, `glue/`, `apicurio/`. Operations: `Register(subject, schema)`, `Get(id)`, `Compatibility(subject, schema)`. Hook into `Publisher`/`Consumer` for auto-validate. Lives here because it's always used alongside a broker — never standalone.
- **`messaging/kafka/cdc/`** — Debezium CDC event parsing. Every CDC consumer reinvents the parser for Debezium's `before`/`after`/`source`/`op` envelope. Typed wrapper `CDCEvent[T]` + dispatcher `Handle[T](msg, fn)` that detects `INSERT`/`UPDATE`/`DELETE`/`SNAPSHOT`. Sub-package of the Kafka driver — only applies when Debezium is upstream.

**Sub-module decision — `messaging/events/`** (Spring `ApplicationEventPublisher` / Modulith Events):

Rather than create a sibling `modules/events/` for lightweight in-process typed pub/sub, expose a thin convenience layer **inside** `modules/messaging/events/` that:

- Provides `Publisher` / `Subscribe[T]` over the existing `DirectChannel` infrastructure.
- Uses nominal types as routing keys (`UserCreated`, `OrderPlaced`) instead of `Message[T]` envelopes.
- Targets in-process domain events without EIP boilerplate.

Same machinery as the rest of `messaging/`, simpler façade. Mirrors the relationship Spring has between `spring-messaging` (channels) and `ApplicationEventPublisher` (pub/sub).

## 1.4. `modules/health/` — runtime + HTTP endpoint

**Status**: Planned
**Pairs with**: `modules/common/health/` primitives — [YA-0077](https://github.com/guidomantilla/yarumo/issues/77) **closed 2026-05-13**, so the leaf side (interfaces, status types, synchronous aggregator) is already in place.

What remains for this module: the runtime side — registers checkers, aggregates status, runs probes on a schedule, exposes an HTTP endpoint (`/healthz`, `/readyz`). Lives here because it has goroutine-driven lifecycle and HTTP integration.

Pulled from go-feather-lib's `health/`:

- Memory stats, uptime, goroutine count.
- HTTP handler for health endpoints.
- `Shutdown()` integrated with `managed.Lifecycle`.

## 1.5. `modules/boot/` — Application Wiring

**Status**: Planned
**Why a new module**: this is the orchestrator of lifecycle. Depends on `managed` and `config`. The opposite of "common".

**Problem**: yarumo has `config` (one-shot bootstrap) and `managed` (lifecycle Start/Stop/Done) but no formal mechanism for connecting components. Wiring is done by hand in `sample/main.go`. This is the missing link.

**What go-feather-lib's `boot/` had** (and what to avoid):

- `ApplicationContext` — god struct with ~30 public fields: app metadata, environment, database (GORM), security (password encoder/generator/manager, token manager, auth service/filter, authz service/filter), HTTP (gin router), gRPC. Monolithic and directly coupled to gin / GORM / grpc.
- `BeanBuilder` — struct with 17 factory functions, one per component. Each receives `*ApplicationContext` and returns the built component. Defaults included (bcrypt, JWT, gin routes `/login`, `/health`, `/info`, `/api`).
- `Init()` — creates `ApplicationContext` via `NewApplicationContext()`, calls a delegate, attaches servers to lifecycle (`qmdx00/lifecycle`), runs `app.Run()`.
- `Enablers` — feature toggles (HttpServerEnabled, GrpcServerEnabled, DatabaseEnabled).
- Fixed sequential order: environment → config → datasource → security → http → grpc.

**Anti-patterns to avoid**:

1. Direct coupling to gin / GORM / grpc — if you don't use one, you still import it.
2. God-struct — `ApplicationContext` with 30+ public fields.
3. Fixed order — initialization sequence is not configurable.
4. No generics — `any` for the gRPC service server.

**Sketched design for yarumo** (informed by the deleted sandbox):

```go
// Container holds wired components. Extensible via typed map.
type Container struct {
    components map[reflect.Type]any
}

// Register[T] registers a component by type.
func Register[T any](c *Container, component T)

// Resolve[T] retrieves a component by type.
func Resolve[T any](c *Container) (T, error)

// BeanFn is a factory that populates the container.
type BeanFn func(ctx context.Context, c *Container) error

// Run orchestrates: config.Default -> BeanFns -> managed lifecycle -> signal wait.
func Run(ctx context.Context, name, version, env string, beans ...BeanFn) error
```

The container must be **generic and unaware of specific types** (must not import gin, GORM, etc.). Concrete frameworks are wired via `BeanFn`s defined by the consumer.

**Internal deps**: `modules/config`, `modules/managed`, `modules/common/crypto/*`.

---

# 2. New tools

## 2.1. `tools/routegen/` — Gin route code generation

**Status**: Planned

Code-generation tool that, given methods annotated with `@route METHOD /path`, generates route definition functions for Gin.

**Input**:

```go
type handlers struct {
    repository Repository
}

// @route POST /events
func (h *handlers) CreateEvent(c *gin.Context) {}

// @route GET /events/:id
func (h *handlers) GetEvent(c *gin.Context) {}
```

**Output** (`zz_routes_gen.go`):

```go
// Code generated by routegen; DO NOT EDIT.
package core

import "github.com/gin-gonic/gin"

func Route_POST_events(h *handlers) (string, string, gin.HandlerFunc) {
    return "POST", "/events", h.CreateEvent
}

func Route_GET_events_id(h *handlers) (string, string, gin.HandlerFunc) {
    return "GET", "/events/:id", h.GetEvent
}
```

**Input rules**:

- Receiver must be the type indicated by `--type`.
- Exact signature: `func(*gin.Context)` (1 param, 0 returns).
- Comment with `@route METHOD PATH` (METHOD ∈ GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, PATH starts with `/`).
- Methods without `@route` are ignored.
- Duplicate `(METHOD, PATH)` is an error.

**Naming convention**:

- Format: `Route_<METHOD>_<sanitized_path>`.
- Sanitization: strip leading `/`, replace `/`, `:`, `{`, `}` with `_`, collapse `_`, empty → `root`, leading digit → `p_` prefix.
- Collisions: counter suffix (`_2`, `_3`).

**Architecture**:

```
cmd/routegen/main.go
internal/routegen/
  reader/    discover .go files, AST parse, inspect receiver methods, parse @route, validate
  generator/ namer (sanitize path), model builder, code emitter, go/format
  shared/    types (RouteModel, RouteDef, MethodNode), errors
```

**CLI**:

```bash
go install github.com/<org>/routegen/cmd/routegen@latest

# In the package:
//go:generate routegen --type handlers --dir . --out zz_routes_gen.go
```

**Flags**: `--dir` (default `.`), `--type` (required), `--out` (default `zz_routes_gen.go`), `--func-prefix` (default `Route_`).

**Out of scope for v1**: `RegisterRoutes(r gin.IRoutes, h *handlers)` opt-in registrar, groups (`/v1`), middlewares via annotation, multiple receiver types.

---

# 4. go-feather-lib migration tracking

Status of the migration from go-feather-lib into yarumo, filtered to items that land **outside** `modules/common/`. The `modules/common/` rows were tracked directly via tickets in [milestone Phase 2 — Modules: Common (non-crypto)](https://github.com/guidomantilla/yarumo/milestones/8) (YA-0035 … YA-0080), all of which closed on 2026-05-13.

## 4.1. Pending

| go-feather-lib | New placement | Section in this doc | Priority |
|---|---|---|---|
| `boot/` (wiring + DI) | **`modules/boot/`** | § 1.5 | High |
| `security/AuthenticationService`, `AuthorizationFilter`, `PrincipalManager` | **`modules/auth/`** | § 1.2 | Medium |
| `datasource/gorm` | `modules/datasource/gorm/` | § 1.1 | Medium |
| `datasource/mongo` | `modules/datasource/mongo/` | § 1.1 | Low |
| `datasource/goredis` | `modules/datasource/goredis/` | § 1.1 | Low |
| `datasource/gocql` | `modules/datasource/gocql/` | § 1.1 | Low |
| `integration/messaging/` (EIP) | **`modules/messaging/`** | § 1.3 | Low |
| `messaging/rabbitmq/amqp` | `modules/messaging/rabbitmq/amqp/` | § 1.3 | Low |
| `messaging/rabbitmq/streams` | `modules/messaging/rabbitmq/streams/` | § 1.3 | Low |
| `health/` (runtime + endpoint) | **`modules/health/`** | § 1.4 | Medium |

## 4.2. Discarded

| go-feather-lib | Reason |
|---|---|
| `web/` | Covered by `common/log` |
| `cache/` (top-level go-feather-lib) | Empty directory in source. If anything needs to ship in this space, design from scratch — see [YA-0079](https://github.com/guidomantilla/yarumo/issues/79) for the in-process decision |
| `messaging/kafka/`, `messaging/nats/` | Empty directories in go-feather-lib |

---


## Related docs

- `modules/common/crypto/*` engineering — closed in [milestone Phase 1 — Modules: Crypto](https://github.com/guidomantilla/yarumo/milestones/11) (30 tickets, consolidated from old phases 1–6).
- `modules/common/` (non-crypto) engineering — tracked directly in [milestone Phase 2 — Modules: Common (non-crypto)](https://github.com/guidomantilla/yarumo/milestones/8) (issues YA-0035 … YA-0080).
- `modules/managed/` engineering — tracked directly in [milestone Phase 3 — Modules: Config / Managed / Telemetry](https://github.com/guidomantilla/yarumo/milestones/9) (issues [YA-0061](https://github.com/guidomantilla/yarumo/issues/61) … [YA-0067](https://github.com/guidomantilla/yarumo/issues/67)).
- `modules/telemetry/otel/` engineering — tracked directly in [milestone Phase 3](https://github.com/guidomantilla/yarumo/milestones/9) (issues [YA-0068](https://github.com/guidomantilla/yarumo/issues/68) … [YA-0075](https://github.com/guidomantilla/yarumo/issues/75) + [YA-0081](https://github.com/guidomantilla/yarumo/issues/81), [YA-0082](https://github.com/guidomantilla/yarumo/issues/82)).
- `modules/config/` engineering — tracked directly in [milestone Phase 3](https://github.com/guidomantilla/yarumo/milestones/9) (issues [YA-0058](https://github.com/guidomantilla/yarumo/issues/58) … [YA-0060](https://github.com/guidomantilla/yarumo/issues/60)).
- `modules/compute/` engineering — tracked directly in [milestone Phase 4 — Modules: Compute](https://github.com/guidomantilla/yarumo/milestones/10) (issues [YA-0085](https://github.com/guidomantilla/yarumo/issues/85) … [YA-0092](https://github.com/guidomantilla/yarumo/issues/92)). [`ROADMAP_COMPUTE.md`](ROADMAP_COMPUTE.md) covers design context.
- [`ROADMAP_DECISIONS.md`](ROADMAP_DECISIONS.md) — `sdks/decisions/` engineering.
- [`ROADMAP_ONTOLOGIES.md`](ROADMAP_ONTOLOGIES.md) — ontology lifecycle.
- Workspace-wide standards proposals — tracked as docs decision-tickets across the Phase 2 milestone: [YA-0041](https://github.com/guidomantilla/yarumo/issues/41) (No Inline Assignments scope, merged), [YA-0098](https://github.com/guidomantilla/yarumo/issues/140) (criterion-4 rewrite, merged), [YA-0083](https://github.com/guidomantilla/yarumo/issues/83) and [YA-0084](https://github.com/guidomantilla/yarumo/issues/84) (wontfix).
- [`STRATEGY.md`](STRATEGY.md) — product strategy.
