# Spring Cloud Vault — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-vault
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Cloud Vault 5.0.1 (with 5.0.2-SNAPSHOT in flight) is the **autoconfig / property-source** layer that sits on top of Spring Vault. Where Spring Vault is a pure client library (templates, operations, session, lease container — analysed separately in [spring-vault.md](./spring-vault.md)), Spring Cloud Vault is the Spring Boot integration that makes secrets behave **as if they were ordinary configuration**: `spring.config.import: vault://` (Spring Boot 2.4+ ConfigData API, replaced the deprecated `bootstrap.yml` mode as of Spring Cloud Vault 3.0) mounts Vault paths as property sources, and `@Value("${spring.datasource.password}")` resolves transparently from a Vault path. Three layered concerns dominate:

1. **Property-source integration** — Vault paths become Spring `PropertySource`s; secrets show up as ordinary properties (`spring.datasource.username`, `spring.rabbitmq.password`, `cloud.aws.credentials.accessKey`) at boot time.
2. **Per-engine autoconfig** — each Vault secret engine (KV v1/v2, Database, AWS, RabbitMQ, Consul, Couchbase, Elasticsearch, MongoDB-deprecated, PostgreSQL-deprecated, MySQL-deprecated, Cassandra-deprecated) ships a dedicated module that knows the engine's path conventions, lease semantics, and target property names. The generic Database backend is the canonical one (Postgres / MySQL / Mongo / Cassandra / Couchbase / Elasticsearch + a generic plugin backend), with multi-database support since 3.0.5 (`spring.cloud.vault.databases.{name}`).
3. **Lease lifecycle in the property-source context** — a `SecretLeaseContainer` (inherited from Spring Vault) is wired automatically; renewals push **updated property values** into the Spring `Environment` so beans that read them stay current. The renewal scheduler exposes `min-renewal: 10s` (floor), `expiry-threshold: 1m` (renew this long before expiry), `lease-endpoints` (Legacy pre-0.8 vs SysLeases 0.8+), and `lease-strategy` (`DropOnError` / `RetainOnError` / `RetainOnIoError`). Default `lease-strategy` is `DropOnError`; `RetainOnIoError` is the recommended prod setting. Sessions have their own lifecycle knobs: `refresh-before-expiry: 10s`, `expiry-threshold: 20s` — lazy login on first use, re-login on expiry, revoke on shutdown.

A reactive variant exists and is **enabled by default** (`spring.cloud.vault.reactive.enabled=true`) when Project Reactor is on the classpath; a handful of authenticators (AppRole pull, GCP IAM) fall back to synchronous. Discovery integration (`spring.cloud.vault.discovery.service-id`, default `vault`), Vault Enterprise namespaces (`X-Vault-Namespace` header), and a custom-backend extension hook (`VaultSecretBackendDescriptor` + `SecretBackendMetadataFactory` via `spring.factories`) round out the surface.

Documented quote that anchors the scope: *"Spring Cloud Vault maintains a lease lifecycle beyond the creation of login tokens and secrets … login tokens and secrets associated with a lease are scheduled for renewal just before the lease expires until terminal expiry. Application shutdown revokes obtained login tokens and renewable leases."* And the well-known limitation: *"Spring Cloud Vault does not support getting new credentials and configuring your `DataSource` with them when the maximum lease time has been reached."* — when `max_ttl` is hit, the app must restart.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Bootstrap-time secret resolution into config keys** — Vault paths are mounted as `PropertySource`s during `ConfigData` phase; `${spring.datasource.password}` resolves from Vault before any bean is built. | Single mental model: secrets are just config. No `vault.Get(...)` sprinkled through the codebase. | Every Go service hand-rolls "read this env var OR call Vault if X is set" branching. A Vault-aware resolver that feeds Viper's key namespace at boot collapses that to one line. **This is the entire point of Spring Cloud Vault.** |
| 2 | **Per-engine property mapping** — Database backend writes `spring.datasource.username` / `spring.datasource.password`; AWS backend writes `cloud.aws.credentials.accessKey` / `cloud.aws.credentials.secretKey` / `cloud.aws.credentials.sessionToken`; RabbitMQ writes `spring.rabbitmq.username` / `spring.rabbitmq.password`; Consul writes `spring.cloud.consul.token`; Couchbase writes `spring.couchbase.username/password`; Elasticsearch writes `spring.elasticsearch.rest.username/password`. Each backend exposes `*-property` knobs to override target keys. | Standard secret engines feed standard config keys without per-engine glue code. The consumer's existing `spring.datasource.*` keeps working. | In Yarumo this is *the* connector between the driver layer (proposed NEW `modules/secrets/vault/` — see spring-vault.md) and the resolver layer (`modules/config/sources/secrets/`). A `Binding` value type lets the database driver in `modules/datasource/gorm/` find its credentials at `spring.datasource.password` without knowing about Vault. |
| 3 | **`spring.config.import: vault://` ConfigData URI scheme** — composable URI form, supports multiple paths (`vault://my/path, vault://other?prefix=foo.`), optional locations (`optional:vault://path`), prefix per-mount. | Declarative composition of multiple Vault paths into one property tree, with per-mount key prefixing for namespace separation. | Translates directly to Viper's `RemoteProvider` pattern — Yarumo's existing `modules/config/` (§ 1 of ROADMAP_NEW_MODULES.md) already has the seams for additional sources. The URI form is a clean DSL: `secrets://vault/path?prefix=db.&optional=true`. |
| 4 | **Multi-database support** (`spring.cloud.vault.databases.{name}` since 3.0.5) — each named database has its own role, backend, and `{username,password}-property` keys, feeding distinct connection pools. | Apps with primary + replica + analytics each need different credentials with different lease TTLs. One Vault role per pool. | Yarumo's `modules/datasource/` (§ 1.1 of ROADMAP_NEW_MODULES.md) plans per-driver subpackages; multi-binding in `modules/config/sources/secrets/` is the natural complement. |
| 5 | **Lease lifecycle wired into property-source updates** — when a database lease is renewed, the new username/password are pushed back into the Spring `Environment`. Connection pool sees fresh credentials on the next checkout. | Closes the loop between "Vault rotated my credentials" and "my running app uses them". Without this, rotation requires either a restart or per-app glue. | The hardest part of secrets infra and the place where Yarumo can win the most. Requires the renewal goroutine in the proposed `modules/secrets/vault/` to publish updates through `secrets.Provider.Subscribe(...)`; `modules/config/sources/secrets/` consumes those events and applies them to the config registry. |
| 6 | **`lease-strategy` enum** (`DropOnError` / `RetainOnError` / `RetainOnIoError`) — explicit policy for what happens when a renewal call fails. Default is `DropOnError` but `RetainOnIoError` is the recommended prod setting: retain on transient network failure, drop on auth/permission failure. | The naive "renew or die" approach causes outages during routine Vault hiccups. The naive "retain forever" approach uses expired credentials. The hybrid is the right answer. | Direct port. Three-value `LeaseStrategy` enum on the `LeaseContainer` from spring-vault.md, surfaced as an option in `modules/config/sources/secrets/`. Coordinates with `modules/common/resilience/` (circuit breaker around Vault calls). |
| 7 | **`min-renewal` + `expiry-threshold`** — `min-renewal: 10s` prevents renewal storms by setting a lower bound on how often any single lease is renewed; `expiry-threshold: 1m` schedules renewal that long before expiry. | Both knobs avoid pathological scheduling. Without `min-renewal`, a misconfigured short TTL hammers Vault. Without `expiry-threshold`, renewals race the expiry deadline. | Direct port. Two `time.Duration` options on the resolver's lifecycle config, plumbed to the driver's `LeaseContainer`. |
| 8 | **`fail-fast: true`** — boot fails if Vault is unreachable. Surface configuration / network problems immediately instead of starting with missing secrets. | Pairs perfectly with `modules/health/` (§ 1.4) startup probes. Default should match: fail loudly at boot, not silently in production. | Direct port. `WithFailFast(bool)` option on the config loader. Recommended Yarumo default is `true` (Spring defaults to `false` but the rest of the doc treats fail-fast as the prod-correct setting). |
| 9 | **ConfigData "optional" prefix** (`optional:vault://`) — Vault path missing → skip silently. | Lets a single config layout work across environments where Vault is sometimes present (prod) and sometimes not (local dev with mocks). | Same problem in Go. Viper's `MergeInConfig` already tolerates missing files; the Vault resolver needs the same "best-effort" mode via `optional:secrets://...` URI prefix. |
| 10 | **`prefix=` per-mount** (`vault://db?prefix=primary.`) — mounted keys are prefixed in the target namespace so multiple Vault paths can feed disjoint config subtrees. | Critical when one app pulls from N engines / N roles — keeps `db.password` from one path from colliding with `db.password` from another. | Direct port. The Vault loader's per-mount config takes a `KeyPrefix string` field surfaced via the URI's `?prefix=` query parameter. |
| 11 | **Custom backend extension** (`VaultSecretBackendDescriptor` + `SecretBackendMetadataFactory`, registered via `spring.factories`) — third-party engines (Transform, custom plugins) can integrate with the property-source pipeline without modifying Spring Cloud Vault itself. | Future-proof: Vault adds new engines regularly, and apps run third-party engines (Snowflake, Databricks, etc.). | Worth keeping the seam. A `Binding` interface in `modules/config/sources/secrets/`: name + raw secret map → target config keys. Consumers register their own bindings; the resolver ships KV, Database, AWS, RabbitMQ, Consul, Couchbase, Elasticsearch. |
| 12 | **Session lifecycle (lazy login, re-login on expiry, revoke on shutdown)** — `session.lifecycle` knobs `refresh-before-expiry: 10s`, `expiry-threshold: 20s`. Tokens are obtained lazily on first use; re-login on expiry; auto-revocation at shutdown. | Login is amortised away (no eager auth at boot). Revoke-on-shutdown is the right hygiene for short-lived workloads. | Direct port. Lives in the driver layer (`modules/secrets/vault/session.go`) as a `managed.Component`; the resolver inherits the behaviour transparently. |

## 3. Long-tail features (skip)

- **Reactive variant** (`ReactiveVaultConfig`, `spring.cloud.vault.reactive.enabled`) — same reasoning as spring-vault.md: Go's `context.Context` + goroutines collapse the imperative/reactive split. Single API.
- **Bootstrap context** (`bootstrap.yml`, `spring-cloud-starter-bootstrap`) — legacy mode predating ConfigData (pre-3.0). Skip; Yarumo only needs the modern single-source model.
- **`DiscoveryClient` integration** (`spring.cloud.vault.discovery.service-id`) — relies on Eureka/Consul service registry, which Yarumo does not adopt. K8s DNS + load balancer covers the same need. Skip.
- **Deprecated per-engine modules** (`spring-cloud-vault-config-mysql`, `-postgresql`, `-mongodb`, `-cassandra`) — these were absorbed by the generic `database` backend in Vault itself (since Vault 0.7.1). Yarumo only implements the generic shape.
- **`AppId` authentication** — deprecated in Vault. Skip.
- **`PCF` (Pivotal CloudFoundry) authentication** — dead platform. Skip.
- **XML namespace config / `@ConfigurationProperties`** — JVM-only.
- **Database `username-property` defaulting heuristic** — every backend has its own defaults (`spring.datasource.username` etc.). Yarumo makes mappings explicit; no defaults that depend on classpath guessing.
- **Cookie/keystore auto-loading conveniences** — handled by `modules/common/crypto/keys` already; nothing engine-specific.
- **`@VaultPropertySource` annotation** — already declared out-of-scope in spring-vault.md. Same here.
- **`SecretBackendMetadataFactory.spring.factories` discovery** — Yarumo prefers explicit registration over service-loader-style discovery. The extension seam stays, but the registration is a function call.
- **`VAULT_NAMESPACE` header for Enterprise namespaces** — kept in spring-vault.md's `WithNamespace` option; no additional surface needed at the Cloud Vault layer.
- **`lease-endpoints: Legacy` mode** — Vault pre-0.8 lease endpoints. Only ship `SysLeases` (0.8+). Documented as default, no knob.
- **`VAULT_AGENT` mode as a flagship feature** — useful but the right place is one more `Authenticator` impl in `modules/secrets/vault/auth/agent.go`, not a separate config-layer code path.

## 4. Mapping to Yarumo

This module is the **bootstrap-time secrets resolver** layer. The split with spring-vault.md is now sharper because `modules/secrets/` is itself a proposal in the cleaned-up roadmap (the deleted § 3 brainstorm is gone, so the driver lands as a NEW top-level module — see spring-vault.md for the proposal).

| Concern | Lives in proposed NEW `modules/secrets/vault/` (driver) | Lives in existing `modules/config/` (resolver) |
|---|---|---|
| Vault HTTP client, auth, session, lease container | ✅ (spring-vault.md) | ❌ |
| Typed engine ops (KV, Transit, PKI, Token) | ✅ | ❌ |
| `Authenticator` interface + impls | ✅ | ❌ |
| `LeaseContainer` lifecycle (renew/rotate/revoke) | ✅ | ❌ |
| **Mounting Vault paths into Viper's key namespace** | ❌ | ✅ (this analysis) |
| **`spring.config.import: vault://`-equivalent URI scheme** | ❌ | ✅ |
| **Per-engine property mapping** (`Database` → `spring.datasource.*`, `AWS` → `cloud.aws.credentials.*`) | ❌ | ✅ |
| **`fail-fast`, `optional`, `prefix` boot semantics** | ❌ | ✅ |
| **Pushing rotated credentials back into config registry** | Driver emits `LeaseEvent`s via `Provider.Subscribe(...)` | Resolver subscribes, updates Viper |
| **Multi-binding for multi-database apps** | ❌ | ✅ |

**Existing § 1 modules with overlap**:

- **`modules/config/`** (existing, § 1 of ROADMAP_NEW_MODULES.md sits adjacent — config is one of the Phase 3 modules in milestone #9). Already viper-driven, one-shot bootstrap. Natural home for the resolver layer: gains a new `sources/secrets/` subpackage. The placement principle in the roadmap puts "Bootstrap (one-shot)" in `modules/config/`; the secrets resolver is exactly that shape.
- **`modules/datasource/` (§ 1.1, planned — gorm, mongo, goredis, gocql, ldap, vector)** — consumers. They keep reading credentials from the existing config keys (`spring.datasource.password`-equivalent); the secrets resolver is transparent to them. Future hook: a `RotationListener` interface in `modules/datasource/` so the gorm/mongo/redis drivers can refresh their connection pools on credential rotation — pairs with the row-level audit hooks already planned in § 1.1.
- **`modules/managed/`** (Phase 3 milestone #9) — the resolver owns a subscription goroutine and is therefore a `managed.Component` (Start/Stop/Done).
- **`modules/common/resilience/`** (closed YA-0076) — circuit breaker around Vault calls; same as in spring-vault.md.
- **`modules/health/` (§ 1.4, planned)** — `fail-fast` at boot dovetails with the runtime health endpoints; once both ship, the resolver's load failure can also surface on `/readyz`.
- **`modules/boot/` (§ 1.5, planned)** — the orchestrator that calls `config.Default(...)` and starts `managed.Component`s. The resolver's lifecycle is wired through `BeanFn`s exactly like any other source.

**Driver-level features** (cross-reference spring-vault.md): live in the proposed NEW `modules/secrets/vault/` module that spring-vault.md is asking the roadmap to add. The resolver in `modules/config/sources/secrets/` depends on the vendor-agnostic `secrets.Provider` interface (`Get` + `Subscribe`) exposed by that new module — *not* on Vault directly. AWS Secrets Manager, GCP Secret Manager, and Doppler plug in symmetrically once they ship as sub-drivers.

**Gaps to fill**:

- **No standard Go pattern for "boot-time secret resolution into config keys"**. Every shop reinvents one of: (a) `envconsul`-style template rendering, (b) custom Viper remote provider, (c) ad-hoc `os.Getenv` with Vault fallback. A first-class `config/sources/secrets/` removes the re-invention.
- **No standard Go pattern for "secret rotation propagated back into config"**. The handful of libraries that exist (e.g. `hashicorp/consul-template` for templating, `aws-secretsmanager-caching-go` for caching) don't push values into a shared config registry.
- **No standard Go pattern for "multi-binding"** — one Vault role per pool, each feeding distinct config keys.
- **No standard Go pattern for "bootstrap-time secret → datasource refresh contract"** — the gap between "credential rotated" and "connection pool sees it" is bridged manually in every Go shop. Yarumo can publish a clean `RotationListener` interface that the datasource drivers implement.

**Anti-patterns to avoid**:

- **No magic `@Value` injection** — secrets resolved at boot are just keys in Viper. Consumers read them via `config.Get("spring.datasource.password")` (or the Yarumo equivalent). No reflection, no struct-field tagging based on Vault paths.
- **No bootstrap-vs-application split** — Spring's pre-2.4 dual context was a historical artefact. Yarumo has one config phase; the secrets source plugs into it like any other.
- **No silent fallback to "missing key returns empty string"** — `fail-fast: true` is the recommended default. The `optional:` URI prefix is the explicit opt-out for "OK if absent".
- **No service-loader-style auto-discovery of backends** — explicit registration via `RegisterBinding(name, factory)`. Consumers ship their own bindings if they need an engine Yarumo doesn't bundle (e.g. Transform engine).
- **No tight coupling to `secrets/vault/`** — the resolver depends on the abstract `secrets.Provider`. Swapping in `aws/secretsmanager` doesn't change the URI scheme or the property-mapping pipeline.
- **No reactive sibling resolver API** — a single `context.Context`-driven `Load(ctx, ...)` is enough; the renewal goroutine handles "long-running" via the existing managed-component model.

## 5. Recommendation

**PARTIAL** — adopt the conceptual model (property-source URI scheme, per-engine property mapping, lease lifecycle propagating to config, multi-binding, fail-fast, optional, prefix) but **as a new subpackage `modules/config/sources/secrets/`** of the existing `modules/config/` module, **not** inside a sibling secrets module. Driver-level concerns (Vault HTTP client, auth, session, lease container, typed engine ops) live in the proposed NEW `modules/secrets/vault/` — see spring-vault.md for that proposal.

The Pareto cut is ~12 features. The most valuable five — bootstrap-time resolution, per-engine property mapping, lease-renewal → config update, `lease-strategy`, `fail-fast` — together cover ~90% of real Spring Cloud Vault deployments. The reactive variant, the legacy bootstrap mode, the discovery integration, and the deprecated per-engine modules are all skipped.

**Sequencing**: this resolver subpackage **only makes sense after the driver ships**. The driver shapes the `secrets.Provider` interface that this resolver consumes. Don't design the resolver first; the driver-layer review (spring-vault.md) is the prerequisite. Once Phase 3 (`modules/config/` work in milestone #9) lands its core refactors and the `modules/secrets/` module is filed, the `sources/secrets/` subpackage is a focused additive ticket.

## 6. Proposed yarumo placement

Two-piece split:

- **Driver-level**: `modules/secrets/vault/` (in proposed NEW `secrets/` — see spring-vault.md). Vault client, authentication, session, lease container, typed engine ops. Exposes vendor-agnostic `secrets.Provider` interface.
- **Resolver-level**: `modules/config/sources/secrets/` (subpackage of existing `config/`). Consumes `secrets.Provider`; mounts secrets into the config registry; supports `secrets://<provider>/<path>?prefix=&optional=&binding=` URI form; propagates rotation events back into the config registry.

**Subpackage layout** for the resolver:

```
modules/config/
  sources/
    secrets/                       NEW — this analysis
      doc.go
      errors.go                    ErrBindingNotFound, ErrPathMissing, ErrProviderUnavailable
      loader.go                    Loader: parses "secrets://<provider>/<path>?prefix=&optional="
                                   URIs, dispatches to registered Providers, applies values to
                                   the Viper-equivalent registry at boot.
      binding.go                   Binding interface — maps a Vault path's raw map to target keys:
                                     Apply(secret map[string]any, target ConfigWriter) error
                                   Built-in bindings:
                                     KVBinding              flat copy with optional prefix
                                     DatabaseBinding        username → spring.datasource.username
                                                            password → spring.datasource.password
                                     AWSBinding             access_key, secret_key, [session_token]
                                                            → cloud.aws.credentials.*
                                     RabbitMQBinding        username → spring.rabbitmq.username
                                                            password → spring.rabbitmq.password
                                     ConsulBinding          token → spring.cloud.consul.token
                                     CouchbaseBinding       username → spring.couchbase.username
                                                            password → spring.couchbase.password
                                     ElasticsearchBinding   username → spring.elasticsearch.rest.username
                                                            password → spring.elasticsearch.rest.password
                                   Custom bindings via Register(name string, b Binding).
      uri.go                       URI parser: extracts provider, path, prefix, optional, binding
      lifecycle.go                 Subscriber: subscribes to Provider.Subscribe(ctx, path);
                                   on rotation, re-applies Binding, fires config-changed event.
                                   Implements managed.Component (Start/Stop/Done).
      options.go                   WithFailFast(bool), WithLeaseStrategy(LeaseStrategy),
                                   WithMinRenewal(d), WithExpiryThreshold(d).
      registry.go                  Provider registry: name → secrets.Provider.
```

**Configuration syntax** (mirrors `spring.config.import: vault://` with Yarumo idioms):

```yaml
# config.yml — loaded by modules/config
sources:
  secrets:
    providers:
      vault:
        endpoint: https://vault.internal:8200
        auth: kubernetes
        kubernetes:
          role: my-app
    bindings:
      - uri: secrets://vault/kv/data/my-app?prefix=app.&binding=kv
      - uri: secrets://vault/database/creds/readonly?binding=database&optional=false
      - uri: secrets://vault/aws/creds/s3-writer?binding=aws
      - uri: optional:secrets://vault/rabbitmq/creds/publisher?binding=rabbitmq
    lifecycle:
      enabled: true
      min-renewal: 10s
      expiry-threshold: 1m
      lease-strategy: RetainOnIoError       # DropOnError | RetainOnError | RetainOnIoError
      fail-fast: true
```

Programmatic equivalent (Yarumo style):

```go
loader := secrets.NewLoader(
    secrets.WithProvider("vault", vaultProvider),
    secrets.WithBinding("kv", secrets.KVBinding{}),
    secrets.WithBinding("database", secrets.DatabaseBinding{
        UserKey: "spring.datasource.username",
        PassKey: "spring.datasource.password",
    }),
    secrets.WithLeaseStrategy(secrets.RetainOnIoError),
    secrets.WithMinRenewal(10*time.Second),
    secrets.WithExpiryThreshold(time.Minute),
    secrets.WithFailFast(true),
)
if err := loader.Load(ctx, cfg, "secrets://vault/database/creds/readonly?binding=database"); err != nil {
    return err
}
```

**Internal deps**:

- `modules/secrets` (NEW, proposed) — consumes the abstract `Provider` interface. Hard dep.
- `modules/config` — host module; the loader writes into the Viper-equivalent registry.
- `modules/common/errs` — domain errors.
- `modules/common/log/slog` — structured logging.
- `modules/common/resilience` (closed YA-0076) — circuit breaker on Subscribe calls; retry with backoff on transient failures (drives the `RetainOnIoError` policy).
- `modules/managed` — `Subscriber` is a `managed.Component` (it owns a goroutine that listens on `Provider.Subscribe(...)`).

**Out of scope for v1**:

- Reactive sibling API (single `context.Context`-driven loader).
- Service-registry-located providers (no Eureka/Consul `DiscoveryClient` analogue).
- Legacy bootstrap mode (no pre-ConfigData dual-context split).
- Per-engine sub-modules in `config/sources/` (e.g. `config/sources/database-vault/`). Bindings are values, not packages.
- Tight integration with `modules/datasource/gorm/`'s connection-pool refresh — that lives in the datasource module, listening on the same rotation events the resolver fires.
- Multi-region / multi-Vault failover at the resolver level — the `Provider` interface can be implemented by a failover composite if a consumer needs it.
- `max_ttl` auto-handling — match Spring's documented limitation (app must restart) for v1; revisit if a consumer asks.

## 7. Open questions

1. **Subpackage placement** — `modules/config/sources/secrets/` (chosen) vs `modules/secrets/config/` vs a new top-level `modules/secretsconfig/`. The choice flows from the dependency direction: this layer depends on both `config` and `secrets`. Placing it under `config` makes `config` the consumer of `secrets.Provider`; placing it under `secrets` makes `secrets` the consumer of `config`. Going with `config/sources/secrets/` because `config/` is the orchestrator of all config sources (env, files, remote) per the roadmap's placement principle, and this is just one more source. ADR-worthy.

2. **URI scheme spelling** — `secrets://vault/path` (chosen) vs `vault://path` (Spring's spelling). Spring is single-backend so its URI scheme can be `vault://`; Yarumo is multi-provider so the scheme must indicate the provider abstractly. `secrets://<provider>/<path>` keeps that explicit. Confirm before committing.

3. **Rotation → config-key update propagation** — when a lease renews and the password changes, the new value lands in the config registry. Two questions: (a) does the registry support live updates at all? Viper does, but only via explicit `OnConfigChange`. (b) does the resolver fire a synthetic config-changed event so consumers can react? Required for `modules/datasource/gorm/` to refresh the pool.

4. **Binding extension model** — explicit `Register(name, Binding)` (chosen) vs Go service-loader pattern (`init()` side-effect registration). Spring uses `spring.factories` (service loader); Yarumo's preference is explicit. Confirm this against the `modules/secrets/vault/auth/*` registration approach to stay consistent across the two layers.

5. **Channel fan-in for subscription goroutines** — Spring's lease container fires events on its own scheduler. Yarumo's `secrets.Provider.Subscribe(ctx, path) (<-chan SecretEvent, error)` returns a channel per path. Does the resolver fan-in multiple channels into a single subscriber goroutine, or one goroutine per binding? Fan-in is cleaner but couples the lifetime of all bindings.

6. **Connection-pool refresh contract** — Spring's docs warn that connection pools may keep using old credentials even after the property is updated. Yarumo's equivalent answer: define a `RotationListener` interface in `modules/datasource/` (§ 1.1) that the gorm/mongo/redis drivers can implement; the resolver calls it after a successful `Binding.Apply(...)`. Alternative: leave pool refresh out of scope and require the consumer to wire it manually. Decide before the gorm subpackage ships.

7. **Should `KVBinding` be implicit (default)?** — most KV paths just need a flat copy with optional prefix. Making `binding=kv` the default for `secrets://...` URIs reduces boilerplate; making it required forces explicit choice. Lean toward "explicit `binding=` parameter required" for clarity.

8. **`max_ttl` policy** — Spring documents the limitation: when Vault's `max_ttl` is hit, the app must restart. Yarumo's options: (a) match Spring's behaviour and document the same limitation, (b) auto-trigger a re-auth + re-fetch when `max_ttl` is approached. Option (b) is more work and slightly more risk (could mask configuration bugs). Lean toward (a) for v1.

9. **`fail-fast` default** — Spring defaults to `false`; this analysis recommends `true`. Confirm with a yarumo "fail loudly at boot" stance, but document the override knob.

10. **Versioned KV reads at boot** — Spring's KV v2 path supports `?version=N` for pinned reads. Should the URI scheme support `secrets://vault/kv/data/path?version=3`? Useful for canary rollouts but adds parsing surface. Defer until a real consumer asks.

11. **Optional binding `prefix=` default** — if `prefix=` is absent, do the secret's keys merge into the root namespace or under a derived prefix (the path's last segment)? Spring merges into root. Root merging risks collisions; deriving from path is safer but surprising. Lean toward "no prefix" as the explicit default and document the collision risk.

12. **Test story** — same as spring-vault.md: `testcontainers-go` Vault module in the future `modules/testing/containers/`. The resolver tests need both a fake `secrets.Provider` (in-process channel) and a real-Vault integration test.

## 8. ROADMAP delta proposed (NOT applied)

This analysis depends on a parallel delta to ROADMAP_NEW_MODULES.md proposed by spring-vault.md (the NEW `modules/secrets/` module). Assuming that lands, the additive delta for this analysis is:

**§ 1 — New modules** — append to `modules/config/` (which is already in scope under Phase 3, milestone #9, but currently has no new-subpackage entry in ROADMAP_NEW_MODULES.md):

> **`modules/config/sources/secrets/`** — Resolver bridge between `modules/secrets/` and `modules/config/`. Mounts secrets from any registered `secrets.Provider` into the config registry using a `secrets://<provider>/<path>?prefix=&optional=&binding=` URI scheme. Built-in bindings: KV, Database, AWS, RabbitMQ, Consul, Couchbase, Elasticsearch. Supports lease-renewal → config update propagation via `Subscribe(...)`. Lifecycle options: `fail-fast`, `lease-strategy` (DropOnError | RetainOnError | RetainOnIoError), `min-renewal`, `expiry-threshold`. Depends on `modules/secrets/`, `modules/managed/`, `modules/common/resilience`. Tickets: file once the driver in `modules/secrets/vault/` is in flight — sequencing is "driver first, resolver second".

No deletions. No conflicts with existing § 1.1–§ 1.5 modules. The resolver is additive: it consumes the abstract `secrets.Provider` interface and writes into the existing config registry.

**Coordination with spring-vault.md**: that doc proposes the NEW `modules/secrets/` top-level module with `vault/` as the anchor sub-driver. Both deltas should land together — neither is useful alone. spring-vault.md is the prerequisite; this analysis is the consumer.
