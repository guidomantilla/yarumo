# Spring Cloud Config — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-config
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL — adopt the client-side patterns (refresh primitive, layered profile resolution, inline `{cipher}` decryption) inside the existing `modules/config/`. Propose a NEW top-level `modules/featureflags/` as the first downstream consumer of the refresh primitive. The secrets-provider story is handed off to the [Spring Vault analysis](spring-vault.md) and the [Spring Cloud Vault analysis](spring-cloud-vault.md) (NEW `modules/secrets/`). **Reject** the Config Server as a Yarumo deliverable.

## 1. Project summary

Spring Cloud Config (SCC) is a centralised externalised-configuration system for the Spring ecosystem. It splits cleanly into two artefacts:

- **Config Server** — a Spring Boot HTTP service that reads configuration from a backend (Git is the default; SVN, Vault, JDBC, Redis, AWS S3 / Parameter Store / Secrets Manager, CredHub, native filesystem are also supported) and exposes it over a REST API at `/{application}/{profile}/{label}`. It can encrypt/decrypt with symmetric or RSA keys, broadcast change notifications via Spring Cloud Bus, and act as a webhook target for git providers (`/monitor` endpoint). The server also supports composite backends with explicit ordering, property overrides imposed on all clients, plain-text / binary file delivery with placeholder resolution, and a health indicator that probes each configured `EnvironmentRepository`.
- **Config Client** — a Spring Boot autoconfiguration that pulls config at bootstrap, merges it with local property sources, supports retries, fail-fast, multiple Config Server URLs, and refreshes `@RefreshScope` beans on `POST /actuator/refresh` or `POST /actuator/busrefresh`. Modern clients use `spring.config.import=configserver:...` (Boot 2.4+); the legacy `bootstrap.yml` mechanism is deprecated.

Resource hierarchy resolves layered profiles in precedence order (`foo-dev-mysql.properties` > `foo-dev.properties` > `foo-mysql.properties` > `foo.properties` > `application-*.properties`), and labels select a git branch / tag / commit. Server-as-library mode (`@EnableConfigServer` + `bootstrap: true`) lets one Boot app embed the server and self-configure from its own backend — useful when the server *is* the first application in a bootstrap chain.

The whole stack assumes a Spring DI runtime: `@RefreshScope` works by proxying beans and clearing a scoped cache so the next access reconstructs them with refreshed `@Value` properties. Without a DI container, refresh has no concept of "which beans need to be rebuilt" — every consumer has to wire that themselves. This single fact is what makes the *client* interesting to Yarumo (the refresh contract is portable) and the *server* uninteresting (it solves a delivery problem that K8s ConfigMaps + Vault + Consul already solve out-of-band).

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Layered profile resolution** | `app-profile1-profile2.yml` cascading over `app-profile1.yml` over `app.yml` over `application.yml`. Profile-specific overrides shared values. | Every Go service reimplements env-based config layering (dev/stage/prod). A first-class precedence model removes one whole class of "why is this value different in prod" bugs. |
| 2 | **Client refresh pattern** (`@RefreshScope` + `/actuator/refresh`) | Identified beans get rebuilt when configuration changes; refresh endpoint mutates a `RefreshScope` cache. | Long-lived Go services need to rotate database passwords, rollover keys, toggle features without restart. The "which subsystems are refreshable" contract is the valuable part. |
| 3 | **External secrets provider abstraction** (Vault / AWS SM / GCP SM via the Config Server) | Server fronts multiple backends; clients consume one URI. | Maps onto the secrets analysis tracked in [spring-vault.md](spring-vault.md) and [spring-cloud-vault.md](spring-cloud-vault.md) — NEW `modules/secrets/`. Out of scope for `modules/config/` itself; cross-referenced here for completeness. |
| 4 | **Server-side encryption with `{cipher}` placeholders** | Values prefixed with `{cipher}...` get decrypted on serve; supports symmetric and RSA, key rotation via `{key:name}` selector. | Storing encrypted-at-rest secrets in plain git is a real pattern. Inline `{cipher}` prefix is composable; "key-id-aware values" idea is portable. |
| 5 | **Composite backends with order/priority** | Multiple backends combined, optional `failOnCompositeError`. | Real-world bootstrap pulls from file + env + Vault + git. A `Source` chain with deterministic precedence is the right model. |
| 6 | **Fail-fast + retry + multiple URIs on client** | `fail-fast`, `initial-interval`, `max-attempts`, comma-separated URI fallback. | Standard resilience patterns; aligns with `common/resilience/` already shipped (YA-0076). |
| 7 | **`overrideNone` / property override discipline** | Clients can mark overrides as defaults vs enforced; explicit precedence between operator-provided and consumer-provided. | The "is this value forced by ops or just a default" distinction is missing in most Go config stacks. Cheap to add. |
| 8 | **Webhook-driven refresh** (`/monitor` endpoint receives git push, broadcasts via Bus) | Auto-refresh on commit, no manual `curl /refresh`. | Decoupling "config changed" from "every pod restarts" is genuinely useful — but the implementation belongs in CI/CD or operator scripts in K8s land, not in a Yarumo server. |
| 9 | **Plain-text and binary file serving** (`/{app}/{profile}/{label}/{path}`) | Server can deliver `nginx.conf`, certificates, anything — with placeholder resolution. | Niche but compact; the `${placeholder}` resolution-on-fetch is portable to any backend. |
| 10 | **Health indicator wired to the config backend** | `/actuator/health` reports whether the backend is reachable. | Pairs naturally with `modules/common/health/` (YA-0077 closed). A config backend would be a `Checker`. |

## 3. Long-tail features (skip)

- **The Config Server itself** as a Yarumo deliverable. The Go ecosystem has no widely-adopted analog because K8s ConfigMaps + Secrets + projected volumes + Consul KV + Vault Agent already cover ~95% of the use case at the infrastructure layer. Re-implementing an HTTP-fronted config service in Go would be a multi-month investment for a feature that is functionally a thinner CI/CD pipeline + a Vault read.
- **Git backend, SVN backend, JDBC backend** — interesting only because they let you store config in your VCS. Valuable in Spring world where the deployment model is "fat JAR + Boot"; in K8s the same outcome is achieved by templating ConfigMaps from git via Argo / Flux / Helm.
- **Spring Cloud Bus integration for refresh broadcasting** — covered by the separate Spring Cloud Bus analysis. The bus is a generic message-broker indirection; if needed, it composes from `modules/messaging/` (§ 1.3 of `ROADMAP_NEW_MODULES.md`).
- **Discovery-first bootstrap** (Eureka / Consul lookup of Config Server) — K8s DNS solves this. Service-registry-based config-server discovery was Spring-Netflix-era infrastructure; mostly obsolete.
- **`bootstrap.yml` legacy mechanism** — replaced even within Spring by `spring.config.import` (Boot 2.4+). No reason to copy a deprecated pattern.
- **`@EnableConfigServer` library mode** — only interesting if we ship a config server.
- **Encryption-as-a-service via `/encrypt` and `/decrypt` REST endpoints** — these exist because Spring teams version-control encrypted secrets. The right place for this in Go is at the developer-tooling level (a CLI in `tools/`), not as a runtime service.
- **`{cipher}` decryption at server (vs client)** — only meaningful when a server intermediates.
- **CredHub backend** — Cloud Foundry-specific; CF is not a Yarumo deployment target.
- **AWS S3 backend** — generic object storage; not a config-specific concern.
- **HTTP Basic auth / SSH-key git auth on the server** — server-side concerns we are not implementing.

## 4. Mapping to Yarumo

### Existing § 1 modules with overlap

- **`modules/config/`** (Phase 3 active, YA-0058..YA-0060) — viper-driven one-shot bootstrap. Currently loads at startup; no refresh. **Closest analog to the Config Client.** Natural home for the refresh primitive, profile cascade types, and `{cipher}` placeholder pipeline.
- **`modules/common/resilience/`** (shipped, YA-0076) — retry / circuit-breaker primitives the refresh-on-change path will lean on.
- **`modules/common/health/`** (shipped, YA-0077) — config-backend health check fits the existing `Checker` interface cleanly (consumed downstream by `modules/secrets/` and `modules/health/`).
- **`modules/common/crypto/`** (Phase 1 closed, milestone #11) — the `{cipher}` placeholder concept maps onto a key-aware decryptor registry; primitives already exist (`crypto/encryption`, `crypto/keys`).

### Cross-references to other Spring analyses

- **NEW `modules/secrets/`** — proposed in [spring-vault.md](spring-vault.md) and elaborated in [spring-cloud-vault.md](spring-cloud-vault.md). Owns `Provider` interface with `vault/`, `aws/secretsmanager/`, `gcp/secretmanager/`, `doppler/`, `file/` impls. Out of scope for `modules/config/`. The refresh primitive proposed here is what `modules/secrets/` would use to expose rotating leases (Vault dynamic credentials) to consumers without a restart.
- **NEW `modules/featureflags/`** — proposed below in § 6 as a top-level consumer of the refresh primitive.
- **`modules/messaging/`** (§ 1.3 of `ROADMAP_NEW_MODULES.md`) — only relevant if/when bus-driven refresh broadcasting becomes a real need. Currently rejected.

### Client lifecycle comparison

The SCC client and Yarumo's `modules/config/` differ in three load-bearing ways. Understanding the deltas is what lets us cherry-pick:

| Concern | Spring Cloud Config Client | Yarumo `modules/config/` (today) |
|---|---|---|
| **Load timing** | Two-phase: bootstrap context fetches `spring.config.import=configserver:...`, then main context wires beans against the resulting `Environment` | One-shot: `config.Default()` loads files + env via viper at startup; no second phase |
| **Refresh model** | `@RefreshScope` proxies + `/actuator/refresh` endpoint + Spring Cloud Bus broadcast | None — config is immutable after `Default()` returns |
| **Backend abstraction** | Pluggable `PropertySourceLocator`; HTTP-fronted Config Server is the canonical one | viper layer (file + env + flag); no remote-provider abstraction |
| **Profile resolution** | Server-side cascade across `app-profile1-profile2`, `app-profile1`, `app`, `application` | Implicit via viper merge of multiple files; no explicit precedence rules |
| **Secret handling** | Server-side: `{cipher}` decrypted on serve, or backend (Vault) consulted | None — secrets are env vars or read by the consumer directly |
| **Resilience** | `fail-fast`, retry budget, multiple URIs, health indicator | viper has none of these because there's no remote load |

The gaps to fill inside `modules/config/` are rows 2, 4, and the local half of row 5 (`{cipher}` decryption inline in the load pipeline). Rows 3 and 6 stay **out of scope** — backend abstraction belongs in `modules/secrets/`; resilience around remote reads is the secrets provider's problem.

### Gaps to fill

1. **Refresh-on-change pattern.** Yarumo has no story for "config or feature flag changed at runtime; rebuild the affected component." Today everyone restarts the pod. The valuable design ingredient from SCC is the **scope of refresh** — not every singleton, only components that opted into a refresh contract. Mapped to Go: a `Refreshable[T]` wrapper (similar shape to `common/resilience/CircuitBreakerRegistry`) that re-derives a value from current config on demand. Lives in `modules/config/refresh/`.
2. **Layered profile resolution as a first-class API.** Viper does this loosely; an explicit `Profile` + ordered cascade with documented precedence is portable and well-tested in SCC. Lives in `modules/config/sources/`.
3. **Operator-vs-consumer override discipline.** The `overrideNone` flag's semantics ("are these values enforced or merely defaults?") is useful regardless of backend. Worth lifting into `modules/config/sources/` as `Source.Priority` or similar.
4. **Encrypted-at-rest config values with inline cipher prefix.** `{cipher}value` is composable across backends. Adding a decryption step to `modules/config/`'s load pipeline (read → resolve placeholders → decrypt prefixed values) is small and high-value. Lives in `modules/config/sources/cipher/`, backed by `modules/common/crypto/`.
5. **NEW `modules/featureflags/`** as the first non-trivial consumer of the refresh primitive. Reuses `modules/config/refresh/`; ships static / GrowthBook / Unleash / Flagsmith providers; replaces the ad-hoc "env var with a bool" idiom that Go services use today.

### Anti-patterns to avoid

- **Don't build a Config Server in Go.** The absence of a popular Go equivalent to SCC Server is a strong signal: K8s ConfigMaps + Vault + Consul KV already partition this problem. Building one would be a Yarumo product, not a Yarumo module — and even within the Spring ecosystem the Config Server is increasingly replaced by Vault Agent / external-secrets operators when teams move to K8s.
- **Don't couple refresh to a DI container.** `@RefreshScope` works because Spring proxies everything. Yarumo is explicitly anti-DI-container in [boot/](../ROADMAP_NEW_MODULES.md#15-modulesboot--application-wiring). Refresh has to be a value-shaped contract (`Refreshable[T].Get()`) not a bean-shaped one. This also means refresh granularity is *explicit* in the consumer's code — there's no "magic, this bean refreshed" mystery.
- **Don't reproduce `bootstrap.yml` / `bootstrap.properties`.** Spring itself deprecated this in 2.4. The right ingestion order is `defaults < file < env < secrets-provider` with explicit precedence — not a two-phase bootstrap. The two-phase model exists in Spring only because the DI container needs to know the Config Server URI before it can wire anything else; with no DI container, the phasing collapses to a single ordered load.
- **Don't ship server-side `/encrypt` and `/decrypt` HTTP endpoints.** This is build-time tooling; ship it as a CLI under `tools/` if/when needed, not as a runtime concern. Runtime decryption (consuming `{cipher}` placeholders) is fine; runtime encryption-as-a-service is a serious lateral-movement risk if the endpoint is ever reachable from inside the cluster.
- **Don't tie refresh to Spring Cloud Bus.** Bus is a separate analysis; coupling refresh to a specific transport (AMQP/Kafka) repeats Spring's overreach. Make refresh a local API; let consumers wire the trigger (HTTP endpoint, file-watcher, K8s ConfigMap reload via inotify, polling) at their discretion. This keeps `modules/config/` free of any messaging dependency.
- **Don't fan out to N backends in `modules/config/`.** Remote backends (Vault, AWS SM, GCP SM, Doppler) belong in NEW `modules/secrets/` — see [spring-vault.md](spring-vault.md) / [spring-cloud-vault.md](spring-cloud-vault.md). `modules/config/` stays one-shot bootstrap aware of files + env + injected secrets-provider results. SCC's "single endpoint fronting many backends" model collapses two responsibilities; the Yarumo split keeps each module narrow.
- **Don't repurpose Vault as a config store for non-secret values.** SCC encourages this because Vault is "just another backend"; in practice it inverts the cost model (Vault is rate-limited and audited). Plain config goes in files / env / ConfigMaps; secrets go in Vault / Secrets Manager. The `{cipher}` placeholder bridges the two when they need to coexist in one file.

## 5. Recommendation

**PARTIAL.** Four concrete absorptions, two explicit rejections, one cross-reference:

**Absorb (P1) — inside existing `modules/config/`**:

1. **Refresh-on-change pattern as a `Refreshable[T]` primitive** — new sub-package `modules/config/refresh/`. This is the load-bearing concept that makes `modules/featureflags/` work and is generally useful (rotating DB creds, JWT public-key sets, throttling thresholds). Issue-worthy.
2. **Layered profile resolution + override priority discipline** in `modules/config/sources/`. Today viper does layering implicitly; making `Profile`, `Source`, and `Priority` explicit types unblocks the override semantics (`overrideNone` analog). Issue-worthy.
3. **`{cipher}` inline placeholder + key registry** in `modules/config/sources/cipher/`. Decryption pipeline step that consults a key-aware decryptor (backed by `modules/common/crypto/`). Small, composable, well-scoped. Issue-worthy.

**Absorb (P2) — new top-level module**:

4. **`modules/featureflags/`** as a NEW top-level module (consumer of `modules/config/refresh/`). Ships `Flag[T]` over `Refreshable[T]`, plus providers: `static/` (in-process), `growthbook/`, `unleash/`, `flagsmith/`. Promote to Planned in `ROADMAP_NEW_MODULES.md` once a real consumer (DaaS or Aluna) surfaces; until then this analysis is the design reference.

**Cross-reference**:

5. **Secrets-provider abstraction** — fully covered by the Spring Vault / Spring Cloud Vault analyses. NEW `modules/secrets/` lives there. The connection point is that `modules/secrets/` will use `modules/config/refresh/` to expose rotating values without coupling back into `modules/config/`.

**Reject**:

- **Config Server as a Yarumo module or product.** K8s + Vault + Consul cover it. The empty Go-ecosystem space for this is a feature, not a bug.
- **Bus-driven refresh broadcasting.** Defer to the separate Spring Cloud Bus analysis; the refresh primitive itself is transport-agnostic.

## 6. Proposed Yarumo placement

```
modules/config/                          EXISTING — Phase 3 active
  refresh/                               NEW sub-package
    refreshable.go                       Refreshable[T] interface + AtomicRefreshable[T]
    registry.go                          Registry: registers refreshables, fans out triggers
    trigger.go                           Trigger interface (file watcher, manual fn, HTTP wrapper)
  sources/                               NEW sub-package — profile / priority resolution
    profile.go                           Profile type, ordered cascade
    source.go                            Source interface, Priority enum (DEFAULT, OVERRIDE, ENFORCED)
    cascade.go                           Resolve(profiles []Profile, sources []Source) → merged map
    cipher/                              NEW leaf sub-package — {cipher} placeholder pipeline
      placeholder.go                     Detect / parse {cipher}{key:name}{secret:...}value
      registry.go                        KeyRegistry: name → Decryptor (backed by common/crypto)
      decrypt.go                         Decrypt step in the load pipeline (fail-fast on missing key)
  loader.go                              EXTEND — call cipher decrypt + profile cascade

modules/featureflags/                    NEW top-level module — § 6 of this doc
  flag.go                                Flag[T] (thin alias over config/refresh.Refreshable[T])
  registry.go                            FlagRegistry — central catalogue, drift detection
  provider.go                            Provider interface (Get(key) → value, Watch(key) → trigger)
  static/                                In-process, file-backed, dev-only
  growthbook/                            GrowthBook adapter
  unleash/                               Unleash adapter
  flagsmith/                             Flagsmith adapter

(modules/secrets/ — covered by spring-vault.md / spring-cloud-vault.md)
```

**Dependency direction**:

```
modules/featureflags/  →  modules/config/refresh/  →  modules/config/
modules/secrets/       →  modules/config/refresh/  →  modules/config/
modules/config/sources/cipher/  →  modules/common/crypto/
```

No new cross-module edges to `managed/`, `boot/`, or `telemetry/`. `modules/config/` keeps no dependency on `modules/secrets/` or `modules/featureflags/` — the dependency points downward only.

**Ticket suggestions for Phase 3 milestone (#9)** — pairs with existing YA-0058..YA-0060:

- **Refresh primitive** (`modules/config/refresh/`): standalone ticket. Independent of YA-0058..YA-0060.
- **Profile + Priority types** (`modules/config/sources/`): could fold into YA-0059 / YA-0060 if their scope allows, otherwise standalone.
- **`{cipher}` decryption pipeline** (`modules/config/sources/cipher/`): standalone ticket. Depends on `common/crypto/encryption` (already shipped).

`modules/featureflags/` stays unticketed until a consumer surfaces; § 8 below lists the roadmap delta that promotes it from "designed here" to "filed under § 1".

## 7. Open questions

1. **Refresh trigger mechanism — file watcher, HTTP, both?** SCC ships an HTTP endpoint as the default; modern K8s deployments use ConfigMap projection + inotify. Yarumo's `Trigger` interface should accommodate both. Decision direction: ship file-watcher + manual-fn first; the HTTP handler variant is a thin wrapper that lives in the consumer's HTTP server (which has lifecycle — `modules/managed/server_http/`), not in `modules/config/` itself. The bus-driven broadcast variant (SCC's `/actuator/busrefresh`) stays out of scope and routes through the separate Spring Cloud Bus analysis.
2. **Is `Refreshable[T]` worth promoting to `modules/common/`?** It has no lifecycle and no external SDK deps — it's the kind of primitive that could live in `common/`. But it has mutable internal state (the cached value) and the only known consumers will be `featureflags/` and `secrets/`. Per the [PACKAGES.md classification guidance](../../modules/PACKAGES.md), mutable internal state alone does not disqualify Shape A, so `common/refresh/` is *plausible*. Defer until a third consumer outside `config/featureflags/secrets/` appears.
3. **Does `modules/config/` need to know about `modules/secrets/`, or do they compose at the consumer level?** SCC has the Config Server *be* the secrets provider, which conflates two concerns. The cleaner split for Yarumo: `modules/config/` reads non-secret values (files, env), `modules/secrets/` reads secret material (Vault, AWS SM, GCP SM), and the consumer wires both into its `Container` via separate `BeanFn`s. Vote for the clean split; the `{cipher}` inline-decryption story makes this trivial because the decryptor itself is just another `BeanFn` plug-in.
4. **Should `{cipher}` decryption block startup, or fall back to plain value?** SCC fails the request if the key is missing. For Yarumo, fail-fast at load time seems right — silent fallback to ciphertext-as-plaintext is a textbook security footgun. Worth confirming with a real use case, but the default should be fail-fast.
5. **Operator overrides — do we need them in v1?** SCC's `overrideNone` exists because Spring deployments are heterogeneous (Spring Boot Admin operators editing values at runtime, etc.). Yarumo consumers are explicit — the operator-vs-consumer distinction is usually CI/CD vs. application code, not two runtime actors. Could defer the `overrideNone` analog until a real need surfaces, but the `Priority` type itself is cheap to introduce now and forecloses future ambiguity.
6. **Health check for config backend — `modules/config/` or `modules/secrets/`?** A failing Vault connection is a secrets problem more than a config problem; `modules/config/` is one-shot bootstrap, so once it returns there's nothing to health-check. Probably `modules/secrets/` owns a `Checker` that reports backend reachability; `modules/config/` exposes no runtime health surface.
7. **Should `modules/featureflags/` ship at all if the only consumer is "in-process toggle"?** The honest answer is no — `static/` alone is a five-line struct, not a module. The case for `featureflags/` only earns its keep when at least one external provider (GrowthBook / Unleash / Flagsmith) is wired in. Until DaaS or Aluna has that need, the module stays in this doc as design-only.
8. **Webhook-driven refresh — Yarumo or out of scope?** SCC's `/monitor` endpoint receives git webhooks and broadcasts via Bus. In K8s the equivalent is "Argo / Flux syncs the ConfigMap; pod sees the file change via inotify." The cleaner placement is **out of scope for Yarumo** — the file-watcher `Trigger` covers it generically, and the orchestrator (Argo / Flux / operator) handles the git-to-K8s edge.
9. **Does `Refreshable[T]` need cancellation / context plumbing?** SCC's refresh is synchronous and blocking. For Yarumo, a `Refresh(ctx) error` shape lets long-running refreshes (e.g., fetching a 10MB key bundle, rotating Vault lease) honour deadlines. Default to `ctx`-aware API; the cost is negligible and the Go idiom is established.

## 8. ROADMAP delta proposed (NOT applied)

The following additions to `docs/ROADMAP_NEW_MODULES.md` would land *if and when* this analysis is acted on. They are listed here for reference; no change is being made to the roadmap as part of this analysis.

**Under § 1 (New modules)** — add a new sub-section:

> ### 1.6. `modules/featureflags/` — Feature-flag refresh contract
>
> **Status**: Planned (designed in [spring-cloud-config.md](spring/spring-cloud-config.md))
> **Why a new module**: depends on `modules/config/refresh/` (also new), ships external-provider SDKs (GrowthBook, Unleash, Flagsmith) with their own lifecycle. Not appropriate for `common/`.
> **Internal deps**: `modules/config/refresh/`, `modules/common/log`, `modules/common/health`.
> **Sub-packages**: `static/`, `growthbook/`, `unleash/`, `flagsmith/`.

**Under § 4.1 (go-feather-lib migration tracking — Pending)**: no change. Feature flags were never part of go-feather-lib.

**Under `modules/config/` engineering** (tracked via milestone #9 — Phase 3): three new tickets to file when work resumes:

| Proposed ticket | Scope | Depends on |
|---|---|---|
| `modules/config/refresh/` — Refreshable[T] + Registry + Trigger | sub-package | none |
| `modules/config/sources/` — Profile, Source, Priority, cascade resolver | sub-package | none |
| `modules/config/sources/cipher/` — `{cipher}` placeholder pipeline + KeyRegistry | leaf sub-package | `modules/common/crypto/encryption` (shipped) |

The `modules/featureflags/` ticket lands separately under § 1.6 above when a consumer (DaaS / Aluna) commits to using it.
