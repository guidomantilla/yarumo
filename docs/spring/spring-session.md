# Spring Session — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-session/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup)
> **Recommendation**: PARTIAL — adopt as a **new top-level module** `modules/sessions/`, scoped to the Redis + Postgres pair and gated behind real demand (DaaS console / Aluna UI). No § 3 home exists after the roadmap trim, so this module must justify itself on its own merits.

## 1. Project summary

Spring Session externalizes HTTP session state out of the servlet container into a shared store (Redis, JDBC, Hazelcast, MongoDB), enabling stateless app servers and horizontal scaling. Latest stable **4.0.3** (4.1.0-RC1 in preview), Java 17+, Apache 2.0.

Scope on the JVM side:

- `Session` interface — id, creation/last-accessed timestamps, `maxInactiveInterval`, attribute map, `isExpired()`.
- `SessionRepository<S>` — `createSession() / save / findById / deleteById`.
- `FindByIndexNameSessionRepository` — `findByPrincipalName(name) → Map<id, Session>` plus `PRINCIPAL_NAME_INDEX_NAME` convention.
- `HttpSessionIdResolver` — `CookieHttpSessionIdResolver` (default `SESSION` cookie) vs `HeaderHttpSessionIdResolver.xAuthToken()` (REST/SPA/mobile).
- `DefaultCookieSerializer` — `Secure` / `HttpOnly` / `SameSite=Lax` defaults, configurable `domain` / `path` / `maxAge`.
- Session events — `SessionCreatedEvent` / `SessionDestroyedEvent` (super of `Expired` + `Deleted`).
- `SessionRepositoryFilter` — wraps `HttpServletRequest` and replaces `getSession()` with the externalized one transparently.
- Repository impls today: `RedisSessionRepository` (non-indexed), `RedisIndexedSessionRepository` (indexed + events via keyspace notifications), `ReactiveRedisSessionRepository`, `JdbcIndexedSessionRepository`, `MapSessionRepository`.

JVM coupling: **medium**. The store contracts (`SessionRepository`, key layout, schema) port cleanly to Go. The ergonomics (`@EnableRedisHttpSession`, request wrapping, `HttpSessionListener` bridge) are servlet-specific and we replace them with a `net/http` middleware.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | `Session` interface | id, creation/lastAccessed time, maxInactiveInterval, attributes map, `isExpired()` | Universal contract. Any HTTP framework can wrap it. |
| 2 | `SessionRepository` interface | `Create / Save / FindByID / DeleteByID` | Pluggable store boundary — the entire abstraction's reason to exist. |
| 3 | Redis backend | Hash per session at `<ns>:sessions:<id>`, native `EXPIRE` for TTL, keyspace notifications (`config set notify-keyspace-events Egx`) for expiry events | The default deployment. Offloading TTL to Redis is the killer feature. |
| 4 | JDBC backend | `SPRING_SESSION` (id PK, creation_time, last_accessed_time, max_inactive_interval, principal_name, expiry_time) + `SPRING_SESSION_ATTRIBUTES` (session_id FK, name, bytes); scheduled `deleteExpiredSessions()` cleanup | Production fallback when Redis isn't available; one less infra dep. Already provisioned via `modules/datasource/gorm/`. |
| 5 | `HttpSessionIdResolver` | Cookie (default `SESSION`) or header (`X-AUTH-TOKEN`) strategies | Same machinery serves browser apps and stateless REST clients (SPA, mobile). |
| 6 | `CookieSerializer` defaults | `Secure`, `HttpOnly`, `SameSite=Lax`, configurable path/domain/maxAge, optional `domainNamePattern` | Secure-by-default cookies — exactly what every Go HTTP server reinvents. |
| 7 | `FindByIndexNameSessionRepository` | `findByPrincipalName(name) → map[id]Session` via `PRINCIPAL_NAME_INDEX_NAME` | Enables "log out all my sessions" and concurrent-session limits. Pairs directly with `modules/auth/` (§ 1.2). |
| 8 | Session events | `Created`, `Destroyed`, `Expired`, `Deleted` | Hook for audit trails, security alerts, downstream cache invalidation. |
| 9 | Delta serialization | Only changed attributes are written per request | Reduces write amplification in Redis under contention. Critical for any non-trivial workload. |
| 10 | Session-fixation protection | `changeSessionId()` on auth boundary | Standard OWASP requirement; trivial to expose as `Store.Rotate(oldID) → newID`. |
| 11 | `FlushMode.ON_SAVE` semantics | Writes batched to end of request, not per-attribute | Sane default; `IMMEDIATE` is a footgun unless explicitly needed. |
| 12 | Filter-based transparent wrapping | `SessionRepositoryFilter` wraps the request before any handler sees it | In Go this becomes a single `http.Handler` middleware — simpler than the servlet equivalent. |

## 3. Long-tail features (skip)

- WebFlux `WebSession` / `ReactiveSessionRepository` — yarumo is non-reactive; `net/http` suffices.
- Hazelcast, MongoDB, Infinispan, Caffeine backends — Redis + Postgres cover ~95% of demand.
- `@EnableSpringHttpSession` / `@EnableRedisHttpSession` / `@EnableJdbcHttpSession` annotations — DI sugar; wire explicitly in Go via constructors.
- `JvmRoute` cookie suffix for sticky sessions — k8s ingress / load balancer concern, not the lib's.
- `domainNamePattern` regex extraction — niche; explicit `domain` setter is enough.
- WebSocket "keep session alive" hooks — coupled to servlet WebSocket; `gorilla/websocket` users wire it manually if needed.
- XML namespace config, SpEL, `@EnableSpringWebSession` — JVM-only ergonomics.
- `@SpringSessionDataSource` qualifier — DI artifact; in Go the consumer passes the `*sql.DB` / `*gorm.DB` directly.
- Concurrent-session control filter — derivable from `FindByPrincipalName`; not v1.
- `RedisFlushMode.IMMEDIATE` flag — `ON_SAVE` is the sane default; expose later only if a real workload demands it.
- `HttpSessionEventPublisher` bridge — that exists to translate Spring Session events to servlet `HttpSessionListener`; we just expose listener funcs directly.
- Reactive Redis variant + `ReactiveMapSessionRepository` — non-reactive Go.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**:

- **`modules/auth/` (§ 1.2)** — sessions are auth-adjacent. `AuthorizationFilter` is the primary consumer: reads the session to recover the `Principal`, writes the session on successful authentication, calls `Rotate` on auth state change (session-fixation), and uses `FindByPrincipal` for "log out all my sessions". The two modules must agree on a contract before either ships.
- **`modules/datasource/goredis/` (§ 1.1)** — provides the Redis client `modules/sessions/redis/` consumes. No direct `github.com/redis/go-redis` import in `sessions/`; the connection is injected.
- **`modules/datasource/gorm/` (§ 1.1)** — provides the SQL handle for `modules/sessions/postgres/`. Schema migrations land in the consumer app, not in this module.
- **`modules/common/uids/`** — cryptographically random URL-safe session-id generation.
- **`modules/common/crypto/random`** — backing entropy for the id resolver.
- **`modules/common/crypto/tokens`** — *alternative* to sessions for stateless deployments (service-to-service). The two coexist: sessions for browser, JWT for back-channel. Documenting when to choose which is part of the module's README.
- **`modules/common/log/slog`** — structured logging in the middleware.
- **`modules/common/errs`** — typed errors (`ErrSessionNotFound`, `ErrSessionExpired`).
- **`modules/managed/` (existing)** — the Postgres backend's expired-row sweeper goroutine integrates with `managed.Lifecycle` Start/Stop/Done.
- **`modules/boot/` (§ 1.5)** — wires the `Store` + middleware via a `BeanFn`.

**Gaps to fill**:

- Today every Go web app reimplements cookie defaults, session-id rotation on auth, and TTL bookkeeping. A shared `sessions.Store` + `sessions.CookieSerializer` ends that.
- `gorilla/sessions` exists but is effectively in maintenance mode and lacks `FindByPrincipalName` semantics, schema migrations, and lifecycle-integrated event hooks.
- DaaS will need server-side sessions for any web console; Aluna's agent UI likewise. Without this module each consumer reinvents the wheel — badly.
- **Session-fixation rotation** — `auth/` needs a single primitive `Store.Rotate(oldID) (newID, error)` to call on login. Without it, every consumer rolls their own and most do it wrong.
- **Cookie security defaults** — Go's `http.Cookie` doesn't default `Secure` / `HttpOnly` / `SameSite`. A `CookieSerializer` that enforces the OWASP-aligned defaults closes a common security gap.

**Anti-patterns to avoid**:

- No annotations / `@EnableXxx` — `sessions.New(store, opts...)` and that's it.
- No god-`SessionManager` struct holding 15 fields. Keep `Store` lean; expose helpers as free functions.
- No fixed init order — `Store` is just a constructor consumed by a `BeanFn`.
- Do not couple to gin / chi / `net/http` specifics in the core. Ship middleware adapters as thin sub-packages (`middleware/nethttp/`, `middleware/gin/`).
- No reactive variant. `net/http` only.
- Do not invent a serialization framework — `encoding/gob` plus a `Codec` interface (with `json` for debugging).
- Do not auto-publish session-fixation rotation as a side effect of some imagined "login hook". Expose `Rotate` and let `auth/` call it explicitly.
- Do not bundle the principal-name index into the base `Store` interface — separate `IndexedStore`. The in-memory store doesn't need it; the Postgres store may not in v1.
- Do not write to Redis on every attribute mutation. `FlushMode.ON_SAVE` (write once at end of request) is the only sane default; `IMMEDIATE` is a footgun.

## 5. Recommendation

**PARTIAL — file as a NEW top-level module `modules/sessions/`. Priority P2: design now, ship when the first real consumer (DaaS console or Aluna UI) is ~1 sprint away.**

Spring Session's two pillars — the `Store` abstraction and secure cookie defaults — are exactly what every Go web app reinvents, and the Redis + Postgres pair is mature enough that a yarumo wrapper has real leverage. The original analysis pointed at `§ 3.2` for placement; that section is gone after the roadmap trim, so the module has to stand on its own. It does: it pairs tightly with the existing `auth/` (§ 1.2) and rides on already-planned `datasource/goredis/` and `datasource/gorm/` (§ 1.1).

What changes vs the previous review:

- **Promote from "tucked into § 3.2 brainstorm" to "new top-level module"** — explicit placement decision, listed in § 1 of `ROADMAP_NEW_MODULES.md`.
- **Coordinate explicitly with `modules/auth/`** — the session-fixation rotation contract (`Store.Rotate`) and the principal-index contract (`FindByPrincipal`) must be agreed before either module ships.
- **Drop reactive across the board** — no `ReactiveStore`, no `WebSession` analog. If `net/http` ever gets a streaming alternative we re-evaluate.
- **Stay P2** — real demand only materializes when DaaS or Aluna ship a browser-facing UI. Until then `common/crypto/tokens` covers service auth. File the module on `ROADMAP_NEW_MODULES.md` as **Brainstorm**, promote to **Planned** when the first consumer locks a date.

## 6. Proposed yarumo placement

**NEW top-level module**: `modules/sessions/`

**Status on `ROADMAP_NEW_MODULES.md`**: Brainstorm initially; promote to Planned when a consumer commits.

**Pairs with**: `modules/auth/` (§ 1.2). Sessions are auth-adjacent — the two modules co-evolve.

**Subpackage layout**:

```
modules/sessions/
  session.go            Session struct (ID, CreatedAt, LastAccessedAt, MaxInactive, Attrs, Expired())
  store.go              Store interface: Create / Save / Find / Delete; IndexedStore: FindByPrincipal, Rotate
  codec.go              Codec interface (gob default; json optional for dev/debug)
  events.go             Event types (Created, Expired, Deleted) + Listener func + Bus
  cookie/               CookieSerializer: secure defaults (Secure, HttpOnly, SameSite=Lax), Encode/Decode, optional WithSigner(...)
  resolver/             IDResolver interface; CookieResolver + HeaderResolver impls (X-AUTH-TOKEN convention)
  middleware/
    nethttp/            stdlib net/http handler wrapper (load → context-inject → save on response)
    gin/                gin middleware (consumer-opt-in subpackage; isolated import of gin)
  redis/                Store impl: <ns>:sessions:<id> hash + EXPIRE; principal index set; keyspace-events listener emits Expired events
  postgres/             Store impl: sessions + session_attributes schema; background sweeper goroutine for expired rows (managed.Lifecycle integration)
  inmemory/             Map-backed Store for tests/dev; concurrent-safe
```

**Internal deps**:

- `modules/common/uids/` — session-id generation.
- `modules/common/crypto/random` — secure entropy for cookie values; HMAC keys for the optional signer.
- `modules/common/crypto/tokens` — *not* a dep; documented as the alternative for stateless service auth.
- `modules/common/log/slog` — structured logging.
- `modules/common/errs` — `ErrSessionNotFound`, `ErrSessionExpired`, `ErrStorageUnavailable`.
- `modules/datasource/goredis/` — Redis client injected (no direct `github.com/redis/go-redis` import in `sessions/` core).
- `modules/datasource/gorm/` — SQL handle for Postgres backend.
- `modules/managed/` — Lifecycle integration for the Postgres sweeper goroutine and the Redis keyspace-events listener.

**Cross-module contracts to nail before ship**:

- **With `modules/auth/`** — agree on `Store.Rotate(ctx, oldID) (newID, error)` signature so `AuthorizationFilter` can invoke it on successful auth without taking a hard dep on a specific backend. Agree on the principal index attribute name (recommend `_sess_principal` to avoid colliding with user-defined attributes).
- **With `modules/boot/` (§ 1.5)** — provide a sample `BeanFn` that wires `Store + middleware + auth filter` in the canonical order.

**Go libraries to wrap** (mature, with repo URL):

- `github.com/redis/go-redis/v9` — Redis client (already wrapped by `datasource/goredis/`). https://github.com/redis/go-redis
- `gorm.io/gorm` — when the consumer already runs GORM (already wrapped by `datasource/gorm/`). https://github.com/go-gorm/gorm
- `github.com/jackc/pgx/v5` — direct driver alternative if we want to skip GORM for the sweeper. https://github.com/jackc/pgx

Reference reading (not wrapped — design guidance only):

- `github.com/gorilla/sessions` — API shape lessons (what to avoid: too coupled to specific stores, no `FindByPrincipal`). https://github.com/gorilla/sessions
- `github.com/alexedwards/scs/v2` — closest active Go equivalent; pluggable stores; clean API. Study the codec and middleware patterns. https://github.com/alexedwards/scs

**Out of scope for v1**:

- Reactive / streaming variant.
- WebSocket session-extension hooks.
- Hazelcast, MongoDB, Cassandra stores.
- Concurrent-session-count limiter middleware (build on top of `FindByPrincipal` post-v1).
- Sticky-session JVM-route cookie suffix.
- Server-Sent Events / WebSocket adapters.
- Schema migrations bundled with the module — consumers manage their own DDL; we publish the canonical schema in the README.
- Distributed-tx-aware Postgres writes — single-row UPSERT is enough; no 2PC.

## 7. Open questions

- Is `Store` the right name, or `Repository` to mirror Spring? Vote: **`Store`** — shorter, idiomatic Go, no Spring baggage.
- Should the principal index live in the base `Store` interface or a separate `IndexedStore`? Vote: **separate** — keeps the minimal interface tiny; not every backend implements it. Mirrors Spring's `FindByIndexNameSessionRepository` split.
- Cookie signing/HMAC inside `cookie/` package, or assume the cookie value is just the opaque id and the store enforces secrecy? Vote: **opaque id by default**; signed cookies as opt-in via `cookie.WithSigner(...)`.
- Postgres schema: ship Spring's exact (`SPRING_SESSION` / `SPRING_SESSION_ATTRIBUTES`) for migration parity, or yarumo-native (`sessions`, `session_attributes`)? Vote: **yarumo-native**; Spring schema parity is a migration-tool concern, not a runtime one.
- Pluggable codec (`gob` / `json` / `msgpack`) or freeze on `gob`? Vote: **`Codec` interface, `gob` default** — JSON needed for inspection in dev.
- Does `modules/cache/` (introduced in Phase 2) already cover the in-process store case well enough to skip `sessions/inmemory/`? Likely yes for prod use; keep `inmemory/` as a test-only impl with explicit doc warning.
- Coordinate with `modules/auth/` (§ 1.2) on the `Store.Rotate` signature and the `PrincipalName` attribute key **before either module ships**. Filing this as a tracked discussion item once both modules promote from Brainstorm to Planned.
- Should the Postgres sweeper run on a fixed interval (default 5 min) or be driven by a `managed.CronWorker`? Vote: **`managed.CronWorker`** — already exists, gives consumers control over cadence.
- Should event delivery be synchronous (called inside `Save` / sweeper) or async (channel-backed)? Vote: **synchronous + non-blocking listener contract** — caller registers a func that must not block; mirrors `slog.Handler` discipline.

## 8. ROADMAP delta proposed (NOT applied)

Add a new entry to `docs/ROADMAP_NEW_MODULES.md` § 1, after § 1.5 (or wherever fits the numbering):

> **§ 1.6. `modules/sessions/` — HTTP session store + secure-cookie defaults**
>
> **Status**: Brainstorm
> **Pairs with**: `modules/auth/` (§ 1.2 — sessions are auth-adjacent); `modules/datasource/goredis/` and `modules/datasource/gorm/` (§ 1.1 — backends).
> **Why a new module**: stateful store with lifecycle (Redis keyspace listener, Postgres expired-row sweeper), depends on `datasource/` for persistence, ships `net/http` middleware. Doesn't fit in `common/` (lifecycle + external SDK) and doesn't belong inside `auth/` (the two pair but compose separately).
>
> Migrated from go-feather-lib: nothing — this is greenfield. `gorilla/sessions` and `alexedwards/scs/v2` are design references, not migration sources.
>
> Scope: `Store` interface (Create/Save/Find/Delete), `IndexedStore` (FindByPrincipal, Rotate), `CookieSerializer` with OWASP defaults, `IDResolver` (cookie + header), `Codec` (gob default), event listeners. Backends: `redis/`, `postgres/`, `inmemory/`. `net/http` middleware; optional `gin/` subpackage.
>
> Promote to **Planned** when the first consumer (DaaS console / Aluna UI) commits a ship date. Until then, `common/crypto/tokens` covers service-to-service auth and the team has no demand signal.

Add a row to the § 4.2 **Discarded** table (or a new explanatory line elsewhere) noting that `gorilla/sessions` is intentionally **not** migrated — it's a reference, not a source, and the module is greenfield.

No § 3 changes (the § 3 brainstorm was deleted; this module would have lived there before, now it's a first-class § 1 entry).
