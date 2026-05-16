# Spring Data REST — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-data/rest
> **Analyzed**: 2026-05-16 (re-analysis after `ROADMAP_NEW_MODULES.md` cleanup)
> **Recommendation**: REJECT at module level; partial salvage of RFC 7232 conditional-request semantics into a future `modules/common/http/conditional/` subpackage.

## 1. Project summary

Spring Data REST (latest stable **5.0.5**, part of the Spring Data umbrella) auto-exports any Spring Data repository (`JpaRepository`, `MongoRepository`, `CassandraRepository`, …) as a HATEOAS / HAL-formatted REST API over Spring MVC. Zero controllers, zero DTOs: the framework introspects the repository's domain type and synthesises collection / item / association / search endpoints, paging+sorting query params, JSON Schema + ALPS metadata, ETag headers from `@Version`, and a typed `ApplicationEvent` stream around every write. Tightly coupled to the JVM (Jackson, JPA / Hibernate metamodel, Spring MVC `DispatcherServlet`, Spring HATEOAS link builder, Bean Validation). The framework's value proposition is **"complete REST API in zero lines"** — exactly the auto-CRUD pattern yarumo rejects on the grounds that every service should author its own contract.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **ETag from `@Version`** | Version field on the entity → automatic `ETag` response header; `If-Match` mismatch → 412, `If-None-Match` match → 304 | Cleanest packaged expression of optimistic-concurrency + cache-revalidation HTTP semantics. Reusable as a convention without the auto-CRUD machinery. |
| 2 | **`Last-Modified` from `@LastModifiedDate`** | Time-based conditional GET (`If-Modified-Since` → 304) | Same family as ETag; saves bandwidth on read-heavy endpoints. |
| 3 | **Lifecycle event hooks** (`BeforeCreateEvent`, `AfterCreateEvent`, `BeforeSaveEvent`, `AfterSaveEvent`, `BeforeDeleteEvent`, `AfterDeleteEvent`, `BeforeLinkSaveEvent`, `AfterLinkSaveEvent`) | 8 typed events fired around repo writes; subscribe via `@RepositoryEventHandler` (typed handler methods) or `AbstractRepositoryEventListener` (untyped) | The **pattern** (typed pre/post-write hooks per aggregate) is solid — outbox, audit, derived-field computation, notifications. The hooks themselves are tied to Spring Data's repository abstraction. |
| 4 | **Repository exposure strategies** (`ALL`, `DEFAULT`, `VISIBILITY`, `ANNOTATED`) + per-method `@RestResource(exported=false)` | Granular control over which repos / methods become HTTP | Only meaningful inside an auto-CRUD framework. Irrelevant once handlers are explicit. |
| 5 | **Projections & Excerpts** | Lightweight per-endpoint views without writing DTOs (`?projection=summary`) | The DTO-free shortcut is exactly what makes auto-CRUD seductive and exactly what couples representation to schema. |
| 6 | **HAL hypermedia (`_links`, `_embedded`)** | Auto-generated navigation between aggregates | Hypermedia is fine; auto-generation isn't. If yarumo ever wants HAL it should be opt-in per-handler, not framework-wide. |
| 7 | **`ValidatingRepositoryEventListener`** | Validators bound by bean-name convention (`beforeCreatePersonValidator`) → standardised 4xx response on JSR-303 failure | The "validate before persist" wiring is reasonable; yarumo already has `modules/validation/` + `common/validation/` invoked explicitly. No need for naming-convention magic. |
| 8 | **JSON Schema + ALPS profile endpoints** | Auto-generated `/profile/{resource}` describing fields and operations | Niche; OpenAPI via `swaggo` / `kin-openapi` already covers this in Go. |

## 3. Long-tail features (skip)

- **`RepositoryDetectionStrategy`** — only meaningful inside the auto-CRUD framework.
- **`EntityLookup` / custom URI segments** — abstracts what should be in the controller.
- **Custom Jackson `ObjectMapper` configuration** — trivially equivalent to `encoding/json` + custom marshalers in Go.
- **CORS configuration** — handled by any Go router (gin, chi, …).
- **Paging+sorting query-param convention** (`?page=0&size=20&sort=name,asc`) — adopt the **convention** if it ever proves useful, but Spring Data REST adds nothing a 30-line helper can't do.
- **Association resource endpoints** (`/orders/1/customer`) — couples REST URL shape to ORM relationships; pure anti-pattern.
- **Spring HATEOAS link builder** — JVM-only.
- **WebFlux variant** — reactive flavour of the same anti-pattern.
- **Spring Security method-level integration on repositories** — equivalent yarumo path is `modules/auth/` middleware on explicit handlers.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**: none directly. Spring Data REST is a framework-on-top-of-framework that has no peer in the current yarumo plan — and that is the intended outcome, not a gap.

**Auto-CRUD rejection (stated inline, no longer a referenced § 3 row)**: yarumo will **not** ship an auto-CRUD module that exposes data repositories as HTTP endpoints. Every service authors its own contract — controllers are explicit, request / response shapes are explicit DTOs, and the REST URL space is owned by the service, not derived from the storage schema. Auto-exposing the repository produces (a) contract drift on every schema change, (b) HATEOAS link shapes coupled to ORM relationship topology, and (c) hidden representation rules that surprise consumers. The previous `ROADMAP_NEW_MODULES.md § 3 Brainstorm` row that held `modules/data-rest/` has been removed from the roadmap; this file is now the canonical record of the rejection and its rationale.

**Cross-references for the patterns Spring Data REST does package well**:

- **Lifecycle event hooks (Before/After × Create/Save/Delete/LinkSave)** — the typed pre/post-write pattern is genuinely useful; yarumo addresses it in two places, both outside this module's scope:
  - **`spring-modulith.md`** evaluates **`modules/outbox/`** (the *new* outbox module proposed there) for the write-then-publish flow that most teams reach for `BeforeSaveEvent` / `AfterSaveEvent` to implement. Idempotent dispatch, retry, transactional outbox — that's the proper home, not a generic event listener on every repository write.
  - **`modules/messaging/events/`** ([ROADMAP § 1.3](../ROADMAP_NEW_MODULES.md#13-modulesmessaging--eip-layer--brokers)) covers in-process typed pub/sub when a handler explicitly publishes a domain event after a successful write.
- **Row-level audit columns** (`CreatedAt`, `UpdatedAt`, `CreatedBy`, `LastModifiedBy`) — handled by **`modules/datasource/gorm/` hooks** at the persistence layer ([ROADMAP § 1.1](../ROADMAP_NEW_MODULES.md#11-modulesdatasource--db-and-cache-adapters)). This is the right altitude for column-level auditing; event-level audit trails (separate concern) belong with the outbox / event design discussed in `spring-modulith.md`.
- **Validation hooks** (`ValidatingRepositoryEventListener` + bean-name convention) — already covered by `modules/common/validation/` (leaf) + `modules/validation/` (composed), shipped in Phase 2 and invoked explicitly inside handlers. No naming-convention DI.
- **Testing patterns** (Spring Data REST has `@WebMvcTest` / `MockMvc` style integration tests for repository endpoints) — `spring-framework.md` proposes a **new `modules/testing/`** for HTTP / handler test harnesses; that is the right place to evaluate the testing recipes from Spring Data REST, not here.

**Anti-patterns to avoid (stated for the record)**:

1. **Auto-exposing the repository / data model** as the REST API. Contract drift = breaking change every time the schema moves.
2. **Hypermedia coupled to ORM relationships** — `/orders/1/customer` style endpoints leak the relational model into the URL space.
3. **Naming-convention DI** (`beforeCreatePersonValidator`) — opaque; yarumo wiring is explicit (`BeanFn` in `modules/boot/`, [ROADMAP § 1.5](../ROADMAP_NEW_MODULES.md#15-modulesboot--application-wiring)).
4. **God-config object** (`RepositoryRestConfiguration`) — same anti-pattern called out for the legacy `boot/` `ApplicationContext` in `modules/boot/` § 1.5.
5. **Projections via query string** — encourages clients to know the schema; consumers should request a **named** resource representation, not assemble views from the storage shape.

## 5. Recommendation

**REJECT** at the module level. The whole framework is the anti-pattern: it sells "API for free" but the bill arrives as data-model lock-in, breaking-change amplification, and HATEOAS coupling between REST URL shape and ORM topology. Yarumo's stance — every service authors its own contract, controllers are explicit, DTOs are not optional — is correct and stays.

**PARTIAL salvage**: the HTTP-level conditional-request recipe (ETag derived from a monotonic version, `If-Match` → 412 for optimistic concurrency, `If-None-Match` → 304 and `If-Modified-Since` → 304 for caching). This is pure RFC 7232 hygiene; Spring Data REST just packages it in one place clearly. Surface it as a small, opt-in middleware in the future `modules/common/http/` track, decoupled from any ORM / version-field source. The version source must be a tiny interface a handler provides, **not** an annotation on a struct.

## 6. Proposed yarumo placement

**Subpackage**: `modules/common/http/conditional/` — to be authored when YA-0042's `common/http/` Phase 2 bundle wakes up (currently deferred together with YA-0043, YA-0044 until a real consumer needs the surface). No new ticket today.

Sketch (illustrative, not a design lock-in):

```go
// VersionSource produces the conditional-request signals for the resource
// targeted by r. Either field may be zero-valued; the middleware only enforces
// the dimensions the source supplies.
type VersionSource interface {
    ConditionalState(r *http.Request) (etag string, lastModified time.Time, err error)
}

// Middleware enforces RFC 7232:
//   GET/HEAD  + If-None-Match matches current ETag       -> 304
//   GET/HEAD  + If-Modified-Since >= LastModified        -> 304
//   PUT/PATCH/DELETE + If-Match mismatch                 -> 412
//   PUT/PATCH/DELETE + If-Match match (or absent + opt-in)
//                                                        -> pass through,
//                                                           response is decorated
//                                                           with the new ETag /
//                                                           Last-Modified.
func Middleware(src VersionSource, opts ...Option) func(http.Handler) http.Handler
```

Design constraints (the things Spring Data REST gets wrong for yarumo's taste):

- **No annotation / reflection** — the handler explicitly registers a `VersionSource`.
- **No coupling to a persistence layer** — `VersionSource` is HTTP-only; the handler decides how a version is computed (row column, hash, monotonic counter, content hash).
- **Composable as vanilla `func(http.Handler) http.Handler`** — plugs into `gin`, `chi`, `net/http`, anything.
- **Strong ETags by default** — Spring Data REST emits weak ETags (`W/"<version>"`) by default because it does not know the semantics of the version field. The yarumo middleware knows: the handler supplied the source, so the default is **strong**, with an option to mark weak when the source declares it (e.g., compressed-content negotiation).

This is a **single-digit-LOC public surface**. It belongs in `common/http/` only because RFC 7232 is HTTP, not because Spring Data REST inspired it.

## 7. Open questions

1. **`VersionSource` ergonomics** — single interface for both ETag and `Last-Modified`, or split? Split is cleaner (different cache semantics) but doubles handler boilerplate. Defer until first consumer.
2. **Where does conditional middleware live relative to a future Aluna / DaaS contract layer?** Probably below it (transport concern), but the answer depends on how Aluna / DaaS package handlers.
3. **Should the paging+sorting query-param convention** (`?page=N&size=M&sort=field,dir`) be standardised across yarumo handlers? Many JS clients expect it. Worth a one-paragraph note in the eventual `common/http/` design doc — not a module.
4. **Interaction with `modules/outbox/`** (proposed in `spring-modulith.md`) — does the outbox dispatcher need to read / propagate ETags for downstream HTTP consumers that re-emit the resource, or is that strictly the gateway's concern? Likely the latter, but worth noting when the outbox design is drafted.
5. **Audit-trail relationship** — `spring-modulith.md` is now the canonical home for the event-level audit-trail discussion (separate from `datasource/gorm/` row-level auditing). Confirm scope split when both module designs activate.

## 8. ROADMAP delta proposed (NOT applied)

- **`ROADMAP_NEW_MODULES.md`** — no change today. The previous § 3 brainstorm row for `modules/data-rest/` has already been deleted; this file is the canonical record of the rejection.
- **`ROADMAP_NEW_MODULES.md § 1`** — no new module. Add a one-line note inside the future `common/http/` design (whenever YA-0042 wakes up) pointing here for the conditional-request middleware sketch.
- **No new ticket** filed today. When a real consumer (most likely DaaS, possibly Aluna gateways) requires conditional requests, file under the `common/http/` track and link this file as the design source.
- **Cross-doc anchoring**: `spring-modulith.md` (outbox / audit / event patterns) and `spring-framework.md` (testing harness) carry the patterns Spring Data REST packages but yarumo factors elsewhere. Confirmed cross-references in § 4 above.
