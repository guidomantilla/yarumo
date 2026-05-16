# Spring → Yarumo Adoption Analyses

Pareto evaluation of 26 Spring projects to decide what Yarumo could absorb as a Go module. This directory is the **brainstorm catalog** that lives separate from the canonical [`docs/ROADMAP_NEW_MODULES.md`](../ROADMAP_NEW_MODULES.md). Each per-project file documents: summary, top-20% features, long-tail to skip, mapping to existing/planned modules, recommendation, proposed placement, ROADMAP delta (NOT applied).

**Original analysis**: 2026-05-16 (25 projects). **Spring AI added**: same day.
**Re-analyzed**: 2026-05-16 after ROADMAP cleanup (deletion of § 3 Brainstorm domain modules, Annex A Spring Messaging reference, Annex B Spring Security feature catalog). **Total**: 26 projects, ~6,500 lines.

## ROADMAP relationship

The canonical roadmap is lean — only what gets built. Spring/* is the catalog of **opinions** from Spring comparison. Every PARTIAL/ADOPT recommendation includes a "ROADMAP delta proposed (NOT applied)" section the user can promote or ignore.

| Source | Status |
|---|---|
| `docs/ROADMAP_NEW_MODULES.md` § 1 (planned) | `datasource/` (1.1), `auth/` (1.2), `messaging/` (1.3), `health/` (1.4), `boot/` (1.5) |
| `docs/ROADMAP_NEW_MODULES.md` § 2 (tools) | `routegen` |
| `docs/spring/*.md` (this catalog) | **Mode: PERMISIVO** — agents propose new top-level modules freely; ROADMAP does not auto-change |

## Recommendation summary

| Verdict | Count | Meaning |
|---|---|---|
| **ADOPT** | 0 | Spawn a new top-level yarumo module (high-confidence, real consumer) |
| **PARTIAL** | 17 | Adopt subset — either refines an existing § 1 module, or proposes a NEW module |
| **REJECT** | 8 | Covered, deprecated, or out-of-scope |
| **DEFER** | 1 | Useful but gated on a real consumer or upstream module |

## Per-project table

| # | Project | Verdict | Lands in | Key contribution |
|---|---|---|---|---|
| 1 | [Spring Session](spring-session.md) | PARTIAL | **NEW** `modules/sessions/` | Store interface, IndexedStore split, secure-cookie defaults, Redis + Postgres backends |
| 2 | [Spring HATEOAS](spring-hateoas.md) | REJECT | — | RFC 7807 + pagination envelope salvage into `common/http/` (~80 LOC) |
| 3 | [Spring Modulith](spring-modulith.md) | PARTIAL | **NEW** `modules/outbox/` | Event Publication Registry: 5-state lifecycle, staleness sweeper, resubmit API |
| 4 | [Spring REST Docs](spring-restdocs.md) | PARTIAL | **NEW** `modules/testing/apidocs/` | Test-driven snippet emission (conditional on testing/ being created) |
| 5 | [Spring AMQP](spring-amqp.md) | PARTIAL | `messaging/rabbitmq/amqp/` (§ 1.3) | Topology builders, retry+DLX, publisher confirms, listener container |
| 6 | [Spring Vault](spring-vault.md) | PARTIAL | **NEW** `modules/secrets/vault/` | Anchor sub-driver: `LeaseContainer`, Transit/PKI/KV-v2 engines |
| 7 | [Spring StateMachine](spring-statemachine.md) | REJECT | — | Features grow `compute/math/fsm/` + `compute/engine/states/` (Phase 4) |
| 8 | [Spring Data Redis](spring-data-redis.md) | PARTIAL | `datasource/goredis/` (§ 1.1) | `WithPipeline`, distributed lock subpackage, OTel hook |
| 9 | [Spring Data MongoDB](spring-data-mongodb.md) | PARTIAL | `datasource/mongo/` (§ 1.1) | Typed aggregation builder, change-stream `managed.Worker` |
| 10 | [Spring Data REST](spring-data-rest.md) | REJECT | — | RFC 7232 conditional requests salvage into `common/http/conditional/` |
| 11 | [Spring Cloud Bus](spring-cloud-bus.md) | DEFER | `messaging/bus/` (§ 1.3) | Gated on messaging + config.Reload() + featureflags |
| 12 | [Spring Cloud CircuitBreaker](spring-cloud-circuitbreaker.md) | PARTIAL | `common/resilience/` extension | Retrier + Bulkhead + TimeLimiter for issue #165 |
| 13 | [Spring Cloud Commons](spring-cloud-commons.md) | REJECT | — | K8s + service mesh cover ~95%; refresh primitive → spring-cloud-config.md |
| 14 | [Spring Cloud Config](spring-cloud-config.md) | PARTIAL | `config/refresh/` + **NEW** `modules/featureflags/` | Refresh primitive + profile resolution + `{cipher}` placeholder + featureflags as first consumer |
| 15 | [Spring Cloud Consul](spring-cloud-consul.md) | REJECT | — | Discovery rejected on own merits; KV/health-check salvage into config/, health/, secrets/ |
| 16 | [Spring Cloud Open Service Broker](spring-cloud-open-service-broker.md) | REJECT | — | No DaaS/Aluna consumer; OSBAPI ecosystem contracted |
| 17 | [Spring Cloud Netflix](spring-cloud-netflix.md) | REJECT | — | All components deprecated; nothing salvageable |
| 18 | [Spring Cloud OpenFeign](spring-cloud-openfeign.md) | PARTIAL | `common/rest/` (YA-0044) | `ErrorDecoderFn[E]`, pluggable encoders, `ResponseInterceptor` |
| 19 | [Spring Cloud Stream](spring-cloud-stream.md) | PARTIAL | `messaging/` (§ 1.3) | `Binding` struct (DLQ + Retry + Partition + ConsumerGroup as deployment policy) |
| 20 | [Stream Applications](stream-applications.md) | REJECT | — | Catalog of deployable apps, not a library; archived 2026-02-26 |
| 21 | [Spring Cloud Task](spring-cloud-task.md) | **PARTIAL** | **NEW** `modules/tasks/` | Finite execution lifecycle tracking (was REJECT, flipped after roadmap trim) |
| 22 | [Spring Cloud Vault](spring-cloud-vault.md) | PARTIAL | `config/sources/secrets/` + cross-ref to **NEW** `modules/secrets/` | Resolver layer: bootstrap-time secret → config-key mapping |
| 23 | [Spring gRPC](spring-grpc.md) | PARTIAL | `common/grpc/` (YA-0043) | Curated interceptor catalog (deadline, validate, ratelimit, errors) |
| 24 | [Spring Framework](spring-framework.md) **DEEP** | PARTIAL | Multi: `datasource/` + `common/http/` + **NEW** `modules/testing/` | RFC 9457 ProblemDetail, `WithTransaction`, middleware order, MockMvc fake, v7 Retry callback |
| 25 | [Spring Integration](spring-integration.md) **DEEP** | PARTIAL | `messaging/` (§ 1.3) | 13+ new files: store/, poller/, advice/, interceptors/, claimcheck/, observability/, routers/, aggregator/, etc. |
| 26 | [Spring AI](spring-ai.md) **DEEP** | PARTIAL | **NEW** `modules/llm/` | 8 sub-modules: memory/cache/guardrails/prompts/parsers + NEW tools/mcp/rag/etl + `Advisor` primitive |

## Proposed NEW top-level modules (consolidated)

After ROADMAP cleanup, every PARTIAL that no longer has a § 3 home is proposed as a NEW top-level module. These are **brainstorm**, not in ROADMAP. The user decides what (if anything) to promote.

| Proposed module | Source analysis | Anchor pattern | Real consumer |
|---|---|---|---|
| `modules/llm/` | [spring-ai.md](spring-ai.md) | ChatClient + ChatModel + 8 sub-modules + Advisor primitive | Aluna (agent platform) |
| `modules/outbox/` | [spring-modulith.md](spring-modulith.md) | Event Publication Registry, externalize sub-package | DaaS async decisioning |
| `modules/tasks/` | [spring-cloud-task.md](spring-cloud-task.md) | Finite execution lifecycle (NOT queueing/retry) | DaaS async + KnowledgeForge + CLI migrations |
| `modules/sessions/` | [spring-session.md](spring-session.md) | Store interface + Redis/Postgres backends | DaaS console (when shipped) |
| `modules/secrets/` | [spring-vault.md](spring-vault.md) + [spring-cloud-vault.md](spring-cloud-vault.md) | Provider + Vault anchor sub-driver | All apps that need rotation/transit/PKI |
| `modules/testing/` | [spring-framework.md](spring-framework.md) + [spring-restdocs.md](spring-restdocs.md) | containers/, fixtures/, fakes/, contracts/, llm/, apidocs/ | All apps |
| `modules/featureflags/` | [spring-cloud-config.md](spring-cloud-config.md) | First consumer of `config/refresh/` primitive | DaaS, Aluna |

## Existing § 1 module refinements (consolidated)

Each existing planned module gets concrete refinements from the analyses:

| Module | Refinements from |
|---|---|
| `datasource/` (§ 1.1) | spring-data-redis (goredis/), spring-data-mongodb (mongo/), spring-framework (WithTransaction core + typed error translation) |
| `auth/` (§ 1.2) | spring-session (Rotate contract for fixation), spring-vault (token-related primitives) |
| `messaging/` (§ 1.3) | spring-integration (13+ new files), spring-cloud-stream (Binding struct), spring-amqp (rabbitmq driver), spring-modulith (events relationship) |
| `health/` (§ 1.4) | spring-cloud-consul (adapters/consul/ opt-in) |
| `boot/` (§ 1.5) | No direct delta — explicit-wiring stance reinforced |
| `common/http/` (existing, YA-0042) | spring-framework (problem/ + response/), spring-hateoas (RFC 7807), spring-data-rest (conditional/ RFC 7232) |
| `common/rest/` (existing, YA-0044) | spring-cloud-openfeign (ErrorDecoderFn, encoders, ResponseInterceptor) |
| `common/grpc/` (existing, YA-0043) | spring-grpc (interceptor catalog) |
| `common/resilience/` (existing, YA-0076 shipped) | spring-cloud-circuitbreaker (Retrier + Bulkhead + TimeLimiter for #165) |
| `config/` (existing) | spring-cloud-config (refresh/, sources/, sources/cipher/), spring-cloud-vault (sources/secrets/) |
| `compute/engine/states/` (Phase 4 planned) | spring-statemachine (design checklist when promoted) |

## Cross-cutting anti-patterns (lock in)

Every analysis identified Spring-specific magic NOT to bring over. Consolidated:

1. **DI container with auto-wiring** — yarumo uses explicit `Container.Register/Resolve` in `boot/`
2. **Annotation-driven configuration** — `@Component`, `@Bean`, `@Configuration`, `@EnableXxx`, `@StreamListener`, `@RabbitListener`, `@Transactional`, `@Cacheable`, `@Scheduled`, `@Tool`, `@PreAuthorize`, etc. → Go functions
3. **AOP / AspectJ** — bytecode proxies → middleware/interceptors
4. **SpEL** — Spring Expression Language → `common/expressions/`
5. **Reactive Mono/Flux** — Spring WebFlux, R2DBC → goroutines + channels + `context.Context`
6. **God-structs / `BeanBuilder` / `ApplicationContext`** — 30+ public fields → small typed components
7. **Classpath scanning** — `@ComponentScan`, `@BindableService` autodiscovery → explicit registry
8. **Bootstrap context** — parent `ApplicationContext` (Spring itself deprecated it)
9. **Annotation-driven docs** — REST Docs is test-driven instead; principle generalizes
10. **Fixed init order** — Spring's auto-config sequence → explicit `BeanFn`s in boot/
11. **Per-provider Spring Boot starters** — consumer imports the driver they need
12. **Dynamic-proxy declarative REST clients** (OpenFeign-style) — code-gen is the Go answer

## Recommended next steps

If the user wants to promote some catalog proposals to ROADMAP, the deltas are ready to paste. Priority order based on real consumer demand:

### Tier 1 — Strategic (real consumer asking)
1. **`modules/llm/`** — Aluna can't ship agents without it. See [spring-ai.md § 9](spring-ai.md). Most strategic: `llm/mcp/` (Model Context Protocol).
2. **`modules/outbox/`** — DaaS async decisioning needs the at-least-once primitive. See [spring-modulith.md § 8](spring-modulith.md).
3. **`modules/secrets/`** — Vault as anchor sub-driver. Every app that needs rotation. See [spring-vault.md § 8](spring-vault.md).

### Tier 2 — High value, no immediate consumer
4. **`modules/sessions/`** — Gated on DaaS console UI shipping. See [spring-session.md § 8](spring-session.md).
5. **`modules/testing/`** — Foundation module that helps everything else. See [spring-framework.md § 6](spring-framework.md).
6. **`modules/tasks/`** — Recurs in 4+ planned contexts. See [spring-cloud-task.md § 8](spring-cloud-task.md).
7. **`modules/featureflags/`** — First consumer of `config/refresh/`. See [spring-cloud-config.md § 8](spring-cloud-config.md).

### Tier 3 — Refinements to existing modules (cheap to apply)
- `common/http/problem/` (RFC 9457) + `common/http/conditional/` (RFC 7232) → YA-0042 Phase 2
- `common/grpc/` interceptor catalog → YA-0043
- `common/rest/` ErrorDecoderFn + ResponseInterceptor → YA-0044
- `common/resilience/` Retrier + Bulkhead + TimeLimiter → issue #165
- `datasource/` core: `WithTransaction(ctx, db, fn)` callback + typed error translation per driver
- `messaging/` (§ 1.3 layout): add `store/`, `poller/`, `advice/`, `interceptors/`, `claimcheck/`, `observability/`, etc. (13+ files from spring-integration.md)
- `messaging/` Binding struct from spring-cloud-stream.md

## Methodology

Each subagent received:
- Full `ROADMAP_NEW_MODULES.md` for context (lean version: § 1, § 2, § 4 only — no § 3 brainstorm, no Annexes)
- A standard template (project summary, Pareto top-20%, long-tail skip, yarumo mapping, recommendation, proposed placement, open questions, ROADMAP delta NOT applied)
- A budget (150–400 lines standard; 500–1000 lines for the three DEEP analyses)
- Instructions to follow 2–3 deeper pages efficiently (Tier 1+2 batches); editorial cleanup only for Tier 3 (no WebFetch)
- **Mode: PERMISIVO** — agents free to propose new top-level modules; ROADMAP does not auto-change

Three projects (`spring-framework`, `spring-integration`, `spring-ai`) received deep-analysis instructions; the rest standard.

Per-project files are independent — each can be read alone without the others. Cross-references between sibling analyses (e.g. `spring-vault.md` ↔ `spring-cloud-vault.md`) replaced what used to live in the now-deleted Annex A / Annex B.
