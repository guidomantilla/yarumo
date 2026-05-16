# Spring Cloud Commons — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-commons/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: REJECT (with one PARTIAL carve-out for `ContextRefresher` patterns — see [spring-cloud-config.md](spring-cloud-config.md))

## 1. Project summary

Spring Cloud Commons is the **shared-abstractions tier** for the entire Spring Cloud
ecosystem — the layer that sits beneath Spring Cloud Netflix (Eureka), Spring Cloud
Config (server/client), Spring Cloud Consul, Spring Cloud Zookeeper, Spring Cloud
Gateway, Spring Cloud Bus, etc. It defines the interfaces that those projects implement
and ships baseline machinery on top of them.

What it ships:

1. **Service discovery abstraction** — `DiscoveryClient` / `ReactiveDiscoveryClient`,
   `ServiceInstance`, `SimpleDiscoveryClient` (property-file fallback),
   `DiscoveryClientHealthIndicator`, `CompositeDiscoveryClient` with ordering,
   `@EnableDiscoveryClient(autoRegister=…)`.
2. **Service registry abstraction** — `ServiceRegistry`, `Registration`,
   `InstancePreRegisteredEvent` / `InstanceRegisteredEvent`, `/serviceregistry`
   actuator endpoint to flip UP / DOWN / OUT_OF_SERVICE.
3. **Spring Cloud LoadBalancer** — client-side LB replacing Netflix Ribbon. Core
   `ReactiveLoadBalancer<T>` + `ServiceInstanceListSupplier`. Built-in strategies:
   round-robin (default), random, weighted (metadata-driven), zone-affinity, periodic
   health-check, same-instance stickiness, request-cookie stickiness, hint-based,
   API-version-based, deterministic subsetting. `@LoadBalanced` on `RestTemplate` /
   `RestClient` / `WebClient.Builder` swaps virtual hostnames for real instances.
   Caffeine/default caching of instance lists. `LoadBalancerLifecycle` hooks.
   Micrometer stats (`loadbalancer.requests.{active,success,failed,discard}`).
   Per-client config via `@LoadBalancerClient` + dedicated `@Configuration` classes.
4. **Bootstrap context** — parent `ApplicationContext` loaded from `bootstrap.yml`
   *before* the main context, used by `spring-cloud-config-client` to fetch external
   config and decrypt `{cipher}*` properties. Replaced in Spring Boot 2.4+ by
   `spring.config.import=…`, but still supported.
5. **Refresh scope** — `@RefreshScope` lazy-proxy beans that get re-instantiated when
   `ContextRefresher.refresh()` runs (triggered by `POST /actuator/refresh`,
   `EnvironmentChangeEvent`, or `RefreshScopeRefreshedEvent`). Re-binds
   `@ConfigurationProperties`, updates `logging.level.*` levels live. Companion
   actuator endpoints: `/actuator/restart`, `/pause`, `/resume`, `POST /env`.
   Carve-outs documented (`HikariDataSource` is `never-refreshable` by default;
   refresh is disabled under Spring AOT / native image).
6. **EnvironmentPostProcessor / PropertySourceLocator** — extension points for plugging
   custom config sources into the bootstrap or main environment. As of Spring Cloud
   2022.0.3+, `PropertySourceLocator` is invoked twice — once without active profiles
   (so it can activate them) and once with — enabling profile-specific sources.
7. **Local encryption / decryption** — RSA-keyed `{cipher}*` property decoding so that
   YAML/properties files can hold ciphertext that's decrypted at bootstrap time.

In short, Spring Cloud Commons is the **glue for "Spring app that wants to play in a
service-mesh-less, Spring-Cloud-Config-driven, discovery-server-based topology"**. It
is opinionated about a deployment model (registry + LB + config server + bus) that
predates Kubernetes-native service mesh by ~5 years.

## 2. Pareto features (top-20%)

These are the features that would matter for *some* Go microservices fleet, ranked by
generality. Most still don't apply to yarumo's intended consumers — see § 4.

| # | Feature | Description | Why it matters (or doesn't) for Go microservices |
|---|---|---|---|
| 1 | `RefreshScope` + `ContextRefresher` pattern | Lazy proxies that re-instantiate on config change; rebind `@ConfigurationProperties`; update log levels live | **Generally useful**: change log level, rotate API keys, swap rate limits without restart. Pattern translates to Go cleanly via `atomic.Pointer[T]` + callback subscribers. Cross-references the parallel re-analysis in [spring-cloud-config.md](spring-cloud-config.md) where the same primitive is the carve-out. |
| 2 | `EnvironmentChangeEvent` / refresh listeners | Pub/sub for "config changed, here are the changed keys" | Decouples "I detected a change" from "I need to re-read property X". Reusable across viper/etcd/Consul/file-watch backends. Same scope as item 1. |
| 3 | `@LoadBalanced` RestTemplate (virtual hostname → real instance) | Apps write `http://stores/api` and the LB resolves at call time | **Marginal in K8s** — `http://stores` already resolves via cluster DNS to a `Service` IP that kube-proxy load-balances. Only useful outside K8s (bare-metal, VMs, ECS without service connect). |
| 4 | LoadBalancer health-check strategy | Periodic `GET /actuator/health` on each candidate, drop unhealthy | **Marginal in K8s** — readiness probes already gate endpoints. Useful only if the registry is faster than the K8s control plane (rare). |
| 5 | LoadBalancer zone-affinity / hint-based / weighted routing | Prefer same-AZ, route by header hint, weight by metadata | **Marginal** — Istio / Linkerd `DestinationRule` and `VirtualService` do this declaratively without app coupling. |
| 6 | LoadBalancer retry integration | "Retry on next instance" with status-code filter | Real value, but yarumo plans this at the `modules/common/http/` level (Retry-After, Idempotency-Key — [YA-0042](https://github.com/guidomantilla/yarumo/issues/42)) — without the LB-tier wrapper. |
| 7 | `ServiceRegistry` events (`InstancePreRegisteredEvent`, `InstanceRegisteredEvent`) | Lifecycle hooks around self-registration | Only relevant if you self-register. Not applicable in K8s where the orchestrator owns lifecycle. |
| 8 | `SimpleDiscoveryClient` (property-file fallback) | Hard-coded URI list for dev / no-registry environments | Trivially replaced by env vars / config — no abstraction needed. |
| 9 | Bootstrap context separation | Two-phase config: external first, app config second | Marketed feature, but Spring itself migrated away from it (Boot 2.4 `spring.config.import`). Yarumo's `modules/config/` is already two-phase by convention (env → file → secrets manager) without a separate context. |
| 10 | `EnvironmentPostProcessor` / `PropertySourceLocator` SPI | Inject custom property sources before binding | Translates to "register a viper backend before `Default()` returns". Yarumo's `modules/config/` already has this shape — viper backends are wired by the consumer. |

## 3. Long-tail features (skip)

- **Encryption / decryption of `{cipher}*` properties** — encryption belongs in the
  secrets backend (Vault, KMS), not in the config-loading library. A future
  `modules/secrets/` (when filed) would integrate Vault / AWS Secrets Manager / GCP
  Secret Manager and never put ciphertext in the YAML file at all.
- **`/actuator/serviceregistry` POST to flip UP/DOWN/OUT_OF_SERVICE** — manual
  deregistration knob; in K8s you drain via readiness probe. Niche.
- **`/actuator/restart`** — restart the `ApplicationContext` in place. Go has no
  equivalent concept; restart is process-level (systemd / K8s / supervisord).
- **`/actuator/pause` / `/resume`** — `ApplicationContext.stop()` / `.start()`. The
  closest Go analog is `managed.Lifecycle.Stop()` / `.Start()` and we already have it.
- **`@LoadBalanced` `WebClient` reactive variant** — Go has no reactive stack.
- **`DeferredSecurityContext` / Reactor context propagation** — Spring-specific.
- **Caffeine-cache of instance lists** — implementation detail, only matters if we
  adopt the discovery + LB tier (we don't).
- **Sticky-session via cookie (`sc-lb-instance-id`)** — server-affinity at the client
  layer; misfit for stateless services + better solved at the gateway / mesh.
- **API-version-based instance selection** — header-routed canaries are an Istio /
  Argo Rollouts concern.
- **Deterministic subsetting** (`spring.cloud.loadbalancer.subset.size=…`) — Google
  SRE technique for >1000-instance fleets; irrelevant at yarumo's intended scale and
  outside K8s territory.
- **`@LoadBalancerClient` annotation + per-client `@Configuration` classes** —
  Spring DI ceremony; would translate to options structs in Go and lose all elegance.
- **Bootstrap context** as a parent `ApplicationContext` — pure Spring container
  mechanics. Even Spring Boot deprecated it as the primary path.
- **Network-interface filtering** (`spring.cloud.inetutils.ignoredInterfaces`,
  `preferredNetworks`) — only matters when the app self-publishes its IP to a
  registry. Not our model.
- **`LoadBalancerLifecycle` hooks + Micrometer LB-tier counters** — only meaningful
  once a client-side LB tier exists. Without one, OTel HTTP-client instrumentation in
  `modules/telemetry/otel/` already captures the same RED metrics one layer down.
- **`spring.cloud.refresh.never-refreshable` / `extra-refreshable` lists** — concrete
  workarounds (`HikariDataSource` can't be refreshed; immutable beans need opt-in)
  that exist precisely because `@RefreshScope` proxying is leaky. A thin
  `atomic.Pointer[T]` swap doesn't need this scaffolding.

## 4. Mapping to Yarumo

### Existing § 1 modules with overlap

| Spring Cloud Commons feature | Yarumo equivalent / decision |
|---|---|
| `RefreshScope` + `ContextRefresher` + `EnvironmentChangeEvent` | **No yarumo equivalent today.** `modules/config/` (viper-based) is one-shot — no live refresh, no listeners. The cross-cutting design ("atomic pointer + callback fan-out") is the same primitive flagged as the carve-out in [spring-cloud-config.md](spring-cloud-config.md); both analyses converge on the same hypothetical `modules/config/refresh/`. |
| `EnvironmentPostProcessor` / `PropertySourceLocator` | Partially covered by viper's backend abstraction inside `modules/config/`. No formal SPI yet — viper backends are wired by the consumer at `Default()` time. |
| Bootstrap encryption (`{cipher}*`) | Out of scope. Belongs in a future secrets module (Vault / AWS Secrets Manager / GCP Secret Manager) and should never appear inline in YAML. |
| `BootstrapContext` (parent context for config-server bootstrap) | Not applicable. Go has no parent-context container model. Yarumo's bootstrap is `modules/config/` → `modules/boot/` (§ 1.5 of `ROADMAP_NEW_MODULES.md`) wiring, single-phase. |
| LoadBalancer retry on next instance | `modules/common/http/` Phase-2 features ([YA-0042](https://github.com/guidomantilla/yarumo/issues/42)): Retry-After header, Idempotency-Key, circuit-breaker hook. Different layer — outbound HTTP client, not an LB-tier wrapper. |
| LoadBalancer health-check strategy | `modules/common/health/` ([YA-0077](https://github.com/guidomantilla/yarumo/issues/77), closed 2026-05-13) + planned `modules/health/` (§ 1.4 of `ROADMAP_NEW_MODULES.md`) for the runtime side. We expose `/healthz`; we don't dial peers. |
| `InstanceRegisteredEvent` / service-registry actuator | Not applicable — orchestrator-owned lifecycle. |

### Discovery / LoadBalancer — rejected on their own merits

- **Service discovery (`DiscoveryClient` / Eureka / Consul / Zookeeper)** — Kubernetes
  cluster DNS + `Service` objects cover ~95% of the use case yarumo would target.
  Outside K8s, service mesh (Istio, Linkerd, Consul-as-mesh) covers the rest. No
  yarumo consumer (DaaS, Aluna, Socratic, KnowledgeForge) targets a non-K8s, no-mesh
  topology. Building and maintaining a Go-side `DiscoveryClient` abstraction for the
  remaining 5% — bare-metal, edge devices, ECS without service-connect — would be a
  long-term maintenance liability with no current consumer.
- **Client-side load balancing (`@LoadBalanced` + Spring Cloud LoadBalancer)** —
  corollary. The whole point of `@LoadBalanced RestTemplate` is to resolve a virtual
  hostname through a discovery client. If we don't ship discovery, we don't need
  client-side LB. Routing concerns (weights, zone affinity, canary by header) belong
  in the mesh (`VirtualService` / `DestinationRule`), not in the application binary.
- **API gateway tier** — same reasoning. K8s `Ingress` + mesh sidecar covers the
  cross-cutting routing / rate-limiting / auth-edge story. If a consumer ever needs
  in-app composition (auth + ratelimit + routing in one binary), it composes
  `modules/managed/server_http` + `modules/common/resilience/` + `modules/auth/` —
  no dedicated gateway module needed.

### Anti-patterns to avoid

1. **Don't ship a `modules/discovery/` abstraction** — K8s + mesh covers 95%; the
   remaining 5% (bare-metal, non-K8s edge) doesn't justify a module yarumo would have
   to maintain forever.
2. **Don't ship client-side load balancing** — same reason. `@LoadBalanced
   RestTemplate` only made sense in 2015 because Spring shipped `RestTemplate` and the
   industry didn't have service meshes yet. We have `net/http` + service mesh.
3. **Don't replicate the Bootstrap-vs-Main context split** — Spring itself moved away
   from it in Boot 2.4 (`spring.config.import` superseded `bootstrap.yml`). Yarumo's
   single-phase config (`modules/config/` → `modules/boot/`) is already the right
   shape.
4. **Don't ship per-client Spring-style `@LoadBalancerClient`-per-service
   configuration** — translates to a viper hellscape of
   `loadbalancer.clients.<name>.<knob>` keys. If we ever do client-side LB, do it via
   a tiny `Pool` interface, not config-driven dispatch.
5. **Don't entangle config refresh with discovery / LB** — Spring Cloud bundles them
   under the same umbrella; in yarumo they should remain orthogonal. Refresh is a
   `modules/config/` concern; LB is a mesh concern.
6. **Don't expose `/actuator/restart` semantics** — restarting an `ApplicationContext`
   in-process is a Spring artifact, not a Go pattern. Process supervision (systemd /
   K8s / docker restart) is the answer.
7. **Don't wrap `@RefreshScope` semantics in a proxy hierarchy** — the Spring
   implementation needs `never-refreshable` / `extra-refreshable` carve-outs precisely
   because lazy-proxy refresh is leaky. A pointer-swap primitive (see § 6) doesn't.

## 5. Recommendation

**REJECT** as a project to adopt.

Of the seven major surfaces Spring Cloud Commons covers:

- **Discovery** → rejected on its own merits (K8s DNS + service mesh cover ~95%;
  no consumer asking for the remaining 5%).
- **Service registry** → corollary of discovery rejection (orchestrator owns
  lifecycle in our deployment model).
- **Load balancer + `@LoadBalanced`** → rejected on its own merits (mesh territory;
  K8s `Service` + Istio `VirtualService` already do this declaratively).
- **Bootstrap context** → architectural mismatch (no parent-context model in Go);
  Spring itself deprecated it as the primary path.
- **Service-registry actuator endpoint** → corollary of discovery rejection.
- **`EnvironmentPostProcessor`** → already covered by viper backends in
  `modules/config/`.
- **`RefreshScope` + `ContextRefresher`** → only genuinely interesting piece.
  No consumer demand. **PARTIAL / DEFER** as a `Brainstorm` row that
  [spring-cloud-config.md](spring-cloud-config.md) tracks; both analyses point at the
  same hypothetical primitive.

**Net**: 6 of 7 surfaces are either architecturally inappropriate or rejected on
their own merits. The remaining one (live refresh) is a small pattern, not a project
to adopt.

The honest read: **Spring Cloud Commons is the foundation of the pre-Kubernetes
microservices stack**. K8s + service mesh + cloud-native config (External Secrets,
Reloader, `kubectl rollout restart`) collapsed most of it into the platform.
Re-implementing it in Go in 2026 would be re-fighting a battle the industry has
already moved past.

## 6. Proposed yarumo placement (PARTIAL carve-out)

Only one item is worth keeping on the radar, and it lives logically under
`modules/config/`, not as a discovery / LB stand-in:

### `modules/config/refresh/` — live configuration refresh (Brainstorm)

- **Inspiration**: Spring Cloud Commons `ContextRefresher` + `@RefreshScope` +
  `EnvironmentChangeEvent`.
- **Status**: Brainstorm — not yet ticketed. Same primitive [spring-cloud-config.md](spring-cloud-config.md)
  identifies as its carve-out. File once if a real consumer materialises; do **not**
  open parallel tickets in both docs.
- **Pain**: log-level changes, API-key rotation, per-tenant flags currently require a
  redeploy. Real but not yet articulated by DaaS / Aluna.
- **Design sketch**:
  - `Refreshable[T]` — `atomic.Pointer[T]` wrapper, `Get()` / `Set()`. No proxy.
  - `Refresher` interface — `Refresh(ctx) (changedKeys, error)` + `OnChange(fn)`.
  - Trigger adapters: HTTP endpoint (`POST /admin/refresh`), SIGHUP, file-watch,
    Consul / etcd watch.
  - Bind layer: same `mapstructure` decode that `modules/config/` uses today,
    applied to a `Refreshable[T]` on change.
- **Promotion trigger**: a consumer (DaaS or Aluna) files a concrete ticket needing
  zero-downtime config change for a non-secret value. (Secrets go through a future
  `modules/secrets/` when filed — its rotation hooks should piggyback on the same
  `Refresher` interface rather than reinventing polling.)
- **Anti-pattern guardrail**: do **not** rebuild the bootstrap-context dichotomy; do
  **not** wrap viper in a `RefreshScope` proxy hierarchy. Keep the pattern thin: a
  pointer swap + a callback list. The proof that this works is that Spring needs
  `never-refreshable` / `extra-refreshable` carve-outs precisely because its proxy
  model leaks; ours wouldn't.

Everything else in Spring Cloud Commons stays out of yarumo.

## 7. Open questions

1. **Does any planned consumer actually need live refresh?** DaaS and Aluna both
   restart cleanly. If neither articulates a concrete need within the next two phases,
   `config/refresh/` stays a `Brainstorm` row indefinitely and is not filed.
2. **If we ever target non-K8s deployments** (edge devices, on-prem bare-metal),
   does the rejection of discovery still hold? Service mesh assumes K8s; without it,
   some form of discovery becomes necessary. Worth re-evaluating only if a product
   line targeting bare-metal materialises.
3. **Should `modules/common/http/` retry-on-error pick up the LoadBalancer concept of
   "different host per retry"?** Today it retries the same URL. Useful only if we have
   multiple hosts to fail over to — which requires a registry, which we don't have.
   Likely irrelevant.
4. **Hot log-level adjustment** (`logging.level.*` via `/actuator/refresh`) — a
   subset of the refresh pattern that's standalone-useful and easy to ship. Worth
   filing on its own against `modules/common/log/slog/` even before the broader
   refresh story? Probably yes if it stays under ~200 LOC.
5. **`modules/secrets/` rotation hooks** — when filed, should they piggyback on the
   `Refresher` interface above, or own their own polling story? Cleaner to share the
   primitive if both materialise; premature to design either in isolation.

## 8. ROADMAP delta proposed (NOT applied)

No changes to `ROADMAP_NEW_MODULES.md` proposed at this time. Rationale:

- The carve-out (`modules/config/refresh/`) is a `Brainstorm` item with no consumer
  demand. Filing it would add noise to a doc whose stated principle is *"scopes new
  modules / tools that no milestone covers yet; once a track absorbs an item it
  leaves this doc"* (cf. doc preamble). Brainstorm items live there only when a
  concrete promotion trigger is realistic.
- The discovery / LB / gateway rejections are now stated inline in this doc rather
  than referenced against a deleted § 3 brainstorm in the roadmap. The roadmap's § 1
  module list is the source of truth for what *is* planned; this doc records what was
  considered and rejected for the Spring Cloud Commons surface specifically.
- If the carve-out is ever promoted, it would slot under `modules/config/` (Phase 3,
  milestone #9) — most likely as a follow-up ticket against `modules/config/` rather
  than a new top-level module. Cross-link with [spring-cloud-config.md](spring-cloud-config.md)
  at promotion time to avoid duplicate tickets.

---
