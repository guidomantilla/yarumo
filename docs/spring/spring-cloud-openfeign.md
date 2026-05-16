# Spring Cloud OpenFeign — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-openfeign
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL — reject the declarative-proxy core, salvage the interceptor / decoder / error-decoder patterns into the existing `common/rest/` track ([YA-0044](https://github.com/guidomantilla/yarumo/issues/44)).

## 1. Project summary

Spring Cloud OpenFeign is the Spring Boot integration of Netflix's Feign library. It exposes **REST clients as Java interfaces** annotated with `@FeignClient` + `@GetMapping`/`@PostMapping`/etc.; a runtime JDK dynamic proxy translates method calls into HTTP requests via configurable `Encoder` / `Decoder` / `Contract` / `Client` beans. Around that core, OpenFeign adds: Spring Cloud LoadBalancer integration, Spring Cloud CircuitBreaker integration (Resilience4j by default), fallbacks / fallback factories, property-based per-client configuration (timeouts, headers, default query params), Micrometer observability, OAuth2 client integration, response caching via `@Cacheable`, per-client refresh via `@RefreshScope`, optional Apache HC5 / OkHttp backends, and request/response gzip compression.

**Crucial status signal**: the project has been **feature-complete since Spring Cloud 2022.0.0**. The official docs explicitly redirect new users to **Spring HTTP Service Clients** (`@HttpExchange` + `HttpServiceProxyFactory`, native Spring Framework 6+). OpenFeign is in maintenance mode, accepting bugfixes and the occasional small community PR. A 2026 evaluation reads a finalised feature set, not a growing one — the right question is "which patterns survived?", not "which extension points should we mirror?".

OpenFeign's value proposition leans entirely on **runtime dynamic proxies over interfaces**. That single Java-platform capability does most of the work; remove it and the abstraction collapses into "a polite façade for `RestClient`/`WebClient`". Go has no equivalent — `reflect.MakeFunc`-style interception exists but is ergonomically painful and toolchain-hostile (lost types, no IDE assists, runtime-only errors).

## 2. Pareto features (top-20%)

The features that actually carry the framework's weight (and have direct Yarumo analogs worth thinking about):

| Feature | What it does | Why it matters |
|---|---|---|
| **`@FeignClient` declarative interfaces** | Annotated Java interface → runtime proxy that performs HTTP. | The framework's reason to exist. Single biggest ergonomic win in Java land. |
| **Pluggable `Encoder` / `Decoder`** | Default `SpringEncoder` / `ResponseEntityDecoder` uses Jackson via `HttpMessageConverters`. Swappable per-client. | Codec abstraction is what makes the interface body-agnostic. |
| **`RequestInterceptor` chain** | `Collection<RequestInterceptor>` bean applied to all outgoing requests; mutates `RequestTemplate` (headers, query params, body). | Standard cross-cutting concern attachment — auth, tracing, correlation IDs, signing. |
| **`ResponseInterceptor`** | Symmetric to `RequestInterceptor` on the response side. | Post-processing hook for decoded bodies / status. |
| **`ErrorDecoder`** | Maps non-2xx `Response` → typed `Exception`. Per-client configurable. | Lets domain code use `try/catch SpecificDomainException` instead of branching on status codes. |
| **`Retryer`** | Pluggable retry policy with backoff (default: `NEVER_RETRY`). | Retry hook anchored at the client boundary, not on every call site. |
| **`Logger.Level` (NONE/BASIC/HEADERS/FULL)** | Stepped log verbosity for request/response. | Operational toggle without code changes. |
| **Property-based per-client config** | `spring.cloud.openfeign.client.config.<name>.{connectTimeout, readTimeout, defaultRequestHeaders, defaultQueryParameters, requestInterceptors, errorDecoder, retryer, ...}`. | Ops-side configuration without recompiling. Most-used part of OpenFeign in real deployments. |
| **`FeignClientsConfiguration`** | Spring `@Configuration` exposing all defaults as overridable beans (`Decoder`, `Encoder`, `Contract`, `Client`, `Feign.Builder`). | Clean override surface for the standard beans. |
| **Spring MVC contract** | Reuses `@RequestMapping`, `@PathVariable`, `@RequestParam`, `@RequestHeader`, `@RequestBody` — same annotations as the server side. | Server/client symmetry: same DTOs, same path templates, same `@Valid`. |

Below the 20% line but tightly coupled to the core: LoadBalancer integration (service-name-as-host), CircuitBreaker integration (`spring.cloud.openfeign.circuitbreaker.enabled` + `fallback` / `fallbackFactory`), Micrometer observation capability (`spring.cloud.openfeign.micrometer.enabled`), request/response gzip compression, OAuth2 token interceptor (`spring.cloud.openfeign.oauth2.enabled`), response caching via `@Cacheable`, `@RefreshScope` for hot-reloading timeouts and URLs, Apache HC5 / OkHttp backend switches (`spring.cloud.openfeign.httpclient.*`).

## 3. Long-tail features (skip)

Several mechanisms are either Spring-platform-specific in a way that does not port, or solve problems Yarumo handles elsewhere:

| Feature | Why skip |
|---|---|
| **Runtime JDK dynamic proxy** | No equivalent in Go that does not destroy ergonomics. `reflect.MakeFunc` exists but produces zero IDE support, runtime-only signature checks, and gnarly stack traces. Code-gen replaces this dimension entirely (see § 4). |
| **Spring MVC contract / annotation reuse** | Yarumo does not use Spring MVC annotations server-side. The "symmetry" argument disappears. |
| **Fallback / `FallbackFactory`** | Fallback-on-circuit-open is a domain decision that should live in the calling code, not be auto-wired by the client. Go's explicit error returns make this trivial — no annotation needed. |
| **Spring Cloud LoadBalancer integration (service name → discovery URL)** | The Yarumo roadmap has not added a discovery module; K8s DNS + service mesh cover ~95% of the in-cluster case. No Eureka/Consul client to integrate with. |
| **CircuitBreaker auto-wrapping (`spring.cloud.openfeign.circuitbreaker.enabled`)** | Yarumo's `common/resilience/CircuitBreakerRegistry` is explicit, goroutine-free, and lazy. Callers wrap their own boundaries; no need for a per-method auto-wrapper. |
| **OAuth2 token interceptor** | Belongs in the planned `modules/auth/oauth2/` (`ROADMAP_NEW_MODULES.md` § 1.2), not in a REST-client module. Injected via a `RequestInterceptor`-equivalent at call site. |
| **`@RefreshScope` per-client refresh** | Yarumo `config` is one-shot bootstrap; dynamic refresh is not on the roadmap as a runtime concern of the HTTP client. |
| **Response caching via `@Cacheable`** | Cross-cuts cleanly with `modules/cache/` and HTTP caching headers in `common/http`. No reason to bind it to the client abstraction. |
| **Gzip compression toggles (`spring.cloud.openfeign.compression.*`)** | `net/http` handles `Accept-Encoding`/`Content-Encoding` transparently. Non-feature. |
| **Apache HC5 / OkHttp backend switches** | Go has one canonical client (`net/http`) plus `http.RoundTripper` middleware. The Spring "swap the backend" property is a JVM-ecosystem artefact. |
| **`@SpringQueryMap`, `@CollectionFormat`, `@MatrixVariable`** | Niche server contract idioms; encode/decode helpers in `common/rest` cover the common case. |
| **HATEOAS support (`CollectionModel`, etc.)** | Yarumo does not ship a HATEOAS dialect; this would precede module work, not follow. |
| **AOT / native image considerations** | Go AOT is the default; nothing to mirror. |
| **`FeignClientConfigurer.inheritParentConfiguration`** | Solves a Spring context inheritance quirk; no Go analog. |
| **`primary = false` disambiguation** | Spring `@Primary` artifact. No analog. |
| **Manual `Feign.builder()` for multi-credential clients** | In Go this is just "construct two clients with different interceptors". |

## 4. Mapping to Yarumo

**Existing modules with overlap**: `common/rest/` (YA-0044 refinements), `common/http/` (YA-0042 retry / circuit-breaker hook).
**OpenFeign-style declarative client**: rejected on own merits — Go has no decent dynamic-proxy facility; code-gen is the Go answer (see § 4.1).
**Anti-patterns to avoid**: god-struct configuration, property-name-keyed client registries, auto-wired fallbacks, runtime URL refresh tied to the HTTP client, server-contract annotation reuse that does not exist in Yarumo.

### 4.1. The declarative-proxy core: rejected on own merits

The clean rejection holds, stated inline rather than by reference:

> **No `modules/clients/` for OpenFeign-style declarative interfaces.** Go has no decent dynamic-proxy facility. `reflect.MakeFunc` is real but produces zero IDE support, runtime-only signature checks, and gnarly stack traces. The abstraction collapses without the proxy.

Two 2026 reinforcements:

1. **Spring itself has moved on.** The recommended replacement (Spring HTTP Service Clients, `@HttpExchange` + `HttpServiceProxyFactory`) is still proxy-based but now lives in Spring Framework core, not Spring Cloud. The Java ecosystem treats "declarative client over interface" as a stable but mature pattern, not a frontier. No momentum to mirror in 2026.
2. **The Go ecosystem already answered with code-gen.** `oapi-codegen` (from OpenAPI), `connectrpc/go` (from Protobuf), `protobuf-go` + `grpc-go` (from `.proto`), and `sqlc` (analogous spirit, different domain) all produce **typed clients/handlers at build time** with full IDE support, zero reflection, zero runtime proxy magic. This is the right answer for Go: shift the "describe shape, generate plumbing" responsibility to a tool that emits real `.go` files.

If Yarumo ever wants `@FeignClient`-grade ergonomics, the route is **OpenAPI → `oapi-codegen` → generated typed client**, optionally wrapped with the Yarumo interceptor / decoder / error-mapper layer (§ 4.2). Yarumo's role in that pipeline is **not** to ship a proxy framework; it is to ship the **composition layer underneath the generated code**. That is already the implicit design of `common/rest` (a small `RequestSpec` → `ResponseSpec[T]` typed call), and the natural place for the salvageable patterns from § 2.

### 4.2. Salvageable patterns — `common/rest/` (YA-0044) and `common/http/` (YA-0042)

What is worth lifting from OpenFeign is **not the proxy**, it is the **shape of the composition surface**: a small fixed cast of pluggable concerns around the HTTP boundary. Mapped to existing tracks:

| OpenFeign concept | Yarumo placement | Status |
|---|---|---|
| `RequestInterceptor` chain | `common/rest/` — interceptor list option (`WithRequestInterceptors(...)`) applied inside `Call` / `CallStream` before invoking `DoFn`. Covers signing, correlation/trace headers, auth header injection. | Already on YA-0044's planned-Phase-2 list as "interceptors". |
| `ResponseInterceptor` | `common/rest/` — symmetric `WithResponseInterceptors(...)` applied after decode. | Add to YA-0044. |
| `ErrorDecoder` (status → typed error) | `common/rest/errors.go` — generalise the existing `DecodeHTTPError[E]` shape into a pluggable `ErrorDecoderFn[E]` option, so each client can map non-2xx to its own domain `Error`. The current `HTTPError` is the right primitive; the missing piece is a **caller-supplied** decoder. | Extend YA-0044 scope (it already mentions request signing + interceptors + path templating; error-decoder fits the same group). |
| `Encoder` / `Decoder` | `common/rest/` — `RequestSpec` and `ResponseSpec[T]` already carry payloads; codec is currently fixed to JSON via `encoding/json`. A pluggable `EncoderFn` / `DecoderFn` option lets clients swap to Protobuf, MessagePack, XML without forking. | Add to YA-0044. |
| `Retryer` | `common/http/` already has retry hooks via `retry-go` (`RetryIfHttpError`, `RetryOn5xxAnd429Response`). No new work; OpenFeign-equivalent functionality exists. | Already covered by `common/http` retry + YA-0042's Retry-After parsing. |
| `Logger.Level` (BASIC / HEADERS / FULL) | `common/rest/` — a `WithLoggingLevel` option that emits structured log events via `slog`. Operationally useful. | Add to YA-0044 as a minor extension. |
| Property-based per-client config | **Reject as a pattern.** Yarumo composes options at construction time, not via a property registry keyed by client name. Consumers wire interceptor lists / timeouts in their `BeanFn`s (`modules/boot/`) using `modules/config` for the values. | No action — the Spring "config-by-name lookup" pattern depends on the DI container Yarumo does not have. |
| Path templating (`/stores/{storeId}`) | `common/rest/` — currently `RequestSpec` carries a URL; YA-0044 lists path templating as a planned feature. Should accept `(template string, vars map[string]any)` and expand safely (URL-escape `{var}` slots). | Add to YA-0044. |
| Default headers / default query params | `common/rest/` — `WithDefaultHeaders` / `WithDefaultQueryParams` options merged into `RequestSpec` at call time. | Add to YA-0044. |
| Request signing | `common/rest/` — already on YA-0044's "Phase-2 features" list (signing). The `RequestInterceptor` plumbing above is the right vehicle. | Already covered by YA-0044. |
| Circuit breaker hook | `common/http/` already has YA-0042's "circuit breaker hook" planned; sits inside the `DoFn` pipeline so any `common/rest` client inherits it transparently. The breaker is supplied by `common/resilience/CircuitBreakerRegistry` (already shipped via YA-0076), not auto-wrapped per method. | Already covered by YA-0042 + `common/resilience`. |
| LoadBalancer integration | Not applicable — no discovery module on the roadmap; K8s DNS + service mesh cover the in-cluster case. | Reject. |
| Fallbacks / `FallbackFactory` | Caller responsibility. Idiomatic Go: `result, err := client.X(...); if err != nil { return fallback() }`. No framework hook needed. | Reject. |
| Spring MVC contract reuse | N/A — Yarumo does not use Spring MVC annotations. The contract layer disappears under code-gen anyway. | N/A. |

The headline is that **YA-0044 already names "interceptors", "path templating", and "request signing"** as planned Phase-2 features for `common/rest/`. The OpenFeign exercise validates that the planned list is the right list and suggests three small extensions:

- **`ErrorDecoderFn[E]`** — generalise the existing `DecodeHTTPError[E]` into a per-client-pluggable option so domain code can `errors.As(err, &MyDomainError{})` directly off the call.
- **Pluggable `EncoderFn` / `DecoderFn`** — codec swap (Protobuf, MessagePack) without forking the package.
- **`ResponseInterceptor` symmetry** — already-present on the request side; the response side rounds out the shape.

None of these create new modules. All three fit within the existing YA-0044 ticket's scope, or as sibling follow-ups if the breadth is worth splitting.

### 4.3. The code-gen angle

OpenAPI-driven code-gen (`oapi-codegen`) is the right "OpenFeign equivalent" for Yarumo, but **it lives outside the modules layer** — it is a build-time concern, not a runtime library. The natural placement, if Yarumo ever wants to standardise on it, is:

- `tools/openapi-client-gen/` — thin wrapper around `oapi-codegen` with Yarumo conventions baked in (use `common/rest` as the underlying transport, emit interceptor / error-decoder hook-points).
- The generated code calls into `common/rest` (interceptors, error decoder, encoder/decoder, retry, circuit breaker) for the runtime side.

This stays consistent with the workspace's `tools/routegen/` pattern (`ROADMAP_NEW_MODULES.md` § 2.1): when code-gen is the right answer, it lives in `tools/`, not `modules/`. **No proposal to file this now** — only if a real consumer (DaaS, Aluna) demands a generated client for an external service whose OpenAPI spec is the single source of truth.

## 5. Recommendation

**PARTIAL.**

Concretely:

1. **REJECT** the OpenFeign-style declarative client core. No `modules/clients/`. The justification ages well: Go's lack of a clean dynamic-proxy facility makes the abstraction ugly; Spring itself has moved away from Feign toward `@HttpExchange`; the Go ecosystem's answer (code-gen via `oapi-codegen`) is structurally superior and already industry-standard.
2. **ADOPT** the surrounding composition patterns into the **already-planned** `common/rest/` and `common/http/` Phase-2 work:
   - `common/rest/` (YA-0044) already lists interceptors, path templating, and request signing. This analysis validates that scope and suggests three small additions worth folding in:
     - `ErrorDecoderFn[E]` (generalise existing `DecodeHTTPError`).
     - Pluggable `EncoderFn` / `DecoderFn` options.
     - `ResponseInterceptor` (symmetry with `RequestInterceptor`).
   - `common/http/` (YA-0042) already lists Retry-After parsing and a circuit breaker hook — OpenFeign's `Retryer` maps cleanly onto existing primitives; no new work. The breaker comes from `common/resilience/CircuitBreakerRegistry` (already shipped).
3. **DEFER** the code-gen track. Do not file `tools/openapi-client-gen/` until a real DaaS or Aluna consumer needs to call an OpenAPI-described external service where the spec is the source of truth. When that demand appears, the placement is clear: `tools/`, not `modules/`.

This is a **PARTIAL** rather than a **DEFER** because the salvage already has a ticket home (YA-0044). The analysis adds three concrete sub-features to consider when YA-0044 is picked up; it does not propose new tickets, modules, or workstreams.

## 6. Proposed yarumo placement (if applicable)

Nothing new to place. The mapping is:

```
common/rest/        (YA-0044, planned Phase-2)
    RequestInterceptor chain   <- OpenFeign RequestInterceptor
    ResponseInterceptor        <- OpenFeign ResponseInterceptor  (new, fold into YA-0044)
    ErrorDecoderFn[E]          <- OpenFeign ErrorDecoder         (new, fold into YA-0044)
    EncoderFn / DecoderFn      <- OpenFeign Encoder/Decoder      (new, fold into YA-0044)
    Path templating            <- @PathVariable / RequestTemplate
    Request signing            <- BasicAuthRequestInterceptor + custom
    WithLoggingLevel           <- OpenFeign Logger.Level         (optional, fold into YA-0044)

common/http/        (YA-0042, planned Phase-2)
    Retry-After honoring       <- OpenFeign Retryer + 429 handling
    Circuit breaker hook       <- OpenFeign + Spring Cloud CircuitBreaker integration
                                   (caller-side, via DoFn wrap, using common/resilience;
                                    not auto-applied per method)

(no new module)     Discovery / LoadBalancer / Fallback / @Cacheable / @RefreshScope /
                    HC5/OkHttp backend swap / gzip toggles: explicitly out of scope.
```

If the three "new" items above are worth tracking separately rather than absorbing them silently into YA-0044, file them as YA-0044 follow-ups when work resumes — same shape as the YA-0154..YA-0161 family that came out of the validation review.

## 7. Open questions

1. **Should `common/rest`'s `ErrorDecoderFn[E]` be generic on `[E]` or return `error`?** Current `DecodeHTTPError[E]` is generic; the question is whether the per-client option should preserve that, or whether returning a plain `error` (and letting callers `errors.As`) is cleaner. Lean toward generic, since the existing helper already proves the shape.
2. **Does Yarumo want a `tools/openapi-client-gen/` parallel to `tools/routegen/`?** Not until a concrete consumer needs it. Possible trigger: DaaS needs to call a third-party API with a rich OpenAPI spec (Stripe, Plaid, etc.); Aluna needs typed clients for several LLM providers behind a unified shape. Today, neither demand is present.
3. **Codec plug-in surface — JSON-only or generic?** YA-0044 currently assumes JSON. Phase-2 follow-ups (Protobuf, MessagePack) are realistic but not currently demanded. Defer until a consumer asks; the `EncoderFn` / `DecoderFn` shape is small enough to retrofit.
4. **Where does `ResponseInterceptor` sit relative to `slogctx` and OTel client instrumentation?** OpenFeign's response interceptor mostly powers logging/metrics. In Yarumo, `slogctx` + the existing HTTP-client OTel instrumentation in `modules/telemetry/otel/` (and the planned RoundTripper decorators in YA-0081) probably cover this already; a dedicated `ResponseInterceptor` may be redundant. Worth pressure-testing before adding the surface.
5. **Should `common/rest` ship a "default interceptors" preset (auth + trace + correlation ID) for the common case?** Convenient, but introduces opinionation that bleeds into `modules/auth/` (still planned). Park until that module lands — composition at the `boot/` layer is the cleaner placement.

## 8. ROADMAP delta proposed (NOT applied)

**Nothing structural.** The analysis lands entirely within existing tickets:

- **YA-0044 (`common/rest/` Phase-2)** — extend scope with three sub-features:
  - `ErrorDecoderFn[E]` (generalise `DecodeHTTPError[E]` into a per-client option).
  - Pluggable `EncoderFn` / `DecoderFn` (codec swap without forking).
  - `ResponseInterceptor` (symmetry with the already-planned `RequestInterceptor`).
  - Optional: `WithLoggingLevel` (NONE/BASIC/HEADERS/FULL) emitted via `slog`.

- **YA-0042 (`common/http/` Phase-2)** — no scope change. Retry-After parsing and circuit-breaker hook already cover OpenFeign's `Retryer` analogue; the breaker primitive is `common/resilience/CircuitBreakerRegistry` (YA-0076, shipped).

- **`ROADMAP_NEW_MODULES.md`** — no new section to add. The "no dynamic proxies in Go" rejection is captured inline in this analysis rather than as a roadmap line item; nothing to track on the roadmap because nothing is being built.

- **`tools/openapi-client-gen/`** — not filed. Trigger to revisit: a real DaaS or Aluna consumer with an OpenAPI-described external service where the spec is the single source of truth.
