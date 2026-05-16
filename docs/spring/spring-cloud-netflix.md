# Spring Cloud Netflix — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-netflix
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: REJECT

## 1. Project summary

Spring Cloud Netflix wraps Netflix OSS components for use in Spring Boot microservices. Born in the 2014-2017 Netflix OSS era, the project has lost most of its surface area to deprecation. The Greenwich release (2018-2019) placed **Hystrix, Ribbon, Zuul 1, and Archaius 1** into maintenance mode and pointed users at modern replacements outside the Netflix lineage. The current 5.x line (5.0.1, January 2026) is effectively an **Eureka-only project**: Eureka client + embedded Eureka server, plus thin glue to `spring-cloud-loadbalancer`. Everything else either lives elsewhere in the Spring Cloud ecosystem now or is dead.

Net: this is no longer a coherent "Netflix stack" — it is a single discovery component (Eureka) carried forward, with the rest cremated and buried under other Spring projects.

## 2. Pareto features (top-20%)

Only two features still have a pulse in current Spring Cloud Netflix:

| Feature | What it is | Status |
|---|---|---|
| **Eureka client** | Registers an instance with a Eureka registry; locally caches the registry; integrates with `spring-cloud-loadbalancer` and Feign for service-to-service calls. | Active. Maintained. |
| **Eureka server** | Embedded `@EnableEurekaServer` registry, peer replication for HA, in-memory store only, no native-image support. | Active. Maintained, but limited (no persistent backend, no AOT). |

Everything Pareto-relevant about Spring Cloud Netflix in 2026 is "do you want a Eureka discovery registry or not?" The answer for Yarumo is **no**: K8s DNS + service mesh cover ~95% of the case, so a top-level `modules/discovery/` is not on the canonical roadmap.

## 3. Long-tail features (skip)

The historic Netflix OSS stack — i.e., the reason the project existed — is **maintenance-mode dead** and has well-known modern replacements that Yarumo either already covers or has explicitly rejected:

| Component | Status | Modern replacement | Yarumo coverage |
|---|---|---|---|
| **Hystrix** (circuit breaker) | Maintenance mode 2018; last release 1.5.18 (Nov 2018). Netflix itself moved to adaptive concurrency limits. | Resilience4j | Covered by `common/resilience/` (CircuitBreaker + RateLimiter registries, lazy goroutine-free). A top-level `modules/resilience/` was explicitly decided against. |
| **Ribbon** (client-side load balancer) | Maintenance mode. | Spring Cloud LoadBalancer | Out of scope. Go uses gRPC client-side LB / mesh sidecars / K8s service routing. |
| **Zuul 1** (edge gateway) | Maintenance mode. | Spring Cloud Gateway | A `modules/api-gateway/` is not on the canonical roadmap — recommended approach is composing `managed/server_http` with future auth/rate-limit middlewares instead of duplicating. |
| **Archaius 1** (dynamic config) | Maintenance mode. | Spring Boot external config + Spring Cloud Config | Covered by `modules/config/` (viper-driven). |
| **Hystrix Dashboard / Turbine** | Maintenance mode. | Micrometer + Grafana | Covered by `modules/telemetry/otel/` + OTel/Prom/Grafana stack. |
| **Eureka registry** | Active. | (itself) | Rejected — see § 4. |
| **Concurrency-limits** | Survived maintenance mode. | (itself) | Adaptive load-shedding; Yarumo has no equivalent and no demand. Out of scope. |

## 4. Mapping to Yarumo

Going feature-by-feature against the Yarumo principle map (`common/`, `managed/`, `config/`, `telemetry/`, planned domain modules):

| Spring Cloud Netflix surface | Yarumo equivalent | Decision |
|---|---|---|
| Eureka client / Eureka server (service discovery + registry) | None planned. A top-level `modules/discovery/` is not on the canonical roadmap — K8s DNS + service mesh cover ~95%. | Reject. |
| Hystrix circuit breaker | `common/resilience/CircuitBreakerRegistry` (lazy, goroutine-free). Closed via YA-0076. | Already covered — no need to revisit. |
| Ribbon client-side LB | Not needed. Go services rely on gRPC `grpc-go` LB policies, K8s Service, or mesh sidecars (Linkerd, Istio). | Reject. |
| Zuul edge gateway | A top-level `modules/api-gateway/` is not on the canonical roadmap. Compose `managed/server_http` + future rate-limit middleware + `modules/auth/` (§ 1.2). | Reject. |
| Archaius dynamic config | `modules/config/` (viper-driven one-shot bootstrap). Dynamic refresh would land in a proposed NEW `modules/featureflags/` (not on the canonical roadmap). | Already covered. |
| Hystrix Dashboard / Turbine | `modules/telemetry/otel/` + downstream Grafana. | Already covered. |
| Concurrency-limits (adaptive throttling) | None. No current consumer demand. Could be a follow-up to `common/resilience/`. | Defer / no immediate action. |

There is **zero net-new functionality** Spring Cloud Netflix offers that Yarumo would want to adopt. Every live component maps either to (a) something Yarumo already shipped, (b) something Yarumo explicitly rejected, or (c) infrastructure (K8s, mesh) that lives below the library layer.

## 5. Recommendation

**REJECT.**

Rationale:

1. **The Netflix stack is mostly dead.** Hystrix, Ribbon, Zuul, Archaius — the four pillars that gave the project its name — have been in maintenance mode since 2018. The project is now Eureka with a thin client.
2. **Eureka itself is out of scope for Yarumo.** A top-level `modules/discovery/` is not on the canonical roadmap: K8s DNS + service mesh cover the realistic deployment targets. Yarumo's consumers (DaaS, Aluna) are K8s-native; building a Eureka client adds a parallel, weaker discovery path with no upside.
3. **All useful surrounding patterns already mapped elsewhere.** Circuit breaker → `common/resilience/` (done). Config → `modules/config/` (done). Observability → `modules/telemetry/otel/` (done). Gateway / load balancer → infra layer, not library. There is nothing left to salvage.
4. **The anti-patterns Yarumo already rejects** — annotation-driven bootstrap (`@EnableEurekaClient`, `@EnableEurekaServer`), DI-coupled auto-configuration, Spring-only filter chain — are exactly what Spring Cloud Netflix doubles down on. Even the salvageable Eureka client would require a substantial Go re-implementation to fit Yarumo's "no DI, no annotations, explicit lifecycle" model, and the deliverable would still solve a non-problem.
5. **Resilience4j-style ideas (adaptive concurrency limits) are interesting but unrelated to Spring Cloud Netflix.** If Yarumo wants them, they extend `common/resilience/` directly — see issue #165 (Retrier + revisit). No need to read Spring Cloud Netflix docs to design that.

## 6. Proposed yarumo placement (if applicable)

None. Nothing to file, nothing to plan, nothing to migrate.

The only adjacent breadcrumb worth keeping in mind is **adaptive concurrency limits** (Netflix's own post-Hystrix direction, also seen in resilience4j's `Bulkhead` / `AdaptiveBulkhead`). If a Yarumo consumer ever reports real pain from static circuit-breaker thresholds, that capability would land inside `common/resilience/` as an additional primitive — **not** a new module, **not** an import of anything Netflix-branded.

## 7. Open questions

1. **Is there any non-K8s deployment scenario for DaaS / Aluna that would re-open the discovery question?** Current answer: no. If one appears, propose a new `modules/discovery/` and evaluate Consul/Nacos before Eureka — Eureka has no persistent backend and no native-image support, both of which are dealbreakers today.
2. **Should `common/resilience/` get an adaptive / token-budget circuit breaker variant?** Tracked by intent on issue #165 (revisit + Retrier). Independent of this analysis; not blocked by Netflix-OSS docs.
3. **Concurrency-limits library** — Netflix's surviving non-maintenance component. Worth a separate one-pager **only if** an internal service demonstrates that static rate limits are causing real outages. No current demand → no action.

## 8. ROADMAP delta proposed (NOT applied)

None. The verdict is REJECT across the board: every live Spring Cloud Netflix component either already maps to existing yarumo coverage (`common/resilience/`, `modules/config/`, `modules/telemetry/otel/`) or sits below the library layer (K8s DNS, service mesh). No new modules, no new sub-packages, no annex entries. If a future non-K8s deployment ever revives the discovery question, a fresh proposal for `modules/discovery/` would evaluate Consul / Nacos before Eureka.
