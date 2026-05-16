# Spring Cloud Open Service Broker — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-open-service-broker
> **Deep pages followed**:
> - https://www.openservicebrokerapi.org/ (OSBAPI spec home)
> - http://docs.spring.io/spring-cloud-open-service-broker/docs/current/reference/ (v5.0.0 reference)
> - http://docs.spring.io/spring-cloud-open-service-broker/docs/ (version index)
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: REJECT

---

## 1. Project summary

Spring Cloud Open Service Broker is a Spring Boot starter that implements the server side of the **Open Service Broker API (OSBAPI)** — a REST contract originally born in Cloud Foundry and later adopted by the Kubernetes Service Catalog (now archived) and various marketplace integrators. The starter exposes a fixed set of `/v2/...` endpoints and lets the developer plug in three Java interfaces:

| Interface | Responsibility |
|---|---|
| `CatalogService` | Returns the catalog of offered services + plans. |
| `ServiceInstanceService` | Create / update / delete / get an instance, plus `getLastOperation` for async polling. |
| `ServiceInstanceBindingService` | Create / delete / get a binding (credentials, route, volume) plus async polling. |

Mechanics:

- **Catalog** advertises services + plans (`bindable`, `plan_updateable`, `instances_retrievable`, `bindings_retrievable` flags).
- **Provisioning** is sync (return `async=false`) or async (return `async=true`; platform polls `GET /v2/.../last_operation` until `SUCCEEDED` / `FAILED`).
- **Bindings** carry credentials back to the platform: an app binding returns `{credentials: {url, username, password, ...}}`, a route binding returns a `routeServiceUrl`.
- **Security** is HTTP Basic on `/v2/**`, configured explicitly via Spring Security (no default).
- **Dashboard SSO** lets the broker hand back a `dashboardUrl` that platforms hook into their own OAuth2 IdP — covered superficially in the spec, not in the Spring starter.
- **Reactor everywhere**: methods return `Mono<...>`; both Spring MVC and WebFlux are supported.

Current line is **5.0.x (Feb 2026)**. The project has been on a steady cadence for ~8 years, but the OSBAPI ecosystem around it has contracted sharply — Cloud Foundry usage is declining, Kubernetes Service Catalog is **archived**, and most cloud marketplaces have moved to native catalogs (AWS Marketplace, GCP Marketplace, Azure Service Catalog) that do not speak OSBAPI.

## 2. Pareto features (top-20%)

If — and only if — a consumer ever needed to expose a yarumo-hosted product (DaaS, ontology-registry) to a Cloud-Foundry-style marketplace, the surface that would actually be used is small:

1. **Catalog endpoint** — `GET /v2/catalog` returning a static or DB-driven service/plan list.
2. **Provision / deprovision** — `PUT/DELETE /v2/service_instances/{id}` mapped to "create tenant", "drop tenant".
3. **Bind / unbind** — `PUT/DELETE /v2/service_instances/{id}/service_bindings/{id}` returning credentials.
4. **Async + last_operation polling** — anything that takes >60s (the OSBAPI sync ceiling) needs this.
5. **HTTP Basic auth on `/v2/**`** — trivially the only platform-trust mechanism.

That's ~5 endpoints, ~3 interfaces, ~12 DTOs. The full Spring starter (event-flow registries, multi-platform path variables, API version negotiation, ETag-aware catalog) is overkill unless multiple marketplaces are integrating at once.

## 3. Long-tail features (skip)

Even inside the Spring starter, most surface area is irrelevant for any plausible Go-microservices use case:

- **Reactor `Mono`/`Flux` API** — Go has no analog; would translate to plain `(T, error)` returns, throwing away half the design.
- **Spring Boot auto-configuration** — yarumo's whole philosophy (no DI, no annotations, no magic) is the antithesis.
- **Event-flow registries** (`CreateServiceInstanceEventFlowRegistry`, etc.) — pre/post/error hook system; over-engineered, replaced in Go by ordinary middleware.
- **Multi-platform path-variable extraction** — only used by brokers serving multiple CF foundations.
- **API version negotiation** (`X-Broker-API-Version` matrix) — single-tenant brokers can hard-code one version.
- **Dashboard SSO via Cloud Foundry UAA** — niche even within CF.
- **Route bindings + volume services** — extremely CF-specific (`routeServiceUrl` injection into the GoRouter).

## 4. Mapping to Yarumo

### 4.1. No real consumer

Per Yarumo context (`ROADMAP_NEW_MODULES.md`, MEMORY.md):

- **DaaS** is the first SDK consumer. Distribution target is SaaS web app / API — **not** a Cloud Foundry marketplace.
- **Aluna** is an agent runtime — same story.
- **Ontology Registry**, **Socrático**, the 6 vertical products — none of them have a CF / K8s-Service-Catalog deployment hypothesis on the roadmap.
- Yarumo's `boot/`, `managed/`, `health/`, `auth/` modules already cover the lifecycle / endpoint primitives a broker would compose from.

There is no documented or implied OSBAPI consumer in the entire canonical roadmap (`ROADMAP_NEW_MODULES.md` § 1 modules + § 2 tools). Related niches that were also evaluated and excluded from the canonical roadmap — `modules/discovery/` (K8s DNS covers it), `modules/api-gateway/` (compose existing pieces), `modules/cloud-function/` (FaaS too thin to wrap) — sit in the same category. An OSBAPI broker module would join them.

### 4.2. Architectural mismatch

| Spring SCOSB design choice | Yarumo's stance |
|---|---|
| Spring Boot auto-config | No DI, no annotations, no magic (explicit wiring via `boot/`) |
| `Mono<T>` reactive types throughout | Plain `(T, error)` Go idioms |
| `@SpringBootApplication` discovers `ServiceInstanceService` bean | Consumer would register impl manually in a `BeanFn` |
| Event-flow registries for hooks | Plain middleware chain |
| Auto-wired catalog from properties | Catalog is one Go struct; no framework needed |

A Go OSBAPI implementation would be ~600 LOC of `net/http` handlers + DTOs — closer to writing the spec by hand than wrapping a framework. There is nothing here that benefits from being a yarumo module rather than ad-hoc code in the (hypothetical) consumer.

### 4.3. Go ecosystem already covers it (if needed)

- **`pmorie/osb-broker-lib`** — Go library implementing OSBAPI server primitives. Last meaningful activity 2019, Kubernetes-Service-Catalog-era. Still functional, but unmaintained.
- **`drewwells/osb-checker-kit-go`** + spec-conformance tooling exists.
- The **OSBAPI spec itself** is small enough (~30 pages) that hand-rolling is reasonable when needed.

Adding `modules/servicebroker/` to yarumo would mean either (a) re-implementing what `pmorie/osb-broker-lib` already did, with no consumer to validate against, or (b) adopting an unmaintained library as a transitive dependency for code we don't need.

### 4.4. Niche-factor of OSBAPI itself

- **Cloud Foundry** is in long-tail maintenance mode (VMware Tanzu deprecations 2024–2025).
- **Kubernetes Service Catalog** project is **archived** (https://github.com/kubernetes-sigs/service-catalog).
- **Crossplane** has effectively replaced "service catalog" thinking with its own `Composition` / `XRD` model — it does **not** consume OSBAPI.
- Cloud providers (AWS / GCP / Azure) expose their own marketplaces with proprietary onboarding flows, not OSBAPI.

The marketplace integration story OSBAPI was built for has fragmented to the point that "expose product via OSBAPI" is no longer a default checkbox for SaaS providers in 2026.

## 5. Recommendation

**REJECT.** Do not create `modules/servicebroker/` and do not add an entry to `ROADMAP_NEW_MODULES.md`.

Rationale:

1. **No real consumer** — neither DaaS, Aluna, Ontology Registry, Socrático, nor any of the planned vertical products has a Cloud-Foundry / K8s-Service-Catalog distribution hypothesis.
2. **Shrinking ecosystem** — Kubernetes Service Catalog is archived; Cloud Foundry is in decline; major cloud marketplaces don't speak OSBAPI.
3. **Pareto failure** — even at 100% utilization, the broker surface (5 endpoints, ~3 interfaces) is small enough to hand-roll in the consumer.
4. **Architectural mismatch** — Spring SCOSB is auto-config + Reactor + annotations; yarumo is explicit, idiomatic Go, no DI.
5. **Better Go alternatives if ever needed** — `pmorie/osb-broker-lib` (unmaintained but functional) or hand-rolled `net/http` handlers.

Apply the same trigger pattern used when ruling out other speculative modules (saga, discovery, api-gateway): revisit only when a concrete consumer with a real marketplace integration appears.

If that day comes, the **right placement** would not be a generic `modules/servicebroker/` framework — it would be a thin `osbapi/` package **inside** the consumer (e.g. `apps/daas/internal/osbapi/`), composed from existing yarumo primitives: `managed/server_http`, `auth/` (Basic Auth), `validation/`, `health/`. No new module.

## 6. Proposed yarumo placement (if applicable)

**None.** No module proposed.

For traceability, if a "Considered and rejected" section is ever added back to ROADMAP_NEW_MODULES.md, the appropriate entry would be:

| Idea | Why not |
|---|---|
| `modules/servicebroker/` (Open Service Broker API server) | No real consumer; OSBAPI ecosystem shrinking (K8s Service Catalog archived, CF declining); spec is small enough to hand-roll inside a consumer if a marketplace integration ever materialises. If revived, would live inside the consuming app, not as a yarumo module. |

## 7. Open questions

- **Does DaaS distribution strategy include any marketplace channel?** (Per `docs/STRATEGY.md` — confirm no CF / K8s-Service-Catalog / Crossplane play.)
- **Is there an Aluna agent-as-a-service offering** where OSBAPI could expose agents-as-services to a managed platform? (Currently speculative.)
- **Crossplane integration** — if yarumo ever exposes products to Crossplane, that's a `XRD` / `Composition` design problem, **not** an OSBAPI one. Different track entirely.
- **Re-evaluation trigger**: a concrete consumer that needs a marketplace listing speaking OSBAPI. Until then, this analysis stands.

## 8. ROADMAP delta proposed (NOT applied)

None. No new module, no new tool, no annex. If/when a marketplace integration becomes real, the implementation lives **inside the consumer** (e.g. `apps/daas/internal/osbapi/`) composed from existing yarumo primitives (`managed/server_http`, `auth/`, `validation/`, `health/`). The canonical ROADMAP stays untouched.
