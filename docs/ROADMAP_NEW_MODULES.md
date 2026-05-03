# Roadmap — New Modules & Tools

Engineering work planned for modules that do not yet exist and for new tools. Each item has a placement decision and is annotated with status. Bigger workspace context — placement principle, status legend, related docs — at the bottom of this file.

> **Crypto and `common/` polish are tracked directly via tickets** — see the open issues at [`github.com/guidomantilla/yarumo/issues`](https://github.com/guidomantilla/yarumo/issues), grouped by milestone (Phase 0–6 for crypto, Phase 7 for non-crypto common, Phase 8 for config/managed/telemetry).

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

## 1.1. `modules/boot/` — Application Wiring

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
- **LDAP authentication provider** lives here. The LDAP-as-directory data source (read users/groups/orgs) lives in `modules/datasource/ldap/` — see § 1.3.

Reference: [Annex B](#annex-b--spring-security-feature-catalog) at the bottom of this doc captures the full Spring Security feature space that informed the scope.

## 1.3. `modules/datasource/` — DB and cache adapters

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
- **Row-level audit hooks** — `CreatedBy` / `LastModifiedBy` / `CreatedAt` / `UpdatedAt` columns auto-populated via `BeforeSave` / `BeforeUpdate` hooks (Spring Data Auditing equivalent). Implemented in `modules/datasource/gorm/`. Different layer from the **event-level** audit trail in `modules/audit/` (§ 3.1) — the two cooperate.
- **Data-lifecycle policies** — retention, archival, soft / hard delete (GDPR-aware). Per-table policy declared as struct tags or programmatic config; background sweeper enforces. Cross-references `best-practices/.../data-design/lifecycle.md`. Decision pending: implement in the gorm driver only, or factor into a `datasource/lifecycle/` sub-package.

## 1.4. `modules/messaging/` — EIP layer + brokers

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

## 1.5. `modules/health/` — runtime + HTTP endpoint

**Status**: Planned
**Pairs with**: `modules/common/health/` primitives ([YA-0077](https://github.com/guidomantilla/yarumo/issues/77)).

The runtime side of health checks: registers checkers, aggregates status, runs probes on a schedule, exposes an HTTP endpoint (`/healthz`, `/readyz`). Lives here because it has goroutine-driven lifecycle and HTTP integration.

Pulled from go-feather-lib's `health/`:

- Memory stats, uptime, goroutine count.
- HTTP handler for health endpoints.
- `Shutdown()` integrated with `managed.Lifecycle`.

## 1.6. `modules/resilience/` — only if needed

**Status**: Decision pending
**Trigger**: open this module **only if `common/resilience/` cannot stay goroutine-free** ([YA-0076](https://github.com/guidomantilla/yarumo/issues/76)). If the lazy-time-check design works, this module is unnecessary.

Would house: circuit breakers with active half-open recovery workers, rate limiters with token-bucket refilling goroutines.

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

# 3. Domain modules — Spring-inspired (brainstorm)

> **Status across this whole section: Brainstorm.** Ideas surfaced from cross-referencing Spring's domain modules with the [`best-practices/backend-engineering/`](https://github.com/guidomantilla/best-practices/tree/main/backend-engineering) guide and yarumo's two real consumers (DaaS, Aluna). None has a real use case yet — they move to **Planned** in [§ 1](#1-new-modules) when concrete demand materializes.
>
> Filter applied:
> 1. Real pain that every consumer reimplements.
> 2. Strategic fit with DaaS and/or Aluna.
> 3. Mature Go library exists to wrap (yarumo is an opinionated wrapper, not a re-implementation).

## 3.1. Tier 1 — high strategic value

### `modules/jobs/` — Background jobs with persistence

- **Inspiration**: Sidekiq (Ruby), Spring Batch (job/step subset).
- **Pain**: "run this later, survive restarts, retry with backoff". Every app reimplements it.
- **Opinionated default**: Postgres-backed via `riverqueue/river`. Redis variant later.
- **Used by**: DaaS async decisioning, Aluna long-running agent tasks.

### `modules/idempotency/` — Server-side Idempotency-Key handling

- **Inspiration**: Stripe API, AWS request signing.
- **Pain**: state-mutating APIs must dedupe retries to avoid double-charge / double-decision.
- **Design**: middleware reads `Idempotency-Key`, looks up store, returns cached response or executes + stores.
- **Opinionated default**: 24 h TTL, Redis store, mandatory key on POST/PUT/PATCH.

### `modules/outbox/` — Transactional Outbox

- **Inspiration**: Spring Modulith outbox, eShop on Containers.
- **Pain**: writing to DB + publishing event atomically. Without it, events leak or duplicate vs DB state.
- **Design**: `outbox_events` table written in the same tx as the business change, a separate worker reads → publishes → marks sent. Reuses `modules/messaging/` (§ 1.4) + `modules/datasource/gorm/` (§ 1.3).
- **Opinionated default**: at-least-once delivery, dedup on the consumer side.

### `modules/tenancy/` — Multi-tenancy

- **Inspiration**: Rails ActsAsTenant, Django-tenants.
- **Pain**: DaaS and Aluna are inherently multi-tenant. Without this every feature reinvents scoping by tenant, leak prevention, tenant attributes in logs/traces.
- **Design**: `TenantContext` propagated via `context.Context`, HTTP/gRPC middleware extracts it from the JWT, automatic GORM scope, tenant attribute injected in logs/traces (coordinates with [YA-0045](https://github.com/guidomantilla/yarumo/issues/45)).
- **Opinionated default**: single-DB with `tenant_id` discriminator. Schema-per-tenant and DB-per-tenant are out of scope for v1.

### `modules/audit/` — Audit trail

- **Inspiration**: Spring Data Auditing, Rails PaperTrail.
- **Pain**: compliance (SOC 2, HIPAA, finance) demands who-did-what-when. Apps improvise with loggers and lose half the data.
- **Design**: append-only `audit_events(actor, action, entity_type, entity_id, before, after, at, tenant_id, request_id)`. Repository decorator emits events. **Different layer** from the row-level audit columns in `modules/datasource/gorm/` (§ 1.3) — the two cooperate.
- **Opinionated default**: append-only, optional hash-chained tamper-evidence as a follow-up.

### `modules/llm/` — LLM provider abstraction

- **Inspiration**: Spring AI (chat, embeddings, function calling, streaming).
- **Pain**: Aluna is an agent platform. Without a provider-agnostic client, every feature reaches for `anthropic-sdk-go` / `openai-go` / `aws-sdk-go-v2/service/bedrockruntime` directly and gets locked in.
- **Design**: `Client` interface with `Chat(ctx, msgs, opts) (Response, error)`, `Embed(ctx, texts) ([]Vector, error)`, `Stream(ctx, msgs, opts) iter.Seq[Chunk]`, `Tools(...)` for function calling. Impls: `anthropic/`, `openai/`, `bedrock/`, `vertex/`, `ollama/` (local dev).
- **Placement decision**: **module**, not SDK companion. The provider abstraction itself is reusable beyond decisions; higher-level patterns (prompt management, RAG, chains, agents) live in `sdks/decisions/companions/llm/` and Aluna.
- **Coordinates with**: `modules/datasource/vector/` (§ 1.3) for RAG storage.

**Sub-modules planned within `llm/`** (each addresses pain that every Spring AI app encounters and every Go app reimplements):

- **`llm/memory/`** — Chat memory / conversation history. Spring AI `ChatMemory` equivalent. `Add(ctx, conversationID, message)` + `Get(ctx, conversationID, opts) []Message`. Impls `inmemory/`, `redis/`, `postgres/`. Truncation strategies: last-N, token-window, summarize-on-overflow. Bloquea cualquier escena conversacional en Aluna.
- **`llm/cache/`** — Semantic cache. Different from `common/cache` (exact-match) and `common/http` (HTTP cache headers): keys by embedding similarity. Indexes `(query_embedding, response)` pairs, retrieves on cosine similarity > threshold. Depends on `modules/datasource/vector/` (§ 1.3). Saves tokens dramatically in chat use cases.
- **`llm/guardrails/`** — Input/output validation pipeline. Pre-guards (PII detector, prompt-injection scanner) + post-guards (toxicity, schema, hallucination). Mirrors Guardrails AI / NeMo / Rebuff catalog from `ai-engineering/secure-coding/`. Each `Guard` is a small interface; pipeline configurable. Required for any agent that runs without human approval.
- **`llm/prompts/`** — Versioned prompt templates. Spring AI `PromptTemplate` equivalent. Prompts as resources (filesystem + frontmatter or DB), `text/template` interpolation with typed variables, `prompts.Get("rag.qa", "v2").Render(vars)`. Hook for A/B testing via `modules/featureflags/` (§ 3.2). Allows ML / product to iterate without Go redeploys.
- **`llm/parsers/`** — Structured output parsers. Spring AI `OutputParser` equivalent. `Parser[T]` interface with impls `json[T]/`, `tagged/` (XML-tagged extraction), `regex/`. Pipeline: parse → validate → retry-with-error-feedback (1–2 attempts) → return error. `ai-engineering/README.md` documents this exact pattern under "Validation layer post-model".

**Instrumentation** (lives outside `llm/` so consumers can opt in):

- **`modules/telemetry/otel/genai/`** (planned, will be filed when `modules/llm/` is promoted from brainstorm to ticket) — Implements [OTel GenAI semantic conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/): `gen_ai.system`, `gen_ai.request.model`, `gen_ai.usage.input_tokens`, `gen_ai.usage.output_tokens`, etc. Adds histograms (tokens-per-request, latency-by-model) and a counter (cost-per-tenant if `modules/tenancy/` (§ 3.1) is active). `WithGenAIInstrumentation()` option on `llm.Client` opts in.

**Rejected adjacent ideas** (kept for traceability; see § 3.3 table for the full list): image generation, audio (TTS/STT), moderation as a separate module, fine-tuning infra, model registry.

## 3.2. Tier 2 — real value

### `modules/secrets/` — Secrets manager abstraction

- **Inspiration**: Spring Cloud Config + Spring Vault.
- **Pain**: env-var secrets don't scale — need rotation, access audit, per-service scoping.
- **Design**: `Provider` interface with impls `vault/`, `aws/secretsmanager/`, `gcp/secretmanager/`, `doppler/`. Cached with TTL.
- **Opinionated default**: 5 min cache TTL, refresh on miss.

### `modules/featureflags/` — Feature flag client

- **Inspiration**: Spring Cloud Config refresh, LaunchDarkly.
- **Pain**: gradual rollouts, kill switches, A/B tests. Without flags, releases are all-or-nothing.
- **Design**: `Provider` interface with impls `growthbook/`, `unleash/`, `flagsmith/`, plus `static/` (file-based) for self-hosted simplicity.
- **Synergy with DaaS**: ruleset versioning can ride this mechanism.

### `modules/ratelimit/` — Server-side rate limiting

- **Inspiration**: Spring Cloud Gateway, express-rate-limit.
- **Pain**: public APIs need per-user / per-API-key / per-IP limits. **Different** from the outbound `common/http` limiter ([YA-0042](https://github.com/guidomantilla/yarumo/issues/42)).
- **Design**: HTTP/gRPC middleware, Redis-backed store for distributed deployments (token bucket or sliding window). Configurable per route.

### `modules/webhooks/` — Outbound webhook delivery

- **Inspiration**: Stripe webhooks, Svix.
- **Pain**: async client notifications — signed, retried with backoff, replay-protected, dead-lettered after N failures.
- **Design**: persistent delivery table, worker picks N at a time, HMAC-signed with per-endpoint secret, exponential backoff, auto-generated idempotency key (coordinates with `modules/idempotency/` in § 3.1).
- **Synergy with DaaS**: async decision-result delivery.

### `modules/objectstore/` — File / object storage

- **Inspiration**: Spring Cloud AWS S3.
- **Pain**: every app needs file upload/download, signed URLs, lifecycle management.
- **Design**: `Store` interface with impls `s3/`, `gcs/`, `minio/`, `local/` (dev). Operations: `Put`, `Get`, `Delete`, `SignedURL`, `Multipart`.
- **Synergy with Aluna**: agent file storage.

### `modules/sessions/` — Server-side sessions

- **Inspiration**: Spring Session, gorilla/sessions.
- **Pain**: web servers that aren't pure APIs need server-side state (not JWT). Cookie → session id → store.
- **Design**: `Store` interface with impls `redis/`, `postgres/`, `inmemory/`. Helpers for secure cookie defaults (`Secure`, `HttpOnly`, `SameSite`).

### `modules/testing/` — Test scaffolding

- **Inspiration**: Spring Boot Test + Spring Boot Testcontainers + Spring Cloud Contract.
- **Pain**: every project re-sets up testcontainers for Postgres / Redis / RabbitMQ / Kafka with the same boilerplate; fixtures are loaded ad hoc; HTTP and gRPC fakes are reinvented per repo; LLM-driven features have no standard eval harness.
- **Design** (subpackages):
  - `testing/containers/` — `testcontainers-go` wrappers per driver (Postgres, MySQL, Redis, RabbitMQ, Kafka, MinIO) returning `(*Container, DSN, error)` with auto-cleanup via `t.Cleanup`.
  - `testing/fixtures/` — YAML/JSON → DB loader (idempotent, transactional).
  - `testing/fakes/` — opinionated fakes for HTTP (`httptest` extensions) and gRPC (`bufconn` helpers).
  - `testing/contracts/` — Pact-go wrappers for consumer-driven contracts (`best-practices/backend-engineering/contract-design/`).
  - `testing/llm/` — LLM eval framework. Spring has no direct equivalent; `ai-engineering/testing/` documents the category (AI-as-judge, golden datasets, adversarial). `Eval` with `Run(prompt, inputs) [][]Result` + `Score` interface (impls `aijudge/`, `exact/`, `regex/`, `semantic/`). Metrics: pass rate, score percentiles, regression vs. baseline. JUnit-friendly output.
- **Differentiated from stdlib `testing`**: no test runner — scaffolding only.

## 3.3. Considered and rejected

The following were on the long list and explicitly removed.

| Idea | Why not |
|---|---|
| `modules/i18n/` | `golang.org/x/text` covers ~90% — wrap adds little |
| `modules/csrf/` | `gorilla/csrf` is thin enough that wrapping doesn't pay |
| `modules/batch/` | Overlaps with `jobs/`; ETL-specific isn't a yarumo target use case |
| `modules/statemachine/` | `compute/engine/fsm` already exists; an engine on top would duplicate |
| `modules/scheduler/` (distributed) | Overlaps with `jobs/` — jobs can cron-schedule themselves |
| `modules/cache/` (distributed) | Use `datasource/goredis` directly, or see [YA-0079](https://github.com/guidomantilla/yarumo/issues/79) for the in-process decision |
| `modules/graphql/` | `gqlgen` generates it all — wrapping doesn't pay |
| `modules/clients/` (OpenFeign-style) | Go has no decent dynamic proxy; without it the abstraction is ugly |
| `modules/saga/` | Real pattern (see `best-practices/backend-engineering/system-design/integration-level.md`) but no v1 DaaS use case |
| `modules/notifications/` | SaaS APIs (SendGrid, Twilio, Postmark) are thin — only worth it with >2 providers to abstract |
| `modules/search/` | Until DaaS / Aluna need full-text search, not a priority |
| `modules/sse/` | Small surface — useful but defers until LLM streaming arrives in Aluna |
| `modules/apidocs/` | `swaggo/swag` and `kin-openapi` cover it well |
| `modules/events/` (standalone) | **Merged into `modules/messaging/events/`** (§ 1.4) — same machinery, simpler façade for nominal-typed pub/sub |
| `modules/admin/` (Spring Boot Admin dashboard) | OTel + Grafana cover the signals; a custom dashboard would re-invent Grafana badly |
| `modules/shell/` (Spring Shell / REPL) | Niche; `cobra` covers admin CLIs; no current demand |
| `modules/api-gateway/` (Spring Cloud Gateway) | Compose `managed/server_http` + `ratelimit/` + `auth/` instead of duplicating |
| `modules/discovery/` (Eureka / Consul) | K8s DNS + service mesh cover ~95%; only justified for non-K8s deployments |
| `modules/saml/` | Enterprise SSO; niche unless DaaS targets SAML-IdP enterprises |
| `modules/cloud-function/` (FaaS abstraction) | Lambda / Cloud Run APIs are simple enough that wrapping adds little |
| `modules/web-services/` (SOAP) | Niche legacy. Point library on demand, no module |
| `modules/data-rest/` (auto-CRUD endpoints, Spring Data REST) | Anti-pattern of exposing the data model directly — every service should define its contract |
| `modules/mobile/` | Niche, mobile-specific |
| `modules/pulsar/` | Yet another messaging driver. Add to `messaging/` only when demand exists, not preemptively |
| `tools/modulith/` (architectural constraints) | `go-arch-lint` / `archtest` already do this via config; wrapping has no value |
| **Kafka Streams equivalent** (Spring Kafka Streams) | Re-implementation enormous. Go has no good base (goka is the best, limited). yarumo adds no value wrapping something half-done. |
| **Pipeline orchestration** (Airflow / Dagster / Prefect / Spring Cloud Data Flow) | Complete product, not library. Out of scope. |
| **Data quality framework** (Great Expectations / dbt tests equivalent) | Overlaps with `common/validation` (planned). Row-level: use validators. Pipeline-level: GE / Soda are tools, not Go libs to wrap. |
| **Lineage emission** (OpenLineage producer) | Just three events (job-start / running / end). Sub-feature of `modules/jobs/` (§ 3.1), not a module. |
| **File formats** (Parquet / Avro / ORC readers) | No concrete demand from DaaS or Aluna. `xitongsys/parquet-go` direct if it appears. |
| **Lakehouse table formats** (Delta Lake / Iceberg) | Go libraries nascent. Premature. |
| **Fine-tuning infrastructure** (Spring AI fine-tune) | Out of scope — yarumo uses models, doesn't train them. |
| **Model registry** (MLflow, W&B) | Out of scope — yarumo is not an ML platform. |
| **Image / Audio generation & transcription** (Spring AI Image, Audio) | No real use case in DaaS / Aluna. If it appears, additional sub-package of `llm/`, not a separate module. |
| **Moderation API** (Spring AI Moderation) | Covered by `llm/guardrails/` (§ 3.1). |

---

# 4. go-feather-lib migration tracking

Status of the migration from go-feather-lib into yarumo, filtered to items that land **outside** `modules/common/`. The `modules/common/` rows are tracked directly via tickets in [milestone Phase 7](https://github.com/guidomantilla/yarumo/milestones/8) (YA-0035 … YA-0080).

## 4.1. Pending

| go-feather-lib | New placement | Section in this doc | Priority |
|---|---|---|---|
| `boot/` (wiring + DI) | **`modules/boot/`** | § 1.1 | High |
| `security/AuthenticationService`, `AuthorizationFilter`, `PrincipalManager` | **`modules/auth/`** | § 1.2 | Medium |
| `datasource/gorm` | `modules/datasource/gorm/` | § 1.3 | Medium |
| `datasource/mongo` | `modules/datasource/mongo/` | § 1.3 | Low |
| `datasource/goredis` | `modules/datasource/goredis/` | § 1.3 | Low |
| `datasource/gocql` | `modules/datasource/gocql/` | § 1.3 | Low |
| `integration/messaging/` (EIP) | **`modules/messaging/`** | § 1.4 | Low |
| `messaging/rabbitmq/amqp` | `modules/messaging/rabbitmq/amqp/` | § 1.4 | Low |
| `messaging/rabbitmq/streams` | `modules/messaging/rabbitmq/streams/` | § 1.4 | Low |
| `health/` (runtime + endpoint) | **`modules/health/`** | § 1.5 | Medium |

## 4.2. Discarded

| go-feather-lib | Reason |
|---|---|
| `web/` | Covered by `common/log` |
| `cache/` (top-level go-feather-lib) | Empty directory in source. If anything needs to ship in this space, design from scratch — see [YA-0079](https://github.com/guidomantilla/yarumo/issues/79) for the in-process decision |
| `messaging/kafka/`, `messaging/nats/` | Empty directories in go-feather-lib |

---

# Annex A — Spring Messaging / Integration reference

Background analysis of Spring Messaging and Spring Integration. Used to inform the design of `modules/messaging/` (§ 1.4). **Reference material** — not engineering work; kept here so the design rationale is colocated.

> **Boot/wiring/DI note**: Spring's auto-wiring is **discarded** for yarumo. Everything explicit.

## A.1. Spring Messaging — base layer

Three abstractions, nothing more:

| Abstraction | What it is |
|---|---|
| **`Message<T>`** | Immutable envelope: generic payload + headers (`map[string]any`) |
| **`MessageChannel`** | The pipe. Two flavors: **Subscribable** (push) and **Pollable** (pull) |
| **`MessageHandler`** | `func(Message) error` — consumes a message |

Built with a copy-on-write builder; messages are never mutated. Pure library, no magic, no DI, no annotations.

### Message

```
Message<T> {
    Payload()  T
    Headers()  MessageHeaders
}
```

- `MessageHeaders` is an immutable `map[string]any`.
- Built-in headers: `ID` (UUID auto), `TIMESTAMP`, `REPLY_CHANNEL`, `ERROR_CHANNEL`, `CONTENT_TYPE`.
- Immutability = thread-safe by design.
- Two concrete implementations: `GenericMessage<T>` and `ErrorMessage`.
- Built via `MessageBuilder` (fluent, copy-on-write from existing messages).

### MessageChannel

```
MessageChannel {
    Send(msg Message, timeout Duration) bool
}
```

Two sub-interfaces define consumption semantics:

- **SubscribableChannel** (push) — subscribers register a `MessageHandler`; the channel pushes messages.
- **PollableChannel** (pull) — adds `Receive()` and `Receive(timeout)`. Messages buffered in a queue.

The fundamental dichotomy: **push vs pull**.

### MessageHandler

```
MessageHandler {
    HandleMessage(msg Message) error
}
```

Contract for any component that processes messages. Functional interface.

## A.2. Spring Integration — EIP on top of Messaging

Implements **Enterprise Integration Patterns** (Hohpe/Woolf) as components connected via channels. Adds `MessageEndpoint` as a third first-class citizen.

Architecture: **Pipes and Filters**.

```
[Endpoint] --Channel--> [Endpoint] --Channel--> [Endpoint]
  (filter)    (pipe)     (filter)     (pipe)     (filter)
```

- **Pipes** = `MessageChannel` implementations.
- **Filters** = `MessageEndpoint` implementations (wrap a `MessageHandler` or `MessageSource`).

### Concrete channels

| Channel | Type | Threading | Buffering | Behavior |
|---|---|---|---|---|
| **DirectChannel** | Point-to-point, subscribable | Sender's thread (synchronous) | None | Default. Supports transactions. Load-balances across subscribers. |
| **QueueChannel** | Point-to-point, pollable | Async (blocking receive) | FIFO in-memory | Decouples sender/receiver timing. Capped capacity. |
| **PublishSubscribeChannel** | Broadcast, subscribable | Sender's thread | None | Every subscriber receives every message. |
| **PriorityChannel** | Point-to-point, pollable | Async | Priority queue | Ordered by `priority` header or custom `Comparator`. |
| **RendezvousChannel** | Point-to-point, pollable | Synchronous handoff | Zero-capacity (`SynchronousQueue`) | Sender blocks until receiver calls `Receive()`. |
| **ExecutorChannel** | Point-to-point, subscribable | Async (via executor) | None | Like DirectChannel but the handler runs in another thread. |
| **PartitionedChannel** | Point-to-point, subscribable | Async (per-partition) | None | Same partition key → same thread. Preserves order per key. |

**DirectChannel** is the default and most common — synchronous, transactional, simple.

### Endpoints (EIP components)

All endpoints are `MessageHandler` implementations connected to input/output channels:

| Endpoint | What it does |
|---|---|
| **Transformer** | Converts payload or structure. Can modify headers. |
| **Filter** | Boolean predicate. Passes matching messages, drops the rest. |
| **Router** | Inspects message and decides which channel(s) to send to. Dynamic dispatch. |
| **Splitter** | 1 message → N messages (fragments composite payloads). |
| **Aggregator** | N messages → 1. Correlation + release strategy. Stateful. |
| **Service Activator** | Invokes business logic (plain object). Extracts payload, calls method, wraps result as a message. |

### External-system connections

**Channel Adapter** — one-way (fire-and-forget):

- **Inbound**: external → channel. Converts external data to `Message`. Polling or event-driven.
- **Outbound**: channel → external. Consumes messages and writes to the external system. No reply.

**Gateway** — two-way (request/reply):

- **Inbound**: external system calls, awaits reply (e.g. HTTP request → pipeline → HTTP response).
- **Outbound**: the flow calls an external system and awaits the reply (e.g. HTTP request → response → message).

**Messaging Gateway (proxy)**: generates a proxy of an interface. Calling a method transparently converts arguments to a `Message`, sends it to the request channel, awaits the reply, returns the result. The caller knows nothing about messaging.

### Supported adapters (30+ protocols)

- **Messaging**: AMQP (RabbitMQ), JMS, Kafka, MQTT, STOMP, ZeroMQ
- **HTTP/Web**: HTTP, WebFlux, Web Services (SOAP), RSocket, GraphQL, WebSockets
- **File systems**: File (local), FTP/FTPS, SFTP, SMB
- **Databases**: JDBC, JPA, MongoDB, R2DBC, Cassandra, Redis
- **Other**: Mail (SMTP/IMAP/POP3), Feed (RSS/Atom), Syslog, TCP/UDP, JMX, Hazelcast, Debezium (CDC)

### IntegrationFlow DSL (composition)

```
from(source) → filter → transform → route → handle(sink)
```

DSL methods: `.channel()`, `.filter()`, `.transform()`, `.route()`, `.split()`, `.aggregate()`, `.handle()`, `.gateway()`, `.log()`, `.wireTap()`, `.intercept()`. Supports sub-flows for routing branches and error handling.

## A.3. Design principles (without Spring's magic)

1. **Loose coupling via channels** — producer and consumer don't know each other.
2. **Immutable messages** — no shared mutable state.
3. **Pipes & filters** — simple components chained.
4. **Transport independence** — same logic with Kafka, RabbitMQ, HTTP, or files.
5. **Business logic in plain objects** — the framework passes payloads, not `Message`.

---

# Annex B — Spring Security feature catalog

Feature-by-feature breakdown of the 6 relevant Spring Security modules (6.x / 7.0). Used to inform `modules/auth/` (§ 1.2) — **reference material**, not a TODO list. Yarumo's `auth` module will pick a subset based on real demand.

> **Boot/wiring/DI note**: discarded. Everything explicit.
> **OAuth2-* and authorization-server**: interesting but too far for now.

## B.1. CRYPTO

### Password Encoding

| # | Feature | Description |
|---|---|---|
| 1.1 | **PasswordEncoder interface** | Contract: `Encode`, `Matches`, `UpgradeEncoding` |
| 1.2 | **DelegatingPasswordEncoder** | Routes by `{id}` prefix in the hash (e.g. `{bcrypt}$2a$...`) — transparent migration |
| 1.3 | **BCrypt** | Hash with random salt, configurable strength (4-31) |
| 1.4 | **Argon2** | Memory-hard, GPU/ASIC-resistant |
| 1.5 | **SCrypt** | Memory-hard + CPU-hard |
| 1.6 | **PBKDF2** | Key-derivation with configurable iterations |
| 1.7 | **MD5/SHA/MD4/LdapSha** | Legacy encoders (insecure, for compatibility only) |
| 1.8 | **NoOpPasswordEncoder** | Plaintext pass-through (testing) |
| 1.9 | **Password4j encoders** | Balloon Hashing and variants via Password4j |

### Encryption

| # | Feature | Description |
|---|---|---|
| 1.10 | **BytesEncryptor / TextEncryptor** | Symmetric encryption interfaces |
| 1.11 | **AES-GCM (Encryptors.stronger)** | 256-bit authenticated encryption |
| 1.12 | **AES-CBC (Encryptors.standard)** | 256-bit non-authenticated (less secure) |
| 1.13 | **BouncyCastle AES** | AES-GCM/CBC via BouncyCastle provider |
| 1.14 | **RSA encryption** | Raw RSA + hybrid (RSA wraps a symmetric key) |
| 1.15 | **KeyStore integration** | Load keys from Java KeyStore files |

### Key Generation

| # | Feature | Description |
|---|---|---|
| 1.16 | **BytesKeyGenerator** | Random bytes via SecureRandom |
| 1.17 | **StringKeyGenerator** | Hex-encoded or Base64-encoded random strings |
| 1.18 | **SharedKeyGenerator** | Static (deterministic) key |

### Utilities

| # | Feature | Description |
|---|---|---|
| 1.19 | **Hex / Utf8 codecs** | Encoding/decoding utilities |

## B.2. LDAP

| # | Feature | Description |
|---|---|---|
| 2.1 | **Bind authentication** | Auth via LDAP bind with the user's credentials |
| 2.2 | **Password comparison** | Auth comparing stored hash |
| 2.3 | **Active Directory provider** | Specialized provider for AD (`user@domain`, AD error codes) |
| 2.4 | **User search / DN resolution** | Search by filter or DN pattern |
| 2.5 | **Authority population from groups** | Map LDAP groups to `GrantedAuthority` |
| 2.6 | **Nested group resolution** | Recursive resolution of nested groups |
| 2.7 | **UserDetails integration** | LdapUserDetails, LdapUserDetailsService, CRUD via LdapUserDetailsManager |
| 2.8 | **Person / InetOrgPerson mapping** | Map LDAP entries to domain objects |
| 2.9 | **Password policy support** | LDAP password-policy controls (expiration, lockout, grace logins) |
| 2.10 | **Embedded LDAP server** | UnboundID in-memory for dev/testing with LDIF loading |
| 2.11 | **Connection management** | `ContextSource` with pool, manager DN, auth source |

## B.3. CORE

### Authentication model

| # | Feature | Description |
|---|---|---|
| 3.1 | **Authentication interface** | Representation of request/principal authenticated |
| 3.2 | **AuthenticationManager** | Central interface `authenticate(Authentication)` |
| 3.3 | **ProviderManager** | Chain of `AuthenticationProvider`s with parent fallback |
| 3.4 | **AuthenticationProvider** | Contract for a specific auth mechanism |
| 3.5 | **UsernamePasswordAuthenticationToken** | Token for username/password |
| 3.6 | **AnonymousAuthenticationToken/Provider** | Anonymous auth |
| 3.7 | **RememberMeAuthenticationToken/Provider** | Remember-me auth |
| 3.8 | **AuthenticationManagerResolver** | Dynamic resolver (multi-tenant) |
| 3.9 | **CredentialsContainer** | Erase sensitive credentials post-auth |
| 3.10 | **One-Time Token auth** | Single-use tokens (magic links): generate, consume, JDBC/in-memory storage |
| 3.11 | **DaoAuthenticationProvider** | Auth against `UserDetailsService` + `PasswordEncoder` |

### UserDetails model

| # | Feature | Description |
|---|---|---|
| 3.12 | **UserDetails interface** | username, password, authorities, account status flags |
| 3.13 | **UserDetailsService** | Load `UserDetails` by username |
| 3.14 | **UserDetailsPasswordService** | Update stored password (encoder migration) |
| 3.15 | **UserDetailsChecker** | Validate account state (locked, disabled, expired) |
| 3.16 | **UserCache / CachingUserDetailsService** | Cache of `UserDetails` |
| 3.17 | **Reactive variants** | `ReactiveUserDetailsService`, `MapReactiveUserDetailsService` |

### User provisioning

| # | Feature | Description |
|---|---|---|
| 3.18 | **UserDetailsManager** | CRUD: create/update/delete user, change password, exists |
| 3.19 | **InMemoryUserDetailsManager** | In-memory implementation |
| 3.20 | **JdbcUserDetailsManager** | JDBC implementation (users, authorities, groups tables) |
| 3.21 | **GroupManager** | Group management with authorities |

### Authorization model

| # | Feature | Description |
|---|---|---|
| 3.22 | **AuthorizationManager interface** | Central authorization contract |
| 3.23 | **AuthorizationDecision / AuthorizationResult** | Decision result |
| 3.24 | **AuthorityAuthorizationManager** | Check individual authorities |
| 3.25 | **AuthenticatedAuthorizationManager** | Check auth level (full, remember-me, anonymous) |
| 3.26 | **AuthorizationManagers.allOf/anyOf** | AND/OR composition of managers |
| 3.27 | **AuthorizationProxyFactory** | Proxies that enforce auth on returned-object methods |
| 3.28 | **Multi-Factor Authorization** | `AllRequiredFactorsAuthorizationManager`, `RequiredFactor` |
| 3.29 | **Observability** | `ObservationAuthorizationManager` with Micrometer |

### Method security interceptors

| # | Feature | Description |
|---|---|---|
| 3.30 | **@PreAuthorize / PreAuthorizeAuthorizationManager** | SpEL pre-execution |
| 3.31 | **@PostAuthorize / PostAuthorizeAuthorizationManager** | SpEL post-execution |
| 3.32 | **@PreFilter / @PostFilter** | Filter input/output collections |
| 3.33 | **@Secured / SecuredAuthorizationManager** | Simple roles |
| 3.34 | **JSR-250 @RolesAllowed** | `Jsr250AuthorizationManager` |
| 3.35 | **@AuthorizeReturnObject** | Security on returned-object fields |
| 3.36 | **Authorization denied handlers** | Custom handling (null, throw, reflective fallback) |
| 3.37 | **Reactive method interceptors** | Reactive variants of all interceptors |

### SecurityContext

| # | Feature | Description |
|---|---|---|
| 3.38 | **SecurityContextHolder** | Static accessor for the current context |
| 3.39 | **ThreadLocal strategy** | Per-thread storage (default) |
| 3.40 | **InheritableThreadLocal strategy** | Propagates to child threads |
| 3.41 | **Global strategy** | Single JVM-wide context |
| 3.42 | **DeferredSecurityContext** | Lazy context loading |
| 3.43 | **SecurityContextChangedEvent** | Context-change events |
| 3.44 | **Reactor context propagation** | Project Reactor integration |
| 3.45 | **ReactiveSecurityContextHolder** | Reactive variant |

### GrantedAuthority & Role Hierarchy

| # | Feature | Description |
|---|---|---|
| 3.46 | **GrantedAuthority / SimpleGrantedAuthority** | Permissions/roles |
| 3.47 | **RoleHierarchy** | Role inheritance (ADMIN > USER) |

### Events

| # | Feature | Description |
|---|---|---|
| 3.48 | **Authentication events** | Success, failure (bad credentials, locked, disabled, expired, …) |
| 3.49 | **Authorization events** | Success/failure of authorization |
| 3.50 | **DefaultAuthenticationEventPublisher** | Exception → event mapping + publish |

### RunAs / Exceptions / PermissionEvaluator / Concurrency

| # | Feature | Description |
|---|---|---|
| 3.51 | **RunAsManager** | Temporary privilege escalation during method invocation |
| 3.52 | **AuthenticationException hierarchy** | BadCredentials, AccountExpired, Locked, Disabled, CredentialsExpired, InsufficientAuth, … |
| 3.53 | **AccessDeniedException** | Authorization failure |
| 3.54 | **PermissionEvaluator** | Object-level permissions for `hasPermission()` |
| 3.55 | **DelegatingSecurityContext wrappers** | `Runnable`, `Callable`, `Executor`, `ExecutorService`, `ScheduledExecutorService` |

## B.4. WEB

### Filter chain

| # | Feature | Description |
|---|---|---|
| 4.1 | **SecurityFilterChain / FilterChainProxy** | Filter-chain architecture |
| 4.2 | **Multiple filter chains** | Multiple chains with distinct matchers |
| 4.3 | **Filter ordering** | Canonical filter order |

### Authentication mechanisms

| # | Feature | Description |
|---|---|---|
| 4.4 | **Form login** | `UsernamePasswordAuthenticationFilter` + login page |
| 4.5 | **HTTP Basic** | `BasicAuthenticationFilter` + WWW-Authenticate challenge |
| 4.6 | **HTTP Digest** | `DigestAuthenticationFilter` (nonce-based) |
| 4.7 | **X.509 certificate** | Client cert auth from SSL handshake |
| 4.8 | **Pre-authenticated (header-based)** | External auth via headers (SiteMinder, request attributes) |
| 4.9 | **WebAuthn / Passkeys** | FIDO2 registration + authentication endpoints |
| 4.10 | **One-Time Token login** | Magic-link endpoints (generate + submit) |
| 4.11 | **Anonymous authentication filter** | Anonymous token when no auth present |
| 4.12 | **Switch User (SU)** | User impersonation by admins |
| 4.13 | **Default login/logout page generation** | Auto-generated HTML |

### Session management

| # | Feature | Description |
|---|---|---|
| 4.14 | **Session fixation protection** | `changeSessionId`, `migrateSession`, `newSession` |
| 4.15 | **Concurrent session control** | Limit sessions per user |
| 4.16 | **Session creation policies** | `IF_REQUIRED`, `NEVER`, `STATELESS`, `ALWAYS` |
| 4.17 | **HttpSessionEventPublisher** | Session events to the event system |

### CSRF / CORS / Headers / Logout / Other

| # | Feature | Description |
|---|---|---|
| 4.18-4.21 | **CSRF** | `CsrfFilter` + token, HttpSession/Cookie repos, BREACH protection, SPA mode |
| 4.22 | **CORS integration** | `CorsConfigurationSource` per path |
| 4.23-4.33 | **Security headers** | Cache-Control, X-Content-Type-Options, HSTS, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy, COOP/COEP/CORP, Clear-Site-Data, custom headers |
| 4.34-4.35 | **Logout** | `LogoutFilter` + handlers, `LogoutSuccessHandler` |
| 4.36 | **RequestCache** | Save/replay original request post-login |
| 4.37-4.38 | **Firewall** | `StrictHttpFirewall` (path traversal, double encoding, null bytes), `RequestRejectedHandler` |
| 4.39-4.41 | **Access denied / entry points** | `AccessDeniedHandler`, `AuthenticationEntryPoint`, success/failure handlers |
| 4.42-4.44 | **Remember-me (web)** | Token-based, persistent token (theft detection), `PersistentTokenRepository` |
| 4.45-4.50 | **Other** | Servlet API integration, HTTPS redirect, URL-based authorization, IP authorization, debug filter, JAAS integration |

## B.5. CONFIG

| # | Feature | Description |
|---|---|---|
| 5.1-5.3 | **@EnableWebSecurity / WebFlux / MethodSecurity** | Activation of security (servlet, reactive, method) |
| 5.4-5.5 | **HttpSecurity / ServerHttpSecurity DSL** | Fluent builder for filter chain |
| 5.6 | **XML namespace** | XML config (`<http>`, `<authentication-manager>`, …) |
| 5.7 | **Security SpEL expressions** | hasRole, hasAuthority, permitAll, isAuthenticated, … |
| 5.8 | **Meta-annotations** | Custom annotations wrapping `@PreAuthorize` with templated SpEL |
| 5.9 | **Custom AuthorizationManager integration** | Replace built-in managers |
| 5.10 | **Modular HttpSecurity** | Compose config from independent modules (7.0) |
| 5.11 | **WebSecurityCustomizer** | Ignore paths (static resources) |

## B.6. ASPECTS

| # | Feature | Description |
|---|---|---|
| 6.1-6.5 | **AspectJ weaving aspects** | `@PreAuthorize`, `@PostAuthorize`, `@PreFilter`, `@PostFilter`, `@Secured` |
| 6.6 | **Self-invocation support** | Security on intra-class calls (which AOP proxy cannot intercept) |
| 6.7-6.8 | **Compile-time / load-time weaving** | AspectJ compiler / Java agent at runtime |

---

## Related docs

- `modules/common/` engineering — tracked directly in [milestone Phase 7](https://github.com/guidomantilla/yarumo/milestones/8) (issues YA-0035 … YA-0080).
- `modules/managed/` engineering — tracked directly in [milestone Phase 8](https://github.com/guidomantilla/yarumo/milestones/9) (issues [YA-0061](https://github.com/guidomantilla/yarumo/issues/61) … [YA-0067](https://github.com/guidomantilla/yarumo/issues/67)).
- `modules/telemetry/otel/` engineering — tracked directly in [milestone Phase 8](https://github.com/guidomantilla/yarumo/milestones/9) (issues [YA-0068](https://github.com/guidomantilla/yarumo/issues/68) … [YA-0075](https://github.com/guidomantilla/yarumo/issues/75) + [YA-0081](https://github.com/guidomantilla/yarumo/issues/81), [YA-0082](https://github.com/guidomantilla/yarumo/issues/82)).
- [`ROADMAP_COMPUTE.md`](ROADMAP_COMPUTE.md) — `modules/compute/` engineering.
- [`ROADMAP_DECISIONS.md`](ROADMAP_DECISIONS.md) — `sdks/decisions/` engineering.
- [`ROADMAP_ONTOLOGIES.md`](ROADMAP_ONTOLOGIES.md) — ontology lifecycle.
- Workspace-wide standards proposals — tracked as docs decision-tickets in [milestone Phase 7](https://github.com/guidomantilla/yarumo/milestones/8): [YA-0041](https://github.com/guidomantilla/yarumo/issues/41) (No Inline Assignments scope), [YA-0083](https://github.com/guidomantilla/yarumo/issues/83) (required params for singletons), [YA-0084](https://github.com/guidomantilla/yarumo/issues/84) (drop receiver asserts).
- [`STRATEGY.md`](STRATEGY.md) — product strategy.
