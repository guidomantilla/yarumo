# Spring Integration — Yarumo Analysis (DEEP)

> **Source**: https://docs.spring.io/spring-integration
> **Analyzed**: 2026-05-16 (re-analysis after `ROADMAP_NEW_MODULES.md` cleanup — Annex A deleted, this document is now the PRIMARY Spring Integration reference for yarumo)
> **Recommendation**: PARTIAL — adopt selected EIP machinery into `modules/messaging/` (§ 1.3 of the roadmap).

## 1. Project summary

Spring Integration is the canonical Java implementation of the Hohpe/Woolf **Enterprise Integration Patterns** (EIP). Latest stable: **7.0.4** (7.0.5 / 7.1.0-RC1 in flight). It builds on `spring-messaging` (the `Message<T>` / `MessageChannel` / `MessageHandler` triad) and adds three layers above it:

1. **Channel implementations** — Direct / Queue / PubSub / Priority / Rendezvous / Executor / Partitioned, plus message-store-backed variants for durability.
2. **EIP endpoints** — Transformer, Filter, Router (six flavours), Splitter, Aggregator, Service Activator, Resequencer, Scatter-Gather, Delayer, Barrier, Idempotent Receiver, Claim Check.
3. **Protocol adapters / gateways** — 40+ channel adapters (HTTP, JMS, AMQP, Kafka, MQTT, JDBC/JPA, FTP/SFTP, File, Mail, MongoDB, Redis, R2DBC, Cassandra, WebSocket/STOMP, GraphQL, Hazelcast, Debezium, Zookeeper, RSocket …).

Around that core: a Java DSL (`IntegrationFlow`), a request-handler advice chain (retry / circuit-breaker / rate-limit / cache / lock / context), pluggable `MessageStore` / `MessageGroupStore` / `MetadataStore` / `LockRegistry`, full Micrometer Observation integration with W3C trace-context propagation (v6.0+), GraalVM native-image support, and a `Control Bus` for runtime management.

This document is **self-contained** — earlier drafts framed everything as "delta against Annex A of `ROADMAP_NEW_MODULES.md`", but that annex was deleted in the 2026-05-15 roadmap trim. The lean roadmap now only sketches `modules/messaging/` at the file/subpackage level in § 1.3; this analysis is the **complete** Spring Integration reference that informs that scope, with anti-pattern guardrails, and a concrete file-by-file refinement of the messaging module layout. Two non-negotiables in yarumo's translation: **no SpEL**, **no annotation-driven discovery** — every wiring is explicit Go code.

---

## 2. Core abstractions

The base triad shared with `spring-messaging`. These are the load-bearing types — every higher concept (channels, endpoints, adapters) sits on them.

### 2.1. `Message`, `MessageChannel`, `MessageHandler`

#### `Message<T>`

A generic wrapper: payload + headers.

```
Message<T>
├── Payload : T          # any object
└── Headers : Map<String, Object>
                          # id, timestamp, correlationId, sequenceNumber,
                          # sequenceSize, replyChannel, errorChannel,
                          # contentType, …
```

- **Headers are immutable** in Spring Integration's `MessageBuilder` pattern (each mutation produces a new `Message`). Yarumo equivalent: a `Builder` that returns a new `Message[T]` per chain.
- Two **reserved headers** drive most plumbing: `CORRELATION_ID` (aggregator / scatter-gather / resequencer), `SEQUENCE_NUMBER` + `SEQUENCE_SIZE` (splitter / resequencer / scatter-gather).
- Two **routing headers**: `REPLY_CHANNEL` and `ERROR_CHANNEL` carry references to channels. Spring uses `HeaderChannelRegistry` to swap actual channel objects with `String` IDs at serialization boundaries; yarumo's equivalent must do the same — channel references **must not cross brokers** as raw pointers.

#### `MessageChannel`

The "pipe" of pipes-and-filters. Two orthogonal axes:

| Semantics | Buffering |
|---|---|
| **Point-to-point** (one consumer wins) | **Subscribable** (push: register a handler) |
| **Publish-subscribe** (all consumers get a copy) | **Pollable** (pull: call `receive()` on a tick) |

Two interfaces:

```java
SubscribableChannel : MessageChannel + subscribe(handler)
PollableChannel     : MessageChannel + receive(timeout) Message
```

#### `MessageHandler`

```java
interface MessageHandler { void handleMessage(Message<?> m); }
```

Single method. Every endpoint type (routers, transformers, aggregators, service activators, splitters, …) reduces to a `MessageHandler`.

### 2.2. Channel taxonomy — full table

| Channel | Semantics | Buffering | Concurrency | Use case |
|---|---|---|---|---|
| **`DirectChannel`** | point-to-point | subscribable | sender thread runs the handler (synchronous, single-threaded handoff) | default; round-robin between subscribers; **propagates exceptions** synchronously to producer |
| **`QueueChannel`** | point-to-point | pollable | internal `BlockingQueue`; consumer thread is a `PollingConsumer` | decouple producer/consumer rate; bounded; FIFO |
| **`PublishSubscribeChannel`** | broadcast | subscribable | sync by default; can be async via `TaskExecutor`; each subscriber gets a copy | event fan-out; with `errorHandler` to swallow per-subscriber failures |
| **`PriorityChannel`** | point-to-point | pollable | `PriorityBlockingQueue` ordered by header `PRIORITY` or custom `Comparator` | quality-of-service queues |
| **`RendezvousChannel`** | point-to-point | pollable | zero-capacity `SynchronousQueue` — sender blocks until receiver ready | back-pressure handoff; never queues |
| **`ExecutorChannel`** | point-to-point | subscribable | dispatches each send to a `TaskExecutor` (pool) | async point-to-point; **decouples** sender thread from handler thread |
| **`PartitionedChannel`** | point-to-point | subscribable | N independent partitions; key extractor maps each message to one partition; each partition has its own worker thread | preserve per-key ordering with parallelism (classic Kafka-consumer-style sharding) |
| **`FluxMessageChannel`** | point-to-point | subscribable | Project Reactor `Sinks.Many<Message>`; pushes through `Flux` | reactive-aware bridge to Reactor |
| **`NullChannel`** | sink | n/a | discards (and counts) | `/dev/null` for unwanted output |

Three more concepts orthogonal to the table:

- **`InterceptableChannel`** — a base trait that lets any channel accept a chain of `ChannelInterceptor`s (§ 3.9).
- **Message-store-backed channels** — `QueueChannel` / `PriorityChannel` can take a `ChannelMessageStore` so unread messages survive restart.
- **`FixedSubscriberChannel`** — bound to **exactly one** handler at construction time; cheaper than `DirectChannel` for chain links.

### 2.3. EIP endpoints — the catalogue

Spring Integration calls every connector between channels a "Message Endpoint". The six classical EIP endpoint types:

| Endpoint | Cardinality | What it does |
|---|---|---|
| **Transformer** | 1 → 1 | Convert payload/headers; pure function. |
| **Filter** | 1 → 0 \| 1 | Boolean `Selector(Message) bool`; drop or pass. Optional `discardChannel`. |
| **Router** | 1 → 1 \| N | Select output channel(s) per message. Six concrete sub-types — see § 3.1. |
| **Splitter** | 1 → N | Emit N messages with `SEQUENCE_NUMBER` / `SEQUENCE_SIZE` / `CORRELATION_ID` headers set. |
| **Aggregator** | N → 1 | Stateful inverse of splitter; correlate + release + combine. See § 3.2. |
| **Service Activator** | 1 → 1 \| 0 | "Invoke business code" — adapter from `MessageHandler` to a domain function. The most common endpoint in real flows. |

Three more endpoints that are not in the original Hohpe/Woolf book but are essential in Spring Integration:

| Endpoint | Cardinality | What it does |
|---|---|---|
| **Resequencer** | N → N | Like aggregator but emits in `SEQUENCE_NUMBER` order, one by one (not combined). |
| **Bridge** | 1 → 1 | No-op handler used to connect channels of different transports / interceptors. |
| **Chain** | 1 → 1 | Linear composition of multiple handlers **without** intermediate channels — efficient pipeline. |

### 2.4. Channel adapters and gateways

Two shapes for "edge" components that bridge `MessageChannel` to the outside world:

- **Channel adapter (one-way)** — `inbound`: external → message → channel; `outbound`: channel → message → external. No reply path.
- **Messaging gateway (two-way)** — request-reply. Inbound gateway: external request → channel + wait for response on reply-channel → external reply. Outbound gateway: channel → external request → reply channel.

Spring Integration ships 40+ of each. The ones with non-trivial machinery beyond "wrap an SDK":

#### 2.4.1. File adapter (`spring-integration-file`)

- **`FileReadingMessageSource`** — polled inbound source. Walks a directory each tick, applies a **filter chain**, returns one `File` per `receive()`.
- **`FileListFilter` chain**:
  - `AcceptOnceFileListFilter` — in-memory dedupe for process lifetime.
  - `PersistentFileListFilter` — same, backed by a `MetadataStore` (survives restarts; race-safe via CAS).
  - `LastModifiedFileListFilter`, `SimplePatternFileListFilter`, `IgnoreHiddenFileListFilter`, `CompositeFileListFilter`.
- **Read modes**: `REF` (default — payload is the `File`), `STREAM` (`InputStream`), `BYTES`, `LINES` (iterator of strings).
- **`FileWritingMessageHandler`** outbound: writes payload to disk; `setUseTemporaryFileName(true)` writes to `<name>.writing` then atomic-renames; modes `REPLACE` / `APPEND` / `IGNORE_IF_EXISTS` / `FAIL`; supports `setPreserveTimestamps`, `setChmod`, `setNewFileCallback`.
- **`FileTailingMessageProducer`** — `tail -f`-style inbound, OS-native or pure-Java implementation.
- **File splitter** — emits one message per line with `START_FILE` / `END_FILE` / `LINE_<n>` markers.
- **File aggregator** — reassembles a split file by correlation header.
- **Locking** — advisory `.lock` file or pluggable `FileLocker` SPI.

#### 2.4.2. HTTP adapter (separate from Spring MVC)

- **`HttpInboundEndpoint`** — registered as a Spring MVC handler; routes a request **into** a channel. Returns the reply from the gathered channel (inbound gateway) or `202 Accepted` (inbound adapter).
- **`HttpRequestExecutingMessageHandler`** — consumes from a channel, sends an HTTP request via `RestTemplate` (sync) or `WebClient` (reactive), maps response back to a `Message`.
- **`DefaultHttpHeaderMapper`** — bidirectional HTTP↔message header mapping; standard HTTP headers (`Authorization`, `Content-Type`, …) plus user-prefixed `X-…` rules.
- URI variables via SpEL (yarumo replaces with explicit `func(Message) string`).
- Multipart support, status-code → channel mapping, timeout, proxy, custom error handler for non-2xx.

This is **HTTP as a message protocol**, not HTTP as an MVC framework: the endpoint dispatch is owned by Spring MVC; Spring Integration only attaches the bridge.

#### 2.4.3. JDBC / JPA adapters (DB-as-broker)

- **`JdbcPollingChannelAdapter`** — runs a `SELECT`, emits rows as messages, then runs an `UPDATE` to mark-as-read **in the same transaction**.
- **`JdbcMessageHandler`** — outbound: writes rows from a message.
- **`JdbcOutboundGateway`** — round-trip: write, then read-back.
- **`StoredProcInboundChannelAdapter`** — stored-procedure version.
- **JPA equivalents** with a `JpaExecutor` (entity-based, named queries, paging, delete-after-poll).
- **PostgreSQL `LISTEN/NOTIFY`** (v6.0+) — `PostgresChannelMessageTableSubscriber` + `PostgresSubscribableChannel` provide push-based notifications using the JDBC channel-message-store as a transactional queue. **This is the killer feature** for "DB-as-broker without external dependencies".

#### 2.4.4. WebSocket / STOMP

- **`IntegrationWebSocketContainer`** — split into `ClientWebSocketContainer` (consume external WS) and `ServerWebSocketContainer` (expose WS).
- Sub-protocol handlers: `PassThruSubProtocolHandler` (raw frames) vs `StompSubProtocolHandler` (full STOMP framing).
- Session registry, `websocket_sessionId` header for fan-out targeting.
- SockJS fallback (legacy browsers).

#### 2.4.5. Broker drivers (AMQP, Kafka, MQTT, JMS, …)

Standard shape: inbound channel adapter consumes from broker → channel; outbound channel adapter publishes channel → broker. Bidirectional gateways for RPC-style. All driven by the underlying Spring `Listener` / `Template` abstractions.

### 2.5. IntegrationFlow DSL

Spring's fluent Java DSL for wiring flows. Three flavours: Java, Kotlin, Groovy.

```java
IntegrationFlow.from("inputChannel")
  .filter(payload -> payload.length > 0)
  .transform(JsonToMap.class)
  .route(Map.class, m -> (String) m.get("type"),
         mapping -> mapping
             .channelMapping("order", "ordersChannel")
             .channelMapping("event", "eventsChannel"))
  .get();
```

Yarumo equivalent: a thin builder that composes `MessageHandler` chains and registers them with a `Registry`. **Not** a string-DSL — it's just Go constructors and method-chains for ergonomics, no annotations and no reflection.

---

## 3. Advanced features — the mechanics that matter

The base catalogue above is necessary; the mechanics below are what separate a "thin broker wrapper" from a real EIP implementation. Every subsection is a concrete mechanism that has design implications for `modules/messaging/`.

### 3.1. Advanced router patterns

Spring exposes **six** router subtypes with different routing-key semantics:

| Router | Routing key | Output cardinality |
|---|---|---|
| **`PayloadTypeRouter`** | `payload.getClass()` | one (per class) |
| **`HeaderValueRouter`** | named header's value | one (per value) |
| **`RecipientListRouter`** | static recipient list with optional `Selector(Message) bool` per recipient | many (broadcast + filter) |
| **`ExceptionTypeRouter`** | exception type chain (cause → root) | one |
| **`ErrorMessageExceptionTypeRouter`** | unwraps `ErrorMessage`, then routes by cause type | one |
| **`XPathRouter`** | XPath result on XML payload | one |

Base abstraction `AbstractMessageRouter` exposes three knobs:

- `defaultOutputChannel` — fallback if no key resolves.
- `resolutionRequired` (bool) — if `true`, **throws** on unmapped key; if `false`, silently sends to default (or drops).
- `channelMapping` — `Map<String, Channel>`.

**Recipient-list with selectors** is under-appreciated. It's the active-broadcast pattern: the router knows its recipients up-front and **filters** per recipient via selector. Distinct from a `PubSubChannel`, where routing is passive (subscribers register themselves with the channel).

### 3.2. Aggregator full semantics

The aggregator is the most complex single endpoint. Four orthogonal axes plus three expiration flags plus a partial-release switch.

#### 3.2.1. CorrelationStrategy — the grouping key

- **`HeaderAttributeCorrelationStrategy`** (default): pulls `CORRELATION_ID` header.
- POJO / function: `func(Message) any` returns the grouping key.
- Multi-aggregator caveat: shared `MessageGroupStore` between aggregators requires per-aggregator **partitioning** (JDBC `region`, Mongo `collectionName`).

#### 3.2.2. ReleaseStrategy — when to fire

- **`SimpleSequenceSizeReleaseStrategy`** (default v5+): waits until `SEQUENCE_NUMBER == SEQUENCE_SIZE`. O(1) per add.
- **`SequenceSizeReleaseStrategy`** (legacy): same + duplicate-sequence detection. O(n) — expensive on large groups.
- **`MessageCountReleaseStrategy`**: fires when group size hits N.
- POJO / function: `func([]Message) bool`.

#### 3.2.3. MessageGroupStore — state backend

- `SimpleMessageStore` (in-memory).
- `JdbcMessageStore`, `MongoDbMessageStore`, `RedisMessageStore`, `HazelcastMessageStore`, `GemfireMessageStore`.
- `MessageGroupFactory` (v4.3+) lets you pick the in-group container: `HASH_SET` (default), `SYNCHRONISED_SET`, `BLOCKING_QUEUE`, `PERSISTENT`, `LIST`.
- `setLazyLoadMessageGroups(false)` (v4.3+) and `streamMessagesForGroup` (v5.5+) optimise for large groups.

#### 3.2.4. LockRegistry — concurrent-add safety

Per-correlation-key lock so two parallel `Send`s for the same key serialise. Backends: `DefaultLockRegistry` (in-process), `RedisLockRegistry`, `JdbcLockRegistry`, `ZookeeperLockRegistry`. Deadlock risk when chaining aggregators sharing a lock registry — mitigations: `releaseLockBeforeSend=true` (v5.1.1+), or async channel hop between aggregators.

#### 3.2.5. Group timeout — forced release / discard

- `group-timeout` — fixed ms idle window before forcing release.
- `group-timeout-expression` — dynamic per-group (e.g. "if size≥2 then 10s else infinite"). Returns `long` (ms-from-now), `Date` (absolute), `0`/negative (immediate), `null` (no-op).
- On timeout: `ReleaseStrategy` gets **one more chance** to release; if still false, the partial-release / discard flag decides.

#### 3.2.6. Expiration flags — group lifecycle

Two orthogonal flags:

- `expire-groups-upon-completion` (default **false**) — after **normal release**, retain empty group metadata so late messages with same correlation key are **discarded**. If `true`, group is fully removed → late messages **start a new group**.
- `expire-groups-upon-timeout` (default **true**) — same behaviour, but after the timeout path.

Two extra knobs for partial release:

- `send-partial-result-on-expiry` (default **false**) — if `true`, on timeout emit accumulated messages; if `false`, discard to `discard-channel` (or drop).
- `discardIndividuallyOnExpiry` (v6.5+) — when discarding, emit each member separately or as one list payload.

#### 3.2.7. Output

- `MessageGroupProcessor` (default `DefaultAggregatingMessageGroupProcessor`) returns `payload = List<payload>` of group members.
- Custom function transforms `[]Message → Message`.
- `headers-function` (`func(MessageGroup) map[string]any`) controls output headers (default keeps non-conflicting headers from all members).
- `popSequence` (default `true`) — removes sequence headers from output, allowing re-splitting downstream.

### 3.3. Scatter-Gather — compound endpoint

`ScatterGatherHandler` is a composition of:

- **Scatterer**: either a `PublishSubscribeChannel` (auction — all subscribers race on the same message) or a `RecipientListRouter` (distribution — explicit recipient list with per-recipient selectors). Mutually exclusive.
- **Gatherer**: an `AggregatingMessageHandler` with `release-strategy = size() == N`.
- **Gather timeout** (default 30s) — wait this long for replies, then aggregate what arrived.
- **Error channel** (`errorChannelName`, v5.1.3+) — per-sub-flow errors flow here instead of failing the whole gather → partial gathering survives individual failures.
- **Async mode** (v6.5.3+) — returns `Mono` so the request thread doesn't block.

Distinct from a plain aggregator because **scatter-gather knows the cardinality** (it created it). With `applySequence=true` on the router, `SEQUENCE_SIZE = recipients.len`, which lets the gatherer release-strategy be a trivial size check.

Use cases: "best quote" (RFQ to N suppliers, take best), parallel data fetches with merge, fan-out + reduce.

### 3.4. Polling consumer mechanics

`SourcePollingChannelAdapter` + `PollerMetadata` form the standard polling envelope. The knobs:

| Knob | Default | Purpose |
|---|---|---|
| Trigger | `FixedDelay(1s)` | Wait N ms **after** processing ends (self-throttling). Alternatives: `FixedRate(N ms)` (regardless of duration), `Cron(expr)`, `DynamicPeriodicTrigger`. |
| `maxMessagesPerPoll` | 1 | How many `receive()` to attempt per tick (-1 = unlimited until empty). |
| `receiveTimeout` | 1000 ms | Block this long on each `receive()`. |
| `taskExecutor` | nil (sync) | Run polls on a pool (parallelism). |
| `errorHandler` | `LoggingHandler` | Default just logs and continues. |
| `transactionManager` | nil | Wrap each poll in a transaction (commit on success, rollback on exception). |
| `adviceChain` | `[]` | AOP advices around `receive + process`. |
| `receiveMessageAdvice` | `[]` | Advices with **access to the polled message** (v5.3+) — enables "slow down when empty, speed up when busy" via `SimpleActiveIdleReceiveMessageAdvice`. |

`fixed-delay` vs `fixed-rate` is the single most important choice: with a slow handler, `fixed-rate` can pile up overlapping ticks; `fixed-delay` self-throttles. **Default to fixed-delay** unless you have a specific real-time requirement.

Long-polling pattern: combine `receiveTimeout=30s` with `fixed-rate=10ms` to emulate event-driven on a `QueueChannel` with minimal CPU.

### 3.5. Request-handler advice chain

`<request-handler-advice-chain/>` wraps an endpoint (typically a `ServiceActivator` or outbound gateway) with AOP-style advices. Catalogue:

| Advice | Mechanism |
|---|---|
| **`RequestHandlerRetryAdvice`** | `RetryTemplate` integration: `maxAttempts`, `FixedBackOff` / `ExponentialBackOff`, `RecoveryCallback` to a fallback channel (`ErrorMessageSendingRecoverer`). **Stateless** = in-flight retry on the same thread; **stateful** = re-enter from upstream broker on each attempt (each attempt is a fresh transaction). |
| **`RequestHandlerCircuitBreakerAdvice`** | Open after `threshold` consecutive failures; fail fast while open; half-open after `halfOpenAfter`. |
| **`RateLimiterRequestHandlerAdvice`** | Resilience4j `RateLimiter` wrapper. |
| **`CacheRequestHandlerAdvice`** | Spring `@Cacheable` semantics for endpoints. |
| **`LockRequestHandlerAdvice`** | Wrap call in a `LockRegistry` lock (per-key serialisation). |
| **`ExpressionEvaluatingRequestHandlerAdvice`** | SpEL on success/failure; routes **original** message to `successChannel` / `failureChannel`. Used for "move FTP file to processed/ on success, to error/ on failure". `trapException=true` swallows exceptions silently. |
| **`ReactiveRequestHandlerAdvice`** | Reactor-friendly variant. |
| **`ContextHolderRequestHandlerAdvice`** | Per-handler `ThreadLocal` / Reactor-context propagation. |
| **`IdempotentReceiverInterceptor`** | Idempotent-receiver advice (§ 3.7). |

Endpoint-scoped advice beats poller-scoped advice: scoping retry to `httpGateway2` doesn't retry `httpGateway1` or the downstream `jdbcOutboundAdapter` on failure.

Two anti-patterns to flag in yarumo's translation:

1. **Backoff on the polling thread blocks the whole pipeline.** Pair RetryAdvice with an `ExecutorChannel` upstream so the poller stays free.
2. **`trapException=true` (silently swallow exception)** is a maintenance trap. Errors should always propagate to an `ErrorChannel` — never disappear.

### 3.6. Message stores — the persistence backbone

Four pluggable interfaces in `org.springframework.integration.store`. All four feed multiple endpoint types (aggregator, idempotent receiver, claim check, delayer, barrier, resequencer, durable queue channels, persistent file filter).

#### 3.6.1. `MessageStore` — generic message persistence

```
addMessage(msg) / getMessage(id) / removeMessage(id)
getMessageCount()
```

Primary use: **claim-check pattern** — replace large payload with `claimCheck` header, store original by ID, restore later. Saves channel space.

#### 3.6.2. `MessageGroupStore` — aggregator state

```
addMessageToGroup(groupId, msg)
getMessageGroup(groupId)
removeMessageGroup(groupId)
iterator()  // for MessageGroupStoreReaper
```

Backends:
- `SimpleMessageStore` (in-memory).
- `JdbcMessageStore` — schema `INT_MESSAGE_GROUP` / `INT_MESSAGE` / `INT_GROUP_TO_MESSAGE` with `region` for tenant isolation.
- `JdbcChannelMessageStore` — specialised for queue-channel backing (`INT_CHANNEL_MESSAGE` table); per-dialect query providers (`PostgresChannelMessageStoreQueryProvider`, `MySqlChannelMessageStoreQueryProvider`, `OracleChannelMessageStoreQueryProvider`, …).
- `MongoDbMessageStore` (collection-per-region).
- `RedisMessageStore` (sorted set + hash).
- `GemfireMessageStore`, `HazelcastMessageStore`.

`MessageGroupStoreReaper` is a scheduled bean that scans for stale groups and expires them — without it, `expire-groups-upon-timeout` is a no-op for in-memory stores.

#### 3.6.3. `MetadataStore` / `ConcurrentMetadataStore` — key-value state

```
get(key) (value, ok)
put(key, value)
remove(key)
// ConcurrentMetadataStore adds:
putIfAbsent(key, value) (existing, ok)  // CAS
replace(key, oldValue, newValue) bool   // CAS
```

The CAS primitives are what make `PersistentFileListFilter` and `IdempotentReceiver` race-safe. Backends: `SimpleMetadataStore`, `JdbcMetadataStore`, `RedisMetadataStore`, `HazelcastMetadataStore`, `GemfireMetadataStore`, `ZookeeperMetadataStore`, `MongoMetadataStore`.

**Gap**: most backends don't support TTL — entries live forever unless explicitly removed. Redis / Mongo native TTL is the exception. JDBC needs a sweeper for bounded storage.

#### 3.6.4. `LockRegistry` — distributed locks

```
obtain(key) Lock     // standard Lock interface
```

Backends: `DefaultLockRegistry` (in-process `Map<key, Mutex>`), `RedisLockRegistry`, `JdbcLockRegistry`, `ZookeeperLockRegistry`.

Cross-aggregator deadlock risk; see § 3.2.4.

#### 3.6.5. Postgres push channel — DB-as-broker without Kafka

`PostgresChannelMessageTableSubscriber` + `PostgresSubscribableChannel` use Postgres `LISTEN/NOTIFY` to convert the polled `JdbcChannelMessageStore` into a **push subscribable channel**. Effect: a Postgres-only deployment gets a transactional broker for free — no Kafka, no RabbitMQ. Highest leverage for DaaS (Postgres-first).

Anti-pattern: using JDBC channels as the **primary** broker on any non-Postgres DB. Postgres MVCC + LISTEN/NOTIFY makes it work; on MySQL or Oracle, poll-based JDBC channels are a performance trap. Use them as **transactional outbox backend**, not primary bus.

### 3.7. Idempotent Receiver

EIP advice that wraps a handler:

1. Compute a key from the message — `func(Message) string`.
2. Look up in `MetadataStore`. **Hit** → either discard, send to `discardChannel`, throw, or mark `duplicateMessage=true` header and pass through.
3. **Miss** → process, then store the key.

Optional `compareValues` predicate for "saw this key but newer value wins" patterns (e.g. last-write-wins on file line numbers).

Pairs with at-least-once consumers (Kafka, AMQP). Spring's outbox + idempotent-receiver combo essentially gives "transactional outbox with replay-safety".

### 3.8. Observability — built into the framework

Spring Integration 6.0+ uses Micrometer Observation natively. Instrumentation points:

| Component | Span | Observation |
|---|---|---|
| `AbstractMessageChannel.send()` | `PRODUCER` | `IntegrationObservation.PRODUCER` |
| `AbstractMessageHandler` | `CONSUMER` | `IntegrationObservation.HANDLER` |
| `MessagingGatewaySupport` (request-reply inbound) | `SERVER` | `IntegrationObservation.GATEWAY` |
| `MessageProducerSupport` (inbound) | `CONSUMER` | `IntegrationObservation.HANDLER` |
| `SourcePollingChannelAdapter` (v6.5+) | `CONSUMER` | `IntegrationObservation.HANDLER` |

Metrics emitted (Micrometer Timer / LongTaskTimer / Gauge / Counter):

```
spring.integration.send           Timer    type=channel|handler  result name exception
spring.integration.receive        Counter  type=channel|source   result name
spring.integration.gateway        Timer + LongTaskTimer          name type outcome
spring.integration.handler        Timer + LongTaskTimer          name type
spring.integration.producer       Timer + LongTaskTimer          name type
spring.integration.channels       Gauge
spring.integration.handlers       Gauge
spring.integration.sources        Gauge
spring.integration.channel.queue.size              Gauge  (per QueueChannel)
spring.integration.channel.queue.remaining.capacity Gauge
```

Trace propagation:

- **Producer side** — `MessageSenderContext` writes trace/span headers into the message **before** the channel hop (so brokers and queue channels carry the trace).
- **Channel** — preserves headers across (possibly async) hop; Kafka channel propagates via Kafka headers, AMQP via headers, JDBC channel via row columns.
- **Consumer side** — `MessageReceiverContext` extracts headers and creates a child span.

Enabled per-channel via `@EnableIntegrationManagement(observationPatterns = "*")` (default: off). Tags: `spring.integration.name`, `spring.integration.type`, `spring.integration.outcome`.

OTel semantic conventions for messaging: `messaging.system`, `messaging.destination.name`, `messaging.operation` (`send` / `receive` / `process`), `messaging.message.id`, `messaging.message.conversation_id` (= correlation_id).

### 3.9. Channel interceptors and wire-tap

`ChannelInterceptor` is the single extension point for channel-level cross-cutting concerns. Six callbacks:

```
preSend(msg, channel) Message            // mutate or drop (return null)
postSend(msg, channel, sent)             // log success
afterSendCompletion(msg, channel, sent, err)  // cleanup (resources)
preReceive(channel) bool                 // pollable only — allow receive?
postReceive(msg, channel) Message        // pollable only
afterReceiveCompletion(msg, channel, err) // pollable only
```

Built-in interceptors:

- **`WireTap`** — copy every message to a secondary channel (audit, debug). Configurable `Selector(Message) bool` filter and send `Timeout`. Implementation: it's just a `ChannelInterceptor` calling `tapChannel.send(msg, timeout)` from `preSend`.
- **`MessageSelectingInterceptor`** — drop messages where `Selector` returns false.
- **`GlobalChannelInterceptor`** — declared with a pattern (`"orders.*"`); auto-applies to all matching channels at registration time.
- **Observation interceptors** — Micrometer adds these silently when `observationPatterns` matches.

Wire-tap is **the** auditing/debugging primitive — clean place to plug instrumentation without polluting business handlers.

### 3.10. Other gaps worth naming

#### 3.10.1. Handler chain

`MessageHandlerChain` composes handlers `H1 → H2 → H3` **without intermediate channels**. Each non-tail element must implement `MessageProducer` (i.e. accept an output channel). Equivalent to Go composition: a `ChainedHandler` struct that calls each in sequence. Cheap; nice ergonomics for "this filter then that transformer then that activator".

#### 3.10.2. Delayer

Postpones message dispatch by N ms (header-driven or fixed). Backed by a `TaskScheduler` and persistable via `MessageGroupStore` (so delayed messages survive restart). Useful for "retry in 5 minutes" inside a flow without resorting to broker DLQ + TTL.

#### 3.10.3. Barrier

`BarrierMessageHandler` blocks a message until a **trigger message** with the same correlation ID arrives. Implements the synchronizer pattern. Niche but interesting for "wait for callback" flows.

#### 3.10.4. Resequencer

Brother of the aggregator: takes N out-of-order messages with `SEQUENCE_NUMBER`, emits them in order one-by-one. Same `MessageGroupStore` plumbing, different output processor.

#### 3.10.5. Control bus

Receives commands as messages (SpEL strings — `"@myEndpoint.start()"`) and invokes lifecycle on integration components. Anti-pattern (SpEL in production), but the **concept** of "send a message to control the flow" is sound. Yarumo equivalent: `managed/Lifecycle` is already operated via Go calls — no need to wrap it in messaging unless remote control is wanted.

#### 3.10.6. Integration graph

`IntegrationGraphServer` snapshots the wired topology (channels + endpoints + handlers + subscriptions) as JSON. Optional `integration-graph-controller` HTTP endpoint. Yarumo equivalent: a `func (r *Registry) Graph() Graph` returning a structured topology for `/actuator`-style diagnostics.

#### 3.10.7. Message history

`MESSAGE_HISTORY` header accumulates entries `{name, type, timestamp}` for each component the message passes through (if `<message-history/>` enabled). Lightweight, cheap, distinct from full distributed tracing — ideal for log breadcrumbs.

#### 3.10.8. Schema-aware messaging

Spring Integration's Kafka driver integrates with Confluent Schema Registry for Avro / Protobuf payloads. Pre-publish validate, post-consume validate, schema-evolution checks. Spring Kafka does the heavy lifting; Spring Integration just hands payloads through.

#### 3.10.9. Debezium CDC

`spring-integration-debezium` consumes Debezium connector events (Postgres / MySQL / MongoDB CDC). Envelope is `before` / `after` / `source` / `op`. Spring exposes this as a regular inbound channel adapter.

---

## 4. Mapping to yarumo `modules/messaging/` (§ 1.3)

The roadmap (`docs/ROADMAP_NEW_MODULES.md` § 1.3) sketches `messaging/` with `message.go`, `channel.go`, `channels/`, `endpoints/`, `schema/`, `events/`, `rabbitmq/`, `kafka/`, `nats/`. That's the **starting** layout. The catalogue below is a concrete refinement that surfaces from the § 3 mechanics — the user decides what to promote to roadmap.

### 4.1. Concrete file layout proposal

```
modules/messaging/
  message.go                Message[T], Headers (immutable Builder)              [in roadmap]
  headers.go                Header constants:                                    [NEW]
                              CORRELATION_ID, SEQUENCE_NUMBER, SEQUENCE_SIZE,
                              REPLY_CHANNEL, ERROR_CHANNEL, CONTENT_TYPE,
                              PRIORITY, MESSAGE_HISTORY, CLAIM_CHECK,
                              DUPLICATE_MESSAGE, EXPIRATION_DATE
  channel.go                Channel (Send/Receive), SubscribableChannel,         [in roadmap]
                            PollableChannel interfaces
  handler.go                MessageHandler interface                             [in roadmap]
  source.go                 MessageSource[T] (pollable producer)                 [NEW]
  registry.go               In-process registry (name → channel,                 [NEW]
                            name → endpoint); Graph() snapshot for /actuator

  channels/
    direct.go               DirectChannel (sync, point-to-point)                 [in roadmap]
    queue.go                QueueChannel (FIFO, bounded, pollable)               [in roadmap]
    pubsub.go               PubSubChannel (broadcast)                            [in roadmap]
    priority.go             PriorityChannel                                      [in roadmap]
    rendezvous.go           RendezvousChannel (zero-cap handoff)                 [in roadmap]
    executor.go             ExecutorChannel (DirectChannel + worker pool)        [NEW]
    partitioned.go          PartitionedChannel (key → worker)                    [NEW]
    fixed_subscriber.go     FixedSubscriberChannel (single-subscriber, cheap)    [NEW]
    null.go                 NullChannel (sink + count metric)                    [NEW]
    intercepting.go         InterceptableChannel decorator that applies a       [NEW]
                            chain of ChannelInterceptors

  interceptors/                                                                  [NEW]
    interceptor.go          ChannelInterceptor interface (6 callbacks)
    wiretap.go              WireTap(target Channel, sel Selector, timeout)
    selecting.go            SelectingInterceptor(Selector) — drop on false
    history.go              MessageHistoryInterceptor — appends
                            {name, type, ts} to MESSAGE_HISTORY header
    global.go               Pattern-based global interceptor registration
                            ("orders.*" auto-applies)

  endpoints/                                                                     [in roadmap — refined]
    transformer.go          Transformer (payload conversion)
    filter.go               Filter (Selector predicate + optional discardChan)
    activator.go            ServiceActivator (call business fn)
    splitter.go             Splitter (1 → N with sequence headers)
    bridge.go               Bridge (no-op, transport adapter)                    [NEW]
    chain.go                ChainedHandler (linear, no intermediate channels)    [NEW]
    delayer.go              Delayer (TaskScheduler-driven delay,                 [NEW]
                            backed by MessageGroupStore for durability)
    barrier.go              BarrierHandler (wait for trigger by correlation ID)  [NEW]
    resequencer.go          Resequencer (orders by SEQUENCE_NUMBER, emits 1×1)   [NEW]
    scatter_gather.go       ScatterGather (scatterer + gatherer + timeout        [NEW]
                            + errorChannel + async)
    idempotent.go           IdempotentReceiver advice (MetadataStore             [NEW]
                            + KeyStrategy + DiscardChannel + ThrowOnReject
                            + compareValues)

    routers/                                                                     [NEW — split from single router.go]
      router.go             AbstractRouter (DefaultOutputChannel,
                            ResolutionRequired, ChannelMapping)
      payload_type.go       PayloadTypeRouter (reflect.TypeOf)
      header_value.go       HeaderValueRouter (header → channel)
      recipient_list.go     RecipientListRouter (Recipient{Channel,Selector};
                            broadcast + filter)
      exception_type.go     ExceptionTypeRouter (cause chain)
      error_message.go      ErrorMessageExceptionTypeRouter (unwrap)

    aggregator/                                                                  [NEW — non-trivial enough for subpkg]
      aggregator.go         Aggregator with all § 3.2 knobs
      correlation.go        CorrelationStrategy[K] type + builtins
                              HeaderCorrelation(headerName)
                              FuncCorrelation(fn)
      release.go            ReleaseStrategy + builtins
                              SequenceSizeRelease()
                              MessageCountRelease(n)
                              FuncRelease(fn)
      processor.go          MessageGroupProcessor + DefaultListProcessor
      headers_function.go   HeadersFunction[K] (group → headers map)
      reaper.go             MessageGroupStoreReaper (one goroutine per store)

  poller/                                                                        [NEW]
    poller.go               Poller config: Trigger, MaxMsgsPerPoll,
                            ReceiveTimeout, TaskExecutor, ErrorHandler,
                            AdviceChain
    trigger.go              Trigger interface + FixedDelay, FixedRate,
                            Cron(expr), DynamicPeriodic, ActiveIdle
    consumer.go             PollingConsumer (Source + Handler, runs on
                            goroutine via TaskExecutor)
    source_adapter.go       SourcePollingChannelAdapter

  advice/                                                                        [NEW]
    advice.go               HandlerAdvice interface (wraps MessageHandler)
    chain.go                ChainAdvice (apply N advices in order)
    retry.go                RetryAdvice (uses common/resilience backoff)
                              stateless + stateful modes + RecoveryCallback
    circuitbreaker.go       CircuitBreakerAdvice (uses common/resilience
                            CircuitBreakerRegistry)
    ratelimit.go            RateLimitAdvice (uses common/resilience
                            RateLimiterRegistry)
    cache.go                CacheAdvice (uses modules/cache)
    lock.go                 LockAdvice (uses LockRegistry)
    hook.go                 HookAdvice (onSuccess/onFailure routes original
                            msg to channel) — Go fns, NOT SpEL
    context.go              ContextAdvice (propagate values across handler
                            boundary; for traces, MDC, etc.)

  store/                                                                         [NEW]
    message_store.go        MessageStore interface
                              Add, Get, Remove, Count
    group_store.go          MessageGroupStore interface
                              AddToGroup, GetGroup, RemoveGroup, Iterator,
                              LastModified, StreamForGroup, Condition
    metadata_store.go       MetadataStore interface  (Get, Put, Remove)
                            ConcurrentMetadataStore (PutIfAbsent, Replace CAS)
                            TTLMetadataStore (optional capability —
                            PutWithTTL; backends no-op if unsupported)
    lock_registry.go        LockRegistry interface  (Obtain(key) Lock)
    inmemory/
      messages.go           SimpleMessageStore (sync.Map)
      groups.go             SimpleMessageGroupStore
      metadata.go           SimpleMetadataStore
      lock.go               LocalLockRegistry (sync.Map of sync.Mutex)
    jdbc/                   (depends on modules/datasource/gorm/)
      messages.go           JdbcMessageStore (INT_MESSAGE schema)
      groups.go             JdbcMessageGroupStore
      channel.go            JdbcChannelMessageStore (INT_CHANNEL_MESSAGE)
                            + QueryProvider per dialect
      metadata.go           JdbcMetadataStore
      lock.go               JdbcLockRegistry
      postgres_notify.go    PostgresSubscribableChannel
                            (LISTEN/NOTIFY push variant — DB-as-broker)
    redis/                  (depends on modules/datasource/goredis/)
      messages.go           RedisMessageStore
      metadata.go           RedisMetadataStore (native TTL)
      lock.go               RedisLockRegistry
    mongo/                  (depends on modules/datasource/mongo/)
      messages.go           MongoDbMessageStore
      groups.go             MongoDbMessageGroupStore
      metadata.go           MongoMetadataStore (native TTL via index)

  claimcheck/                                                                    [NEW]
    claimcheck.go           IntoClaimCheck(store), FromClaimCheck(store)
                            — Transformers backed by MessageStore.
                            Pattern: large payload → small ref.

  observability/                                                                 [NEW]
    conventions.go          OTel semantic conventions
                            (messaging.system, messaging.destination.name,
                            messaging.operation, messaging.message.id,
                            messaging.message.conversation_id)
    interceptor.go          ObservationInterceptor — instruments Send/Receive
                            via ChannelInterceptor (PRODUCER/CONSUMER spans)
    propagation.go          Inject/Extract trace headers via Headers
                            (works across QueueChannel hop & broker drivers)
    metrics.go              Standard meters (queue size gauge, send timer,
                            receive counter)

  schema/                   Schema Registry client                               [in roadmap]
    registry.go             Registry interface: Register, Get, Compatibility
    confluent/              Confluent Schema Registry client
    glue/                   AWS Glue Schema Registry client
    apicurio/               Apicurio Registry client

  events/                                                                        [in roadmap]
    publisher.go            Publisher façade over DirectChannel
    subscriber.go           Subscribe[T](fn) registers a typed handler
                            (nominal-typed pub/sub, no Message[T] envelope)

  rabbitmq/
    amqp/                   driver implementing Channel interfaces               [in roadmap]
    streams/                driver implementing Channel interfaces               [in roadmap]
  kafka/
    driver.go               Kafka inbound/outbound channel adapters              [in roadmap]
    cdc/                    Debezium CDC event parsing                           [in roadmap]
  nats/                                                                          [in roadmap]

  file/                                                                          [NEW]
    source.go               FileSource (poll directory + filter chain)
    handler.go              FileWriter (atomic move, modes, charset)
    filter.go               FileListFilter interface +
                              AcceptOnce, LastModified, Pattern,
                              Composite, IgnoreHidden builtins
    persistent.go           PersistentFileListFilter(MetadataStore)
                            — CAS-based dedupe across restarts
    tail.go                 FileTailSource (tail -f; OS or pure-Go impl)
    splitter.go             FileSplitter (line-by-line + markers)
    aggregator.go           FileAggregator (reassemble split file)
    lock.go                 FileLocker SPI + .lock impl

  http/                                                                          [NEW]
    inbound.go              HttpInboundEndpoint (net/http or gin handler
                            → Channel) — registers on shared mux
    outbound.go             HttpOutboundHandler (Channel → HTTP request)
                            sync via net/http; mapping per status code
    headers.go              HttpHeaderMapper (HTTP ↔ Message header)
                            bidirectional + X-* prefix rules
```

### 4.2. Why splitting `endpoints/router.go` matters

In Spring, `AbstractMessageRouter` is one base class with six+ subtypes; in Go without inheritance, that's a flat ecosystem of small structs. Keeping them in `routers/` prevents `router.go` from becoming a 600-line file. Each router gets its own test file and its own `Options` struct, and the consumer picks one explicitly by constructor.

### 4.3. Why `aggregator/` is a subpackage, not a file

The four orthogonal axes (correlation × release × store × locks) plus expiration semantics plus the partial-release / discard switch produce a non-trivial cartesian product. Folding them all into one `aggregator.go` would be impenetrable. A subpackage gives each axis a file and lets options compose cleanly:

```go
agg := aggregator.New(
  aggregator.WithCorrelation(aggregator.HeaderCorrelation("order_id")),
  aggregator.WithRelease(aggregator.MessageCountRelease(5)),
  aggregator.WithGroupTimeout(10 * time.Second),
  aggregator.WithGroupStore(redisGroups),
  aggregator.WithLockRegistry(redisLocks),
  aggregator.WithProcessor(aggregator.SumProcessor),
  aggregator.WithExpireOnCompletion(false),
  aggregator.WithSendPartialOnExpiry(true),
  aggregator.WithDiscardChannel(deadLetterChan),
)
```

### 4.4. Why `store/` is at module root

`MessageStore` / `MessageGroupStore` / `MetadataStore` / `LockRegistry` are pluggable interfaces consumed by **multiple** packages: `aggregator/` (group store), `endpoints/idempotent.go` (metadata store), `claimcheck/` (message store), `file/persistent.go` (metadata store), `endpoints/delayer.go` (group store), `endpoints/barrier.go` (group store), `endpoints/resequencer.go` (group store), `poller/` (metadata store for "remember where we left off"), and `store/jdbc/postgres_notify.go` (itself a channel impl backed by a channel message store).

Putting `store/` at module root keeps each dependent package from re-declaring its own ad-hoc state interface. Backend subpackages (`store/jdbc/`, `store/redis/`, `store/mongo/`) compose with the matching `modules/datasource/*` driver — they depend on the driver, not vice versa.

### 4.5. Why `observability/` lives inside `messaging/`

Two reasons. First: it's a tight coupling — observation only makes sense in terms of channel / endpoint / source / sink concepts that don't exist outside this module. Second: yarumo's `modules/telemetry/otel/` is the **SDK setup** (provider, exporter, sampler); per-domain instrumentation lives with the domain. Same pattern as `telemetry/otel/genai/` planned next to `modules/llm/`.

### 4.6. Why `file/` and `http/` live inside `messaging/`

The roadmap mentions broker drivers (`rabbitmq/`, `kafka/`, `nats/`) but not file or HTTP. Both file and HTTP share the **channel adapter** shape (one-way) and the **gateway** shape (two-way) — same `Send` / `Receive` interfaces, same `Headers` handling, same poller plumbing. Promoting them to siblings of `rabbitmq/` is the cleanest fit.

For HTTP specifically: `modules/managed/server_http` already wraps `net/http` for **lifecycle** (Start/Stop). `messaging/http/` is **content adaptation** (HTTP request → `Message` → channel). They compose: the managed HTTP server hosts handlers that `messaging/http/inbound.go` registered. The `tools/routegen/` tool (§ 2.1 of the roadmap) can emit route definitions that messaging-http composes with channel-publishing handlers — three layers cooperating cleanly.

### 4.7. Conscious omissions

Worth being explicit about what is **not** in the refinement:

- **SpEL** — Spring's SpEL drives correlation / release / key extraction. Yarumo uses Go functions everywhere. `modules/common/expressions` exists but is reserved for user-facing decision rules in `sdks/decisions/`, not framework wiring.
- **Annotations** — `@Aggregator` / `@CorrelationStrategy` / `@ServiceActivator` are explicit anti-patterns. Everything is wiring code.
- **XML namespace** — irrelevant; Go has no XML config tradition.
- **Reactive Streams** — Spring Integration 6.x has heavy Reactor support (`MessageChannels.flux()`, `FluxAggregatorMessageHandler`, async `Mono` reply). Go's `iter.Seq` covers the **iteration** case; full reactive (subscription / backpressure / hot-cold) is **out of scope** for `messaging/`. If consumers need it, that lives in user code.
- **WebSocket / STOMP / SockJS** — defer until a real consumer appears. Same defer bucket as SSE.
- **Control bus** — anti-pattern (SpEL). Lifecycle is already operated via `managed/` calls — no need to wrap it in messaging.
- **GraphQL inbound adapter** — niche; user code can publish to a channel from a resolver if needed.
- **`@MessagingGateway` proxy generation** — Go has no proxies. Replace with explicit constructors.

---

## 5. Anti-patterns to avoid

- **SpEL strings as routing keys / correlation strategies / release predicates.** Yarumo uses Go functions.
- **Annotation-driven endpoint discovery** (`@Aggregator`, `@CorrelationStrategy`, `@ServiceActivator`, `@Splitter`). Yarumo wires endpoints explicitly via constructors.
- **God-class registry** holding every channel and every endpoint and deciding everything (Spring's `ApplicationContext` flavour). Yarumo's `Registry` is a flat lookup — names → instances — and exposes nothing else.
- **SpEL "Control Bus"** that lets remote callers invoke `start()` / `stop()` via expression strings. Lifecycle is operated by `managed/` calls, period.
- **Cross-aggregator `MessageGroupStore` sharing without partitioning.** Aggregators get dedicated stores or partition by `region` / collection name.
- **`trapException=true`** semantics (silently swallow exceptions and return null). Errors propagate or hit the `ErrorChannel` — never disappear.
- **Backoff-in-handler-thread on the polling thread.** RetryAdvice with exponential backoff must be paired with an `ExecutorChannel` upstream; otherwise the poller goroutine blocks and the whole flow stalls.
- **JDBC channel as primary broker** on non-Postgres DBs. Use as outbox backend, not primary bus. Postgres MVCC + `LISTEN/NOTIFY` is the exception that makes the pattern viable.
- **Wire-tap to a `QueueChannel` with a slow consumer** — backs pressure into the main flow. Wire-tap targets should be fire-and-forget (`DirectChannel` to an async writer, or `ExecutorChannel`).
- **`fixed-rate` polling on a slow handler** — ticks pile up; default to `fixed-delay`.
- **Untimed `MetadataStore` entries on JDBC backends.** Without TTL or a sweeper, `IdempotentReceiver` storage grows unbounded.
- **Raw channel references in headers crossing brokers.** Use `HeaderChannelRegistry`-equivalent string IDs for `REPLY_CHANNEL` / `ERROR_CHANNEL` at serialisation boundaries.
- **Inline assignment in business handlers** (yarumo convention: `if err := fn(); err != nil` is forbidden — must split). Applies in particular to handler bodies that consume / produce messages.

---

## 6. Overall recommendation

**PARTIAL** — adopt selected EIP machinery into `modules/messaging/` (§ 1.3 of the roadmap).

The core abstractions (`Message`, `Channel`, `Handler`, channel taxonomy, EIP endpoint catalogue) are the **uncontroversial base**. The **mechanics** identified in § 3 — router taxonomy, full aggregator surface, scatter-gather, poller knobs, request-handler advice, message / group / metadata / lock stores, idempotent receiver, claim check, wire-tap, observability conventions — are the actual differentiator between a "thin broker wrapper" and a real EIP layer.

Adopt:

- **Yes**: full channel taxonomy; router subtypes as separate structs; full aggregator surface (correlation × release × store × locks + expiration flags + partial release); scatter-gather as compound endpoint; poller config with all knobs; request-handler advice chain hooked into `common/resilience` and `modules/cache`; pluggable `store/` interfaces; idempotent receiver; claim check; wire-tap and channel interceptors; observability via OTel with W3C trace-context propagation; file and HTTP adapters; Postgres-push channel; integration graph snapshot; message history header.
- **Selectively**: handler chain (ergonomics — cheap), delayer (real use case in DaaS scheduled decisions), resequencer (if any Kafka consumer needs re-ordering), bridge (transport adapter; cheap), `FixedSubscriberChannel` (perf for chain links).
- **Defer**: WebSocket / STOMP adapters (no current consumer demand), barrier handler (niche), control bus (anti-pattern), reactive support (Go's `iter.Seq` covers realistic cases), SockJS, GraphQL inbound, RSocket, MQTT (file when a need surfaces).
- **Reject**: SpEL anywhere in routing / correlation / release / key extraction; annotation-driven endpoint discovery; XML namespace; `@MessagingGateway` proxy generation; `trapException=true` semantics; god-class application-context.

Implementation slicing — first useful subset is **core + channels + atomic endpoints + poller + HTTP adapter** (a minimum-viable in-process EIP layer with HTTP-bridged ingestion). Second slice adds **store + aggregator + idempotent receiver + claim check**. Postgres-push (`store/jdbc/postgres_notify.go`) is the highest-leverage add for DaaS specifically. Broker drivers (rabbitmq / kafka / nats) come last — per the roadmap, they are tagged "Low" priority.

---

## 7. Open questions

1. **Postgres-push channel: ship in `messaging/store/jdbc/` or in a separate `messaging/postgres/` subpackage?** Blurs the line between "store" (passive state) and "channel" (active dispatch). Recommendation: keep in `store/jdbc/` with a `Subscribe()` method on the message store; the channel impl is a thin wrapper. Avoids a separate package with only ~200 lines.

2. **Should `messaging/events/` (roadmap) reuse the full EIP machinery or be a flat `Publisher[T] / Subscribe[T](fn)` façade?** The roadmap leans toward "thin façade over `DirectChannel`." Question: does it accept advice chains? Recommendation: **no** — keep `events/` as the simple façade; users who want retry / circuit-breaker for domain events switch to the full `endpoints/activator.go` API. Two clear surface levels.

3. **Wire-tap as channel interceptor or as endpoint?** Spring exposes both flavours (`.wireTap()` DSL method + `WireTap` interceptor). Recommendation: **interceptor only**. DSL method is cosmetic; interceptor is sufficient.

4. **Sequence headers (`SEQUENCE_NUMBER`, `SEQUENCE_SIZE`) — header constants or typed `SequenceDetails` value type?** Spring uses headers (string-keyed map). Yarumo's `Message[T]` could enforce typed access via `msg.Sequence() (n, total int, ok bool)`. Recommendation: **typed accessor on `Message`** plus header backing for serialisation (broker-agnostic). The accessor reads / writes the headers.

5. **`MESSAGE_HISTORY` header — slice of structs in header (serialisation issue at broker hop) or write-only span?** Spring stores a `MessageHistory` list as a header. For cross-broker durability, serialise to portable format (JSON / protobuf). Recommendation: store as **`[]HistoryEntry`** struct slice; serialise via the broker driver's encoder (JSON for Kafka, AMQP table for RabbitMQ).

6. **Aggregator group-store reaper — separate goroutine per aggregator, or shared reaper?** Spring has `MessageGroupStoreReaper` that scans periodically. Per-aggregator simplifies lifecycle but multiplies goroutines on a busy app; shared centralises but couples aggregators. Recommendation: **shared reaper bean**, registered once per store, with `Register(aggregator)` API. Single goroutine per store, regardless of aggregator count.

7. **Should `messaging.Channel` be generic `Channel[T]` or `Channel` over `Message`?** Generic over `T` would be more type-safe but limits a channel to one payload type — kills routing flexibility (a router by definition mixes payload types). Recommendation: **`Channel` over `Message`** (untyped at channel level) + `Message[T]` for payload typing + a thin generic helper `Send[T](ch, payload, headers)` for producer-side ergonomics. Matches Spring's `Message<?>` channels.

8. **Idempotent receiver: where does the `MetadataStore` TTL live?** Spring's `MetadataStore` is untimed — entries live forever unless explicitly removed. In production you need TTL to bound storage. Recommendation: add `TTLMetadataStore` capability (optional; implementations can no-op TTL if backend doesn't support it). Redis / Mongo support TTL natively; JDBC needs a sweeper.

9. **HTTP-inbound endpoint registration — does it own the route or share with `managed/server_http`?** Owning is simpler but creates two competing route-registration paths. Sharing is cleaner but requires messaging to know about gin (or whichever router). Recommendation: **share** — `messaging/http/inbound.go` takes a `gin.IRoutes` (or `*http.ServeMux`) and registers handlers; doesn't own the server. Composes with `tools/routegen/` (§ 2.1 of the roadmap) cleanly.

10. **Recovery callback message shape — `ErrorMessage(cause, originalMessage)` or `Message[error]` with original as header?** Spring uses `ErrorMessage(Throwable payload, Message originalMessage)` accessed via `errMsg.GetOriginalMessage()`. In Go, a typed `ErrorMessage` struct mirroring this is clearer than overloading `Message[error]`. Recommendation: **dedicated `ErrorMessage` type** implementing `Message`, with `Payload() error` and `OriginalMessage() Message`. Errors are first-class.

11. **Schema registry placement — `messaging/schema/` or top-level `modules/schema/`?** Roadmap places it inside messaging. It's always used alongside a broker (never standalone), so keep inside messaging. If a non-messaging consumer appears (e.g. event-sourcing without a broker), promote to top-level then.

12. **Per-channel `errorChannel` resolution — by name (string) or by reference (Channel pointer)?** Spring uses names + a registry. References force eager wiring. Recommendation: **by name** with lazy resolution at first send, so error channels can be late-bound (and renamed in tests).

13. **`PartitionedChannel` semantics — fixed partitions at construction, or dynamic resize?** Spring fixes them. Recommendation: **fixed** — dynamic resize requires rebalancing logic that's out of scope. Users who need elasticity should compose with a Kafka consumer group instead.

---

## 8. Roadmap delta proposed (NOT applied — user decides)

If user wants to promote any of § 4 into `ROADMAP_NEW_MODULES.md` § 1.3, the suggested additions to the `modules/messaging/` layout in that file:

```
modules/messaging/
  message.go            Message[T], Headers, Builder       [already listed]
  headers.go            Standard header constants          [ADD]
  channel.go            Channel interface (Send/Receive)   [already listed]
  handler.go            MessageHandler = func(Message)     [already listed]
  source.go             MessageSource[T]                   [ADD]
  registry.go           Name → channel/endpoint lookup,    [ADD]
                        Graph() snapshot

  channels/
    direct.go           DirectChannel                      [already listed]
    queue.go            QueueChannel                       [already listed]
    pubsub.go           PubSubChannel                      [already listed]
    priority.go         PriorityChannel                    [already listed]
    rendezvous.go       RendezvousChannel                  [already listed]
    executor.go         ExecutorChannel                    [ADD]
    partitioned.go      PartitionedChannel                 [ADD]
    fixed_subscriber.go FixedSubscriberChannel             [ADD]
    null.go             NullChannel                        [ADD]

  interceptors/         ChannelInterceptor + WireTap +     [ADD]
                        SelectingInterceptor + History +
                        global pattern-based registration

  endpoints/            EIP components — refined           [refine bullet]
    transformer.go, filter.go, activator.go, splitter.go,
    bridge.go, chain.go, delayer.go, barrier.go,
    resequencer.go, scatter_gather.go, idempotent.go
    routers/            Six router subtypes              [ADD subpackage]
    aggregator/         Full aggregator surface          [ADD subpackage]

  poller/               Poller + Trigger + PollingConsumer [ADD]
  advice/               HandlerAdvice chain hooked into    [ADD]
                        common/resilience + modules/cache
  store/                MessageStore, MessageGroupStore,   [ADD]
                        MetadataStore, LockRegistry
                        + inmemory/ + jdbc/ + redis/ + mongo/
                        + jdbc/postgres_notify.go
  claimcheck/           Claim-check transformers           [ADD]
  observability/        OTel semconv + ObservationInterceptor [ADD]
  file/                 File adapter (poll dir + write +   [ADD]
                        tail + persistent filter)
  http/                 HTTP inbound + outbound + mapper   [ADD]

  schema/               Schema Registry client             [already listed]
  events/               Nominal-typed pub/sub façade       [already listed]
  rabbitmq/             AMQP + Streams drivers             [already listed]
  kafka/                + cdc/ Debezium                    [already listed]
  nats/                 future                             [already listed]
```

When the refinements land in the roadmap:

1. Replace the current `endpoints/` bullet (single line) with the subpackage breakdown (`routers/`, `aggregator/`, plus standalone files).
2. Add bullets for `store/`, `poller/`, `advice/`, `interceptors/`, `claimcheck/`, `observability/`, `file/`, `http/`.
3. Cross-reference this document (`docs/spring/spring-integration.md`) as the Spring Integration source-of-truth for the design.

Cross-module dependencies to wire (in order):

| Hook | From | To | Mechanism |
|---|---|---|---|
| Resilience | `messaging/advice/` | `common/resilience/` | `CircuitBreakerRegistry.Get(name)`, `RateLimiterRegistry.Get(name)` (already shipped, YA-0076) |
| Cache | `messaging/advice/cache.go` | `modules/cache/` | `Cache[K,V]` interface |
| JDBC store | `messaging/store/jdbc/` | `modules/datasource/gorm/` | `*gorm.DB` injection |
| Redis store | `messaging/store/redis/` | `modules/datasource/goredis/` | `redis.UniversalClient` injection |
| Mongo store | `messaging/store/mongo/` | `modules/datasource/mongo/` | `*mongo.Client` injection |
| Lifecycle | `messaging/poller/` + every long-running endpoint | `modules/managed/` | implements `Lifecycle` (Start/Stop/Done) |
| Observability | `messaging/observability/` | `modules/telemetry/otel/` | `TracerProvider` / `MeterProvider` |
| HTTP transport | `messaging/http/inbound.go` | `modules/managed/server_http` | shares the gin engine; messaging registers routes |
| Schema validation | `messaging/schema/` | `messaging/kafka/`, `messaging/rabbitmq/` | optional pre-publish / post-consume validation hook |
| CDC | `messaging/kafka/cdc/` | `messaging/kafka/` | wraps the Kafka driver's `Source` |

Ticket ordering (first useful slice highlighted in bold):

1. **Core**: `message.go`, `headers.go`, `channel.go`, `handler.go`, `source.go`, `registry.go`.
2. **Channels**: each channel impl in `channels/`.
3. **Store interfaces + inmemory**: `store/MessageStore`, `MessageGroupStore`, `MetadataStore`, `LockRegistry` + `store/inmemory/*`. **Blocks aggregator, idempotent, claim-check.**
4. **Interceptors**: `interceptors/` package — wire-tap, selecting, history, global pattern.
5. **Endpoints — atomic**: `Transformer`, `Filter`, `ServiceActivator`, `Splitter`, `Bridge`, `Chain`, `Routers/*`.
6. **Aggregator subpackage**: depends on Store.
7. **Resequencer**: depends on Store.
8. **Scatter-Gather**: depends on RecipientListRouter + Aggregator.
9. **Idempotent Receiver advice**: depends on MetadataStore.
10. **Poller**: `poller/` package — independent of brokers.
11. **Advice**: `advice/*` — depends on `common/resilience` and `modules/cache`.
12. **Claim check**: thin layer over MessageStore.
13. **File adapter**: `file/` — depends on Poller and MetadataStore.
14. **HTTP adapter**: `http/` — composes with `managed/server_http`.
15. **Observability**: `observability/` — depends on `telemetry/otel`.
16. **Store backends**: `store/jdbc/`, `store/redis/`, `store/mongo/` — depend on respective `datasource/*` drivers (Phase 3+ in yarumo's plan).
17. **Postgres-push channel**: `store/jdbc/postgres_notify.go` — depends on `datasource/gorm` and `INT_CHANNEL_MESSAGE` schema. **Highest leverage for DaaS.**
18. **Broker drivers**: rabbitmq/amqp, rabbitmq/streams, kafka, nats — roadmap-listed, lowest priority per § 4.1 of `ROADMAP_NEW_MODULES.md`.
19. **`events/` façade**: thin wrapper over `DirectChannel` + nominal types.
20. **`schema/` registry client**: standalone, can land independently.

Minimum-viable EIP layer (no broker drivers) = items **1 → 5 → 10 → 14**. Sufficient for in-process flows and HTTP-bridged ingestion.

Storage-backed flows ("filesystem-driven" pipelines, idempotent at-least-once) = add items **3, 6, 9, 13**.

DaaS-Postgres-best-bet = add item **17** (Postgres-push channel) — replaces "would need Kafka" with "Postgres alone is enough".

Ordering rationale: store is the load-bearing abstraction; every advanced endpoint depends on it.
