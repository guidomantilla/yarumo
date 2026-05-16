# Spring Data MongoDB — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-data/mongodb
> **Analyzed**: 2026-05-16
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Data MongoDB (5.0.5 stable, 5.1.0-RC1 preview) is the Spring Data sub-project that layers a `MongoTemplate` (and reactive variant), a repository abstraction with derived queries, an aggregation-pipeline builder, change-stream/transaction/GridFS support, index annotations, and client-side field-level encryption on top of the official MongoDB Java driver. Scope is broad: from low-level template ops to declarative repository interfaces, AOT-optimized at build time. JVM-coupled (relies on `@Document`, `@Indexed`, AspectJ proxies, AbstractMongoClientConfiguration, Pageable, SpEL inside `@Query` strings). Not directly portable to Go — only the **conceptual layering** survives translation, since `mongo-driver` for Go already provides typed BSON marshalling, transactions, change streams, and aggregation builders.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | `MongoTemplate` / `MongoOperations` | High-level facade over the driver with `find/save/update/remove/aggregate/executeCommand` and a fluent query API. | Centralizes connection/collection access; the Yarumo equivalent is a thin `Connection` + transaction helper, not a wrapper of every CRUD verb. |
| 2 | Connection lifecycle (`MongoClient` factory + per-DB `MongoOperations`) | Sets up auth, server URI, replica-set discovery, pooling. | Direct match for `modules/datasource/mongo/`: Context (URI/credentials), Connection (open/close), wired into `managed.Lifecycle`. |
| 3 | `WithTransaction` / `ClientSession` callback | Wraps multi-document transactions; rolls back on exception. | Mirrors the planned cross-driver `WithTransaction(ctx, db, fn)` helper. Worth implementing as a thin wrapper of `mongo.Session.WithTransaction`. |
| 4 | Aggregation pipeline DSL (`Aggregation.newAggregation(match(...), group(...), project(...))`) | Type-safe builder over the 24+ pipeline stages and 60+ operators. | `go.mongodb.org/mongo-driver` exposes `bson.D{}` pipelines, which are verbose and error-prone. A typed DSL is the single biggest Spring value-add Yarumo could replicate. |
| 5 | Index management annotations + resolver | `@Indexed`, `@CompoundIndex`, `@TextIndexed`, `@HashIndexed`, `@WildcardIndexed`, TTL. `IndexResolver` materializes them at startup. | Go has no annotations. Equivalent is a tag-based `IndexSpec` registry + `EnsureIndexes(ctx, collection, model)` helper. Useful but not load-bearing. |
| 6 | Change streams (imperative `MessageListenerContainer` + reactive `Flux<ChangeStreamEvent>`) | Resumable via token/timestamp, filter via pipeline. | Native driver already exposes `Collection.Watch()`. Wrapping pays off only when paired with `managed.Lifecycle` (Start/Stop a long-running consumer). |
| 7 | Repository abstraction (derived queries) | `findByLastnameAndAgeGreaterThan(...)` → query is inferred from method name. | **Reject for Go.** Requires reflection-driven query generation; conceptually clashes with explicit Go idioms. Same anti-pattern as ORM. |
| 8 | `@Query` JSON queries with `?0` placeholders | Custom queries declared as JSON strings, parameter substitution by index. | Same anti-pattern category. Direct BSON literals are already idiomatic in Go. |
| 9 | Lifecycle events + auditing (`@CreatedDate`, `@LastModifiedBy`) | Hooks fire before save/update; populates audit columns. | Mirrors the row-level audit hooks listed in § 1.3 of ROADMAP_NEW_MODULES. Worth a small `auditing/` sub-package shared with `gorm/` via interface. |
| 10 | Multi-document transactions integrated with templates | `@Transactional` participates in the same session for repository + template calls. | Driver already handles this with `Session.WithTransaction`. The win is making it ergonomic (single `WithTransaction(ctx, fn)` call). |
| 11 | GridFS template | Upload/download large files, query metadata. | Niche but real (binary attachments, model files). The driver provides `gridfs.Bucket` — a thin facade is enough. |
| 12 | Geospatial query helpers (`near`, `withinSphere`, `geoWithin`) | Typed wrappers around `$near`, `$geoWithin`. | Same pipeline-DSL category. Worth including in the query builder. |
| 13 | Projections (interface- and DTO-based) | Map a subset of fields into a typed projection. | Go achieves this via separate structs + `FindOptions{Projection: ...}`. Helper that derives projection from struct tags = small ergonomic win. |
| 14 | Client-side field-level encryption (CSFLE) / queryable encryption | `AutoEncryptionSettings`, key vault, deterministic vs random encryption per field. | The Go driver supports this directly via `options.AutoEncryption()`. Skip wrapping; document the recipe instead. |
| 15 | Observability (Micrometer/OTel spans & metrics) | Tracing per operation, command name + collection labels. | Direct match for `modules/telemetry/otel/`. The Go driver exposes `event.CommandMonitor` — wire it to OTel in a `telemetry/otel/mongo/` adapter. Out of scope for `datasource/mongo/` v1. |

## 3. Long-tail features (skip)

- **Repository derived queries** — relies on parsing Java method names; no Go-idiomatic equivalent.
- **`@Query` JSON-with-placeholder syntax** — direct BSON construction is cleaner in Go.
- **CDI/Spring config integration** (`@EnableMongoRepositories`, `AbstractMongoClientConfiguration`) — Yarumo wires explicitly via `BeanFn`.
- **Reactive variant** (`ReactiveMongoTemplate`, `Flux`/`Mono`) — Go has no Reactor; native context+goroutines cover async needs.
- **Kotlin coroutine support** — N/A.
- **AOT / native-image hints** — Go has no equivalent runtime.
- **SpEL inside aggregation expressions** — Spring's expression language; `modules/common/expressions/` is the closest Yarumo analogue and is unrelated to query construction.
- **AspectJ-driven `@Transactional` / lifecycle events** — Go has no proxy/AOP; explicit `WithTransaction(fn)` covers it.
- **Tailable cursors** — niche; driver native method `Find` with `CursorType: Tailable` is enough.
- **JSON Schema validation annotations on documents** — overlap with `modules/common/validation/`; use that instead of duplicating.
- **CDI integration** — N/A.
- **Sharding annotations (`@Sharded`)** — operational concern, not application concern.
- **MongoDB Atlas Search / Vector Search wrappers** — vector search belongs in `modules/datasource/vector/` (§ 1.3), not `datasource/mongo/`.

## 4. Mapping to Yarumo

**Existing/planned modules with overlap**: `modules/datasource/mongo/` (§ 1.3 of `docs/ROADMAP_NEW_MODULES.md`). This is the natural and exclusive placement. Cross-driver concerns (`WithTransaction`, audit hooks, lifecycle policies) live in `modules/datasource/` core per § 1.3.

**Gaps this could fill**:
- **Typed aggregation pipeline builder**. The single biggest ergonomic gap left by `go.mongodb.org/mongo-driver`. Writing complex pipelines as nested `bson.D{}` literals is verbose and error-prone. A small DSL (`Pipeline().Match(...).Group(...).Project(...)`) directly inspired by Spring's `Aggregation.newAggregation(...)` would pay rent.
- **Index spec registry from struct tags**. `EnsureIndexes(ctx, db, &Model{})` derived from struct tags (`bson:"email" index:"unique"`, `index:"text,weight=5"`, `index:"ttl,expire=86400"`). Same role as Spring's `IndexResolver`. Optional; consumers opt in.
- **Transaction helper** matching the cross-driver `WithTransaction(ctx, fn)` shape declared in § 1.3. Thin layer over `Session.WithTransaction` that participates in `context.Context`.
- **Change-stream consumer as a `managed.Worker`**. Long-running consumer with resume-token persistence, OTel-instrumented, plugged into `Start/Stop/Done`. Mirrors `MessageListenerContainer`.
- **Audit-field hooks** (`CreatedAt`/`UpdatedAt`/`CreatedBy`/`UpdatedBy`) populated via a `Mutator` interface invoked from a wrapper `Collection.InsertOne/UpdateOne`. Implementation shared with `gorm/`.

**Anti-patterns to avoid**:
- **No repository abstraction with derived queries**. Reflection-driven query generation is the ORM smell § 1.3 explicitly calls out.
- **No `@Query`-style string-templated queries**. BSON literals beat embedded query DSLs.
- **No DI / annotation-driven config**. Wired explicitly via `BeanFn`.
- **No god-template** (`MongoOperations` with 80+ methods). Keep the surface minimal: `Connection` (open/close), `Database()`, transaction helper, optional pipeline builder, optional index registry. The raw `*mongo.Collection` stays accessible.
- **No reactive variant**. `context.Context` is the cancellation model.
- **No JSON-Schema validation overlap** with `modules/common/validation/`. If schema validation is needed at the document level, route through that module.

## 5. Recommendation

**PARTIAL** — adopt the *conceptual layering* and a small set of concrete helpers; reject the repository abstraction, the declarative `@Query` mechanism, the DI machinery, and reactive support. The Go driver already provides type-safe marshalling, transactions, change streams, GridFS, CSFLE, and aggregations; wrapping each verb adds no value. The legitimate wins are: (a) a small typed pipeline builder, (b) a `WithTransaction` helper consistent with other drivers, (c) struct-tag-driven index registration, (d) audit-field mutators, (e) `managed.Worker`-based change-stream consumer. Everything else stays as the raw `*mongo.Client`/`*mongo.Database`/`*mongo.Collection` accessible via `Connection.Database()`.

Honest assessment: ~80% of Spring Data MongoDB's value comes from the repository abstraction + DI integration, which we **deliberately don't want**. The remaining 20% (template ergonomics, pipeline DSL, lifecycle hooks, index management) is genuinely useful and aligns with the per-driver pattern already declared in § 1.3. That's enough to justify the module, but the module should stay thin — closer to a `mongo-driver` ergonomic layer than to a Spring-Data port.

## 6. Proposed yarumo placement

**Module**: `modules/datasource/mongo/`

**Subpackages**:
- `mongo/` (root) — `Context` (URI, credentials, replica set, options), `Connection` (Start/Stop, exposes `Client()`, `Database()`).
- `mongo/tx/` — `WithTransaction(ctx, conn, fn)` helper. Re-exported at root once stable.
- `mongo/pipeline/` — typed aggregation builder. `Pipeline().Match(criteria).Group(...).Project(...).Sort(...).Build() []bson.D`. Mirrors Spring `Aggregation`.
- `mongo/query/` — criteria builder (`Where("age").Gte(38).And("active").Eq(true)`) used by pipeline and find. Optional; raw `bson.M` is always acceptable.
- `mongo/index/` — struct-tag-driven index registry + `EnsureIndexes(ctx, db, models...)`. Supports unique, compound, text, hashed, TTL, geospatial.
- `mongo/audit/` — `BeforeSave`/`BeforeUpdate` field mutators. Pluggable; implementation shared with `gorm/audit/` via an interface in `modules/datasource/` core.
- `mongo/changestream/` — `Consumer` as a `managed.Worker`; resumable via token; pluggable persistence for resume token.
- `mongo/gridfs/` — thin convenience over `gridfs.Bucket` (optional; promote out of v1 if no consumer needs it).

**Internal deps**:
- `modules/common/errs` — typed errors (`ErrConnectionNil`, `ErrSessionInactive`, `ErrIndexExists`).
- `modules/common/assert` — struct invariants only (never on external inputs).
- `modules/managed` — `Connection` and `changestream.Consumer` are `managed.Component`s.
- `modules/config` — URI/credentials loaded via viper-driven config.
- `modules/datasource/` core — shared `WithTransaction` contract, shared audit interfaces.
- `modules/telemetry/otel/mongo/` (separate module) — command monitor → OTel adapter, wired by consumer if telemetry is on.

**Go libraries to wrap**:
- `go.mongodb.org/mongo-driver/mongo` — primary; everything is a thin layer on top.
- `go.mongodb.org/mongo-driver/mongo/options` — options builders re-exported where useful.
- `go.mongodb.org/mongo-driver/bson` — type used in pipeline outputs; never replaced.
- `go.mongodb.org/mongo-driver/mongo/gridfs` — wrapped by `mongo/gridfs/` only.

**Out of scope for v1**:
- Repository abstraction / derived queries (never).
- Reactive variant (never).
- Vector search (lives in `modules/datasource/vector/`).
- Schema validation (lives in `modules/common/validation/`).
- Client-side field-level encryption — document the recipe; the driver supports it directly via `options.AutoEncryption()`.
- Geospatial helpers beyond what the pipeline/query builders already cover.
- GridFS (defer to v1.1 unless an early consumer needs it).
- Telemetry integration (lives in `modules/telemetry/otel/mongo/`; consumer wires it).

## 7. Open questions

1. **Pipeline builder scope**: implement all 24+ stages day one, or start with the 8 most-used (`match`, `group`, `project`, `sort`, `unwind`, `lookup`, `limit`, `skip`) and let consumers fall back to `bson.D` for the rest? Spring covers everything; the Yarumo Pareto answer is probably the smaller set.
2. **Audit interface location**: shared `datasource/audit` package or per-driver subpackage with a common interface? The cross-driver helper declared in § 1.3 implies a shared package.
3. **Index registry: struct tags vs explicit `IndexSpec` builders**? Spring uses annotations; Go tags work but couple the model. An explicit `IndexSpec{...}` registered next to the model is more idiomatic and matches existing `modules/cache/` patterns.
4. **Change-stream resume-token persistence**: in-memory only for v1, or pluggable `Store` interface backed by another collection / Redis from day one? `modules/messaging/kafka/cdc/` will hit the same problem — coordinate the design.
5. **Should the connection support multiple databases per `*mongo.Client`?** Spring's `MongoTemplate` is per-DB. Yarumo's `Connection` could expose `Database(name)` and let consumers fan out. Driver pooling makes this safe.
6. **Where do schema-validation rules go** when the team wants Mongo's native `validator: {$jsonSchema: ...}`? `modules/common/validation/` produces JSON Schema; wire as part of `index/` collection-create options, or a separate `mongo/schema/` package?
7. **CSFLE recipe**: document it in `mongo/README.md` only, or ship a typed `WithEncryption(keyVault, schemaMap)` constructor option? Real demand from DaaS is unclear.
8. **Migration path from go-feather-lib `datasource/mongo`**: any existing consumer surface to preserve? § 4.1 marks the migration as Low priority — fresh design is acceptable.
