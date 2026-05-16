# Spring gRPC — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-grpc
> **Analyzed**: 2026-05-16
> **Recommendation**: PARTIAL — validates the `modules/common/grpc/` direction; adopts a **small, curated interceptor catalog** as an extension of [YA-0043](https://github.com/guidomantilla/yarumo/issues/43). No autoconfig, no annotation surface, no DSL.

## 1. Project summary

Spring gRPC (1.0.x stable, 1.1.0-M1 preview) is a **Spring Boot façade over `grpc-java`**. It does not add new gRPC features — it wraps the existing transport, channel, interceptor, and security APIs with Spring's conventions: autoconfiguration of a `Server` from `BindableService` beans, a `GrpcChannelFactory` to materialize channels from `spring.grpc.client.channels.<name>.*` properties, a `@GlobalServerInterceptor` / `@GlobalClientInterceptor` discovery pattern, a `GrpcSecurity` DSL bridging to Spring Security, automatic Micrometer-driven observability via Spring Boot Actuator, and an `@AutoConfigureInProcessTransport` helper for tests.

Three server transports are offered: Netty (default standalone), Servlet (run gRPC alongside HTTP in a servlet container via `spring-grpc-server-web-spring-boot-starter`), and in-process (test-only). Client-side adds `@ImportGrpcClients` for scanning generated stubs and registering them as beans, plus `BasicAuthenticationInterceptor` and `BearerTokenAuthenticationInterceptor` helpers.

The architectural reality: **the value-add is Spring container integration**. The `grpc-java` library itself already exposes everything Spring gRPC surfaces — `ServerBuilder`, `ManagedChannelBuilder`, `ServerInterceptor`, `ClientInterceptor`, `ServerServiceDefinition`, TLS via `NettyServerBuilder.sslContext()`, observability via `OpenTelemetryMetrics`. Spring gRPC saves Java developers from writing factory beans by hand. In Go that pain does not exist: `google.golang.org/grpc` is **already idiomatic**, already composable, already first-class, and there is no DI container to wire it into.

What does transfer to Go is the **interceptor catalog mindset** — Spring gRPC ships with built-in observability, security, and exception-handler interceptors that consumers opt into via properties. The Go analog is `grpc-ecosystem/go-grpc-middleware/v2`, which already has a mature catalog. Yarumo's job under [YA-0043](https://github.com/guidomantilla/yarumo/issues/43) is to **curate**, not invent: pick a small set of interceptors that every microservice needs, ship them as composable `grpc.UnaryServerInterceptor` / `grpc.StreamServerInterceptor` factories under `modules/common/grpc/`, and leave the rest to consumers.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Server interceptors — recovery & logging** | Panic recovery → `codes.Internal`; structured access log with method, duration, status, peer | **Already in `modules/common/grpc/`** (`RecoveryInterceptor`, `LoggingInterceptor`, plus stream variants). Reference point: every microservice needs both. |
| 2 | **Server interceptor — tracing / metrics** | OTel spans per RPC; histograms for duration, counters for status codes | **Gap** — belongs in `modules/telemetry/otel/grpc/`. `otelgrpc` from OTel-Go already exposes `UnaryServerInterceptor` + `StreamServerInterceptor`; yarumo wraps with sensible defaults (resource attributes, exemplars). |
| 3 | **Server interceptor — auth (token validation)** | Extract bearer token from metadata, validate, inject `Principal` into context | **Gap** — lives in `modules/auth/` (§ 1.2 of [ROADMAP_NEW_MODULES.md](../ROADMAP_NEW_MODULES.md)). gRPC variant of `AuthorizationFilter`. Reuses `common/crypto/tokens`. |
| 4 | **Server interceptor — request timeout / deadline propagation** | Enforce a server-side timeout if caller did not set a deadline; honor caller deadline otherwise | **Gap (YA-0043)** — Go-idiomatic: read `ctx.Deadline()`, wrap with `context.WithTimeout` when absent. Single small file. |
| 5 | **Server interceptor — rate limiting (per-method or per-peer)** | Reject requests with `codes.ResourceExhausted` when over budget | **Composition** — wraps `common/resilience/RateLimiter`. Keyed by `:authority` header, peer, or method. Belongs in [YA-0043](https://github.com/guidomantilla/yarumo/issues/43) catalog. |
| 6 | **Server interceptor — validation** | Validate request messages (protovalidate / buf-validate annotations) before reaching handler | **Composition** — wraps `bufbuild/protovalidate-go`. Returns `codes.InvalidArgument` with field paths in the `BadRequest` detail. Ship as `interceptors/validate/`. |
| 7 | **Exception handler chain** | Translate handler errors to `status.Error` with appropriate `codes.Code`; preserve sentinel errors; attach error details | Spring's `GrpcExceptionHandler` is the right shape: a small chain that maps domain errors → gRPC status. Single-file Go equivalent: `errs.TypedError → codes.Code` mapping function + interceptor that calls it. |
| 8 | **Client interceptor — retry with backoff** | Re-invoke `Unavailable` / `DeadlineExceeded` with exponential backoff + jitter | **Composition** — gRPC has native service-config retry (`MethodConfig.RetryPolicy`); also wraps `common/resilience/Retrier` once [#165](https://github.com/guidomantilla/yarumo/issues/165) lands. Prefer service-config retries; offer interceptor as fallback. |
| 9 | **Client interceptor — auth (bearer / basic token attachment)** | Attach `authorization` metadata header on every outbound call | **Composition** — small helpers under `common/grpc/interceptors/` that read from a token source. Mirrors Spring's `BasicAuthenticationInterceptor` + `BearerTokenAuthenticationInterceptor`. |
| 10 | **Client interceptor — tracing / metrics** | OTel spans + duration histograms on outbound calls | **Gap** — same `otelgrpc` package; `modules/telemetry/otel/grpc/` exports both server and client side. |
| 11 | **Client interceptor — deadline default** | Force a max deadline on every outbound call if caller did not set one | **Gap (YA-0043)** — paranoid default that prevents accidental no-deadline calls hanging on an unhealthy peer. Trivial wrapper. |
| 12 | **TLS / mTLS via `credentials.TransportCredentials`** | Server: load cert + key, optional client CA; Client: trust peer cert, optional client identity | **Configuration shape** — exposed via `common/grpc/Options` (`WithTLS(cfg *tls.Config)`). The crypto inputs come from `common/crypto/certs` and `common/crypto/keys`. |
| 13 | **In-process transport for tests** | `bufconn.Listener` + `grpc.WithContextDialer` — server and client in-memory, no port | Already idiomatic in Go via `google.golang.org/grpc/test/bufconn`. Yarumo can offer a `common/grpc/test/` helper: `NewInProcessServer()` + `NewInProcessClientConn()`. Small surface, real value. |
| 14 | **gRPC health service** | Standard `grpc.health.v1.Health` server with status per service | **Composition** — `google.golang.org/grpc/health` already provides it; `managed/server_grpc_builder` adds a `WithHealthCheck()` option that wires `common/health/Aggregator` (closed [YA-0077](https://github.com/guidomantilla/yarumo/issues/77)) to the gRPC health server status. |
| 15 | **gRPC reflection service** | Allow `grpcurl` and Insomnia to introspect at runtime | **Composition** — `google.golang.org/grpc/reflection.Register(s)`. One-liner option `WithReflection()` on the server builder. |
| 16 | **Graceful shutdown** | `GracefulStop()` with a deadline that falls back to `Stop()` | **Already in `modules/managed/`** (`BuildGrpcServer`) — fits the `StopFn(ctx, timeout)` shape. |

## 3. Long-tail features (skip)

- **Servlet transport (`spring-grpc-server-web-spring-boot-starter`)**. Spring's bridge that runs gRPC inside a servlet container. Go has no servlet container; gRPC always runs as a standalone listener. Skip entirely.
- **Reactive support (Reactor / Mono / Flux variants)**. Go uses `context.Context` + channels / iterators; no Project Reactor analog. Skip.
- **Spring Security `GrpcSecurity` DSL** (`grpc.authorizeRequests().methods("Simple/StreamHello").hasAuthority(...)`). The DSL is a Spring-ergonomics solution; in Go authorization is a normal interceptor that consults `modules/auth/`. Express rules as code in the handler or as a small per-method map — no DSL needed.
- **OAuth2 Resource Server autoconfig**. Translation: an interceptor that validates a JWT from the `authorization` metadata against a JWKS endpoint. Lives in `modules/auth/oauth2/` (already planned § 1.2). Spring just wires it; the value is in the JWT validator, which `common/crypto/tokens` already covers.
- **`@GlobalServerInterceptor` / `@GlobalClientInterceptor` annotation discovery**. Annotation-driven bean scanning has no Go equivalent. Consumers register interceptors explicitly via `grpc.UnaryInterceptor(chain(...))` / `grpc.WithUnaryInterceptor(...)` — which is **better**, not worse: explicit ordering, no surprise interception.
- **`@ImportGrpcClients` stub scanning**. Spring's classpath scanner that finds generated stubs and registers them as beans. Go has no classpath introspection; stubs are just types instantiated from a `*grpc.ClientConn`. Wire them in `modules/boot/` `BeanFn`s (§ 1.1) when that module lands. One line per stub. Skip the abstraction.
- **`ServerServiceDefinitionFilter` to selectively register services**. Useful in JVM mega-monoliths where many `BindableService` beans live in one context. In Go you import what you use; if a service is in the binary, it is intentional. Skip.
- **`@AutoConfigureInProcessTransport` annotation**. Replaced by the `common/grpc/test/` helper functions noted under Pareto #13.
- **`spring.grpc.server.health.actuator.health-indicator-paths`** (sync Boot Actuator health → gRPC health service). Replicated by composing `common/health/` (closed) with `WithHealthCheck()` on the gRPC builder. The property hierarchy itself doesn't transfer.
- **`ServerBuilderCustomizer` / `GrpcChannelBuilderCustomizer` beans**. Spring's pattern for "give me access to the builder before it's built". In Go this is just `Option` functions accepting `*grpc.Server` / `grpc.DialOption`. Already the idiom.
- **Multi-channel autoconfig (`spring.grpc.client.channels.<name>.address`)**. The string-keyed factory is useful if you have ~10+ stub-to-target mappings to manage. For typical Go microservices (2-5 outbound dependencies), explicit `grpc.NewClient(addr, opts...)` is clearer. If a real consumer needs named-channel registry, design it then — premature now.
- **`@LocalGrpcPort`**. Test ergonomics for ephemeral ports. Trivial in Go: `lis, _ := net.Listen("tcp", "127.0.0.1:0"); lis.Addr().(*net.TCPAddr).Port`. No abstraction needed.
- **`GrpcSecurity.preauth()` (mTLS pre-authentication)**. The TLS handshake itself gives you the client cert via `peer.FromContext(ctx).AuthInfo.(*credentials.TLSInfo).State.PeerCertificates`. A one-liner; not a framework feature.

## 4. Mapping to Yarumo

**Existing modules with overlap**:

- **`modules/common/grpc/`** — already ships:
  - `Server` interface + `server` impl wrapping `*grpc.Server`.
  - `RecoveryInterceptor`, `LoggingInterceptor` + stream variants — two of the most-used Pareto items.
  - `Options` pattern (`WithService`, `WithServerOption`).
  - `Error` type with `ServerType` for sentinel-translated failures.
  - Type compliance vars (`_ Server = (*server)(nil)`, `_ UnaryInterceptorFn = RecoveryInterceptor`).
- **`modules/managed/`** — already ships:
  - `BuildGrpcServer(ctx, name, internal, errChan)` lifecycle wrapper.
  - `server_grpc_builder.go` + `server_grpc_adapter.go` — `Component[GrpcServer]` with `StopFn` returning a graceful-stop hook.
  - Symmetric to the HTTP server builder, fitting the `managed.Lifecycle` contract.
- **`modules/common/health/`** — leaves shipped by [YA-0077](https://github.com/guidomantilla/yarumo/issues/77) closed 2026-05-13. Provides primitives the gRPC health-check option will consume.
- **`modules/common/resilience/`** — registries shipped by [YA-0076](https://github.com/guidomantilla/yarumo/issues/76) closed 2026-05-13. Provides `CircuitBreaker` + `RateLimiter` that rate-limit interceptors compose against. Retrier still pending (#165).

**Open ticket alignment**:

- **[YA-0043](https://github.com/guidomantilla/yarumo/issues/43)** (Phase-2 follow-up, deferred) — "common/grpc interceptor library: retry, deadline, logging, metrics". This analysis **confirms the scope is correct** and refines the list: ship 4-6 server interceptors (recovery, logging, deadline, validate, rate-limit, panic-translate) and 3-4 client interceptors (auth, deadline-default, retry, propagate-headers). Tracing/metrics interceptors should land in `modules/telemetry/otel/grpc/` rather than `common/grpc/` so `common/` stays SDK-free.

- **`modules/auth/`** AuthorizationFilter for gRPC (§ 1.2) — directly mirrored by Spring's `AuthenticationProcessInterceptor`. The Spring approach (interceptor that runs through an `AuthorizationManager` chain) maps cleanly: extract token from `metadata.MD`, validate via `common/crypto/tokens`, inject `auth.Principal` into context.

**Where Spring's structure does NOT translate**:

- **No autoconfig**. Yarumo's wiring is explicit via `modules/boot/` `BeanFn`s (§ 1.1).
- **No annotation-driven interceptor discovery**. Interceptors registered explicitly in build order. This is a feature, not a regression — implicit chains are a frequent Spring foot-gun.
- **No DSL for authorization rules**. The `grpc.authorizeRequests().methods(...).hasAuthority(...)` builder is replaced by a normal interceptor that consults a per-method permission map or `modules/auth/` resolver.

## 5. Recommendation

**PARTIAL.** Spring gRPC validates the direction of `modules/common/grpc/` and confirms [YA-0043](https://github.com/guidomantilla/yarumo/issues/43)'s scope. Adopt **the interceptor catalog idea** as a curated subset; reject autoconfig, annotation discovery, the security DSL, the multi-channel registry, and the servlet transport. Concretely:

1. **Expand `common/grpc/` under YA-0043** with the interceptor catalog from §2 items #4-7 server-side and #8-9, #11 client-side. Each interceptor is a small `*.go` file in `modules/common/grpc/interceptors/<name>/` exporting a factory function plus options. Recovery and logging are already in place.
2. **Add `modules/telemetry/otel/grpc/`** wrapping `otelgrpc` for the tracing/metrics interceptors (server + client). Keeps `common/grpc/` free of OTel imports.
3. **Add `common/grpc/test/`** with `NewInProcessServer()` / `NewInProcessClientConn()` helpers wrapping `bufconn`. ~50 lines, replaces Spring's `@AutoConfigureInProcessTransport` ergonomics.
4. **Extend `managed/server_grpc_builder.go`** with `WithHealthCheck(agg health.Aggregator)` and `WithReflection()` options.
5. **Defer** Spring-style multi-channel registry until a real consumer asks. Today, explicit `grpc.NewClient(addr, opts...)` per dependency in `boot/` `BeanFn`s is clearer.
6. **Don't ship a `GrpcSecurity` DSL**. The gRPC auth interceptor in `modules/auth/` (§ 1.2) takes a small policy map (method → required scopes/roles) — no fluent builder.

The net effect: yarumo ends up with a thin, idiomatic gRPC stack roughly **20% the size** of Spring gRPC's surface, covering 80% of the real microservice needs. The other 80% of Spring's surface is Spring-container ergonomics that Go does not need.

## 6. Proposed yarumo placement

**Module**: `modules/common/grpc/` (extension via [YA-0043](https://github.com/guidomantilla/yarumo/issues/43))

**Subpackages**:

```
modules/common/grpc/
  server.go                       Server interface + impl              [EXISTS]
  options.go                      Functional options for NewServer     [EXISTS]
  types.go                        Type aliases + compliance vars       [EXISTS]
  errors.go                       ServerType errors                    [EXISTS]
  functions.go                    RecoveryInterceptor, LoggingInterceptor + stream variants  [EXISTS]
  interceptors/
    deadline/                     Server + client deadline defaults    [NEW — YA-0043]
    validate/                     protovalidate-go integration         [NEW — YA-0043]
    ratelimit/                    Composes common/resilience/RateLimiter [NEW — YA-0043]
    retry/                        Client-side retry — falls back to common/resilience/Retrier (#165) [NEW — YA-0043]
    auth/                         Bearer / basic-token attachment (client); token extraction (server) [NEW — partial; full validator in modules/auth/]
    errors/                       Domain-error → gRPC status mapping   [NEW — YA-0043]
  test/
    inprocess.go                  bufconn-based server + client helper [NEW]
```

**Sibling module additions**:

```
modules/telemetry/otel/grpc/
  server.go    otelgrpc unary + stream server interceptors with yarumo-defaulted attribute config
  client.go    otelgrpc unary + stream client interceptors

modules/managed/
  server_grpc_builder.go    EXISTS — extend with WithHealthCheck / WithReflection
  server_grpc_adapter.go    EXISTS

modules/auth/  (planned § 1.2)
  grpc/
    authorization_interceptor.go  Server-side interceptor that validates the token and injects Principal
```

**Internal deps**:

- `modules/common/grpc/` → `google.golang.org/grpc`, `modules/common/assert`, `modules/common/errs`, `modules/common/log`.
- `modules/common/grpc/interceptors/ratelimit/` → `modules/common/resilience/`.
- `modules/common/grpc/interceptors/retry/` → `modules/common/resilience/` (#165).
- `modules/common/grpc/interceptors/validate/` → `buf.build/go/protovalidate`.
- `modules/telemetry/otel/grpc/` → `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc`, `modules/common/grpc/`.
- `modules/managed/server_grpc_builder.go` → `modules/common/grpc/`, `modules/common/health/`, `google.golang.org/grpc/health`, `google.golang.org/grpc/reflection`.
- `modules/auth/grpc/` → `modules/common/crypto/tokens`, `modules/common/grpc/`.

**Go libraries to wrap**:

- `google.golang.org/grpc` — primary transport / server / channel.
- `google.golang.org/grpc/health` — standard health-check service (composed in `managed`).
- `google.golang.org/grpc/reflection` — reflection service (composed in `managed`).
- `google.golang.org/grpc/test/bufconn` — in-process transport for tests (`common/grpc/test/`).
- `grpc-ecosystem/go-grpc-middleware/v2` — **reference only**. Read the catalog; cherry-pick patterns. Yarumo writes its own interceptors so they speak the yarumo error / log / context shape consistently. Pulling the dep adds ~30 transitive packages for code we'd be rewriting anyway.
- `bufbuild/protovalidate-go` — request validation (`interceptors/validate/`).
- `go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc` — tracing + metrics (`modules/telemetry/otel/grpc/`).
- `sony/gobreaker`, `golang.org/x/time/rate` — already deps of `common/resilience/`; reused transitively.

## 7. Open questions

1. **Should validation interceptor be opt-in per service or default-on?** Spring's stance is default-on if `spring.grpc.server.validation.enabled=true`. Yarumo lean: opt-in via explicit `WithInterceptor(validate.New())` — silent rejection of malformed requests has bitten too many users.
2. **Client retry: service-config vs interceptor?** gRPC supports retry via `MethodConfig.RetryPolicy` JSON in `grpc.WithDefaultServiceConfig`. This is the **idiomatic Go path**. An interceptor that wraps `common/resilience/Retrier` only makes sense if [#165](https://github.com/guidomantilla/yarumo/issues/165) lands and we want a unified retry knob across HTTP + gRPC. Decision deferred to #165.
3. **gRPC reflection in production?** Spring auto-enables on classpath presence. Yarumo lean: **off by default**, opt-in via `WithReflection()`. Reflection leaks proto schemas — fine in dev, debatable in prod.
4. **OTel attribute defaults.** What `gen_ai.system`-style namespacing do we apply for gRPC method attributes? Probably `rpc.system="grpc"`, `rpc.service`, `rpc.method`, status code as a numeric attribute (already otelgrpc defaults). Confirm against `modules/telemetry/otel/`'s existing convention when the package lands.
5. **Servlet-style transport?** Hard skip — but worth a note: if yarumo ever ships a unified HTTP+gRPC server (e.g. `connectrpc/connect-go` or `improbable-eng/grpc-web`), that's a **new design**, not a Spring port. File as brainstorm if a real use case emerges.
6. **`@ImportGrpcClients` analog via `tools/`?** Spring's stub scanner has a cousin in Go: `tools/routegen/` (§ 2.1). A `tools/grpcstubgen/` that walks a directory of `*.pb.go`, finds `*Client` types, and emits `BeanFn`s for `modules/boot/` is conceivable but premature — file under brainstorm only if `boot/` lands and consumers complain about stub-wiring boilerplate.
7. **Connect protocol support (`connectrpc/connect-go`)?** Connect = gRPC-over-HTTP/1.1+JSON with the same proto schemas. Spring gRPC doesn't cover it; it's a Go-native ecosystem gap. **Out of scope** for this analysis; track separately if Aluna or DaaS need browser clients.
