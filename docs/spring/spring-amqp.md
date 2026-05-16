# Spring AMQP — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-amqp
> **Analyzed**: 2026-05-16
> **Recommendation**: PARTIAL

## 1. Project summary

Spring AMQP 4.0.3 (stable; 4.1.0-RC1 preview) is Spring's RabbitMQ abstraction: `RabbitTemplate` for publishing, listener containers (`SimpleMessageListenerContainer`, `DirectMessageListenerContainer`, `StreamListenerContainer`) for consuming, `RabbitAdmin` for declarative topology (exchanges/queues/bindings), plus integrations with `spring-retry`, Micrometer, and the RabbitMQ Stream plugin. Scope is purely RabbitMQ — there is no broker-agnostic API. JVM coupling is **high** at the framework level (Spring beans, `@RabbitListener`, SpEL, `BeanPostProcessor`s) but **low at the conceptual level** (template + listener container + topology declaration + retry interceptor + DLX patterns map cleanly to Go).

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | `RabbitTemplate.send/convertAndSend` | Publish API hiding channel/connection management, with default exchange/routingKey, message-properties builder, and a `MessageConverter` hook. | Every Go producer reimplements this on top of `amqp091-go`'s raw `Channel.PublishWithContext`. Single typed primitive removes the boilerplate. |
| 2 | **Publisher Confirms + Returns + mandatory** | Async confirm callbacks per-message via `CorrelationData`; returns when `mandatory=true` and no queue is bound. | Without confirms, "publish" is fire-and-forget — you don't know the broker accepted it. Required for at-least-once. amqp091-go exposes the primitive (`Confirm`, `NotifyPublish`, `NotifyReturn`) but the bookkeeping is non-trivial. |
| 3 | **Listener container with concurrency + prefetch + ack modes** | `SimpleMessageListenerContainer` (thread-pool, dynamic scaling `concurrentConsumers..maxConcurrentConsumers`) and `DirectMessageListenerContainer` (consumer-thread direct). `prefetchCount`, `AcknowledgeMode.{NONE,AUTO,MANUAL}`, `globalQos`. | This is the single most error-prone piece of any RabbitMQ consumer in Go: spinning N goroutines, each owning a channel, with QoS prefetch, manual ack/nack, and graceful drain on shutdown. |
| 4 | **Declarative topology** — `Exchange`, `Queue`, `Binding`, `QueueBuilder`, `ExchangeBuilder`, `BindingBuilder` | Fluent builders for direct/topic/fanout/headers/delayed exchanges and queues with TTL, DLX args, overflow, lazy, quorum/stream type, max-length, single-active-consumer. `RabbitAdmin` declares on connect and re-declares on recovery. | Go consumers declare via raw `Channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)` — positional args, magic `amqp.Table`. Builder + re-declare on reconnect are essential. |
| 5 | **Dead-Letter Exchange / DLQ pattern** | Queues with `x-dead-letter-exchange` / `x-dead-letter-routing-key`; recoverer publishes to DLX after retry exhaustion with `x-death` / `x-exception-*` headers. | Canonical "poison message" handling. Every production RabbitMQ deployment needs it; ad-hoc reimplementation is error-prone. |
| 6 | **Retry interceptor with backoff + recoverer** | `RetryInterceptorBuilder.stateless()/.stateful()` with `maxAttempts`, exponential `backOffOptions(initial, multiplier, max)`, and a pluggable `MessageRecoverer` (`RejectAndDontRequeueRecoverer`, `RepublishMessageRecoverer`). Includes retry-count header pattern for client-side retry. | The synchronous "retry N times in-process before DLQ" pattern is what most Go services need. `cenkalti/backoff` covers the math; the message-shaped wrapping is what's missing. |
| 7 | **Connection recovery** | `CachingConnectionFactory` reconnects automatically, listeners re-subscribe, `RabbitAdmin` re-declares topology, in-flight unacked messages are redelivered by the broker. | amqp091-go has `NotifyClose` and recovery patterns, but no built-in supervisor. This is the second-most-error-prone piece after the listener container. |
| 8 | **Micrometer observation hooks** | `template.setObservationEnabled(true)` + `container.setObservationEnabled(true)` emit standard `spring.rabbit.template` / `spring.rabbit.listener` spans + timers; OTel-compatible via Micrometer's OTel bridge. | OTel messaging semantic conventions (`messaging.system=rabbitmq`, `messaging.destination`, `messaging.rabbitmq.routing_key`) are the right target — yarumo already has `telemetry/otel`. |
| 9 | **Stream plugin support** (`RabbitStreamTemplate`, `StreamListenerContainer`, `SuperStream`) | Separate `spring-rabbit-stream` artifact wrapping `rabbitmq-stream-java-client`. Offset specification (`first/last/timestamp/offset`), manual offset tracking, super-streams with partitioning. | Streams have **different semantics** (offset-based, append-only, partitioned) and a different wire protocol — they belong in a **sibling driver**, not the AMQP one. Validates the planned `messaging/rabbitmq/streams/` separation. |
| 10 | **Message conversion** (`Jackson2JsonMessageConverter`) | Body ↔ struct via a pluggable `MessageConverter`; content-type negotiated via `content_type` header. | Pluggable encoder/decoder is essential; consumers should not deal with `[]byte` for JSON payloads. |

## 3. Long-tail features (skip)

- **`@RabbitListener` annotation** + `BeanPostProcessor` registration — replaced by explicit construction in Go.
- **SpEL** in `mandatoryExpression`, recoverer routing — replaced by Go closures.
- **XML schema config** (`<rabbit:queue/>`, `<rabbit:listener-container/>`) — JVM-only.
- **AMQP 1.0 support** — RabbitMQ-side is preview; AMQP 1.0 vs 0.9.1 is a different driver entirely.
- **`AMQPAppender` for logging frameworks** — yarumo `log/slog` does not need a RabbitMQ sink.
- **Native image / GraalVM hints** — N/A in Go.
- **`BatchingRabbitTemplate`** with in-memory client-side batching — defer until concrete demand (batching loses messages on crash; most consumers don't want this).
- **Polling consumer** (`RabbitTemplate.receive(queue, timeout)`) — niche pattern; the listener-container model covers 99%.
- **`Switch User` / impersonation, RememberMe** — not AMQP-related; mis-pasted from security; N/A.
- **`StatefulRetryOperationsInterceptor`** (the stateful variant requiring `MessageId`) — the stateless variant is sufficient for v1.
- **Delayed Message Exchange** — broker plugin; declare via `args` if needed, no special API.
- **`AnonymousQueue` naming strategies** (Base64Url, UUID) — broker-generated names suffice.
- **`AuthorizationProxyFactory`-style proxies on returned objects** — N/A.
- **Conditional declaration across multiple admins** — single-broker is enough for v1.
- **JMX / multi-broker management endpoints** — out of scope.

## 4. Mapping to Yarumo

**Existing/planned modules with overlap**: `modules/messaging/rabbitmq/amqp/` and `modules/messaging/rabbitmq/streams/` are both explicitly planned in **§ 1.4** of `docs/ROADMAP_NEW_MODULES.md`. The split (amqp vs streams) directly mirrors Spring's split between `spring-rabbit` and `spring-rabbit-stream` and is validated by this analysis — they are different wire protocols and different semantic models (queue vs log). The driver also plugs into the planned `messaging/` core (`Message[T]`, `Channel`, `MessageHandler`, `endpoints/`) as the AMQP-backed `Channel` implementation.

**Gaps this could fill**:

- **Topology declaration as data** — Go consumers usually call `QueueDeclare`/`ExchangeDeclare` imperatively. Declarative `Topology` value plus a `Declare(ctx, topology)` admin that re-runs on reconnect is the single largest ergonomic win.
- **Listener container abstraction** — currently nothing in yarumo owns "N goroutines, each with a channel, with prefetch QoS, ack/nack, graceful drain." This is what `Consumer` should be.
- **Retry + DLX pattern as one unit** — `WithRetry(maxAttempts, backoff, recoverer)` decorator that composes `common/resilience` retry with a recoverer that either republishes to a DLX or `Nack(requeue=false)` to trigger broker-side `x-dead-letter-exchange`.
- **Publisher confirms wrapper** — `Publisher.Publish(ctx, msg) (ack <-chan bool, err error)` (or sync `PublishConfirm`) layered on top of `amqp091-go`'s `NotifyPublish`.
- **Connection supervisor** — `Connection` as a `managed.Component` with lifecycle `Start/Stop/Done`, owning reconnect + re-declaration + listener re-subscription.

**Anti-patterns to avoid** (from § 1.4 placement principle and Spring's failure modes):

- **No DI container / no annotation magic** — no `@RabbitListener` equivalent, no `BeanPostProcessor`. Consumers are constructed with `NewConsumer(conn, queue, handler, opts...)`.
- **No SpEL** — recoverer routing keys are `func(*Message, error) (exchange, routingKey string)` closures.
- **No god-struct `CachingConnectionFactory`** — `Connection` is one struct; cache mode `CHANNEL` is the only supported semantic (channel-per-consumer); `CHANNEL+CONNECTION` complexity is dropped.
- **No XML / no `Declarables` collection sugar** — topology is a Go value; one struct, one builder, done.
- **No `AmqpTemplate` interface beside `RabbitTemplate`** — Spring has both because of JMS legacy. One concrete `Publisher` type.
- **No publisher-side batching** for v1 — `BatchingRabbitTemplate` loses messages on crash; defer until demand.

## 5. Recommendation

**PARTIAL** — adopt the conceptual model (template + listener container + topology declarations + retry/DLX + confirms + observation) but rewrite idiomatically in Go on top of `amqp091-go`. The Spring API surface is far larger than what most consumers need; Pareto gives ~10 features that cover 95% of cases. The planned split in § 1.4 (`messaging/rabbitmq/amqp/` and `messaging/rabbitmq/streams/`) is validated and should proceed. Concrete refinements to § 1.4: (a) name the AMQP top-level types `Connection`, `Publisher`, `Consumer`, `Admin` (drop the `Rabbit-` prefix — the package path already says it); (b) treat `messaging/rabbitmq/streams/` as a fully separate driver, not a sub-mode of amqp (different protocol, different client library); (c) the AMQP driver implements `messaging.Channel` from the core layer, so EIP endpoints (`Transformer`/`Filter`/`Router`) compose with it without special-casing RabbitMQ.

## 6. Proposed yarumo placement

**Module**: `modules/messaging/rabbitmq/amqp/` (sub-module of § 1.4)

**Subpackages**:

```
modules/messaging/rabbitmq/amqp/
  doc.go
  errors.go               Domain errors (errs.TypedError pattern): ErrConnectionClosed,
                          ErrPublishNack, ErrPublishReturned, ErrConsumerCancelled,
                          ErrTopologyDeclare, ErrAckFailed.
  connection.go           Connection (managed.Component): Dial, Start, Stop, Done,
                          NotifyClose. Owns reconnect supervisor + channel pool.
  publisher.go            Publisher: Publish(ctx, exchange, routingKey, msg) error,
                          PublishConfirm(ctx, ...) error (waits for ack), Returns()
                          <-chan Return. Mandatory flag per-publish.
  consumer.go             Consumer (managed.Component): N goroutines, each with own
                          channel + QoS prefetch, AckMode {None,Auto,Manual},
                          graceful drain on Stop. Implements messaging.Channel
                          (subscribe side).
  admin.go                Admin: Declare(ctx, Topology) error. Re-runs on reconnect
                          via Connection's listener hook.
  topology/
    topology.go           Topology = {Exchanges, Queues, Bindings}. Value type.
    exchange.go           ExchangeBuilder: Direct/Topic/Fanout/Headers/Delayed,
                          .Durable(), .AutoDelete(), .Internal(), .Arg(k, v),
                          .Alternate(name).
    queue.go              QueueBuilder: .Durable(), .Exclusive(), .AutoDelete(),
                          .TTL(d), .Expires(d), .MaxLength(n), .MaxLengthBytes(n),
                          .Overflow(strategy), .DLX(exchange), .DLRK(key),
                          .MaxPriority(n), .Lazy(), .Quorum(), .Stream(),
                          .SingleActiveConsumer(), .Arg(k, v).
    binding.go            BindingBuilder: Bind(queue).To(exchange).With(key).Arg(k,v).
  message.go              Message envelope (compatible with messaging.Message[T]):
                          Body []byte, Headers map[string]any, ContentType,
                          MessageID, CorrelationID, ReplyTo, DeliveryMode, Priority,
                          Expiration, Timestamp, UserID, AppID.
  conversion/             Pluggable Converter[T] interface.
    json.go               JSONConverter[T] (encoding/json, content-type
                          application/json). Default.
    proto.go              ProtoConverter[T] (google.golang.org/protobuf). Optional.
  retry/                  Retry decorator over Consumer handler.
    retry.go              WithRetry(handler, maxAttempts, backoff.Backoff,
                          Recoverer) — composes with common/resilience.
    recoverer.go          Recoverer interface; RejectAndDontRequeueRecoverer
                          (broker-side DLX via Nack(false, false)),
                          RepublishMessageRecoverer (publisher.Publish to error
                          exchange with x-exception headers).
  observation/            Optional OTel instrumentation.
    otel.go               WithOTel(opts...) wraps Publisher/Consumer with spans +
                          messaging.* semantic-convention attributes. Imported only
                          if opted in (no hard dep from core driver).
  internal/
    channel_pool.go       Channel allocation, lifecycle, recovery.
    supervisor.go         Reconnect loop, exponential backoff, re-declare hook.
```

**Internal deps**:

- `modules/common/errs` — `TypedError` pattern for domain errors.
- `modules/common/resilience` — `Retrier` + backoff for the retry decorator.
- `modules/common/log/slog` — structured logging.
- `modules/common/assert` — struct invariant checks (nil receiver).
- `modules/managed` — `Component` lifecycle for `Connection` and `Consumer`.
- `modules/messaging/` (core, § 1.4) — `Message[T]`, `Channel` interface implementation.
- `modules/telemetry/otel` — optional, only via `observation/` sub-package.

**Go libraries to wrap** (mature, with repo URL):

- `github.com/rabbitmq/amqp091-go` — official RabbitMQ Go client for the 0.9.1 protocol. Maintained by the RabbitMQ team. The only sane choice; everything else (e.g. streadway/amqp) is deprecated in favor of it.
- `github.com/rabbitmq/rabbitmq-stream-go-client` — official Go client for the stream plugin; wrapped by the sibling `messaging/rabbitmq/streams/` driver, not by this one.
- (already wrapped indirectly) `cenkalti/backoff/v4` via `common/resilience` for retry math.

**Out of scope for v1**:

- AMQP 1.0 protocol support (preview in RabbitMQ; different wire format).
- `BatchingPublisher` (in-memory client-side batching) — loses messages on crash; defer until concrete demand.
- Polling consumer (`Receive(queue, timeout)`) — listener-container model is sufficient.
- Stateful retry (the `MessageId`-keyed variant) — stateless retry covers 95% of cases.
- Delayed Message Exchange API — declare via raw `Arg("x-delayed-type", "direct")` if the plugin is installed; no special wrapper.
- Multi-broker / conditional admin / per-admin queue scoping — single connection per `Connection` instance for v1.
- `RabbitListenerContainerFactory` equivalent — explicit construction is enough; no factory factory.
- Logging framework appender (slog → RabbitMQ sink).
- `AnonymousQueue` naming strategies — broker-generated names suffice; consumers can pass `""` for the broker default.

## 7. Open questions

- **Listener container model**: SMLC (thread-pool, dynamic scale `m..n`) vs DMLC (one goroutine per consumer)? Go's goroutines are cheap enough that DMLC-style (one goroutine per concurrent consumer, fixed count) is the natural default. Should `concurrency: m..n` dynamic scaling exist at all in v1? Recommend **no** — fixed `concurrency: n` only, revisit if a real use case appears.
- **Publisher confirms API**: synchronous (`PublishConfirm(ctx, ...) error` blocking until ack) vs async (`Publish(ctx, ...) (<-chan Confirmation, error)`)? Stripe-style sync is simpler; async is more efficient for high-throughput. Default sync, expose async via separate method?
- **DLX strategy default**: broker-side via `x-dead-letter-exchange` queue arg (simpler, no extra Publisher) vs application-side via `RepublishMessageRecoverer` (more control, custom headers)? Spring offers both. Recommend broker-side as the default with `RepublishMessageRecoverer` as an explicit opt-in.
- **Where do EIP endpoints live?** If `messaging/endpoints/` (Transformer/Filter/Router/Splitter/Aggregator) only operates on `messaging.Channel`, the AMQP driver gets them "for free" by implementing that interface. Confirms § 1.4's design intent — but worth verifying with a sketch before committing.
- **Reconnect supervisor inside `Connection`, or as a separate `managed/` primitive?** Other drivers (kafka, nats) will need the same pattern. Possible refactor: `managed/supervisor.go` with reconnect+redeclare hooks, used by all `messaging/<broker>/*` drivers. Defer until at least two drivers exist.
- **Topology re-declaration on reconnect**: should `Admin.Declare` register a `ConnectionListener` automatically (Spring-style auto-re-declare), or should the consumer re-call `Declare()` on reconnect events? Auto seems safer; explicit is more honest. Recommend auto with an opt-out flag.
- **Stream driver scope**: should `messaging/rabbitmq/streams/` also implement `messaging.Channel`, or expose a distinct offset-aware API? Stream semantics (offsets, replay, partitions) don't fit cleanly behind a `Channel` that hides them. Recommend a sibling interface (`messaging.Log` or similar) — out of scope for this evaluation but worth flagging for § 1.4's design.
