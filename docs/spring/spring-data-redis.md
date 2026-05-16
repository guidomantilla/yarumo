# Spring Data Redis — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-data/redis
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup; § 3 brainstorm DELETED)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Data Redis (current stable **4.0.5**, preview 4.1.0-RC1; previous line 3.5.x) is a Spring-flavoured Redis client built on top of either **Lettuce** (default, Netty-based, reactive) or **Jedis** (TCP-only, blocking). The surface area covers:

- Low-level `RedisConnection` and high-level `RedisTemplate` with typed sub-operations (`ValueOperations`, `ListOperations`, `HashOperations`, `StreamOperations`, `ZSetOperations`, `GeoOperations`, `HyperLogLogOperations`).
- Pluggable serializers (`StringRedisSerializer`, `Jackson2JsonRedisSerializer`, `GenericJackson2JsonRedisSerializer`, JDK serializer).
- Connection topologies: standalone, sentinel, cluster, master-replica.
- Pub/Sub via `RedisMessageListenerContainer` (lifecycle-managed subscriber pool, pattern subscriptions, lazy connect).
- Streams via `StreamMessageListenerContainer` (consumer groups, `XREADGROUP`, manual/auto-ack, polling).
- Lua scripting (`RedisScript`, automatic EVALSHA caching with EVAL fallback, typed result binding).
- Pipelining and MULTI/EXEC + WATCH transactions.
- Repository pattern over hashes with object-to-hash mapping, secondary indexes, and TTL.
- `RedisCacheManager` as the Spring Cache abstraction backend.
- Observability hooks (Micrometer).

It is **deeply coupled to the JVM** (Netty, Reactor, Spring Context, AOP `@Transactional`, `@Cacheable`) — only the *patterns* port to Go, not the API.

## 2. Pareto features (top-20%)

| #  | Feature | Description | Why it matters for Go microservices |
|----|---|---|---|
| 1  | `RedisTemplate` typed operation views | One client surface, sub-interfaces per data structure (string, list, hash, set, zset, stream, geo, HLL) | `go-redis/v9` already exposes typed methods (`Set`, `LPush`, `HSet`, `XAdd`, `ZAdd`); yarumo only needs a thin Connection wrapper that exposes the `*redis.Client` and lifecycle. **No bespoke template needed.** |
| 2  | Serializer pluggability | Decouples wire format from app objects | In Go this is just `encoding/json`/`vmihailenco/msgpack` around `[]byte`. The driver stays byte-oriented; consumers own their `Codec[T]`. |
| 3  | Topology abstractions: standalone / sentinel / cluster / master-replica | Single config switches deployment topology | Critical for prod. `go-redis/v9` supports all four (`redis.NewClient`, `redis.NewFailoverClient`, `redis.NewClusterClient`, `NewUniversalClient`). `goredis.Context` must model the topology in its options (URL list, sentinel master name, route-by-latency, etc.). |
| 4  | Connection pooling + SSL/TLS + AUTH/ACL | Production hygiene | `go-redis` has pool + TLS + `Username/Password` baked in. `goredis.Context` exposes these. |
| 5  | Lettuce reactive + thread-safe shared connection | Non-blocking, single connection per client | Irrelevant in Go — `go-redis` is already concurrent-safe and pooled. **Skip reactive analogue.** |
| 6  | `RedisMessageListenerContainer` (pub/sub) | Managed subscription lifecycle, listener pool, pattern subs, lazy connect | Maps to a `Subscriber` lifecycle component that owns `PubSub.Channel()` consumer goroutine, dispatches to `Handler`, integrates with `managed.Lifecycle` (Start/Stop). Belongs in `goredis/pubsub/`. |
| 7  | `StreamMessageListenerContainer` (streams) | Polling loop + consumer groups + ack + retry | Same pattern as pub/sub but with `XREADGROUP` + manual/auto-ack + claim-on-stale via `XPENDING`/`XCLAIM`/`XAUTOCLAIM`. Goes in `goredis/streams/`. |
| 8  | Lua scripting (`RedisScript`, EVALSHA caching) | Compute on the server side atomically | `go-redis` has `redis.NewScript(...).Run(...)` with EVALSHA fallback built in. Provide a thin `goredis/script/` helper that pins script bytes + return type. |
| 9  | Pipelining (`executePipelined`) | Batch dispatch to reduce RTT | `go-redis` exposes `Pipeline()`/`Pipelined(ctx, fn)`. Wrap in `WithPipeline(ctx, conn, fn)`. |
| 10 | MULTI/EXEC + WATCH (optimistic locking) | Atomic multi-command, optimistic CAS | `go-redis` has `TxPipelined` and `Watch`. Wrap in `WithTransaction(ctx, conn, keys, fn)` following the `datasource/` cross-driver convention. |
| 11 | `RedisCacheManager` (TTL, key prefix, null caching, SCAN batch, statistics, TTI) | Spring Cache abstraction over Redis | **Do not** ship a `cache/redis` backend. `modules/cache/` is in-process by design (YA-0079); a distributed cache is built by consumers calling `goredis` directly. **Adopt the patterns** (TTL, key prefix, lock-on-write opt-in, statistics counters, SCAN-based eviction) as **knobs in consumers**, not as a generic cache. |
| 12 | Object-to-hash repositories + secondary indexes + TTL | `CrudRepository<T, ID>` over hashes with derived queries | **Reject as a module-level feature.** Redis is not a primary data store in yarumo's stack (`datasource/gorm` and `mongo` are). Object-to-hash repository is an anti-pattern when the same domain is already persisted in SQL/Mongo. |
| 13 | Distributed locks (Spring Integration `RedisLockRegistry`) | `Lock` interface backed by SET NX EX + Lua release | **Not in Spring Data Redis itself** — lives in Spring Integration / Redisson. Worth shipping in yarumo as `goredis/lock/` (SET NX PX + Lua-released, fencing token optional). General-purpose primitive — any consumer that needs mutual exclusion across replicas can compose it. |
| 14 | Observability (Micrometer integration) | Per-command metrics + tracing | `go-redis` ships an OTel hook (`redisotel.InstrumentTracing`, `redisotel.InstrumentMetrics`). `goredis.Connection` opts in via `WithOTel()`. Aligns with `modules/telemetry/otel/`. |
| 15 | Cluster pipelining via Lettuce | Slot-aware multi-command batching | `go-redis` cluster client routes per-slot. Document the cross-slot caveats (rename, sort, multi-key set ops) but no wrapper needed. |

## 3. Long-tail features (skip)

- **Reactive `ReactiveRedisTemplate` / `StreamReceiver` (Reactor Flux)** — Go is already concurrent; no Reactor port.
- **`@Transactional` AOP for Redis** — Go has no AOP; the callback-shaped `WithTransaction` covers it.
- **`MessageListenerAdapter` POJO bridge** — annotation-driven reflection. In Go use plain `func(Message) error`.
- **Repository pattern + derived queries** (`findByLastnameAndAge`) — see Pareto #12.
- **Object-to-hash mapping with `@RedisHash`, `@Indexed`, `@Reference`** — same rationale.
- **`QueryByExampleExecutor`** — Spring-specific reflection magic.
- **OXM (XML mapping) serializer** — irrelevant.
- **JDK serialization** — actively dangerous (RCE on untrusted input) and irrelevant outside JVM.
- **`SessionCallback` vs `RedisCallback`** — both collapse into a single `func(ctx context.Context, c *redis.Client) error` callback shape.
- **`HttpSessionEventPublisher` / spring-session integration** — covered by the parallel re-analysis (see § 4 cross-references), not by this driver.
- **Bound operation interfaces (`BoundListOperations`, `BoundHashOperations`)** — sugar over key-prefix; Go closures cover it ad-hoc.
- **Embedded test Redis server** — use `testcontainers-go` redis module directly from tests; no yarumo wrapper needed at this stage.
- **`RedisCacheWriter` lock toggle, TTI via `GETEX`** — useful patterns but expose as knobs in whichever consumer builds a distributed cache, not in the driver.
- **XML namespace config** — N/A.

## 4. Mapping to Yarumo

**Existing § 1 modules with overlap**:

- **`modules/datasource/goredis/`** (§ 1.1, Planned, Low priority) — **the natural home** for the Redis driver. Follows the per-driver convention: `Context` (URL/servers/credentials/topology), `Connection` (open/close, lifecycle-aware), `TransactionHandler` (callback-based MULTI/EXEC + WATCH; also pipeline). This re-analysis refines that placement.
- **`modules/cache/`** (in-process only, YA-0079 closed) — **explicitly does not absorb Redis**. Distributed caching is built by consumers calling `goredis` directly. This re-analysis reaffirms that decision.
- **`modules/messaging/`** (§ 1.3, Planned) — **not a Redis broker home**; messaging's brokers are RabbitMQ/Kafka/NATS. Redis pub/sub and streams stay in `goredis/` as primitives. A `messaging/` Redis adapter is conceivable later if demand appears but is not on the roadmap.
- **`modules/health/`** (§ 1.4, Planned) — could ship a `RedisChecker` once `goredis` exists; trivial `PING` + RTT probe. Not part of the driver itself.

**Cross-references to consumer modules** (proposed in parallel re-analyses; each is a NEW top-level module proposal, not a § 3 brainstorm):

- `spring-session.md` → **NEW `modules/sessions/`** with a Redis-backed `Store` impl (HSET/HGETALL/EXPIRE on session blobs). Consumes `goredis` for transport.
- `spring-ai.md` → **NEW `modules/llm/memory/`** and **NEW `modules/llm/cache/`** with Redis backends (chat history, semantic cache). Consume `goredis` directly.
- `spring-retry.md` / `spring-cloud-circuit-breaker.md` → resilience patterns; a *distributed* variant is **not in scope** — `modules/common/resilience/` stays in-process per its current charter.
- `spring-integration.md` / `spring-cloud-stream.md` → Redis Streams as a broker substrate for **NEW `modules/outbox/`** or **NEW `modules/jobs/`** if those modules are proposed in those re-analyses. They would consume `goredis/streams/` + `goredis/lock/`.
- `spring-batch.md` → if a job-state store is proposed there, Redis is one backend option via `goredis`.
- `spring-cloud-gateway.md` / `bucket4j.md` (rate limiting) → if a **NEW `modules/ratelimit/`** is proposed there, the Redis-backed token-bucket / sliding-window impl uses `goredis/script/`.
- `spring-modulith.md` (event publication, outbox patterns) → cross-references **NEW `modules/outbox/`** and **NEW `modules/idempotency/`** if those are proposed; both can be Redis-backed (streams + `SET NX EX`).

**Gaps this fills inside `modules/datasource/goredis/`**:

1. **Connection lifecycle** — a `Connection` bean that wires `*redis.Client`/`*redis.ClusterClient`/`*redis.SentinelClient` into `managed.Lifecycle` Start/Stop. Today every consumer would reinvent this.
2. **Topology-agnostic config** — `Context` options accept `URL`, `Addrs []string`, `MasterName`, `RouteByLatency`, `ReadOnly` and pick the right `go-redis` constructor (`NewUniversalClient` covers most cases).
3. **TransactionHandler** — `WithTransaction(ctx, fn, watch ...string)` over `client.Watch` + `TxPipelined`. Matches the convention in `gorm/` and `mongo/`.
4. **Pipeline helper** — `WithPipeline(ctx, fn)` over `client.Pipelined`.
5. **Script registry** — `goredis/script/` for `*redis.Script` instances pinned once at construction (EVALSHA caching already built in).
6. **Distributed lock primitive** — `goredis/lock/` — `SET NX PX` + fencing-token + Lua-released. Spring Data Redis lacks this; Spring Integration / Redisson supply it. Yarumo provides it once.
7. **PubSub subscriber** — lifecycle-aware goroutine that owns `PubSub.Channel()`, dispatches to `Handler`, supports patterns. Maps to `RedisMessageListenerContainer`.
8. **Streams consumer** — lifecycle-aware `XREADGROUP` consumer with manual/auto-ack, retry on `XPENDING`, claim-on-stale via `XCLAIM`/`XAUTOCLAIM`. Maps to `StreamMessageListenerContainer`.
9. **OTel instrumentation hook** — `WithOTel()` wires `redisotel.InstrumentTracing` + `redisotel.InstrumentMetrics`. One line for consumers.

**Anti-patterns to avoid**:

1. **Do not invent a "RedisTemplate" wrapper**. `go-redis` is already idiomatic; wrapping it adds tokens and friction.
2. **Do not invent a "serializer abstraction"** inside `goredis/`. Stay byte-oriented; let consumers own their `Codec[T]`.
3. **Do not ship object-to-hash repositories**. Anti-pattern in a polyglot stack; Redis is a cache/queue/lock substrate, not a primary store.
4. **Do not bundle a Spring Cache equivalent inside `goredis/`**. `cache/` stays in-process per YA-0079; consumers compose `goredis` directly when they need a distributed cache.
5. **Do not couple to a specific topology**. Use `NewUniversalClient` for the common case and expose explicit constructors for cluster-only / sentinel-only when needed.
6. **Do not implement reactive variants**. Pure noise in Go.
7. **Do not wrap pub/sub with annotation-driven listeners**. Plain `func(Message) error` + lifecycle goroutine.
8. **Do not build a distributed cache atop `modules/cache/`**. YA-0079 already decided.
9. **Do not ship JDK-style binary serialization**. Default to JSON in consumer codecs; document msgpack as an option.
10. **Do not put pub/sub or streams in `modules/messaging/`**. Those are Redis primitives — they belong in `goredis/`. `messaging/` is for AMQP/Kafka/NATS-class brokers.

## 5. Recommendation

**PARTIAL** — adopt the Spring Data Redis **patterns** (topology-aware connection, transaction/pipeline callbacks, pub/sub listener container, stream listener container with consumer groups, Lua script registry, observability hook, distributed lock) inside `modules/datasource/goredis/`. **Reject** the Java-coupled surface (RedisTemplate, repository/object-hash mapping, AOP `@Transactional`, reactive Flux, serializer registry, Spring Cache abstraction integration as a separate cache backend).

This stays consistent with prior decisions: YA-0079 (in-process cache), datasource per-driver convention, `modules/cache/` is not pluggable to remote backends.

**Concrete refinements for the `datasource/goredis/` design** (§ 1.1):

- **Priority gate**: keep at Low until the first cross-referenced consumer module (`sessions/`, `outbox/`, `idempotency/`, `llm/memory`, `llm/cache`, `ratelimit/`) gets ticketed as Planned in its respective re-analysis. At that point `goredis` becomes a blocker and should be promoted to Medium.
- Lock in the `Context` / `Connection` / `TransactionHandler` triplet as per the cross-driver convention.
- Add `WithPipeline` next to `WithTransaction` — pipelining is a first-class Redis idiom, unlike SQL.
- Carve out four sub-packages from day one: `goredis/pubsub`, `goredis/streams`, `goredis/script`, `goredis/lock`. None is optional once consumers materialize.
- Bake OTel instrumentation into `Connection` via a `WithOTel()` option (consumes `telemetry/otel/`).
- Document explicitly that distributed cache is built by consumers, **not** as a `cache/` backend.

## 6. Proposed yarumo placement

**Module**: `modules/datasource/goredis/` (§ 1.1 — existing planned slot, refined by this analysis)

**Subpackages**:

- `goredis/` (root) — `Context`, `Connection`, `TransactionHandler`, `WithTransaction`, `WithPipeline`, options (URL/Addrs/MasterName/TLS/ACL/PoolSize/ReadTimeout), `WithOTel()`. Topology auto-detected via `UniversalOptions` or explicit `NewStandalone`/`NewSentinel`/`NewCluster`.
- `goredis/pubsub/` — `Subscriber` lifecycle component, `Handler` callback, channel + pattern subscription, dispatch goroutine, integrates with `managed.Lifecycle`.
- `goredis/streams/` — `Consumer` with consumer groups, `XREADGROUP` polling, manual/auto-ack, claim-on-stale via `XPENDING` + `XCLAIM`/`XAUTOCLAIM`. `Producer` is just `XAdd` on the connection — no wrapper.
- `goredis/script/` — `Script` registry (`redis.NewScript` wrappers pinned at construction), typed return helpers.
- `goredis/lock/` — `Locker` interface, `Lock(ctx, key, ttl) (Lease, error)`, fencing-token optional, Lua-released. Backed by `SET NX PX`.

**Internal deps**:

- `modules/common/errs` — typed errors (`ErrConnectionClosed`, `ErrLockTaken`, `ErrScriptFailed`).
- `modules/common/log/slog` — structured logging.
- `modules/managed` — lifecycle for `Connection`, `Subscriber`, `Consumer`.
- `modules/telemetry/otel` — opt-in instrumentation via `WithOTel()`.
- `modules/common/resilience` — optional circuit-breaker / rate-limiter around `Connection` calls (consumer-side composition; in-process scope only).

**Go libraries to wrap**:

- `redis/go-redis/v9` — https://github.com/redis/go-redis — primary client; standalone, sentinel, cluster, pub/sub, streams, scripts, pipeline, TLS, ACL, OTel hooks.
- `redis/go-redis/extra/redisotel/v9` — https://github.com/redis/go-redis/tree/master/extra/redisotel — official OTel instrumentation for tracing + metrics.
- `bsm/redislock` — https://github.com/bsm/redislock — battle-tested distributed-lock implementation (SET NX PX + Lua release). Either wrap or inline the same ~30-line algorithm inside `goredis/lock/` to avoid the extra dep.

**Out of scope for v1**:

- Reactive / streaming API surface.
- Object-to-hash repository / secondary indexes.
- Generic codec / serializer registry inside the driver.
- Distributed cache backend (`cache/` stays in-process — YA-0079).
- RediSearch, RedisTimeSeries, RedisJSON modules — file as separate roadmap items if real demand emerges.
- Redisson port — Redisson is JVM-only and far broader than what yarumo needs.
- Redis Cluster cross-slot multi-key pipeline workarounds — document caveats, no wrappers.
- Spring Integration `RedisLockRegistry`-style hierarchical lock namespaces.

## 7. Open questions

1. **Should `goredis/lock/` wrap `bsm/redislock` or inline the algorithm?** Inlining keeps the dep graph minimal; wrapping inherits maintenance. Decision when ticket is filed.
2. **Streams consumer back-pressure**: bounded inbox channel size or caller-driven goroutine pool? Affects any future consumer module that uses streams.
3. **PubSub message envelope**: keep raw `[]byte` payload + headers, or expose `Message[T]` like `messaging/` does? Recommend raw — pub/sub is unreliable and short-form; consumers add codec as needed.
4. **Fencing tokens in `Lock`**: monotonic counter requires a Redis-side `INCR`-on-acquire. Worth the extra RTT, or leave to consumers who need it?
5. **`Connection` vs `Connections`**: support multi-tenant pools (one client per tenant) at driver level, or push into a consumer module? Lean toward the latter — keep `goredis` single-cluster.
6. **Streams ack semantics**: does the consumer auto-`XACK` on handler success, or always require explicit ack? Recommendation: explicit-ack default + `WithAutoAck()` for fire-and-forget.
7. **When to promote priority Low → Medium?** When the first cross-referenced consumer module (`sessions/`, `outbox/`, `idempotency/`, `llm/memory`, `llm/cache`, `ratelimit/`) gets ticketed for Planned in its parallel re-analysis.
8. **Cluster pipelining**: `go-redis` v9 supports cluster pipelining per-slot. Worth exposing as a separate `WithClusterPipeline` helper, or fold into `WithPipeline` with auto-routing? Probably auto-route — single API for consumers.
9. **`messaging/` Redis adapter — defer or reject?** Redis pub/sub and streams already live in `goredis/`. A `messaging/redis/` driver would only make sense if a consumer needs the EIP `Channel` abstraction over Redis specifically. Recommend defer until demand appears.

## 8. ROADMAP delta proposed (NOT applied)

Suggested edits to `/Users/raven/Workspace/guidomau/yarumo/docs/ROADMAP_NEW_MODULES.md` § 1.1 once the first consumer module gets ticketed:

1. **Promote `modules/datasource/goredis/` priority Low → Medium** at that point — it becomes a blocker for the consumer.
2. **Document the four sub-packages explicitly** in the § 1.1 table: `pubsub/`, `streams/`, `script/`, `lock/`. Currently only the driver root is listed.
3. **Add a cross-driver note**: pipelining helper (`WithPipeline`) is Redis-specific and lives alongside `WithTransaction` in `goredis/`, not in the cross-driver `datasource/` core.
4. **Reaffirm in § 4.2 (Discarded)**: cite this re-analysis as the rationale for keeping `cache/` in-process and not introducing a Redis cache backend.
5. **Add a Spring-Data-Redis reference link** to § 1.1's `goredis/` row, alongside this analysis file, so future readers see the design provenance.
6. **No new module needed**: `goredis/` already exists in the roadmap; this analysis only refines its surface. Consumer-side modules (`sessions/`, `outbox/`, `idempotency/`, `ratelimit/`, `llm/memory`, `llm/cache`) are proposed in parallel re-analyses (`spring-session.md`, `spring-modulith.md`, `spring-ai.md`, `spring-cloud-gateway.md`, etc.), not here.

No tickets to file from this analysis in isolation — `goredis/` is already a Planned slot and remains uncontroversial. Filing happens when a consumer triggers it.
