# Spring Cloud Consul — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-consul/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: REJECT (the project as a unit) — three patterns salvaged into existing § 1 modules (`config/`, `health/`, and the NEW `modules/secrets/` whose KV-as-secrets story is owned by `spring-vault.md`).

## 1. Project summary

Spring Cloud Consul wraps HashiCorp Consul (`https://www.consul.io`) in Spring Boot starters that perform four jobs at app bootstrap:

1. **Service discovery & registration** — the app auto-registers itself with a local Consul agent on startup (`spring.application.name` → service name, `server.port` → port). Other apps resolve services through `DiscoveryClient` or a `@LoadBalanced RestTemplate` / `WebClient`.
2. **Distributed configuration** — Consul's KV store is mounted as a Spring `PropertySource` hierarchy (`config/testApp,dev/` > `config/testApp/` > `config/application,dev/` > `config/application/`). YAML, PROPERTIES, KEY_VALUE, and FILES (git2consul) formats are supported.
3. **Health checks** — the Consul agent polls `/actuator/health` (HTTP) every 10 s by default, or the app pushes TTL heartbeats. Failed checks deregister or mark the instance unhealthy. Multi-port (management vs. main) double-registration is supported.
4. **Watch / refresh** — a blocking long-poll against Consul KV publishes a Spring `RefreshEvent` when a key changes; `@RefreshScope` beans rebuild. A `ThreadPoolTaskScheduler` (pool size 1) drives the loop.

Starters: `spring-cloud-starter-consul-discovery`, `spring-cloud-starter-consul-config`, `spring-cloud-starter-consul-bus`, plus the shared `spring-cloud-consul-core`. ACL token, retry/backoff, header-based auth, custom `ApplicationStatusProvider`, and multi-datacenter routing are all single-property toggles.

The Go counterpart is `hashicorp/consul/api` (the official Go SDK). Spring's value-add over the raw SDK is the Boot auto-configuration glue: `Environment` integration, `DiscoveryClient` abstraction, `@LoadBalanced` clients, `/actuator/health` hook-up, and the `spring.config.import=consul:` lifecycle phase.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | Service auto-registration with health check | App registers itself + an HTTP / TTL health check on startup; agent deregisters on stop | Removes manual `agent.ServiceRegister` boilerplate; couples lifecycle to `managed/` + `health/` |
| 2 | KV-as-config-source | KV keys become a tiered property hierarchy (app+profile / app / shared+profile / shared) | Centralised config without git+CI redeploy; survives pod restarts |
| 3 | Long-poll watch + refresh event | Background goroutine blocks on KV index, fires refresh on change | The one feature you cannot get from env vars or one-shot viper |
| 4 | `DiscoveryClient` abstraction | Single interface over Consul / Eureka / ZooKeeper / K8s | In Spring's world, lets one app run inside or outside K8s unchanged |
| 5 | Format pluggability (YAML / PROPERTIES / FILES) | One key holds a whole YAML doc — easier to edit than 200 leaf keys | UX win for ops; `git2consul` reuses git as source of truth |
| 6 | Fail-fast vs optional import | `spring.config.import=optional:consul:` tolerates Consul outage at boot | Without this, a Consul outage takes down the whole fleet |
| 7 | ACL token plumbed once | Single property propagates auth to all KV / agent calls | Mandatory in any prod Consul cluster |

These seven features cover ~95% of day-to-day usage. Everything below is long tail.

## 3. Long-tail features (skip)

- **`spring-cloud-starter-consul-bus`** — broadcasts refresh events across the fleet via Consul events. Replaced in practice by per-instance long-poll watch (#3 above). Spring kept it for legacy reasons.
- **Hystrix / Turbine integration** — Hystrix is end-of-life; modern stack uses Resilience4j (Java) or `modules/common/resilience/` (yarumo).
- **`@LoadBalanced RestTemplate` / Feign client-side LB** — Spring-specific HTTP wrapper. Go's grpc-go resolver/balancer plugins and `http.RoundTripper` cover the same surface.
- **Catalog watch heartbeat events** — niche; only useful if you want to be told the topology changed *outside* a call path.
- **Multi-datacenter routing (`datacenters: STORES: dc-west`)** — federation-grade; ~0% of Yarumo's target deployments will federate Consul DCs.
- **Management-port-as-separate-service** — Spring quirk born from Boot Actuator on a different port; Yarumo would expose `/healthz` on the main server (or a sibling port) without registering twice.
- **Bootstrap.yml legacy path** — Spring-internal lifecycle wart; the `spring.config.import` API replaced it in Boot 2.4+.
- **`Encryptors.*` / encrypted KV values** — handled separately at the agent / Vault layer.
- **Consul Connect / service mesh** — Spring barely supports it; in Yarumo this is a deployment concern (Istio / Linkerd / Consul Connect sidecar), not a library concern.
- **Auto-generated metadata (`secure`, `zone`, `group`)** — Spring conventions that don't translate; a Yarumo wrapper would expose plain `map[string]string`.
- **`ApplicationStatusProvider` SPI** — Boot-specific health-aggregation extension; Yarumo's `common/health/` aggregator covers the underlying need.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**:

- **`modules/config/`** — one-shot bootstrap via Viper. Viper has a built-in `viper.AddRemoteProvider("consul", …)` already, so the KV-as-config story is technically achievable today; the gap is *dynamic* refresh, not the initial load. A backend-agnostic `RefreshableSource` + `Watcher` primitive (file / Consul / etcd / S3) would be the right shape — Consul becomes one optional adapter, not a dedicated module.
- **`modules/health/` (§ 1.4)** — pairs with `modules/common/health/` (YA-0077, closed). Spring's HTTP / TTL check pattern maps cleanly: `health/` already needs to expose `/healthz` and `/readyz`; *registering* that endpoint with an external registry (Consul agent) would be an opt-in adapter inside `modules/health/`, not a core feature. Gated on a non-K8s deployment use case actually appearing.
- **`modules/secrets/` (see `spring-vault.md`)** — Vault is the primary backend for the new secrets module. Consul KV could fit as a P2 Provider alongside Vault, AWS Secrets Manager, GCP Secret Manager, Doppler — but it is a weaker fit than Vault (no native rotation, no leasing, no fine-grained ACL UX) and overlaps with `config/`. The detailed shape of `modules/secrets/` lives in `spring-vault.md`; this analysis only contributes the optional `secrets/consul-kv/` row.
- **`modules/messaging/` (§ 1.3)** — `spring-cloud-starter-consul-bus` would in principle land here, but Consul-events-as-bus is an anti-pattern. Not adopted.
- **`modules/boot/` (§ 1.5)** — the Spring auto-configuration that ties everything together has no analogue planned; the `Container` / `BeanFn` shape sketched in § 1.5 deliberately stays generic and unaware of Consul.

**Discovery**: rejected on its own merits. Kubernetes DNS plus a service mesh (Istio / Linkerd / Consul Connect sidecar) cover ~95% of the service-resolution problem for Yarumo's target deployments. A `DiscoveryClient` abstraction with a Consul implementation is feature creep that would justify itself only if a real consumer asked for Nomad / VM / bare-metal service resolution — none has. The discovery half of Spring Cloud Consul therefore lands nowhere in the Yarumo roadmap.

**Anti-patterns to avoid** (if any salvage code is written):

1. **Don't build a `modules/consul/` god-package** — Spring's auto-configuration ties registration + config + watch + bus into one starter. In Yarumo, each concern lives in its own module (`config/`, `health/`, optional `secrets/consul-kv/`). No aggregate Consul wrapper.
2. **Don't ship a `DiscoveryClient` abstraction** — see the discovery rejection above. A discovery facade is feature creep until a non-K8s consumer materialises.
3. **Don't reinvent `hashicorp/consul/api`** — the Go SDK is idiomatic and stable. Salvage adapters wrap it; they do not re-wrap it.
4. **Don't introduce a `bootstrap.yml` lifecycle phase** — Yarumo's `config/` is deliberately one-shot. Dynamic config = a separate stateful component started by `managed/`, not an earlier-lifecycle bootstrap.
5. **Don't blanket-couple `@LoadBalanced` semantics** — Go's HTTP/gRPC clients have their own resolver / balancer extension points; a Yarumo `LoadBalancedClient` would shadow that ecosystem.
6. **Don't make Consul mandatory at boot** — Spring defaulted `fail-fast: true` for years. Any Yarumo Provider must default to `optional:` semantics so a Consul outage doesn't cascade.
7. **Don't conflate KV-config with secrets** — different rotation cadence, different ACL model, different blast radius. Spring conflates them; Yarumo splits along `config/` vs `secrets/` (the latter scoped in `spring-vault.md`).
8. **Don't double-register management vs. main port** — pick one `/healthz` surface and register it once.

## 5. Recommendation

**REJECT** the Spring Cloud Consul project as a dedicated `modules/consul/` wrapper. It bundles five jobs (registration, config, health, watch, bus); for Yarumo each is either rejected (discovery, bus), already covered (`modules/config/` one-shot, `modules/common/health/` primitives), or better delivered by a backend-agnostic primitive (`config/refresh/`, `health/` self-registration adapter, `secrets/consul-kv/`).

**PARTIAL salvage** — adopt three *patterns* into existing § 1 modules, no new top-level module created:

1. **Dynamic-config-refresh pattern** → file a ticket under the `modules/config/` track. Design: `RefreshableSource` interface + a goroutine-driven `Watcher` (lifecycle-aware via `managed/`) that emits change events; backends implement long-poll (Consul KV) or periodic re-fetch (S3, etcd, file). Default backend = file watcher; Consul ships as one opt-in adapter (`modules/config/refresh/consul/`).
2. **Consul KV as a `secrets/` Provider** → add a P2 row to the secrets backend list owned by `spring-vault.md`: `secrets/consul-kv/` alongside `vault/`, `aws/`, `gcp/`, `doppler/`. Vault remains the primary story.
3. **Health-check self-registration** → add a brainstorm note inside § 1.4 `modules/health/` for an optional `modules/health/adapters/consul/` thin adapter that registers `/healthz` with a Consul agent. Gated on a real non-K8s consumer; do not implement preemptively.

The discovery half — `DiscoveryClient`, `@LoadBalanced`, multi-DC routing, catalog watch — stays **REJECTED**. The bus half — `spring-cloud-starter-consul-bus` — stays **REJECTED** (anti-pattern: Consul events are not a message bus).

## 6. Proposed yarumo placement (if applicable)

```
modules/config/
  (existing — one-shot bootstrap, unchanged)
  refresh/                        NEW (planned, Phase 3 follow-up)
    refresher.go                  RefreshableSource interface, Watcher loop
    file/                         file-watch backend (default, no external dep)
    consul/                       OPTIONAL — depends on hashicorp/consul/api
    etcd/                         OPTIONAL — future
    s3/                           OPTIONAL — future

modules/secrets/                  (scoped in spring-vault.md)
  vault/                          (primary backend)
  consul-kv/                      NEW row — P2, opt-in
  aws/, gcp/, doppler/            (already planned in spring-vault.md)

modules/health/                   (§ 1.4 — runtime side)
  adapters/
    consul/                       NEW thin adapter — register /healthz with Consul agent.
                                  Gated on non-K8s use case; otherwise do not implement.
```

Hard rules for the salvage:

- No `modules/consul/` aggregate package.
- No transitive import of `hashicorp/consul/api` from `modules/config/` core, from `modules/common/`, or from `managed/`. The SDK only enters via opt-in subpackages.
- `optional:` semantics by default — a Consul outage must not break startup.
- ACL token is a first-class option on every adapter, never a global.
- Watch loops use `context.Context` cancellation; lifecycle integrates with `managed.Lifecycle.Start/Stop/Done`.
- Adapter packages stay free of business logic — they translate between the raw Consul SDK and a Yarumo interface (`RefreshableSource`, `SecretsProvider`, `HealthRegistrar`).

## 7. Open questions

1. **Is dynamic-config-refresh a real Yarumo need yet?** DaaS (rulesets versioned by SDK) and Aluna (agent prompts versioned via `llm/prompts/`) both have natural reload stories that don't require Consul. Until a consumer asks for sub-minute config rollout on a running pod, `config/refresh/` stays speculative.
2. **Vault vs Consul KV for `modules/secrets/`** — `spring-vault.md` owns the secrets module scope. Is Consul KV worth supporting as a poor-man's secrets backend, or skip on v1 and tell users to run Vault? Lean: skip on v1.
3. **Non-K8s deployment volume** — the discovery rejection is conditional on K8s being the deployment target. Has any internal consumer (DaaS demo, Aluna self-host) credibly asked for Nomad / VM / bare-metal? If yes, the calculus on health-check self-registration changes.
4. **Service mesh stance** — Yarumo currently has no opinion on Istio / Linkerd / Consul Connect. If `modules/messaging/` or `modules/auth/` ever needs mTLS identity from a mesh sidecar (SPIFFE / SVID), that integration belongs in `auth/` or `telemetry/`, not in a Consul module.
5. **Overlap with Viper's built-in remote-config** — `modules/config/` already uses Viper, and Viper has `AddRemoteProvider("consul", …)` baked in. Does `config/refresh/consul/` reuse that, or call `hashicorp/consul/api` directly? Viper's remote support is thinly maintained — leaning toward the raw SDK.
6. **Bus rejection durability** — reconfirm under § 1.3 `modules/messaging/` that Consul-events-as-bus is not a planned adapter. Long-poll watch per instance is the canonical replacement.
7. **Long-poll vs. blocking-query etiquette** — Consul's blocking-query model maps cleanly onto a Go goroutine with `context.Context`; need to pick a default index-wait timeout that balances responsiveness against agent load. Spring defaults to 55 s wait + 1 s delay; copying those numbers is a reasonable starting point.

## 8. ROADMAP delta proposed (NOT applied)

The salvage triggers three small edits to `ROADMAP_NEW_MODULES.md` if/when the work is picked up:

1. **§ 1 (Modules)** — add `modules/config/refresh/` as a planned sub-track of the existing `modules/config/` work, with `file/` as the default backend and `consul/`, `etcd/`, `s3/` as opt-in adapters. Status: Brainstorm until a consumer asks. Owns the pattern; not a new top-level module.
2. **§ 1.4 `modules/health/`** — add a one-line note: *"Optional `modules/health/adapters/consul/` thin adapter for non-K8s deployments; gated on real use case appearing."* Status: Brainstorm.
3. **`modules/secrets/` (owned by `spring-vault.md`)** — append one row to the planned backends table: `secrets/consul-kv/`, P2 priority, "weaker than Vault but useful when Consul is already deployed and Vault isn't". Cross-reference back to this analysis.

No new milestone, no ticket filed yet. Discovery and bus produce **no roadmap delta** — they remain rejected.
