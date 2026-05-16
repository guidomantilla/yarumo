# Spring Framework — Yarumo Analysis (DEEP)

> **Source**: https://docs.spring.io/spring-framework/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup; supersedes 2026-05-16 prior)
> **Spring Framework version**: 7.0.7 (current stable as of analysis)
> **Recommendation**: PARTIAL (per-sub-module verdict)

> **Re-analysis context**: `docs/ROADMAP_NEW_MODULES.md` was trimmed to 323 lines on 2026-05-15. Three sections were **deleted entirely**: § 3 Brainstorm domain modules, Annex A (Spring Messaging / Integration reference), and Annex B (Spring Security feature catalog). The roadmap now scopes only § 1 (`datasource/`, `auth/`, `messaging/`, `health/`, `boot/`), § 2 (`routegen` tool), and § 4 (go-feather-lib migration tracking). This rewrite (a) re-homes every cross-reference into the lean roadmap, (b) routes messaging analysis to the dedicated `spring-integration.md` companion (which still covers the deleted Annex A material), and (c) **proposes a NEW top-level `modules/testing/`** that previously lived as § 3.5 brainstorm and has no concrete home anymore.

---

## 1. Project summary

Spring Framework is the foundation of the entire Spring ecosystem — every other Spring project (Boot, Cloud, Security, Data, AMQP, Integration, Modulith, AI) builds on top of its **Core Container**, its **Data Access** abstractions, and one of its two **Web stacks** (Servlet or Reactive). At v7.0.7 the framework spans nine top-level reference sections: Overview, Core Technologies, Data Access, Web on Servlet Stack, Web on Reactive Stack, Testing, Integration, Language Support (Kotlin/Groovy), and Appendix.

Two structural changes since the previous analysis are worth flagging:

1. **Spring Framework 7 introduces native resilience primitives** (`@Retryable`, `@ConcurrencyLimit`, `RetryTemplate`) inside `core/resilience.html`. This is a notable scope expansion — historically resilience was the territory of Spring Cloud / Resilience4j. Yarumo's `modules/common/resilience/` (CB + RL registries, lazy, goroutine-free — YA-0076 closed 2026-05-13) covers the same ground from a different angle.
2. **Testing is now a top-level section** (no longer a sub-chapter of Core). Spring's testing surface has crystallized around MockMvc, WebTestClient, RestTestClient, TestRestTemplate, the TestContext framework, and `@MockitoBean`/`@MockitoSpyBean`. This validates promoting yarumo's testing concerns into their own module — see § 6 ROADMAP delta below.

For yarumo the value is **almost entirely conceptual, not transactional**. Spring's defining feature — the IoC container with auto-wiring, classpath scanning, annotation-driven configuration, AspectJ-style proxies, and SpEL — is exactly the "framework magic" that yarumo's `modules/boot/` design (§ 1.5 of the lean roadmap) explicitly inverts. What remains useful are the **patterns** Spring crystallized over twenty years of enterprise Java: the callback-template pair (JdbcTemplate, TransactionTemplate), exception translation, the `Cache`/`CacheManager` split, interceptor chains, ResponseEntity + ProblemDetail (RFC 9457), `@Async`/`@Scheduled` ergonomics, MockMvc's no-server HTTP testing, and the new `@Retryable`/`RetryTemplate` shape.

The big picture: **Spring is a finished encyclopedia of every concern a server-side application must address.** Yarumo's job is not to translate Spring to Go but to lift the *abstractions* (Cache interface, Transaction callback, problem-detail responses, scheduled-task ergonomics, retry template) while leaving behind the *implementation mechanics* (proxies, annotations, classpath scanning). Roughly 60% of Spring is reinventing Java's deficits (no first-class functions, weak DI, no goroutines, no `context.Context`); Go solves those at the language level. The remaining 40% is genuine cross-language insight worth absorbing.

---

## 2. Sub-module analysis

### 2.1. Core (IoC, Beans, ApplicationContext, AOP, SpEL, AOT, Resilience)

**Overview**: Spring Core is the IoC container — `ApplicationContext` reads bean definitions (annotations, Java config, XML), instantiates them, wires dependencies via reflection, applies `BeanPostProcessor` extensions (including AOP proxies), runs lifecycle callbacks (`@PostConstruct`/`@PreDestroy`), and publishes `ApplicationEvent`s. SpEL is an embedded expression language; AOP intercepts method calls via JDK or CGLIB proxies; AOT generates reflection-free code for GraalVM native images. **New in v7**: resilience primitives (`@Retryable`, `@ConcurrencyLimit`, `RetryTemplate`) — see § 2.1.x below.

**Pareto features (universal value)**:
- **Dependency Injection as a discipline**: components declare what they need, an orchestrator wires them. Universal across languages.
- **Bean lifecycle phases**: explicit init/destroy hooks — yarumo already has this in `managed.Lifecycle` (`Start`/`Stop`/`Done`).
- **Profile / Environment abstraction**: load different config per environment without code changes — yarumo's `config/` (viper-driven) already covers this.
- **Event publishing (ApplicationEventPublisher)**: in-process observer pattern — yarumo will cover this via `modules/messaging/events/` (§ 1.3 of the lean roadmap; the `events/` sub-package is **explicitly enumerated** in the proposed `modules/messaging/` layout).
- **Resource abstraction**: `ClassPathResource`, `FileSystemResource` — Go's `io/fs.FS` covers it.
- **Validation interface + BindingResult**: collect multiple field errors instead of throwing on first — yarumo has `modules/validation/` and `common/validation/` (leaves).
- **Type conversion (ConversionService)**: pluggable converters between types — partial relevance for HTTP binding and config parsing.
- **Resilience primitives** (new in v7): retry with backoff/jitter, concurrency limiting. See § 2.1.x.

**Map to yarumo**:
- IoC container → **rejected**. `modules/boot/` (§ 1.5) uses an explicit, typed `Container` with `Register[T]`/`Resolve[T]` generics. No reflection-based auto-wiring. No annotations. BeanFn factories are plain Go functions the consumer writes.
- BeanPostProcessor → **rejected**. The "inject yourself between framework and user code" hook is exactly the magic yarumo refuses. Decorator pattern at the call site instead.
- AOP / AspectJ / proxies → **rejected**. Go has no proxy primitive; interceptors are written as middleware (HTTP) or function wrappers. `modules/managed/` decorators (e.g. `WithMetrics`, `WithLogging`) are the idiomatic equivalent.
- SpEL → **rejected**. `modules/common/expressions/` (lexer/parser/eval) already covers Go-native expressions.
- AOT → **N/A**. Go is already AOT-compiled.
- ApplicationEvent → **already covered** by the planned `modules/messaging/events/` (§ 1.3 lean roadmap — uses nominal-typed pub/sub façade over `DirectChannel`).
- Profiles / Environment → **already covered** by `modules/config/`.
- Validator interface → **already covered** by `modules/validation/` + `common/validation/`.
- ConversionService → **partial relevance**. Gin's binding tags + `encoding/json` cover most of it; if `common/http/` ever needs custom URL-param converters, the *interface* (`Convert[S, T](src S) (T, error)`) is worth borrowing.

#### 2.1.x. Resilience (new in Spring Framework 7.0)

Spring 7 surfaces three primitives directly in `org.springframework.core.retry` and `org.springframework.core.concurrent`:

| Primitive | API | Notes |
|---|---|---|
| **`@Retryable`** | Annotation on a method | maxRetries, delay, jitter, multiplier, maxDelay, includes/excludes exception types. Publishes `MethodRetryEvent`. |
| **`@ConcurrencyLimit(n)`** | Annotation on a method | Throttle concurrent invocations. Set to 1 = exclusive lock. Pairs with virtual threads. |
| **`RetryTemplate`** | Programmatic | `template.invoke(() -> ...)` callback. Same backoff/jitter knobs as `@Retryable`. Returns or throws `RetryException` with `getExceptions()` history. |

What Spring 7 deliberately **does not** include in core: **circuit breaker, rate limiting, timeout**. Those remain in Spring Cloud Resilience4j.

**Map to yarumo**:
- Yarumo's `modules/common/resilience/` (YA-0076 — closed 2026-05-13) ships **CircuitBreaker + RateLimiter** registries (lazy, goroutine-free). This is **the opposite scoping** of Spring 7: Spring kept retry+concurrency-limit in core and pushed CB+RL to Cloud; yarumo keeps CB+RL in `common/` and **does not yet have a retry primitive**.
- **Recommendation**: file a follow-up to add `modules/common/resilience/retry` — a `RetryPolicy` value type + a `Retry(ctx, policy, fn) (T, error)` callback. The Spring 7 `RetryTemplate` shape (callback-style + structured `RetryException` history) is the cleanest borrow. Lives in `common/resilience/` alongside the existing CB+RL. Skip the annotation surface; the callback alone covers the use case in Go.
- The `@ConcurrencyLimit` primitive is essentially "semaphore around the method." Go has `semaphore.NewWeighted(n)` in `golang.org/x/sync/semaphore` — already idiomatic. No new yarumo work.

**Verdict**: **REJECT** the container/AOP/SpEL surface (with one tiny carve-out: `Converter[S, T]` as an interface shape if `common/http/` ever needs it). **ADOPT** the new resilience-retry shape into `modules/common/resilience/retry/` as a follow-up. Keep this section as the **canonical cautionary reference** when newcomers ask "why not a DI container like Spring?".

---

### 2.2. Data Access (JDBC, Transactions, DAO, ORM, R2DBC, OXM)

**Overview**: Spring Data Access splits into three layers. (a) **Transaction management** — `PlatformTransactionManager` interface with implementations per resource (JDBC/JPA/Hibernate/JTA), `@Transactional` declarative wrapper, and `TransactionTemplate` programmatic callback. (b) **JDBC** — `JdbcTemplate` and `NamedParameterJdbcTemplate` reduce boilerplate via callback-driven `RowMapper`/`ResultSetExtractor`/`PreparedStatementSetter`; exception translation converts checked `SQLException`s into a hierarchical `DataAccessException` tree. (c) **DAO support** — base classes (`JdbcDaoSupport`) and the `@Repository` stereotype that triggers exception translation. ORM (Hibernate/JPA) and R2DBC build on the same transaction abstraction.

**Pareto features (highest Go relevance)**:

1. **TransactionTemplate callback pattern** — `template.execute(status -> { ... })`. This is *the* pattern yarumo's `modules/datasource/` core plans (§ 1.1 of the lean roadmap, "cross-driver features": "`WithTransaction(ctx, db, fn)` helper — Go has no `@Transactional` AOP; a single function-shaped helper that runs `fn` inside a tx and rolls back on error covers 90% of the use case"). The Spring shape is:
   ```java
   <T> T execute(TransactionCallback<T> action) throws TransactionException;
   ```
   Go equivalent (planned in § 1.1):
   ```go
   func WithTransaction[T any](ctx context.Context, db DB, fn func(ctx context.Context, tx Tx) (T, error)) (T, error)
   ```
   Key insights to copy: (i) the callback receives a `status` handle so user code can `setRollbackOnly()` without throwing; (ii) panic in fn → rollback + repanic (Go's `defer recover()` equivalent); (iii) nested `WithTransaction` should either join (REQUIRED) or open savepoint (NESTED) — start with REQUIRED-only, savepoints behind a flag.

2. **Propagation enum (REQUIRED / REQUIRES_NEW / NESTED / SUPPORTS / NOT_SUPPORTED / NEVER / MANDATORY)** — 90% of real usage is REQUIRED + REQUIRES_NEW + NESTED. Yarumo's `WithTransaction` should expose those three. Skip MANDATORY/NEVER (anti-patterns in disguise) and SUPPORTS/NOT_SUPPORTED (Go-explicit code doesn't need them).

3. **Isolation levels** — direct passthrough to the SQL driver (`sql.TxOptions{Isolation: sql.LevelReadCommitted}` in Go). The Spring contribution is just naming the levels; `database/sql` already does this.

4. **Exception translation** — the abstraction that maps vendor-specific error codes (MySQL #1062, Postgres "23505") into typed exceptions (`DataIntegrityViolationException`, `EmptyResultDataAccessException`, `CannotAcquireLockException`). **High Go value**. Yarumo's `modules/datasource/gorm/` should ship a `TranslateError(err) error` helper that returns typed sentinel errors:
   ```go
   var (
       ErrUniqueViolation     = errs.New("unique violation")
       ErrForeignKeyViolation = errs.New("foreign key violation")
       ErrLockNotAvailable    = errs.New("lock not available")
       ErrSerializationFailure = errs.New("serialization failure")
       ErrEmptyResult         = errs.New("empty result")
   )
   ```
   Driver-specific subpackages (`postgres/`, `mysql/`, `mongo/`) own the actual mapping table. Pairs with `modules/common/errs/`. **Add to the design of every `modules/datasource/<driver>/` package** as it lands.

5. **RowMapper / ResultSetExtractor / PreparedStatementSetter callback trio** — Go's `database/sql.Rows.Scan(&dst...)` already gives you most of this. The conceptual takeaway is: **a small interface for "row → T" composes much better than a fat ORM**. If yarumo ever ships a `datasource/jdbc/` (raw SQL) sub-package alongside `gorm/`, this shape is canonical:
   ```go
   type RowMapper[T any] func(rows *sql.Rows) (T, error)
   func Query[T any](ctx context.Context, db DB, sql string, args []any, mapper RowMapper[T]) ([]T, error)
   ```

6. **NamedParameterJdbcTemplate (`:email` instead of `?`)** — niche but useful when SQL has many parameters. Go has no built-in; libraries like `sqlx` cover it. Don't reinvent; `gorm` already supports `Where("email = ?", email)`.

7. **`@Repository` + automatic exception translation** — rejected (annotation-driven). The translator function (point 4) is what survives.

8. **Declarative `@Transactional`** — **rejected**. AOP-driven, swallows control flow, hides commit boundaries. Yarumo standard already prohibits inline-assignment-in-`if` for the same reason: explicit > magic.

9. **Data lifecycle policies** (`@CreatedDate`, `@LastModifiedBy`, soft delete, retention) — Spring Data Auditing covers row-level auditing. Yarumo's lean roadmap (§ 1.1, "cross-driver features") already enumerates: *Row-level audit hooks — `CreatedBy` / `LastModifiedBy` / `CreatedAt` / `UpdatedAt` columns auto-populated via `BeforeSave` / `BeforeUpdate` hooks. Implemented in `modules/datasource/gorm/`* and *Data-lifecycle policies — retention, archival, soft / hard delete (GDPR-aware).* Spring confirms this scoping. Note the lean roadmap pending decision: implement in the gorm driver only, or factor into a `datasource/lifecycle/` sub-package — Spring's split (`spring-data-commons` for the auditing primitives, per-store implementations) suggests **a small `datasource/lifecycle/` is the right factoring** when a second driver needs it.

**Map to yarumo**:
- `WithTransaction(ctx, db, fn)` → already designed in `modules/datasource/` core (§ 1.1). Spring's `TransactionTemplate` confirms this is the right shape; no design changes needed.
- Exception translation → **add to design** of every `modules/datasource/<driver>/` package. Postgres + MySQL drivers in v1; others when ticketed.
- Propagation enum → **scope down to REQUIRED + REQUIRES_NEW + NESTED**. Add to `WithTransaction` options as `WithPropagation(...)`.
- Isolation levels → **passthrough** to `sql.TxOptions`.
- RowMapper-style typed query helper → **defer**; let consumers use `gorm` or `database/sql` directly until a real use case appears.
- Data lifecycle policies (`@CreatedDate`/`@LastModifiedBy`) → already in scope for `modules/datasource/gorm/` (§ 1.1, "Row-level audit hooks"). Note: the event-level audit trail Spring Modulith covers (different concern entirely) **has no current home in the lean roadmap** — § 3 (Brainstorm) was deleted. If audit trail surfaces as a real DaaS need, it requires either a new module or a sub-package decision; **not blocking here**.
- DAO base classes → **rejected**. Embedding/composition in Go is explicit; no need for a template-method base.

**Verdict**: **PARTIAL — adopt aggressively**. The two patterns yarumo should hardwire into `modules/datasource/` are (i) the `WithTransaction(ctx, db, fn)` callback and (ii) typed error translation per driver. These two cover ~90% of the Java app-developer pain that Spring Data Access addresses, and they're both genuinely language-neutral wisdom.

---

### 2.3. Web on Servlet Stack (DispatcherServlet, MVC, Filters, Interceptors, REST, WebSocket)

**Overview**: Spring MVC is built around `DispatcherServlet` — the front controller that dispatches requests to handler methods (annotated `@Controller`/`@RestController` or functional `RouterFunction`s) selected by `HandlerMapping`, executes them via `HandlerAdapter`, optionally renders a view, and routes errors through `HandlerExceptionResolver`. Around the handler sit two interception points: **Servlet filters** (servlet-spec layer, can run before Spring even sees the request) and **`HandlerInterceptor`** (Spring layer, knows the chosen handler, has three hooks: `preHandle`/`postHandle`/`afterCompletion`). Beyond routing, the section covers content negotiation, CORS, validation, `ResponseEntity`, `ProblemDetail` (RFC 9457), WebSocket/STOMP, server-sent events, and MockMvc testing. **New in v7**: a dedicated "API Versioning" sub-section and refined "Error Responses" guidance around ProblemDetail.

**Pareto features (highest Go relevance, ranked)**:

1. **Filter chain composition** — Spring's filter chain matters because (a) ordering is canonical and well-documented, and (b) filters compose without knowing about each other. Gin middleware chains already give yarumo this. Yarumo's `managed/server_http` should ship a documented **canonical middleware order** (recovery → request-id → logging → tracing → cors → metrics → auth → ratelimit → tenancy → handler) so consumers don't reinvent it per service. This is the highest-leverage borrow.

2. **HandlerInterceptor's three-hook contract (`preHandle`/`postHandle`/`afterCompletion`)** — Gin middleware is a single function with `c.Next()` separating before/after, which collapses to two hooks. The third hook (`afterCompletion` — runs *even on exception*, after the response is committed) is genuinely useful for cleanup, audit emission, and metric finalization. Yarumo's `managed/server_http` middleware authors should be guided to use `defer` for the `afterCompletion` slot — a documentation pattern, not new code.

3. **ResponseEntity** — Spring's fluent builder for status + headers + body:
   ```java
   ResponseEntity.status(201).header("Location", uri).body(user);
   ResponseEntity.notFound().build();
   ResponseEntity.ok(body);
   ```
   In Gin, this is `c.JSON(201, body)` + manual header setting. **Worth wrapping** in `common/http/` as a small builder:
   ```go
   http.Response().Status(201).Header("Location", uri).JSON(c, user)
   ```
   Low LOC, high ergonomic value, no framework lock-in. **Needs a concrete home** — `common/http/response/` is the natural location. The lean roadmap doesn't currently mention this; file as a follow-up when `common/http/` materializes.

4. **ProblemDetail (RFC 9457, formerly 7807)** — standardized error response shape:
   ```json
   {"type": "...", "title": "...", "status": 400, "detail": "...", "instance": "/req/123", "<extension>": "..."}
   ```
   Yarumo's `common/errs/` already has typed errors; **adding a `common/http/problem` package** that renders any `errs.TypedError` to a ProblemDetail JSON body is the cleanest mapping. Spring's `ProblemDetail.forStatusAndDetail(status, detail)` factory shows the shape. This is one of the cleanest borrows in the whole analysis — pure data format with an RFC behind it, no framework baggage. **Needs a concrete home** — propose `modules/common/http/problem/`.

5. **`@ControllerAdvice` + `@ExceptionHandler` (global exception handling)** — the *mechanism* is annotation-driven (rejected), but the *pattern* (centralized exception → HTTP mapping) is universally useful. In yarumo: a single Gin middleware that catches `panic`s and known `errs.TypedError` types, maps them to status codes via a registered table, and emits ProblemDetail. Ships in `managed/server_http/middleware/` (which pairs naturally with the `DefaultStack()` builder below).

6. **Content negotiation** — Spring picks response format from `Accept` header, URL extension, or query parameter (`?format=json`). 99% of yarumo consumers will be JSON-only. **Defer** until a multi-format consumer appears (XML, MessagePack, protobuf-over-HTTP).

7. **CORS** — first-class in Spring (global config + per-route `@CrossOrigin`). Gin has `gin-contrib/cors`. Yarumo should standardize the configuration shape (`AllowedOrigins`, `AllowedMethods`, `AllowedHeaders`, `MaxAge`, `AllowCredentials`) and ship a default-strict preset in `managed/server_http/`. Low effort.

8. **`HandlerExceptionResolver` chain** — the *strategy* (multiple resolvers tried in order, first match wins) is the same pattern as middleware. Already covered by Gin middleware chains.

9. **REST clients (RestTemplate / WebClient / RestClient)** — Spring 7 stabilized `RestClient` as the modern synchronous shape:
   ```java
   restClient.get().uri("/users/{id}", id).retrieve().body(User.class);
   ```
   Go has `net/http.Client` + many wrapper libraries. **Needs a concrete home** — propose `modules/common/http/client/` for retries, circuit breakers, timeouts, structured logging. Spring's `exchange()` fluent shape is worth borrowing. Pairs with `modules/common/resilience/` for the CB/RL integration.

10. **MockMvc** — covered under § 2.6 (Test).

11. **API Versioning** (new sub-section in v7) — Spring 7 documents three strategies: URL path (`/v1/users`), media-type (`Accept: application/vnd.example.v1+json`), query parameter (`?version=1`). Yarumo should pick **one** convention per service and document it; `managed/server_http/` can ship helpers for path-based versioning (most common, most cacheable).

12. **WebSocket / STOMP / SSE** — WebSocket isn't in yarumo scope yet; SSE is point-feature potential in `managed/server_http/`. Spring's STOMP-over-WebSocket convention is enterprise-Java-specific. **Defer** — no current consumer.

13. **Functional endpoints (RouterFunction DSL)** — Spring's alternative to annotation-driven controllers. Looks like Go HTTP routing already. Nothing to borrow.

14. **HTTP message converters (`HttpMessageConverter<T>`)** — automatic body marshaling. Go has `encoding/json`. Niche cases (XML, MsgPack) handled per consumer.

**Map to yarumo**:
- **Canonical middleware order** → ship as documentation in `managed/server_http/CODING_STANDARDS.md` or as a `DefaultStack()` constructor that wires the standard chain. **HIGH VALUE.**
- **ResponseEntity-style builder** → propose new `modules/common/http/response/` (small package). Optional; gin's `c.JSON()` works.
- **ProblemDetail renderer** → propose new `modules/common/http/problem/` (small package). Maps `errs.TypedError` → RFC 9457 JSON. **HIGH VALUE — explicitly recommended.**
- **Global exception handler middleware** → recovery + typed-error mapping. Ships in `managed/server_http/middleware/` (lives there alongside other middleware).
- **CORS configuration shape** → standardize options struct + default preset in `managed/server_http/`.
- **HandlerInterceptor's three hooks** → documentation pattern (use `defer` for afterCompletion).
- **API Versioning** → small helper in `managed/server_http/` for path-based versioning.
- **REST client wrapper** → propose new `modules/common/http/client/` (depends on `modules/common/resilience/`).
- **Content negotiation, WebSocket, STOMP, SSE** → DEFER.

**Verdict**: **PARTIAL — patterns inform middleware design**. The big takeaways are (a) canonical middleware order, (b) ProblemDetail rendering, (c) ResponseEntity-style builder, (d) HTTP-client wrapper. None require new top-level modules; all fit cleanly under a new `modules/common/http/` sub-tree (`response/`, `problem/`, `client/`, `middleware/`) and into `managed/server_http/`. The deeper Servlet abstractions (DispatcherServlet, HandlerMapping, HandlerAdapter, view resolution) are Java-specific machinery that Gin/Echo replace at the language level.

---

### 2.4. Web on Reactive Stack (WebFlux, WebClient, RSocket, reactive WebSocket)

**Overview**: Spring's alternative web stack built on Project Reactor (`Mono<T>`, `Flux<T>`). Same conceptual surface as Servlet MVC (controllers, routers, exception handlers, interceptors) but every return type is a reactive publisher; the entire stack is non-blocking and back-pressure-aware. Pairs with R2DBC for reactive database access. Spring 7 retains a separate WebClient documentation under WebFlux, including `retrieve()`, `exchange()`, filters, attributes, context propagation, and a "synchronous use" sub-section that admits the reactive abstraction is overkill for many cases.

**Pareto features**: nothing transferable. The reactive paradigm exists in Java specifically because the JVM thread-per-request model is expensive; reactive lets one OS thread multiplex thousands of in-flight requests. (Note: Java 21 virtual threads largely eliminate this motivation; Spring's reactive stack survives mostly for backward compatibility and existing reactive ecosystems.)

**Map to yarumo**: Go solves the underlying problem differently and natively:
- One **goroutine per request** is cheap (~4 KB stack vs. Java's 1 MB OS thread). Multiplexing a handler over an event loop is unnecessary.
- **Back-pressure** in Go is provided by channels with bounded capacity (`make(chan T, N)`).
- **Streaming responses** use `http.ResponseWriter` + `Flush()`; SSE is direct.
- **Async cancellation** is `context.Context` everywhere — far more ergonomic than Reactor's `Mono.timeout()`.

There is no `Mono`/`Flux` abstraction worth borrowing because Go's concurrency primitives subsume the use cases at the language level.

**Verdict**: **REJECT**. Goroutines + channels + `context.Context` cover every concern WebFlux exists to solve. If yarumo ever ships SSE or HTTP/2 server push, those are point features in `managed/server_http/`, not a parallel reactive stack.

---

### 2.5. Integration (JMS, JMX, Email, Tasks, Scheduling, Cache, Observability, REST Clients)

**Overview**: A grab-bag of enterprise integrations and cross-cutting features layered on the core container. The single most reusable abstractions are the **Cache** interface family and the **task scheduling** annotations (`@Async`, `@Scheduled`). REST clients (RestTemplate, WebClient, RestClient) live under this section in v7 — reorganized from "Web" for v6+. JMS is the messaging precursor to Spring AMQP and Spring Kafka. JMX exports beans for monitoring (largely superseded by Micrometer/OTel). Email is `JavaMailSender`. Observability integrates Micrometer for metrics + tracing.

**Pareto features (ranked by Go relevance)**:

1. **Cache abstraction (`Cache` + `CacheManager`)** — the clean two-interface split that Spring crystallized:
   ```java
   public interface Cache {
       ValueWrapper get(Object key);
       <T> T get(Object key, Class<T> type);
       void put(Object key, Object value);
       void evict(Object key);
       void clear();
   }
   public interface CacheManager {
       Cache getCache(String name);
       Collection<String> getCacheNames();
   }
   ```
   The shape is provider-neutral: backends include Caffeine, Redis, EhCache, Hazelcast — `@Cacheable("books")` works against any. Yarumo's `modules/cache/` (added in Phase 2 per CLAUDE.md) **already follows this split** via the generic `Cache[K, V]` interface (criterion 4 of CODING_STANDARDS.md). The mapping is direct: Spring `Cache` ↔ yarumo `Cache[K, V]`, Spring `CacheManager` ↔ a typed registry. **Validates the existing design.**

2. **`@Async` (background execution)** — Spring annotation that runs a method on a different thread:
   ```java
   @Async
   public Future<String> doWork() { ... }
   ```
   Go's equivalent is `go fn()` — language-native, no annotation needed. The conceptual gap is **structured execution** (executor pool, propagation of MDC/trace context). Yarumo's `managed/` should standardize a `WorkerPool` primitive with bounded concurrency + context propagation (this exists in part via `managed/BaseWorker`).

3. **`@Scheduled` (cron + fixed-rate + fixed-delay scheduling)** — Spring annotation:
   ```java
   @Scheduled(fixedDelay = 5000) ...
   @Scheduled(fixedRate = 5000) ...
   @Scheduled(cron = "0 0 * * * *") ...
   ```
   Yarumo has `managed/CronWorker` — direct equivalent. Spring's contribution is the **dual support for `fixedDelay`/`fixedRate`/`cron` in one annotation**. `CronWorker` should expose all three modes via constructor options. **Low-effort improvement** — file as a `modules/managed/` enhancement (Phase 3 — see milestone #9). Open question: does the current `CronWorker` already support `fixedDelay` and `fixedRate`, or only cron? Check before filing.

4. **`@Cacheable` / `@CachePut` / `@CacheEvict` annotations** — declarative cache aside. The **annotation-driven mechanism is rejected** (AOP magic). The *pattern* is `cache.GetOrCompute(key, fn)` — already idiomatic in Go and likely already in `modules/cache/`. Worth confirming the API exposes this.

5. **Observability** — Spring delegates entirely to Micrometer (metrics) and Micrometer Tracing (formerly Sleuth). Yarumo delegates to OTel via `modules/telemetry/otel/`. **Direct parallel** — no new work, just confirming the architecture mirrors.

6. **REST clients (RestClient, RestTemplate, WebClient)** — covered under § 2.3 point 9. The fluent `RestClient.get().uri(...).retrieve().body(T.class)` shape is the modern Spring 7 default.

7. **JMS / AMQP / Kafka** — **covered by `modules/messaging/`** (§ 1.3 of the lean roadmap). See the companion `spring-integration.md` for the full Spring Integration / Spring Messaging analysis (the deleted Annex A content was relocated there). JMS is one of the brokers `messaging/` can plug into; the EIP layer (channels, endpoints, transformers, filters, routers, splitters, aggregators) is already enumerated in `messaging/endpoints/`.

8. **JMX** — superseded by OTel metrics. Reject.

9. **Email (`JavaMailSender`)** — niche; yarumo has no current use case. The deletion of § 3 Brainstorm removed the "notifications" placeholder; **no current home** in the lean roadmap. If a real consumer surfaces, file as a new `modules/notification/` proposal at that time.

10. **JVM checkpoint restore (CRaC)** — Java-specific; Go has fast startup natively.

**Map to yarumo**:
- Cache `Cache`/`CacheManager` split → **already implemented** in `modules/cache/`. Spring's split validates the design.
- `@Async` → covered by goroutines + `managed/BaseWorker` for bounded concurrency.
- `@Scheduled` → `managed/CronWorker` already covers cron; **file an enhancement** for unified `fixedDelay`/`fixedRate`/`cron` modes if not yet present.
- `@Cacheable` → reject mechanism; confirm `cache.GetOrCompute(key, fn)` API exists.
- Observability/Micrometer → covered by `telemetry/otel/`.
- JMS/AMQP/Kafka → covered by `messaging/` (§ 1.3) — see `spring-integration.md` companion.
- JMX → reject.
- Email → defer; no home in lean roadmap; revisit when a consumer materializes.

**Verdict**: **PARTIAL — mostly already done**. The Cache abstraction validates `modules/cache/`'s shape; `@Scheduled` informs a small `CronWorker` enhancement (fixed-delay/fixed-rate modes if missing). JMS/AMQP/Kafka coverage was Annex-A territory and now lives in `modules/messaging/` per the lean roadmap. Everything else is already covered elsewhere in yarumo or rejected on principle.

---

### 2.6. Testing (TestContext, MockMvc, WebTestClient, RestTestClient, TestRestTemplate)

**Overview**: Spring 7's testing arm is now a top-level reference section (no longer a sub-chapter of Core). It crystallizes around five client tools — **MockMvc** (servlet stack, no-server), **WebTestClient** (reactive), **RestTestClient** (testing REST clients via mock server), **TestRestTemplate** (real-server integration with random port), and the **TestContext framework** (context caching, listener extension points). Bean overriding in v7 standardized on `@MockitoBean` and `@MockitoSpyBean` (succeeding the older `@MockBean`/`@SpyBean`). Database-test ergonomics include `@Transactional` auto-rollback, `@Sql` script execution, `@SqlConfig`, and `@DynamicPropertySource` (the Testcontainers integration point).

**Pareto features**:

1. **MockMvc — the canonical pattern.** Set up a controller with mocked dependencies, fire HTTP requests against the in-memory mux, assert on status + headers + body. Spring's shape:
   ```java
   mockMvc.perform(get("/users/{id}", 1).contentType(MediaType.APPLICATION_JSON))
       .andExpect(status().isOk())
       .andExpect(jsonPath("$.name").value("John"));
   ```
   Go has `httptest.NewRecorder()` + `httptest.NewRequest()` as primitives. The Pareto gap is a **fluent assertion builder** on top. **HIGH VALUE — recommended.** Concrete home: see § 6 below (propose new `modules/testing/`).

2. **TestRestTemplate (full integration with a real server)** — Spring spins a random-port server and exposes a client preconfigured to hit it. Go equivalent: `httptest.NewServer(handler)` + a typed client wrapper. Worth a small helper:
   ```go
   srv := testing.NewIntegrationServer(t, router)
   resp, err := srv.Client().Get("/users/1")
   ```
   `t.Cleanup` handles shutdown.

3. **RestTestClient (new in v7)** — tests outbound REST clients via a mock server with recorded responses. Go equivalent: `httptest.NewServer(handler)` with hand-rolled response stubs, or library wrappers like `httpmock`. Worth a thin canonical wrapper in `modules/testing/` if `modules/common/http/client/` ships.

4. **WebTestClient** — reactive variant. N/A.

5. **TestContext framework** — Java-specific context-caching machinery. Go test runs are per-package processes; nothing direct to port. The conceptual takeaway — **package-level fixture caching** — maps to Go's `TestMain` + package-level state.

6. **`@MockitoBean` / `@MockitoSpyBean`** — replaces a bean with a Mockito mock in the context. Go uses interface mocks (handwritten or gomock-generated); no framework hook needed.

7. **`@Transactional` on tests (automatic rollback)** — wrap each test in a tx that's rolled back at the end, so the DB is clean. Yarumo equivalent: **`testing/database.WithRollback(t, db, fn)`** helper for tx-scoped tests. Pairs with testcontainers (`testing/containers/postgres`) for the heavier "fresh container per test class" alternative.

8. **`@Sql` script execution** — declarative SQL fixture loading. Yarumo equivalent: a `testing/database.LoadSQL(t, db, files...)` helper.

9. **`@DynamicPropertySource`** — register runtime properties (e.g. testcontainer JDBC URL) before context starts. Yarumo equivalent: `t.Setenv()` + `config.Reload()` or constructor-time wiring. Less ceremonial in Go.

10. **Test slicing (`@WebMvcTest`, `@DataJpaTest`)** — Spring Boot pattern. Yarumo equivalent: explicit `BeanFn` composition (via `modules/boot/`) — wire only the slice you want.

**Map to yarumo**:
- MockMvc-style fluent HTTP fake → propose new **`modules/testing/fakes/http/`** package (see § 6).
- TestRestTemplate-style integration server → **`modules/testing/fakes/http/`** integration variant.
- RestTestClient-style outbound stub → **`modules/testing/fakes/http/server.go`** with stub helpers.
- Test-rollback transaction helper → **`modules/testing/database/WithRollback(t, db, fn)`**.
- Testcontainers wrappers → **`modules/testing/containers/`** (postgres, mongo, redis, kafka, rabbitmq, etc.).
- `@Sql`-style SQL loader → **`modules/testing/database/LoadSQL(t, db, files...)`**.
- LLM eval harness (deterministic evaluator + golden traces) → **`modules/testing/llm/`** when Aluna or DaaS need it.
- Property-based testing primitives → **`modules/testing/property/`** when needed.
- Contract testing (Pact-style) → **`modules/testing/contracts/`** when consumers materialize.
- TestContext, `@MockitoBean` → reject (Java-specific machinery).

**Verdict**: **PARTIAL — informs the design of a new `modules/testing/` module**. The MockMvc pattern is the single most copyable testing idea in Spring. Go has the primitives (`httptest`) but no canonical fluent layer. Yarumo can ship one in ~300 LOC and consumers will use it everywhere. **See § 6 below — this is the headline ROADMAP delta from this re-analysis.**

---

### 2.7. Language Support (Kotlin, Apache Groovy)

**Overview**: First-class Kotlin support (DSL for bean registration, coroutine-friendly suspend functions, null-safety annotations) and Apache Groovy support for dynamic/script-style configuration. Spring 7 leans heavily into Kotlin; Groovy is legacy.

**Pareto features**: none for yarumo. Yarumo is Go-only.

**Verdict**: **REJECT** (N/A).

---

### 2.8. Other sub-areas (Aspects, Instrumentation, JCL, OXM, AOT)

| Sub-area | Description | Verdict |
|---|---|---|
| **Aspects** | AspectJ integration (compile-time / load-time weaving for `@Transactional`, `@Cacheable`, etc.) | **REJECT** — Go has no proxies, no AspectJ |
| **Instrumentation** | Java agents for bytecode transformation | **REJECT** — Go's pprof + OTel autoinstrumentation cover the use case |
| **JCL (Jakarta Commons Logging)** | Logging facade | **REJECT** — yarumo uses `log/slog` |
| **OXM (Object/XML Mapping)** | JAXB, Castor, XStream marshalling | **REJECT** — Go uses `encoding/xml` directly; no use case |
| **AOT optimizations** | GraalVM native image support, reflection-free codegen | **N/A** — Go is AOT-compiled natively |

---

## 3. Cross-cutting findings

### 3.1. Patterns worth adopting (mapped to lean-roadmap homes)

| Pattern | Spring source | Yarumo home (lean roadmap) | Priority | Notes |
|---|---|---|---|---|
| **`WithTransaction(ctx, db, fn)` callback** | `TransactionTemplate.execute` | `modules/datasource/` core (§ 1.1) | **HIGH** — already planned, design validated |
| **Typed DB error translation** | `SQLExceptionTranslator` → `DataAccessException` | `modules/datasource/<driver>/` per-driver translator | **HIGH** — file as design item for first gorm driver |
| **`ProblemDetail` (RFC 9457) renderer** | `ProblemDetail` + `forStatusAndDetail` | **propose new `modules/common/http/problem/`** | **HIGH** — clean RFC mapping, no framework baggage |
| **Canonical middleware order** | `SecurityFilterChain` ordering | `managed/server_http/` (docs + `DefaultStack()` constructor) | **HIGH** — pure documentation/builder win |
| **`ResponseEntity`-style builder** | `ResponseEntity.status().header().body()` | **propose new `modules/common/http/response/`** | MEDIUM — optional ergonomics |
| **Global exception handler middleware** | `@ControllerAdvice` + `@ExceptionHandler` | `managed/server_http/middleware/recover_and_map/` | **HIGH** — covers panic + typed-error → ProblemDetail |
| **REST-client wrapper (retries, CB, RL, logging)** | `RestClient` | **propose new `modules/common/http/client/`** (uses `common/resilience/`) | MEDIUM |
| **MockMvc-style HTTP fake builder** | `MockMvc.perform().andExpect()` | **propose new `modules/testing/fakes/http/`** | **HIGH** — single most copyable testing idea |
| **Integration server with auto-cleanup** | `@SpringBootTest(RANDOM_PORT)` + `TestRestTemplate` | **new `modules/testing/fakes/http/` integration variant** | MEDIUM |
| **Test-rollback DB helper** | `@Transactional` on test methods | **new `modules/testing/database/WithRollback(t, db, fn)`** | MEDIUM |
| **Testcontainers fixtures** | Boot Testcontainers integration | **new `modules/testing/containers/`** (per backend) | MEDIUM — pairs with `datasource/` drivers |
| **`CronWorker` dual-mode (`fixedDelay` + `fixedRate` + `cron`)** | `@Scheduled(fixedDelay/fixedRate/cron)` | `modules/managed/CronWorker` enhancement (Phase 3) | LOW — confirm current API first |
| **Propagation enum on `WithTransaction` (REQUIRED / REQUIRES_NEW / NESTED)** | Propagation enum | `modules/datasource/` core options | MEDIUM — scope down to 3 modes |
| **`Retry(ctx, policy, fn)` callback** | `RetryTemplate.invoke` (new in v7) | `modules/common/resilience/retry/` (follow-up) | MEDIUM — pairs with existing CB+RL |
| **`Converter[S, T]` interface** | `Converter<S, T>` + `ConversionService` | `modules/common/conv/` (only if a consumer surfaces) | DEFER — speculative |
| **API versioning helper (path-based)** | API Versioning section (new in v7) | `managed/server_http/` | LOW — pick one convention, document |

### 3.2. Patterns already covered (validation of existing design)

| Pattern | Yarumo location | Spring counterpart |
|---|---|---|
| `Cache` / `CacheManager` split | `modules/cache/` | `Cache` + `CacheManager` |
| Environment-driven config | `modules/config/` (viper) | `Environment` + `@Profile` |
| Validator with multi-error accumulation | `modules/validation/` + `common/validation/` | `Validator` + `BindingResult` |
| Event publishing (in-process pub/sub) | `modules/messaging/events/` (§ 1.3) | `ApplicationEventPublisher` |
| Background workers with lifecycle | `modules/managed/BaseWorker`, `CronWorker` | `@Async`, `@Scheduled`, `TaskExecutor` |
| Observability (metrics + tracing) | `modules/telemetry/otel/` | Micrometer + Micrometer Tracing |
| JMS / AMQP / Kafka integration | `modules/messaging/` (§ 1.3) | Spring JMS + Spring Messaging + Spring Integration (see `spring-integration.md`) |
| Resilience CB + RL | `modules/common/resilience/` (lazy, goroutine-free) | (Spring 7 has retry+concurrency-limit in core; CB+RL still in Spring Cloud Resilience4j) |
| Resource abstraction | Go `io/fs.FS` + `embed.FS` | `Resource` hierarchy |
| Expression language | `modules/common/expressions/` | SpEL |
| Application wiring | `modules/boot/` (§ 1.5 — Container + BeanFn + Run, explicit) | `ApplicationContext` (rejected mechanism, same goal) |
| Authentication / Authorization | `modules/auth/` (§ 1.2) | Spring Security (out of Framework scope; see separate `spring-security.md`) |
| OAuth2 client + resource server | `modules/auth/oauth2/` (§ 1.2) | Spring Security OAuth2 |
| LDAP directory access | `modules/datasource/ldap/` + `auth/` provider (§ 1.1, § 1.2) | Spring LDAP + Spring Security LDAP |
| Health endpoints | `modules/health/` runtime + `modules/common/health/` leaves (§ 1.4) | Spring Boot Actuator `/health` |

### 3.3. Anti-patterns to lock in (cautionary reference)

Yarumo's design deliberately rejects these Spring features. Treat this section as a checklist for code review:

| Spring feature | Why rejected | Yarumo's idiomatic replacement |
|---|---|---|
| **IoC container with auto-wiring** | Reflective wiring is opaque, slow at startup, breaks compile-time guarantees | Explicit `Container.Register[T]` + `Resolve[T]` in `modules/boot/` (§ 1.5). Wiring is code, not metadata |
| **Annotation-driven configuration** (`@Component`, `@Bean`, `@Autowired`) | Annotations are metadata; Go functions are first-class. Functions are easier to read, debug, test | `BeanFn func(ctx, c *Container) error` in `modules/boot/` |
| **AOP / AspectJ / `@Aspect`** | Proxy-based interception hides control flow. No Go equivalent without runtime codegen | Middleware (HTTP) + function decorators (`WithMetrics`, `WithLogging`) + manual wrappers |
| **`BeanPostProcessor` hooks** | Lets third-party code mutate your bean graph silently | Compose explicitly at construction site |
| **Classpath scanning / component discovery** | Magic registration based on package layout | Explicit `BeanFn` registration in `Run(...)` |
| **SpEL (Spring Expression Language)** | Yet another expression dialect, Java-bytecode-bound | `modules/common/expressions/` (Go-native lexer/parser/eval) |
| **`@Transactional` declarative transactions** | AOP-driven, hides commit boundaries, surprising rollback rules (checked exc not rolled back by default) | Explicit `WithTransaction(ctx, db, fn)` callback (§ 1.1 + § 2.2 above) |
| **`@Async` annotation** | Spawns a thread implicitly; loses caller context | `go fn()` (explicit) + `managed/BaseWorker` for bounded pools |
| **`@Cacheable` annotation** | AOP-driven; hides which calls hit cache vs. compute | `cache.GetOrCompute(key, fn)` (explicit call site) |
| **`@Retryable` annotation** (new in v7) | AOP-driven; the *pattern* is good, the *mechanism* is rejected | `resilience.Retry(ctx, policy, fn)` callback (proposed follow-up) |
| **`@ConcurrencyLimit` annotation** (new in v7) | AOP-driven semaphore-around-method | `semaphore.NewWeighted(n).Acquire(ctx, 1)` explicit |
| **Reactive `Mono`/`Flux`** | Solves a JVM thread-cost problem Go doesn't have | Goroutines + channels + `context.Context` |
| **`ApplicationContext` lifecycle events** | Hooks anyone can intercept, side-channel comms | `managed.Lifecycle` explicit `Start`/`Stop`/`Done` |
| **`BeanFactoryPostProcessor`** | Mutate bean definitions before instantiation — pure metaprogramming | N/A in yarumo's explicit model |
| **`@MockitoBean` / context-replacement testing** | Java-context-specific | Interface mocks + dependency injection at constructor (planned in `modules/testing/`) |
| **DAO base classes (`JdbcDaoSupport`)** | Inheritance-based reuse, brittle | Composition (struct embedding) at the call site |
| **`@ControllerAdvice` global handlers via annotation** | The *pattern* is good, the *mechanism* is rejected | Recovery + typed-error → ProblemDetail **middleware** instead |

### 3.4. Observations on Spring's overall structure

A few cross-cutting observations after re-reading the entire Framework reference at v7.0.7:

1. **Spring's biggest insight is the template-callback duo** — `JdbcTemplate(sql, RowMapper)`, `TransactionTemplate(callback)`, `RestTemplate(uri, ResponseExtractor)`, and now `RetryTemplate(callback)` in v7. The pattern is: "we own the resource lifecycle and boilerplate; you write the interesting bit as a function." Go has this natively via higher-order functions; yarumo's `WithTransaction(ctx, db, fn)` is the same idea. **Adopt this as a yarumo-wide convention** for any resource-acquiring API. Note v7's `RetryTemplate` extends the pattern to retry, which yarumo's `common/resilience/` should mirror.

2. **Spring's worst insight is annotation-driven everything** — every clean idea (`@Cacheable`, `@Transactional`, `@Async`, `@Scheduled`, `@Retryable`, `@ConcurrencyLimit`) is bolted onto a method via annotation that triggers a runtime proxy. This breaks the call graph for static analysis, debuggers, and IDEs. Go's idiomatic equivalent — wrap the function at the call site — is uglier per-line but radically clearer in aggregate.

3. **Spring confuses two layers** — "the framework" (DI container, lifecycle) and "the libraries" (JDBC abstraction, REST client, Cache, Validator, Retry). Almost every Spring concept has a usable form that doesn't require the framework. Yarumo's bet is that the libraries are the value; the framework can be deleted. So far, every Phase 1/2 ticket has confirmed this.

4. **Spring's documentation is encyclopedic in the best sense** — every concern (CORS, content negotiation, savepoints, propagation, exception translation, API versioning, retry jitter) has a section. Yarumo's docs can borrow this *topic taxonomy* even when the implementation differs. The promotion of "Testing" to a top-level v7 section is itself a strong signal: **yarumo should treat `modules/testing/` as a first-class module, not a sub-package**.

5. **What's missing from Spring (interesting gaps)** — no first-class **multi-tenancy** (must use third-party libraries), no **outbox pattern** (Spring Modulith finally added it), no **idempotency-key handling** (DIY), no native **LLM client abstraction** (Spring AI is separate). The deletion of yarumo's § 3 Brainstorm means these have **no current home in the lean roadmap** — they survive as "file new module when a consumer materializes." That's the right call: the lean roadmap deliberately scopes only what's already designed.

6. **The Spring 7 native resilience addition is a tell** — historically Spring punted CB/RL/Retry to Spring Cloud. v7 pulling retry+concurrency-limit into core suggests the Spring team finally accepted that "resilience" is too universal to gatekeep behind a separate module. Yarumo arrived at the same conclusion earlier with `modules/common/resilience/` housing CB+RL — and should now extend it with `retry/` to close the gap.

---

## 4. Overall recommendation

**PARTIAL**, with a precise breakdown:

- **ADOPT** (specific patterns, file as enhancements to existing or already-planned modules, or as new sub-packages):
  - `WithTransaction(ctx, db, fn)` callback → `modules/datasource/` (§ 1.1) — *already planned*
  - Typed DB error translation → `modules/datasource/<driver>/` per driver
  - `ProblemDetail` (RFC 9457) renderer → **new `modules/common/http/problem/`**
  - Canonical middleware order + `DefaultStack()` builder → `managed/server_http/`
  - Global exception handler middleware (panic + typed-error → ProblemDetail) → `managed/server_http/middleware/`
  - `ResponseEntity`-style fluent response builder → **new `modules/common/http/response/`** (optional)
  - REST client wrapper (retries, CB, RL, logging) → **new `modules/common/http/client/`**
  - `Retry(ctx, policy, fn)` callback → `modules/common/resilience/retry/` (follow-up to YA-0076)
  - MockMvc-style HTTP fake → **new `modules/testing/fakes/http/`**
  - Test-rollback DB helper → **new `modules/testing/database/`**
  - Testcontainers fixtures → **new `modules/testing/containers/`** (per backend)
  - `CronWorker` dual-mode → `modules/managed/CronWorker` enhancement (confirm/file in Phase 3)

- **VALIDATE** (Spring confirms yarumo's existing design):
  - `Cache`/`CacheManager` split — matches `modules/cache/`
  - Lifecycle phases — match `modules/managed/`
  - Profiles + Environment — match `modules/config/`
  - Validator + BindingResult — match `modules/validation/`
  - ApplicationEventPublisher — matches `modules/messaging/events/`
  - Observability — matches `modules/telemetry/otel/`
  - CB + RL primitives in `common/resilience/` — pre-empted Spring 7's own move
  - Health endpoints split (leaves in `common/health/`, runtime in `modules/health/`) — matches the Boot Actuator split

- **REJECT** (lock in as cautionary anti-patterns):
  - IoC container with auto-wiring, classpath scanning, annotation-driven config
  - AOP / AspectJ / proxy-based interception
  - SpEL (yarumo has `common/expressions/`)
  - Declarative `@Transactional` / `@Cacheable` / `@Async` / `@Scheduled` / `@Retryable` / `@ConcurrencyLimit` annotations (the *mechanism*, not the *pattern*)
  - WebFlux / Reactor / `Mono`/`Flux` (Go-native concurrency subsumes)
  - Language support (Kotlin, Groovy)
  - Aspects, Instrumentation, JCL, OXM sub-modules

- **DEFER** (no current consumer demand):
  - Content negotiation (multi-format response selection) — JSON-only suffices
  - WebSocket / STOMP / SSE — no Aluna/DaaS use case yet
  - Email (`JavaMailSender`) — no home in lean roadmap (§ 3 Brainstorm deleted); revisit when consumer surfaces
  - `Converter[S, T]` interface — speculative

**Net effect on the roadmap**: **one new top-level module proposed** (`modules/testing/`, see § 6.1) plus three new sub-package proposals under `modules/common/http/` (`problem/`, `response/`, `client/`) and one inside `modules/common/resilience/` (`retry/`). The Spring Framework analysis reinforces the existing module taxonomy and contributes ~10 sub-features distributed across already-planned modules (`datasource/`, `managed/server_http/`, `common/resilience/`) plus the new `testing/` + `common/http/` homes.

---

## 5. Open questions

1. **ProblemDetail extension fields** — RFC 9457 allows arbitrary additional fields (`fieldName`, `traceId`, `correlationId`). Should `common/http/problem/` ship a `WithField(key, value)` builder, or expose the underlying `map[string]any` directly? Probably the former — typed setters for `TraceID`, `RequestID`, and `WithExtension(key, val)` for anything else.

2. **`WithTransaction` propagation** — start with REQUIRED-only, or ship REQUIRED + REQUIRES_NEW + NESTED from v1? Real DaaS use cases will inform. Given § 3 Brainstorm is gone, the audit-on-rollback use case has no concrete consumer yet — **decide when a real driver lands**.

3. **Canonical middleware order** — is there a single canonical order, or does it vary by service type (public API vs. internal vs. webhook receiver)? Likely two presets: `DefaultStack()` and `WebhookStack()` (which omits auth, adds HMAC verification + idempotency).

4. **Typed DB error translation** — should the translation map live in the driver subpackage (`gorm/errors.go`) or in a shared `datasource/translate/` with per-driver registrations? Inverse-of-control concern: a shared registry is more Springy. **Per-driver is simpler and explicit — recommend that.**

5. **MockMvc fluent builder** — wrap Gin's router directly, or wrap any `http.Handler`? The latter is more general; the former is more ergonomic for the 95% case (Gin-based services). Likely ship both layers: `testing.HTTP(t, http.Handler).GET(...)` as the base and `testing.Gin(t, *gin.Engine)` as a thin convenience.

6. **`ResponseEntity` builder vs. direct Gin calls** — does the builder pay for itself? Probably only if it composes well with the `ProblemDetail` renderer (e.g. `http.Response().Status(400).Problem(typedErr).Send(c)`). If yes, ship; if it's just a wrapper around `c.JSON(...)`, skip.

7. **`@Scheduled` `fixedDelay`/`fixedRate` mode in `CronWorker`** — does the current `managed/CronWorker` API expose fixed-delay (delay between completions) and fixed-rate (between starts) in addition to cron expressions? If only cron, file a small enhancement ticket. **Action: check `modules/managed/` source before Phase 3 absorbs Managed tickets.**

8. **TestContext caching equivalent** — Go test runs share state within a package via `TestMain`. Worth a `modules/testing/context/` for shared fixtures across tests in a package (e.g. one Postgres container per test package, not per test)? Defer until repeated boilerplate across consumer projects materializes.

9. **Where does `common/http/problem` live alongside `common/errs/`?** — does the renderer reach into `errs.TypedError` directly (tight coupling), or accept an `interface { HTTPStatus() int; ProblemType() string; ProblemTitle() string }` adapter (loose coupling)? **Recommend the adapter** so non-typed errors can still render, and `errs.TypedError` types implement the interface in `common/errs/http.go` (single file, no cross-module change).

10. **Should `boot/` ship a `Cautions` doc?** — every Spring concept rejected in § 3.3 is something a Java-fluent engineer will reach for instinctively. A 1-page `modules/boot/CAUTIONS.md` enumerating "if you were going to write a `BeanPostProcessor`, do X instead" pre-empts a lot of code review.

11. **`modules/testing/` placement principle** — the lean roadmap's placement table says "Pure library, no lifecycle, no external SDK deps → `modules/common/`." Most of `testing/` is pure library, but `testing/containers/` depends on testcontainers-go (external SDK) and orchestrates Docker containers (lifecycle). The mixed nature justifies a top-level `modules/testing/` rather than a `common/testing/` sub-package. **See § 6.1.**

12. **Retry callback ergonomics** — Spring 7's `RetryTemplate.invoke(() -> ...)` is a `Supplier<T>`. The Go shape is `Retry[T any](ctx, policy, fn func(ctx) (T, error)) (T, error)`. Pass the context into `fn` so it can early-exit; alternatively, have `Retry` enforce timeout/deadline itself. **Recommend passing ctx in** for symmetry with `WithTransaction`.

---

## 6. ROADMAP delta proposed (NOT applied)

This section enumerates concrete additions to `docs/ROADMAP_NEW_MODULES.md` that the re-analysis surfaces. They are NOT applied here — they're proposals for a future roadmap revision.

### 6.1. Propose NEW `modules/testing/` as top-level

**Rationale**: The deletion of § 3 (Brainstorm) removed the `modules/testing/` placeholder. Spring 7's promotion of "Testing" to a top-level reference section validates treating it as a first-class module rather than a sub-package of `common/`. Mixed nature (some pure-library sub-packages, some lifecycle-bearing for containers) further justifies a dedicated top-level home.

**Proposed scope**:

```
modules/testing/
  fakes/
    http/             MockMvc-style fluent builder + integration server
    grpc/             in-memory grpc.Server + bufconn client wrapper (future)
  fixtures/
    factories.go      generic factory pattern (NewBuilder[T]() ... .Build())
    seeds/            JSON/SQL fixture loaders
  containers/
    postgres/         testcontainers-go wrapper, fresh DB per test class
    mongo/
    redis/
    rabbitmq/
    kafka/
  database/
    rollback.go       WithRollback(t, db, fn)
    load_sql.go       LoadSQL(t, db, files...)
  contracts/          consumer-driven contract testing (Pact-style, when consumer materializes)
  property/           property-based testing helpers (testing/quick wrapper, when needed)
  llm/                deterministic evaluator + golden trace harness (for Aluna/DaaS, when needed)
```

**Why a new module (not `common/testing/`)**:
- `containers/` depends on external SDKs (testcontainers-go, docker client) — fails `common/` placement criterion.
- `containers/` orchestrates Docker containers with lifecycle — fails "no lifecycle" criterion.
- `fakes/http/` and `database/` are pure library and could in principle live under `common/`, but co-location with the lifecycle-bearing sub-packages keeps the testing surface together.

**Priority**: file when DaaS or Aluna's first integration test materializes. Until then, `fakes/http/` is the most-requested sub-package and can ship standalone.

### 6.2. Propose NEW `modules/common/http/` sub-tree

Three small sibling sub-packages under a new `modules/common/http/`:

| Sub-package | Purpose | Spring counterpart |
|---|---|---|
| `common/http/problem/` | RFC 9457 ProblemDetail renderer + `HTTPProblem` interface | `ProblemDetail` |
| `common/http/response/` | Fluent response builder (`Status().Header().JSON()`) | `ResponseEntity` |
| `common/http/client/` | `net/http.Client` wrapper with retries, CB, RL, structured logging, timeouts | `RestClient` |

**Why `common/`**: all three are pure libraries, no external SDK deps (Gin imports happen at the consumer side; `client/` uses `net/http` from stdlib + `modules/common/resilience/`). All satisfy `common/` placement criteria.

**Dependencies**:
- `common/http/problem/` → `common/errs/` (for `HTTPProblem` adapter).
- `common/http/response/` → no internal deps.
- `common/http/client/` → `common/resilience/` (CB+RL+retry), `common/log/`, `common/telemetry/` (optional).

**Priority**: ProblemDetail renderer first (cleanest win, recommended in every § 4 list). Response builder + client wrapper file when `managed/server_http/` work resumes.

### 6.3. Propose `modules/common/resilience/retry/` sub-package

**Rationale**: Spring 7 added `@Retryable` + `RetryTemplate` to core resilience. Yarumo's `common/resilience/` already houses CB + RL (YA-0076 closed 2026-05-13). Adding retry closes the gap and presents a unified resilience surface.

**Proposed API**:

```go
package retry

type Policy struct {
    MaxAttempts int           // total attempts including first; default 3
    Delay       time.Duration // initial delay; default 100ms
    Multiplier  float64       // backoff multiplier; default 2.0
    MaxDelay    time.Duration // cap; default 10s
    Jitter      time.Duration // random ± jitter; default 0
    Retryable   func(error) bool // predicate; default = err != nil
}

func Do(ctx context.Context, policy Policy, fn func(ctx context.Context) error) error
func DoT[T any](ctx context.Context, policy Policy, fn func(ctx context.Context) (T, error)) (T, error)
```

Mirrors Spring 7's `RetryTemplate` shape (callback + structured retry-history error) without the annotation surface.

**Priority**: medium — file as a follow-up issue when next Phase 2 retrospective surfaces "we keep hand-rolling retry loops."

### 6.4. § 1.1 `datasource/` API hardening

Concrete additions to the existing § 1.1 design (no new sub-package, just spec refinement):

- **Explicit `WithTransaction(ctx, db, fn)` signature** with `TxOptions` (Propagation enum + Isolation + ReadOnly).
- **Propagation enum scoped to three modes**: `Required` (default), `RequiresNew`, `Nested`.
- **`TranslateError(err) error` helper required for every driver**, returning sentinel errors (`ErrUniqueViolation`, `ErrForeignKeyViolation`, `ErrLockNotAvailable`, `ErrSerializationFailure`, `ErrEmptyResult`, `ErrQueryCanceled`).

These are pattern requirements, filed as design notes when first driver ticket lands.

### 6.5. `managed/server_http/middleware/` order convention

Document a canonical middleware order as either prose in `managed/server_http/CODING_STANDARDS.md` or as a `DefaultStack()` constructor. Recommended order:

```
1. RecoverAndMap      (panic → 500 + ProblemDetail; typed-error → ProblemDetail)
2. RequestID          (inject request_id into ctx + response header)
3. Trace              (start OTel span; inject trace IDs)
4. Logging            (request log line with request_id + trace_id)
5. Metrics            (http_server_duration_seconds histogram)
6. CORS               (preflight handling; before auth so OPTIONS doesn't 401)
7. Auth               (extract Principal; inject into ctx)
8. RateLimit          (per-Principal; after auth so anonymous attackers can't burn the limiter)
9. Tenancy            (extract tenant from Principal; inject into ctx)
10. (Handler)
```

Plus a `WebhookStack()` variant: replaces Auth + Tenancy with HMAC verification + idempotency-key handling.

### 6.6. § 1.3 `messaging/` cross-references

The deletion of Annex A means `spring-integration.md` (companion file in this directory) now owns the full Spring Messaging / Spring Integration feature reference. The `messaging/` module design enumerated in § 1.3 of the lean roadmap (channels, endpoints, schema/, events/, cdc/, rabbitmq/, kafka/) **stands as-is**; no changes from this Framework re-analysis. JMS is one of the brokers `messaging/` can plug into.

### 6.7. Summary table — proposed roadmap revisions

| Change | Type | Location | Priority |
|---|---|---|---|
| New top-level `modules/testing/` | New module | placement table + new § 1.6 in roadmap | MEDIUM |
| New `modules/common/http/problem/` | New sub-package | follow-up to Phase 2 closure | HIGH |
| New `modules/common/http/response/` | New sub-package | follow-up to Phase 2 closure | MEDIUM |
| New `modules/common/http/client/` | New sub-package | follow-up to Phase 2 closure; depends on resilience | MEDIUM |
| New `modules/common/resilience/retry/` | New sub-package | follow-up to YA-0076 | MEDIUM |
| § 1.1 `WithTransaction` spec refinement | Design note | inside § 1.1 of roadmap | HIGH (when first driver lands) |
| `managed/server_http/middleware/` order convention | Doc + constructor | inside Phase 3 milestone scope | HIGH |
| `CronWorker` dual-mode (`fixedDelay`/`fixedRate`/`cron`) | Enhancement | Phase 3 milestone scope | LOW (verify current API first) |

---

## 7. Appendix A — concrete code sketches for the ADOPT items

These are illustrative shapes, not final APIs. They're included so reviewers can sanity-check the analysis against real Go syntax before tickets are filed.

### 7.1. `WithTransaction` (`modules/datasource/`)

```go
package datasource

import (
    "context"
    "database/sql"
    "errors"
)

// Propagation controls how an inner WithTransaction relates to an outer one.
type Propagation int

const (
    // Required joins an existing transaction or starts a new one (default).
    Required Propagation = iota
    // RequiresNew always starts a new transaction; suspends any outer one.
    RequiresNew
    // Nested uses a savepoint within an existing transaction; standalone otherwise.
    Nested
)

// TxOptions configures a transaction's behavior.
type TxOptions struct {
    Propagation Propagation
    Isolation   sql.IsolationLevel
    ReadOnly    bool
}

// WithTransaction runs fn inside a transaction. On non-nil error or panic, the
// transaction is rolled back; otherwise it is committed.
//
// fn receives a *sql.Tx that must NOT be committed or rolled back manually —
// this function owns the lifecycle. To force rollback from inside fn, return
// a non-nil error.
func WithTransaction[T any](ctx context.Context, db *sql.DB, opts TxOptions, fn func(ctx context.Context, tx *sql.Tx) (T, error)) (T, error) {
    var zero T
    if db == nil {
        return zero, ErrDBIsNil
    }
    // Propagation handling elided for brevity — Required path shown.
    tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: opts.Isolation, ReadOnly: opts.ReadOnly})
    if err != nil {
        return zero, err
    }
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()
    result, err := fn(ctx, tx)
    if err != nil {
        if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
            return zero, errors.Join(err, rbErr)
        }
        return zero, err
    }
    if cmErr := tx.Commit(); cmErr != nil {
        return zero, cmErr
    }
    return result, nil
}
```

### 7.2. Typed error translation (`modules/datasource/postgres/errors.go`)

```go
package postgres

import (
    "errors"

    "github.com/jackc/pgx/v5/pgconn"
    "github.com/guidomantilla/yarumo/modules/datasource"
)

// TranslateError maps Postgres-specific errors to datasource sentinel errors.
// Unknown errors pass through unchanged.
func TranslateError(err error) error {
    if err == nil {
        return nil
    }
    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) {
        return err
    }
    switch pgErr.Code {
    case "23505": // unique_violation
        return errors.Join(datasource.ErrUniqueViolation, err)
    case "23503": // foreign_key_violation
        return errors.Join(datasource.ErrForeignKeyViolation, err)
    case "23502": // not_null_violation
        return errors.Join(datasource.ErrNotNullViolation, err)
    case "40001": // serialization_failure
        return errors.Join(datasource.ErrSerializationFailure, err)
    case "55P03": // lock_not_available
        return errors.Join(datasource.ErrLockNotAvailable, err)
    case "57014": // query_canceled
        return errors.Join(datasource.ErrQueryCanceled, err)
    }
    return err
}
```

### 7.3. ProblemDetail renderer (`modules/common/http/problem/`)

```go
package problem

import (
    "encoding/json"
    "net/http"
)

// Detail represents an RFC 9457 problem-detail body.
type Detail struct {
    Type       string         `json:"type,omitempty"`
    Title      string         `json:"title,omitempty"`
    Status     int            `json:"status,omitempty"`
    Detail     string         `json:"detail,omitempty"`
    Instance   string         `json:"instance,omitempty"`
    Extensions map[string]any `json:"-"`
}

func New(status int) *Builder { return &Builder{d: Detail{Status: status}} }

type Builder struct{ d Detail }

func (b *Builder) Type(uri string) *Builder      { b.d.Type = uri; return b }
func (b *Builder) Title(t string) *Builder        { b.d.Title = t; return b }
func (b *Builder) Detail(d string) *Builder       { b.d.Detail = d; return b }
func (b *Builder) Instance(p string) *Builder     { b.d.Instance = p; return b }
func (b *Builder) WithExtension(k string, v any) *Builder {
    if b.d.Extensions == nil {
        b.d.Extensions = map[string]any{}
    }
    b.d.Extensions[k] = v
    return b
}

// WriteJSON writes the problem detail using the RFC 9457 "application/problem+json"
// content type and merges extension fields into the top-level JSON object.
func (b *Builder) WriteJSON(w http.ResponseWriter) error {
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(b.d.Status)
    merged := map[string]any{}
    if b.d.Type != "" {
        merged["type"] = b.d.Type
    }
    if b.d.Title != "" {
        merged["title"] = b.d.Title
    }
    if b.d.Status != 0 {
        merged["status"] = b.d.Status
    }
    if b.d.Detail != "" {
        merged["detail"] = b.d.Detail
    }
    if b.d.Instance != "" {
        merged["instance"] = b.d.Instance
    }
    for k, v := range b.d.Extensions {
        merged[k] = v
    }
    return json.NewEncoder(w).Encode(merged)
}

// HTTPProblem is implemented by errors that know their HTTP shape.
// errs.TypedError types implement this in common/errs/http.go.
type HTTPProblem interface {
    HTTPStatus() int
    ProblemType() string
    ProblemTitle() string
}

func From(err HTTPProblem, instance string) *Builder {
    return New(err.HTTPStatus()).
        Type(err.ProblemType()).
        Title(err.ProblemTitle()).
        Detail(err.(error).Error()).
        Instance(instance)
}
```

### 7.4. Global exception handler middleware (`managed/server_http/middleware/`)

```go
package middleware

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/guidomantilla/yarumo/modules/common/http/problem"
)

// RecoverAndMap returns a Gin middleware that:
//   1. Recovers from panics, logs them, and returns 500 + problem detail.
//   2. After the handler runs, inspects c.Errors for any HTTPProblem-compatible
//      error and renders it as a problem detail (RFC 9457).
func RecoverAndMap(logger Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                logger.Error("panic", "value", r, "path", c.Request.URL.Path)
                _ = problem.New(http.StatusInternalServerError).
                    Title("Internal Server Error").
                    Instance(c.Request.URL.Path).
                    WriteJSON(c.Writer)
                c.Abort()
            }
        }()
        c.Next()
        if len(c.Errors) == 0 {
            return
        }
        err := c.Errors.Last().Err
        var hp problem.HTTPProblem
        if errors.As(err, &hp) {
            _ = problem.From(hp, c.Request.URL.Path).WriteJSON(c.Writer)
            return
        }
        _ = problem.New(http.StatusInternalServerError).
            Title("Internal Server Error").
            Instance(c.Request.URL.Path).
            WriteJSON(c.Writer)
    }
}
```

### 7.5. Retry callback (`modules/common/resilience/retry/`)

```go
package retry

import (
    "context"
    "errors"
    "math/rand/v2"
    "time"
)

// Policy configures retry behavior. Zero value is intentionally invalid;
// use DefaultPolicy() or build one explicitly.
type Policy struct {
    MaxAttempts int
    Delay       time.Duration
    Multiplier  float64
    MaxDelay    time.Duration
    Jitter      time.Duration
    Retryable   func(error) bool
}

func DefaultPolicy() Policy {
    return Policy{
        MaxAttempts: 3,
        Delay:       100 * time.Millisecond,
        Multiplier:  2.0,
        MaxDelay:    10 * time.Second,
        Retryable:   func(err error) bool { return err != nil },
    }
}

// Exhausted carries the per-attempt error history when retry budget is exceeded.
type Exhausted struct {
    Attempts []error
}

func (e *Exhausted) Error() string  { return "retry: budget exhausted" }
func (e *Exhausted) Last() error    { return e.Attempts[len(e.Attempts)-1] }
func (e *Exhausted) Unwrap() []error { return e.Attempts }

// DoT runs fn up to policy.MaxAttempts times, applying exponential backoff
// with jitter between attempts. ctx cancellation aborts immediately.
func DoT[T any](ctx context.Context, p Policy, fn func(ctx context.Context) (T, error)) (T, error) {
    var zero T
    var history []error
    delay := p.Delay
    for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
        result, err := fn(ctx)
        if err == nil {
            return result, nil
        }
        history = append(history, err)
        if !p.Retryable(err) || attempt == p.MaxAttempts {
            break
        }
        jitter := time.Duration(0)
        if p.Jitter > 0 {
            jitter = time.Duration(rand.Int64N(int64(p.Jitter)))
        }
        select {
        case <-ctx.Done():
            return zero, errors.Join(ctx.Err(), &Exhausted{Attempts: history})
        case <-time.After(delay + jitter):
        }
        delay = time.Duration(float64(delay) * p.Multiplier)
        if delay > p.MaxDelay {
            delay = p.MaxDelay
        }
    }
    return zero, &Exhausted{Attempts: history}
}
```

### 7.6. MockMvc-style HTTP fake (`modules/testing/fakes/http/`)

```go
package http

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
)

// Client wraps an http.Handler for in-memory request testing.
type Client struct {
    t       *testing.T
    handler http.Handler
}

func NewClient(t *testing.T, handler http.Handler) *Client {
    t.Helper()
    return &Client{t: t, handler: handler}
}

func (c *Client) GET(path string) *Request    { return c.build("GET", path) }
func (c *Client) POST(path string) *Request   { return c.build("POST", path) }
func (c *Client) PUT(path string) *Request    { return c.build("PUT", path) }
func (c *Client) DELETE(path string) *Request { return c.build("DELETE", path) }

type Request struct {
    c       *Client
    method  string
    path    string
    headers http.Header
    body    io.Reader
}

func (c *Client) build(m, p string) *Request {
    return &Request{c: c, method: m, path: p, headers: http.Header{}}
}

func (r *Request) Header(k, v string) *Request { r.headers.Set(k, v); return r }
func (r *Request) JSON(body any) *Request {
    b, err := json.Marshal(body)
    if err != nil {
        r.c.t.Fatalf("marshal request body: %v", err)
    }
    r.body = bytes.NewReader(b)
    r.headers.Set("Content-Type", "application/json")
    return r
}

func (r *Request) Expect() *Expectation {
    r.c.t.Helper()
    req := httptest.NewRequest(r.method, r.path, r.body)
    req.Header = r.headers
    rec := httptest.NewRecorder()
    r.c.handler.ServeHTTP(rec, req)
    return &Expectation{t: r.c.t, rec: rec}
}

type Expectation struct {
    t   *testing.T
    rec *httptest.ResponseRecorder
}

func (e *Expectation) Status(want int) *Expectation {
    e.t.Helper()
    if e.rec.Code != want {
        e.t.Fatalf("status: got %d want %d (body=%s)", e.rec.Code, want, e.rec.Body.String())
    }
    return e
}

func (e *Expectation) HeaderEquals(key, want string) *Expectation {
    e.t.Helper()
    if got := e.rec.Header().Get(key); got != want {
        e.t.Fatalf("header %s: got %q want %q", key, got, want)
    }
    return e
}

func (e *Expectation) JSONField(key string, want any) *Expectation {
    e.t.Helper()
    var body map[string]any
    if err := json.Unmarshal(e.rec.Body.Bytes(), &body); err != nil {
        e.t.Fatalf("unmarshal body: %v (body=%s)", err, e.rec.Body.String())
    }
    if body[key] != want {
        e.t.Fatalf("body[%q]: got %v want %v", key, body[key], want)
    }
    return e
}
```

Usage:

```go
func TestGetUser(t *testing.T) {
    t.Parallel()
    router := setupRouter(t)
    httpfake.NewClient(t, router).
        GET("/users/42").
        Header("Authorization", "Bearer test-token").
        Expect().
        Status(200).
        HeaderEquals("Content-Type", "application/json").
        JSONField("name", "John")
}
```

---

## 8. Appendix B — Spring → Yarumo directory mapping

A reference table for engineers familiar with one ecosystem reading code in the other.

| Spring package | Yarumo equivalent | Notes |
|---|---|---|
| `org.springframework.context.ApplicationContext` | `modules/boot.Container` | Explicit registration, no scanning |
| `org.springframework.beans.factory.annotation.Autowired` | Constructor parameters / `BeanFn` | Wiring is code |
| `org.springframework.context.annotation.Bean` | `BeanFn func(ctx, c *Container) error` | Plain Go function |
| `org.springframework.context.annotation.Profile` | `config.Env()` switch | Env-driven |
| `org.springframework.core.env.Environment` | `modules/config.Config` | viper-driven |
| `org.springframework.context.ApplicationEventPublisher` | `modules/messaging/events.Publisher` (§ 1.3) | Channel-backed |
| `org.springframework.context.ApplicationListener` | `modules/messaging/events.Subscribe[T]` (§ 1.3) | Typed subscriber |
| `org.springframework.scheduling.annotation.Async` | `go fn()` / `managed.BaseWorker` | Goroutine-native |
| `org.springframework.scheduling.annotation.Scheduled` | `modules/managed.CronWorker` | Cron + fixed-delay/rate |
| `org.springframework.cache.Cache` | `modules/cache.Cache[K, V]` | Generic, typed |
| `org.springframework.cache.CacheManager` | `modules/cache` registry | Typed lookup |
| `org.springframework.cache.annotation.Cacheable` | `cache.GetOrCompute(key, fn)` | Explicit call site |
| `org.springframework.transaction.PlatformTransactionManager` | `modules/datasource.TxRunner` (§ 1.1) | Per-driver impl |
| `org.springframework.transaction.support.TransactionTemplate` | `datasource.WithTransaction(ctx, db, fn)` (§ 1.1) | Callback |
| `org.springframework.transaction.annotation.Transactional` | `WithTransaction` at call site | Explicit |
| `org.springframework.jdbc.core.JdbcTemplate` | `database/sql` + helpers | Stdlib |
| `org.springframework.jdbc.core.RowMapper` | `func(*sql.Rows) (T, error)` | Higher-order fn |
| `org.springframework.dao.DataAccessException` | `modules/datasource.Err*` | Typed sentinels |
| `org.springframework.jdbc.support.SQLExceptionTranslator` | `datasource/<driver>.TranslateError` | Per-driver |
| `org.springframework.core.retry.RetryTemplate` | `common/resilience/retry.DoT(ctx, policy, fn)` (proposed) | Callback |
| `org.springframework.retry.annotation.Retryable` | `retry.DoT` at call site | Explicit |
| `org.springframework.core.concurrent.ConcurrencyLimit` | `semaphore.NewWeighted(n).Acquire(ctx, 1)` | stdlib |
| `org.springframework.web.servlet.DispatcherServlet` | Gin `*Engine` | HTTP router |
| `org.springframework.web.servlet.HandlerInterceptor` | Gin middleware | `c.Next()` + defer |
| `jakarta.servlet.Filter` | Gin middleware | Single layer in Go |
| `org.springframework.web.bind.annotation.RestController` | Plain Go struct + handler methods | Routes registered explicitly |
| `org.springframework.http.ResponseEntity` | `common/http/response.Builder` (proposed) | Optional ergonomics |
| `org.springframework.http.ProblemDetail` | `common/http/problem.Builder` (proposed) | RFC 9457 |
| `org.springframework.web.bind.annotation.ControllerAdvice` | `middleware.RecoverAndMap` | Global mw |
| `org.springframework.web.bind.annotation.ExceptionHandler` | Typed-error → status mapping in `RecoverAndMap` | Adapter pattern |
| `org.springframework.web.cors.CorsConfiguration` | `gin-contrib/cors` + standard preset | Already mature |
| `org.springframework.validation.Validator` | `modules/validation` | Multi-error |
| `org.springframework.validation.BindingResult` | `validation.Errors` collector | Accumulating |
| `org.springframework.expression.Expression` (SpEL) | `modules/common/expressions` | Go-native |
| `org.springframework.aop.framework.ProxyFactory` | Function decorator at call site | No proxy |
| `org.springframework.web.client.RestClient` | `common/http/client` (proposed) | net/http wrapper |
| `org.springframework.test.web.servlet.MockMvc` | `modules/testing/fakes/http.Client` (proposed) | `httptest`-based |
| `org.springframework.boot.test.web.client.TestRestTemplate` | `testing/fakes/http` integration variant | `httptest.NewServer` |
| `org.springframework.test.web.servlet.client.RestTestClient` | `testing/fakes/http` server-stub variant | `httptest.NewServer` |
| `org.springframework.test.context.ContextConfiguration` | `TestMain` package-level setup | Go test model |
| `org.springframework.test.context.jdbc.Sql` | `testing/database.LoadSQL(t, db, files...)` (proposed) | helper |
| `org.springframework.test.annotation.Rollback` | `testing/database.WithRollback(t, db, fn)` (proposed) | tx scope |
| `org.springframework.beans.factory.config.BeanPostProcessor` | (no equivalent — by design) | Anti-pattern |
| `org.springframework.aop.aspectj.annotation.AspectJProxyFactory` | (no equivalent — by design) | Anti-pattern |
| WebFlux `Mono<T>` / `Flux<T>` | Goroutines + channels + `context.Context` | Language-native |
| Spring JMS `JmsTemplate` / `@JmsListener` | `modules/messaging/` (§ 1.3) + broker drivers | See `spring-integration.md` |
| RSocket / STOMP | (deferred) | No use case yet |
| `JavaMailSender` | (no home in lean roadmap) | Revisit when a consumer surfaces |

---

## 9. Conclusions

The Spring Framework re-analysis (against the lean roadmap of 2026-05-15) confirms three structural decisions yarumo has already made, surfaces one new top-level module proposal, and reinforces the wisdom of the roadmap trim:

1. **Reject the framework, keep the libraries.** Every Spring concept that survives review (Cache, TransactionTemplate, Validator, problem details, scheduled tasks, retry template) is a *library pattern*; everything yarumo rejects (IoC, AOP, annotations, SpEL, reactive types, concurrency-limit annotation) is *framework machinery*. The split is clean.

2. **Templates with callbacks are the single most reusable Java→Go pattern.** `JdbcTemplate(sql, RowMapper)`, `TransactionTemplate(callback)`, `RestTemplate(uri, ResponseExtractor)`, and the new `RetryTemplate(callback)` all decompose to `func DoX(resource R, fn func(handle H) T) T`. Yarumo's `WithTransaction(ctx, db, fn)` already encodes this; the pattern should be the **default API shape** for any resource-acquiring helper across yarumo.

3. **RFC 9457 ProblemDetail is the single cleanest borrow.** Pure data format, framework-neutral, IETF standardized. The 60-line `modules/common/http/problem` package proposed in § 7.3 will benefit every yarumo-built service for ~zero ongoing maintenance.

4. **`modules/testing/` deserves first-class module status.** The deletion of § 3 Brainstorm removed the placeholder; Spring 7's promotion of Testing to a top-level reference section validates treating it as a peer of `datasource/`, `auth/`, `messaging/`, `health/`, `boot/`. The headline ROADMAP delta from this re-analysis is § 6.1: file `modules/testing/` as a new § 1.6 of the lean roadmap when the first integration-test consumer (DaaS or Aluna) materializes.

5. **The lean roadmap trim was the right call.** Re-reading Spring Framework end-to-end did not surface any feature category that the deleted § 3 Brainstorm needed to cover. Email, multi-tenancy, idempotency, outbox, LLM client — all are real concerns, but none are *Spring Framework* concerns. The lean roadmap correctly scopes to what's already designed, leaving brainstorm items to surface as new modules when consumers materialize. **No re-expansion of § 3 is justified by this analysis.**

The eight-item ADOPT list in § 4 sits mostly inside existing module boundaries — one new top-level module (`testing/`), three new sub-packages under `common/http/`, one new sub-package under `common/resilience/`. Phase 3 (Config/Managed/Telemetry) and Phase 4 (Compute) remain unaffected. The work surfaces as follow-up tickets under `modules/datasource/`, `modules/common/http/`, `modules/common/resilience/`, `managed/server_http/`, and the new `modules/testing/` once those tracks are activated.
