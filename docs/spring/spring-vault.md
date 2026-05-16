# Spring Vault — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-vault
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Vault 4.0.2 (stable; part of Spring Data) is the **foundational** Java client for HashiCorp Vault: `VaultTemplate` for imperative ops, `ReactiveVaultTemplate` for non-blocking, `ClientAuthentication` for ~14 auth backends, `SessionManager` for token lifecycle, dedicated operations classes for KV v1/v2, Transit, PKI, Token and System engines, plus `SecretLeaseContainer` for dynamic-secret lease management. Scope is **client-library only** — no auto-configuration, no Spring Boot starter, no `bootstrap.yml` property-source magic. Those higher-level concerns live in the separately analysed **Spring Cloud Vault** (autoconfig wrapper that consumes Spring Vault) — that doc covers the resolver/property-source layer. JVM coupling is **medium**: the API itself is mostly POJO-shaped (templates + operations interfaces + value types like `VaultToken`, `Versioned<T>`, `CertificateBundle`, `Plaintext`/`Ciphertext`), but the bean-config (`AbstractVaultConfiguration`) and reactive stack (Project Reactor, `WebClient`) are JVM-specific. The conceptual model — typed engine operations + pluggable auth + background session/lease renewal — maps cleanly to Go.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **`VaultTemplate` / `VaultOperations`** — generic `read/write/delete/list` plus `opsForKeyValue`, `opsForTransit`, `opsForPki`, `opsForToken`, `opsForSys` accessor methods. | Typed entry point that hides path conventions and JSON marshalling per engine. | Every Go consumer of `hashicorp/vault/api` writes the same `client.Logical().Read(path)` + map-juggling boilerplate per engine. One typed facade per engine is the single largest ergonomic win. |
| 2 | **`ClientAuthentication` interface + ~14 backends** — `TokenAuthentication`, `AppRoleAuthentication`, `KubernetesAuthentication`, `AwsIamAuthentication`, `AwsEc2Authentication`, `AzureMsiAuthentication`, `GcpIamAuthentication`, `GcpComputeAuthentication`, `JwtAuthentication`, `TlsCertAuthentication` (a.k.a. `ClientCertificateAuthentication`), `UserPasswordAuthentication` (userpass/ldap/okta/radius), `CubbyholeAuthentication`, `PcfAuthentication`, `GitHubAuthentication`. | One uniform contract returning a `VaultToken` regardless of mechanism. Login becomes a strategy plug. | `hashicorp/vault/api/auth/*` ships separate sub-packages per method with **inconsistent** APIs (each one a different struct + `Login(ctx, client)` shape). A single `Authenticator` interface that returns `(token, leaseInfo, err)` removes the inconsistency. |
| 3 | **`LifecycleAwareSessionManager`** — background thread renews the session token before TTL expiry, falls back to re-login when the token reaches its max-TTL, revokes `LoginToken`s on shutdown, leaves externally-provided `VaultToken`s alone. | Closes the "token expired silently mid-request" failure mode. | The Go client exposes `LifetimeWatcher` for renewal but every consumer wires the goroutine + restart-on-max-TTL + clean-revoke-on-shutdown by hand. Yarumo's `managed.Component` lifecycle is the natural host. |
| 4 | **`SecretLeaseContainer` + `RequestedSecret` + `LeaseListener`** — declarative "rotate this dynamic credential on a schedule"; events `BeforeSecretLeaseRenewed`, `AfterSecretLeaseRenewed`, `SecretLeaseExpiredEvent`, `SecretNotFoundEvent`, `SecretLeaseCreatedEvent`. `RequestedSecret.rotating(path)` vs `RequestedSecret.renewable(path)` selects rotation semantics. | Dynamic DB credentials, AWS STS, RabbitMQ creds — all leased. Without this, secrets expire in production. | The Go client has lease primitives but **no container** that owns a set of leases and re-fetches on expiry. This is the second-largest ergonomic gap after auth. |
| 5 | **KV v2 versioning** — `VaultVersionedKeyValueOperations.put/get/get(path, Version.from(n))/delete/undelete/destroy/list`, `Versioned<T>` envelope with `Metadata{version, createdAt, deletionTime, destroyed}`, check-and-set (CAS) on write. | KV v2 is the default backend on modern Vault deployments. Version pinning + CAS + soft-delete are first-class needs for config-as-data flows. | `hashicorp/vault/api/kv2.go` exists but is positional and doesn't model `Versioned<T>` cleanly. Typed wrapper with `Get[T any]` and explicit `Version` value type is idiomatic in Go generics. |
| 6 | **Transit engine** — `VaultTransitOperations.createKey/encrypt/decrypt/rewrap/sign/verify/hmac/hash/rotateKey/exportKey`, with binary-safe `Plaintext`/`Ciphertext`/`Signature` value types and `VaultTransitContext` for context+nonce. | Encryption-as-a-service: encrypt PII at the app layer without managing keys locally. Key versioning + rewrap supports zero-downtime key rotation. | Common ask in DaaS (encrypt decision payloads with tenant-scoped keys) and Aluna (encrypt agent memory). Direct `client.Logical().Write("transit/encrypt/key", ...)` works but loses type-safety on the response. |
| 7 | **PKI engine** — `VaultPkiOperations.issueCertificate(role, request)/revoke(serial)/signCertificateRequest/getCrl`, with `VaultCertificateRequest` builder, `CertificateBundle.{createKeyStore, getPrivateKeySpec, getX509Certificate, getX509IssuerCertificate}`. | Dynamic mTLS certs for service-to-service, short-lived. Removes the cert-management burden from each service. | Required by Aluna's agent-to-agent comms if mTLS becomes the default. `crypto/x509` + `client.Logical().Write("pki/issue/role", ...)` works but needs a typed builder + bundle extractor. |
| 8 | **Token operations** — `VaultTokenOperations.create(request)/createOrphan/renew/revoke/revokeOrphan/lookup/lookupSelf`, with `VaultTokenRequest.builder().withPolicy().renewable().ttl().displayName()`. | Apps that themselves create child tokens for downstream services (token sandboxing, scoped delegation). | Niche but standard in service-mesh-adjacent designs. Cheap to wrap once. |
| 9 | **`SslConfiguration` + `KeyStoreConfiguration`** — first-class TLS config: trust store, client cert, PEM support, optional KeyStore. | Vault is almost always TLS-terminated; mTLS to Vault is a common hardening. | Go has `crypto/tls.Config` natively, but loading PEM bundles + optional client cert with sensible defaults is the wrapper layer. Reuse `common/crypto/keys` rather than building anew. |
| 10 | **`AuthenticationSteps` DSL** — composable, declarative login flow (`fromHttpRequest → map → login`). Same flow runs synchronously (`AuthenticationStepsExecutor`) or reactively (`AuthenticationStepsOperator`). | Lets complex auths (AWS IAM signing, JWT exchange) be expressed as data, not procedural code. | In Go this is a function chain anyway, but the **separation of "describe" from "execute"** lets a single auth flow be tested without a real Vault, run with a mock client, or replayed for audit. |
| 11 | **Generic `read/write/delete/list` on `VaultOperations`** — escape hatch for engines without a dedicated operations class (Transform, AWS dynamic, GCP dynamic, Database, RabbitMQ, etc.). | Vault keeps adding engines; the long tail doesn't get a typed wrapper but is still reachable. | Critical for Yarumo: ship typed wrappers for the top 4 (KV, Transit, PKI, Token) and leave the rest behind `vault.Client.Read/Write` exactly the way `hashicorp/vault/api` already exposes `client.Logical()`. |

> Note: the previous draft listed a 12th item — `VaultPropertySource` integration. That's the **Spring Cloud Vault** resolver layer. See the parallel `spring-cloud-vault.md` analysis; it is intentionally out of scope for the `secrets/` driver here.

## 3. Long-tail features (skip)

- **`AbstractVaultConfiguration` / `AbstractReactiveVaultConfiguration`** bean-config base classes — JVM `@Configuration` lifecycle, N/A.
- **Reactive stack** (`ReactiveVaultTemplate`, `ReactiveSessionManager`, `WebClient` + Reactor Netty) — Go's `context.Context` + goroutines covers the same need without a separate API; do not ship a sibling "reactive" type.
- **`Vault Repositories`** (Spring Data repository abstraction over Vault) — anti-pattern in Go; storage repositories don't generalise across engines cleanly. Skip.
- **`@VaultPropertySource` annotation** — annotation magic; if a consumer wants property-source semantics, that lives in the Spring Cloud Vault analysis on top of `modules/config`, not in `modules/secrets/`.
- **XML namespace config** — JVM-only.
- **Spring Security integration** (`VaultAuthenticationProvider`, `VaultAuthenticationManager`) — auth-into-the-app via Vault is niche; pairs with `modules/auth/` (§ 1.2), not `modules/secrets/`. Defer.
- **`SimpleSessionManager`** (no-renew variant) — only used for one-shot scripts; the lifecycle-aware manager covers 100% of long-running apps.
- **`Cubbyhole` response unwrapping** as a flagship auth — useful only with Vault's response-wrapping tokens; can be added later as one more `Authenticator` impl.
- **`AppId` authentication** — deprecated in Vault itself, skip.
- **`PcfAuthentication`** — Pivotal Cloud Foundry; effectively dead platform.
- **`GitHubAuthentication`** — niche.
- **`Transform` engine (Enterprise)** — Vault Enterprise feature, gated by license; reach via generic `read/write` if a consumer needs it.
- **`KeyStore` materialisation** from PKI bundle — Java-specific format; Go consumers want `*tls.Certificate` directly.
- **Native image / GraalVM hints** — N/A.
- **`ClientHttpRequestFactory` abstraction over Java HTTP clients** — Go uses `net/http.Client`; the abstraction collapses.
- **`VaultEndpointProvider` (multi-endpoint failover)** — Vault clusters are usually fronted by a load balancer; defer until a consumer needs client-side failover.
- **Sys backend operations** (`opsForSys.health/seal/unseal/initialize/policy/audit`) — operator concerns, not app concerns; consumers can reach via generic `Read/Write`. Skip in v1.
- **`EnvironmentVaultConfiguration`** property-driven bootstrap — overlaps with Spring Cloud Vault and `modules/config`; not in scope for `modules/secrets/`.

## 4. Mapping to Yarumo

> **Roadmap context (post-cleanup, 2026-05-16)**: the prior § 3 brainstorm section (with `modules/secrets/` at § 3.3) was deleted along with Annexes A/B. The current roadmap covers only § 1 modules (datasource/auth/messaging/health/boot), § 2 routegen, and § 4 migration tracking. As a result `secrets/` has no landing zone in the current roadmap — this analysis **proposes it as a new top-level module** (see § 6 and § 8).

**Existing § 1 modules with overlap**:

- **`modules/common/crypto/`** — closed in Phase 1 (milestone #11). TLS material loading already lives in `common/crypto/keys`; `secrets/vault/` should consume it for the `SslConfiguration` equivalent, not duplicate. `common/crypto/tokens` covers JWT validation that `JwtAuthentication` parallels (but Vault's JwtAuth submits a token, doesn't validate one — different direction).
- **`modules/common/resilience/`** (closed YA-0076) — `CircuitBreakerRegistry` + `RateLimiterRegistry`. Vault calls go behind a circuit breaker; lease-renewal goroutines use retry with backoff from the same module.
- **`modules/common/health/`** (closed YA-0077) — leaf interfaces. The Vault driver can expose a `Checker` that pings `sys/health` and feeds into the future `modules/health/` runtime (§ 1.4).
- **`modules/config/`** (Phase 3, active) — viper-driven bootstrap. The Spring Vault driver itself does **not** depend on `config/`. The "vault as a property source" feature is the Spring Cloud Vault resolver layer and is analysed in `spring-cloud-vault.md`; if that lands, it composes `modules/secrets/vault/` with `modules/config/` from the outside.
- **`modules/managed/`** (Phase 3, active) — `Session` and `LeaseContainer` are textbook `managed.Component` candidates (Start/Stop/Done lifecycle, own goroutines, must revoke on shutdown).
- **`modules/auth/`** (§ 1.2, planned) — only loosely related. Auth-**into**-the-app via Vault tokens is a `modules/auth/` concern (`VaultAuthenticationProvider` equivalent). Auth-**to**-Vault (AppRole, Kubernetes, etc.) is a `modules/secrets/vault/` concern. Different directions, different modules.
- **`modules/telemetry/otel/`** (Phase 3, active) — optional `WithObservation()` decorator emits spans for Vault calls (path, engine, lease id).

**Gaps to fill** (none of these have a home in the current roadmap):

- **Provider abstraction** — vendor-agnostic interface (`Get`, `Subscribe`, `Close`) so AWS Secrets Manager / GCP Secret Manager / Doppler can slot in alongside Vault later.
- **Typed engine facade** — wrap `hashicorp/vault/api`'s `client.Logical()` map-shaped responses with `kv.Get[T any](ctx, path, version) (Versioned[T], error)`, `transit.Encrypt(ctx, key, plaintext) (Ciphertext, error)`, `pki.Issue(ctx, role, request) (*CertificateBundle, error)`.
- **Unified `Authenticator` interface** — collapse the inconsistent per-method APIs in `hashicorp/vault/api/auth/*` into one contract: `Authenticator.Login(ctx, *vault.Client) (*Token, error)` with implementations `TokenAuth`, `AppRoleAuth`, `KubernetesAuth`, `AwsIamAuth`, `AwsEc2Auth`, `AzureMsiAuth`, `GcpIamAuth`, `GcpComputeAuth`, `JwtAuth`, `TlsCertAuth`, `UserPassAuth` (userpass/ldap/okta/radius), `CubbyholeAuth`.
- **`Session` as a `managed.Component`** — owns the background renewal goroutine, exposes `Token() string` that always returns a non-expired token, revokes on `Stop()`. Replaces ad-hoc `LifetimeWatcher` wiring.
- **`LeaseContainer` as a `managed.Component`** — registers `RequestedSecret`s, fetches them lazily, renews/rotates them on schedule, emits events on a channel (Go's `chan LeaseEvent` instead of Spring's `ApplicationEventPublisher`).
- **Caching layer with TTL** — TTL cache decorator over the top-level `Provider`. Will need `modules/cache` once that exists (YA-0162 / YA-0079 backlog); until then ship a minimal in-package cache.

**Anti-patterns to avoid**:

- **No `AbstractVaultConfiguration` bean-config base class** — there is no DI in Yarumo. Construction is `vault.NewClient(opts...)`, `vault.NewSession(client, authenticator, opts...)`, `vault.NewLeaseContainer(session, opts...)`.
- **No `@VaultPropertySource` / no auto-injection** — secrets are fetched via explicit calls, not magically materialised into struct fields. Property-source semantics are Spring Cloud Vault's job and stay out of the driver.
- **No reactive sibling API** — `ReactiveVaultTemplate` is justified in Java because blocking IO is expensive on the JVM; Go's goroutines + `context.Context` already deliver the same outcome with one API.
- **No Spring Data repository pattern on Vault** — engines are too heterogeneous; "find by ID" doesn't generalise across KV/Transit/PKI/database.
- **No god-struct** — `VaultTemplate` aggregates everything by accident-of-history (`opsForX` accessor methods). Yarumo splits into `secrets/vault/kv/`, `secrets/vault/transit/`, `secrets/vault/pki/`, `secrets/vault/token/` — one engine per sub-package, all sharing one `*vault.Client`.
- **No `Versioned<T>` carrying Spring's full metadata bag** — keep `Versioned[T]` minimal: `{Value T, Version int, CreatedAt time.Time, Destroyed bool}`. Drop nullable JVM idioms.
- **No catch-all `SimpleSessionManager` parallel** — the lifecycle-aware session is the only session type. Short-lived scripts construct the session, call `Stop()`, done.
- **No fixed bean-init order** — the boot-order problem that bit go-feather-lib's `boot/` doesn't appear here because `secrets/vault/` is a library, not an orchestrator. Wiring stays in the consumer.

## 5. Recommendation

**PARTIAL** — adopt the conceptual model (typed engine facades + uniform authenticator + lifecycle-aware session + lease container with events) but rewrite idiomatically in Go on top of `github.com/hashicorp/vault/api`. The Pareto cut is ~11 features that cover 95% of real Vault use.

Because the roadmap no longer carries a `modules/secrets/` placeholder, this analysis **proposes a new top-level module** with Vault as the **anchor sub-driver** (its operations surface — Transit, PKI, dynamic secrets, leases — is far richer than AWS Secrets Manager / GCP Secret Manager / Doppler, so its design constraints should shape the `Provider` interface). The core `Provider` interface stays minimal (`Get(ctx, key) (Secret, error)` + `Subscribe(ctx, key) (<-chan SecretEvent, error)`) and the rich Vault-specific surface lives **inside** the vault sub-driver as concrete types, not behind the `Provider` interface. `Session` and `LeaseContainer` are `managed.Component`s, so app shutdown revokes tokens cleanly. Spring Cloud Vault's "vault as a property source" feature is **explicitly out of scope** — it composes on top via `modules/config` if a consumer ever needs it, and is analysed in the parallel `spring-cloud-vault.md`.

## 6. Proposed yarumo placement

**NEW top-level module**: `modules/secrets/` (anchor sub-driver: `modules/secrets/vault/`)

**Why a new module** (per the Placement principle in `ROADMAP_NEW_MODULES.md`):

- Has lifecycle (`Session` and `LeaseContainer` own renewal goroutines; Start/Stop/Done) → not `common/`.
- Pulls external SDKs (`hashicorp/vault/api`, future `aws-sdk-go-v2/secretsmanager`, etc.) → not `common/`.
- Not a one-shot bootstrap → not `config/`.
- Not pure observability → not `telemetry/`.
- Not application wiring → not `boot/`.
- Distinct domain (secrets management) from `auth/` (which authenticates principals into the app, opposite direction).

Therefore: own module, sibling to `datasource/`, `auth/`, `messaging/`, `health/`.

**Subpackages**:

```
modules/secrets/
  doc.go
  errors.go                 Domain errors (errs.TypedError):
                            ErrSecretNotFound, ErrAccessDenied, ErrProviderUnavailable,
                            ErrLeaseExpired, ErrAuthFailed.
  provider.go               Provider interface (vendor-agnostic):
                              Get(ctx, key) (Secret, error)
                              GetVersioned(ctx, key, version int) (Secret, error)  // optional capability
                              Subscribe(ctx, key) (<-chan SecretEvent, error)
                              Close() error
  secret.go                 Secret value type {Data map[string]any, Version int, Lease *LeaseInfo}
  cache.go                  TTL cache decorator over Provider (uses modules/cache once available)
  vault/                    HashiCorp Vault sub-driver (anchor impl)
    doc.go
    errors.go
    client.go               NewClient(opts...) wraps *vault.Client from hashicorp/vault/api
    options.go              Options pattern: WithEndpoint, WithTLS, WithNamespace,
                            WithHTTPClient, WithObservation.
    session.go              Session (managed.Component): owns token + renewal goroutine
    authenticator.go        Authenticator interface; Login(ctx, *vault.Client) (*Token, error)
    auth/
      token.go              TokenAuth (static token)
      approle.go            AppRoleAuth (roleID + secretID, push/pull, wrapped secretID)
      kubernetes.go         KubernetesAuth (service-account JWT)
      aws.go                AwsIamAuth + AwsEc2Auth
      azure.go              AzureMsiAuth
      gcp.go                GcpIamAuth + GcpComputeAuth
      jwt.go                JwtAuth (OIDC / external JWKS)
      tls.go                TlsCertAuth (mTLS to Vault)
      userpass.go           UserPassAuth (username/password — userpass/ldap/okta/radius)
      cubbyhole.go          CubbyholeAuth (wrapped-response unwrap)
    kv/
      v1.go                 KV v1: Get/Put/Delete/List
      v2.go                 KV v2: Get[T]/Put[T]/Delete/Undelete/Destroy/List
                            with Versioned[T] envelope and CAS.
    transit/
      operations.go         CreateKey, Encrypt, Decrypt, Rewrap, Sign, Verify, HMAC, RotateKey
      types.go              Plaintext, Ciphertext, Signature, KeyType, KeyConfig
    pki/
      issue.go              Issue, Revoke, Sign, GetCRL
      bundle.go             CertificateBundle{Cert, PrivateKey, CA, Chain}
                            with TLSCertificate() (*tls.Certificate, error)
      request.go            CertificateRequest builder
    token/
      operations.go         Create, CreateOrphan, Renew, Revoke, Lookup, LookupSelf
    lease/
      container.go          LeaseContainer (managed.Component): owns lease set, runs renew loop
      requested.go          RequestedSecret = {Path, Mode (renewable|rotating), MinRenewal, ExpiryThreshold}
      event.go              LeaseEvent {Kind, Path, Lease, Error};
                            Kind in {Created, BeforeRenew, AfterRenew, Expired, NotFound}
  aws/secretsmanager/       (Future) Provider impl backed by AWS Secrets Manager
  gcp/secretmanager/        (Future) Provider impl backed by GCP Secret Manager
  doppler/                  (Future) Provider impl backed by Doppler
```

**Internal deps**:

- `modules/common/errs` — domain errors.
- `modules/common/log/slog` — structured logging.
- `modules/common/resilience` — circuit breaker around Vault calls; retry+backoff inside the renewal loop.
- `modules/common/crypto/keys` — TLS material loading (PEM, certificate chain).
- `modules/common/health` (leaf) — `Checker` impl for `sys/health`, surfaced into `modules/health/` runtime later.
- `modules/managed` — `Session` and `LeaseContainer` are `Component`s (Start/Stop/Done).
- `modules/cache` (planned / YA-0162) — TTL cache decorator for the top-level `Provider`. Until that lands, ship a minimal in-package cache.
- `modules/telemetry/otel` (optional) — `WithObservation()` emits spans for Vault calls.

**Go libraries to wrap** (mature, with repo URL):

- `github.com/hashicorp/vault/api` (https://github.com/hashicorp/vault/tree/main/api) — **official** Go client. Stable. Carries `Logical()`, `Auth()`, `Sys()`, `KVv1()`, `KVv2()`, and the `LifetimeWatcher` for token/lease renewal. Foundation everything else wraps.
- `github.com/hashicorp/vault/api/auth/approle` (https://github.com/hashicorp/vault/tree/main/api/auth/approle), `.../auth/kubernetes`, `.../auth/aws`, `.../auth/azure`, `.../auth/gcp`, `.../auth/userpass`, `.../auth/ldap`, `.../auth/cert`, `.../auth/jwt` — per-method login helpers. Wrap behind one `Authenticator` interface.
- `github.com/hashicorp/vault-client-go` (https://github.com/hashicorp/vault-client-go) — **BETA, do not use yet**. Generated from OpenAPI. Track for v2; for v1 stay on the stable `vault/api`.

**Out of scope for v1**:

- Reactive sibling API (single `context.Context`-driven API only).
- Spring Data repository pattern.
- `Sys` operations (init/seal/unseal/policy management) — reach via raw `client.Sys()` if needed.
- `Cubbyhole` response unwrapping as a flagship auth — ship as one of many `Authenticator` impls, no special wiring.
- Vault Enterprise features (`Transform`, namespaces beyond simple `WithNamespace`, performance replication).
- Multi-endpoint client-side failover (`VaultEndpointProvider`) — assume a load balancer in front.
- `PropertySource`-style integration with `modules/config` — Spring Cloud Vault's job; tracked in `spring-cloud-vault.md`.
- `modules/auth/` integration (auth-into-the-app via Vault tokens) — separate module, separate concern.
- AWS/GCP/Doppler sub-drivers — ship after `vault/` is in production for one consumer; the `Provider` interface is shaped by Vault first.

## 7. Open questions

1. **`Provider` interface granularity** — does the vendor-agnostic `Provider` expose just `Get`/`Subscribe`, or also `Versioned` and `Lease` capabilities? Two options:
   - **Minimal core** (`Get` + `Subscribe`) and rich types behind the concrete driver (Vault has more surface than AWS Secrets Manager anyway).
   - **Capability interfaces** (`Provider`, `VersionedProvider`, `LeasedProvider`) with consumers type-asserting. Picking 1 vs 2 changes the consumer experience significantly.
2. **Lease event channel vs callback** — Spring uses `ApplicationEventPublisher` + listener interfaces. Go-idiomatic options: `<-chan LeaseEvent` (one consumer), `Subscribe(func(LeaseEvent))` (multi-consumer), or both. Pick before locking the `LeaseContainer` API.
3. **`Versioned[T]` vs untyped `Versioned`** — generic KV `Get[T any]` requires a marshaller decision (encoding/json default, pluggable?). Untyped `Versioned[map[string]any]` is simpler but loses static typing. Spring keeps both APIs side-by-side; Yarumo should pick one default.
4. **PKI bundle output** — Java returns `KeyStore` + raw fields. Go natural form is `*tls.Certificate` + `*x509.Certificate` + `*rsa.PrivateKey`/`*ecdsa.PrivateKey`. Decide whether the bundle exposes a `TLSCertificate()` helper or raw fields only.
5. **`vault-client-go` migration window** — the generated OpenAPI client is BETA today (2026-05-16). When it goes GA, do we migrate, dual-target, or stay on `vault/api`? Track HashiCorp's GA announcement; design the internal client wrapper so the swap is one-file.
6. **Caching defaults** — for dynamic secrets with their own lease TTL, the cache TTL must be `min(configured_ttl, lease_ttl - renewal_threshold)`. Confirm the cache decorator can subtract from the lease TTL, not just respect a static config.
7. **Vault dev / test story** — `testcontainers-go` has a Vault module. Decide whether the testcontainer helper lives inside `secrets/vault/` or in a future `modules/testing/containers/`.
8. **Observability scope** — pick the OTel semantic-convention attribute set (`vault.path`, `vault.engine`, `vault.lease_id`, `vault.namespace`) before wiring `WithObservation()`.
9. **Multi-tenant Vault namespaces** — Vault Enterprise namespaces matter for tenancy. Decide whether `vault.Client` carries a single namespace or per-call `WithNamespace(ns)` is supported. Single-namespace is simpler; per-call is required for cross-tenant flows.
10. **AppRole secret-id rotation** — AppRole `SecretID` itself is a short-lived credential. Some deployments rotate it via a wrapped response (Cubbyhole). Does the `AppRoleAuth` authenticator need a "secret-id resolver" hook (a `func(ctx) (string, error)` that talks to a wrapping endpoint), or is a static secret-id sufficient for v1?
11. **Placement in the roadmap** — should `modules/secrets/` land in Phase 3 (alongside config/managed/telemetry, since it depends heavily on `managed`), or wait for a new phase after the § 1 modules track? See § 8.

## 8. ROADMAP delta proposed (NOT applied)

Add to `docs/ROADMAP_NEW_MODULES.md` § 1 (new modules):

```
## 1.6. `modules/secrets/` — Secret retrieval, rotation, and engine access

**Status**: Planned
**Why a new module**: stateful (sessions own renewal goroutines, lease containers own background renewal). External SDKs per backend. Distinct from `auth/` (auth-INTO-app) and `config/` (one-shot bootstrap).
**Granularity**: vendor-agnostic core + one sub-driver per backend; each driver independently shippable. Vault is the anchor (richest surface — KV, Transit, PKI, leases — drives the core API shape).

**Core (vendor-agnostic)**:
- `Provider` interface (`Get`, `Subscribe`, `Close`). Capability interfaces for `Versioned` / `Leased` TBD (open question).
- `Secret` value type, `LeaseInfo`, `SecretEvent`.
- TTL cache decorator (uses `modules/cache` once available).

**Sub-drivers planned within `secrets/`**:

| Subpackage | Backend | Priority | Status |
|---|---|---|---|
| `modules/secrets/vault/` | HashiCorp Vault — KV v1/v2, Transit, PKI, Token, dynamic leases, ~11 auth backends | High (anchor) | Planned |
| `modules/secrets/aws/secretsmanager/` | AWS Secrets Manager | Low | Planned (later) |
| `modules/secrets/gcp/secretmanager/` | GCP Secret Manager | Low | Planned (later) |
| `modules/secrets/doppler/` | Doppler | Brainstorm | Speculative |

**Internal deps**: `modules/common/{errs,log/slog,resilience,crypto/keys,health}`, `modules/managed`, future `modules/cache`, optional `modules/telemetry/otel`.

**Cross-module relationships**:
- **Pairs with `modules/auth/` (§ 1.2)** — opposite direction. `auth/` authenticates principals INTO the app; `secrets/` authenticates the app TO a secrets backend.
- **Pairs with `modules/config/`** via the Spring Cloud Vault resolver layer (see `docs/spring/spring-cloud-vault.md`) — not via the secrets driver itself.
- **Pairs with `modules/health/` (§ 1.4)** — `secrets/vault/` exposes a `health.Checker` that pings `sys/health`.

**Reference analyses**: `docs/spring/spring-vault.md` (this doc, the client-library equivalent), `docs/spring/spring-cloud-vault.md` (the autoconfig/property-source equivalent — out of scope for the driver).
```

Add a row to § 4.1 migration tracking only if any go-feather-lib code maps here (likely none — Vault was not in go-feather-lib).

**Phase placement options** (pick one, file as a follow-up ticket):

- **Option A — extend Phase 3**: add `modules/secrets/` as a stretch goal of milestone #9. Pro: depends heavily on `managed/`, which is already in flight. Con: scope creep on a phase already at 20 issues.
- **Option B — new Phase 5 "Modules: Domain (auth/secrets/datasource/messaging/health/boot)"**: own milestone for the § 1 module track. Pro: clean scope. Con: long-lived milestone.
- **Option C — file as standalone tickets unmilestoned** until a consumer shows up: keep the § 1 module track free-floating, file `modules/secrets/` tickets only when DaaS or Aluna asks for Vault. Pro: avoids speculative work. Con: design can drift without a deadline.

Recommended: **Option C** until a real consumer surfaces (DaaS encrypting payloads via Transit, or Aluna issuing short-lived mTLS via PKI). Until then this analysis stays as a design doc.
