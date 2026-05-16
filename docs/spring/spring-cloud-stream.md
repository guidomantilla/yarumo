# Spring Cloud Stream — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-stream
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup; previous version referenced the now-deleted Annex A)
> **Recommendation**: PARTIAL

## 1. Project summary

Spring Cloud Stream sits **on top of** Spring Messaging + Spring Integration (see [`spring-integration.md`](spring-integration.md) for the underlying EIP layer) and contributes one big idea: **the Binder** — a broker-side adapter that turns plain `java.util.function.{Function,Consumer,Supplier}` beans into running consumers/producers against Kafka / RabbitMQ / Pulsar / Kinesis / etc. without the user touching the broker SDK.

Three layered abstractions:

| Layer | What it does |
|---|---|
| **Destination** | The external name on the broker (Kafka topic, AMQP exchange). |
| **Binding** | The bridge between an in-app function input/output and a destination. Implicit names: `<fn>-in-<idx>` / `<fn>-out-<idx>`. Carries all the runtime policy (group, content-type, DLQ, retry, partition). |
| **Binder** | The pluggable component that knows how to subscribe/publish on a specific broker. Selected by classpath artifact (`spring-cloud-stream-binder-kafka`, `-rabbit`, …) and configurable per-binding via `defaultBinder` / `binder` properties. Exposes `bindProducer(name, channel, props)` and `bindConsumer(name, group, channel, props)`. |

The function the user writes is **business logic only**. The framework supplies:

- consumer groups (durable competing consumers, distributed across instances)
- partitioning (producer-side key extraction + consumer-side instance routing — Kafka native, RabbitMQ simulated via queue-per-partition)
- content-type negotiation (`application/json` ↔ struct via `MessageConverter`)
- error handling (per-binding error channel `<destination>.<group>.errors`, `RetryTemplate`, DLQ — broker-specific materialisation)
- batch consumers / batch producers
- polling consumers (sync `Receive()` with manual ack)
- reactive support (`Flux<I> → Flux<O>`)
- functional composition (`fn1|fn2|fn3` via `spring.cloud.function.definition`)
- dynamic destinations via `StreamBridge.send(destination, payload)`
- header-driven routing (`spring.cloud.stream.sendto.destination`)
- schema-registry integration (Confluent / AWS Glue / Apicurio for Avro/Protobuf)
- a **Kafka Streams binder** that swaps the abstraction from messages to `KStream`/`KTable`/`GlobalKTable` with native windowing, joins, branching, interactive queries.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **Binder SPI** | `bindProducer` / `bindConsumer` decouples function from broker. Same `Function<I,O>` runs on Kafka or RabbitMQ by swapping the binder artifact. Selection via `defaultBinder` or per-binding `binder=` property. | Yarumo's planned `messaging/` already has driver-as-`Channel` — but the **input/output binding** layer (config-driven destination + group + content-type + DLQ) is the Stream-specific glue worth borrowing as a uniform API. |
| 2 | **Consumer group as a first-class binding property** | `spring.cloud.stream.bindings.<x>.group=myGroup`, broker-neutral. Kafka maps to consumer group; RabbitMQ maps to a named durable queue bound to the destination exchange. | Every Go consumer reimplements this with broker-specific gotchas. The **binding-level** consumer-group concept normalises the two brokers' wildly different topologies. |
| 3 | **DLQ as a binding property** | `enableDlq=true` (Kafka) / `autoBindDlq=true` (Rabbit) materialises the DLQ topic/queue automatically. Failed payloads carry `x-original-topic`, `x-exception-message`, `x-exception-stacktrace` headers. Group is required for DLQ naming (`<dest>.<group>.dlq`). | The DLQ pattern is universal; Go shops reimplement it per broker and lose the exception metadata. A normalised `DLQ` policy on the binding is high value. |
| 4 | **Retry template at the binding level** | `maxAttempts` (default 3), `backOffInitialInterval` (1000 ms), `backOffMultiplier` (2.0), `backOffMaxInterval` (10000 ms), `retryableExceptions` map — declarative per-binding. After exhaustion, the message lands in DLQ (if enabled) or the error channel. | Most services want "retry N with backoff, then DLQ". A binding-level policy avoids hand-rolling. Plays directly with yarumo's already-shipped `common/resilience` ([YA-0076](https://github.com/guidomantilla/yarumo/issues/76)). |
| 5 | **Per-binding error channel + custom error handler** | `<destination>.<group>.errors` is a real Spring Integration channel; alternatively register `Consumer<ErrorMessage>` and point `error-handler-definition` at it. | Decouples error-handling logic from happy-path logic. Maps cleanly to a Go `OnError func(ctx, msg, err)` callback on the binding. |
| 6 | **Partitioning normalised across brokers** | Producer extracts `partitionKeyExpression` (SpEL); consumer declares `instanceIndex`/`instanceCount`. Kafka uses native partitions; Rabbit simulates with `<dest>-<partition>` queues. | Partitioning is essential for ordered/stateful processing (sticky-key consumers). Spring's normalisation is genuinely useful — Rabbit-native partition simulation is not obvious and worth porting. |
| 7 | **Content-type negotiation** | `Content-Type: application/json` header drives payload conversion in/out of the function. Native Kafka serializers can be opted into via `useNativeEncoding` / `useNativeDecoding`. | Go consumers handle `[]byte`-to-struct conversion in every handler. A `MessageConverter` hook (JSON default, plug your own for Avro/Protobuf) is the right abstraction. |
| 8 | **`StreamBridge.send(destination, payload)`** | Dynamic-destination producer without pre-declared bindings. Caches up to N destinations. Used heavily by webhook-style / routing-style flows. | Maps directly to "send to a destination computed at runtime" — a common Go pattern in outbox / webhook workers. The planned `Publisher.Send(dest, msg)` in `messaging/` already covers this shape. |
| 9 | **`requiredGroups` on the producer** | Producer pre-declares which consumer groups must be materialised so that messages published before any consumer starts are not lost. | Solves the "publish before subscribe" gap declaratively. In a broker-neutral binding it's a simple list of groups to pre-create. |
| 10 | **Schema-registry integration** | `kafka-avro-serializer` + `schema.registry.url` + `subject.name.strategy`. Pluggable for Confluent / AWS Glue / Apicurio. | Avro/Protobuf payloads with a schema registry are mainstream in CDC / event-sourcing pipelines. Validates the planned `messaging/schema/` sub-module from § 1.3 of the roadmap. |

## 3. What Spring Cloud Stream adds beyond Spring Messaging + Spring Integration

`spring-integration.md` already covers the EIP base — `Message<T>`, `MessageChannel`, `MessageHandler`, EIP endpoints (Transformer / Filter / Router / Splitter / Aggregator / Service Activator), channel adapters, the Java DSL, request-handler advices, message stores, and observability. **Cloud Stream is not a replacement** for that layer; it's a deployment / packaging wrapper on top of it. The net contributions over the EIP layer are:

| Cloud Stream contribution | Present in Spring Integration? | Verdict for yarumo |
|---|---|---|
| **Binder SPI** — a single SPI (`bindProducer` / `bindConsumer`) that broker drivers implement so a function can be wired to any broker by swapping the artifact. | No. Spring Integration has **channel adapters** (per-protocol code) but no SPI to swap brokers behind the same in-app abstraction. | **Net add**. The "uniform driver shape" is the most portable idea in Stream. |
| **Binding** — a config struct that names a destination, a group, a content-type, a DLQ flag, a retry policy, a partition policy and is materialised at boot. | Spring Integration wires endpoint → channel imperatively. No declarative deployment policy. | **Net add**. Binding-as-deployment-policy is the second most portable idea. |
| **Consumer group at the binding layer** | Per-driver only. | **Net add at the abstraction level** — normalisation across brokers is what's new. |
| **DLQ as a binding property** | Spring Integration models DLQ via routing + dead-letter exchange manually inside a flow. | **Net add** (declarative on/off; framework materialises the topology). |
| **`requiredGroups` (pre-declare consumer groups for a producer)** | No. | **Net add** — solves "publish before subscribe" without an extra config step. |
| **Schema-registry integration** | No — Spring Integration has converters but no first-class registry plug. | **Net add** (and matches the `messaging/schema/` plan in roadmap § 1.3). |
| **Functional programming model (`Function<I,O>`)** | Spring Integration endpoints are `MessageHandler`s. | **Convention shift, not a structural net add**. Trivial in Go with plain functions. |
| **Content-type negotiation at the boundary** | Partially — converters are pluggable in Spring Integration too. Stream makes it config-driven and broker-aware. | **Refinement, not net add.** |
| **`StreamBridge` dynamic destinations** | Spring Integration's Outbound Channel Adapter does this imperatively. | **No net add** — naming convenience only. |
| **Kafka Streams binder** | No — KStream is its own model. | **Net add inside Stream**, but **out of scope** for yarumo (Go has no equivalent runtime). |
| **Reactive functional model** (`Flux<I> → Flux<O>`) | No. | **Net add inside Stream**, but **N/A in Go** (goroutines + channels are the equivalent; no reactive layer needed). |

The portable kernel is the same five items from the previous analysis: **Binder + Binding + DLQ-as-property + consumer-group-at-binding + schema-registry**. Everything else either lives in the EIP layer already (see `spring-integration.md`) or is JVM-only sugar.

## 4. Mapping to Yarumo

### 4.a Existing planned modules with overlap

`modules/messaging/` (roadmap § 1.3, status **Planned**) already covers both layers in one module:

- `message.go` (`Message[T]`, `Headers`, `Builder`)
- `channel.go` (`Channel` interface: Send / Receive)
- `channels/` (direct, queue, pubsub, priority, rendezvous)
- `endpoints/` (Transformer, Filter, Router, Splitter, Aggregator, Activator)
- `schema/` (schema registry client — already planned)
- `events/` (nominal-typed pub/sub façade)
- `rabbitmq/amqp/`, `rabbitmq/streams/`, `kafka/` (drivers — already planned)
- `kafka/cdc/` (Debezium envelope parsing — already planned)

The drivers under `rabbitmq/` and `kafka/` **are binders in Stream's sense** — broker adapters implementing `Channel`. Yarumo's term is "driver". No new naming, but the **binding artefact between config and driver is missing** today.

### 4.b Gaps to fill

1. **`Binding` struct** — the single most useful piece of Stream to port. A typed, explicit struct that bridges a config payload to a running consumer/producer. Fields: `Destination`, `Group`, `ContentType`, `DLQ`, `Retry`, `Partition`, `RequiredGroups`, `OnError`. Drivers implement a `BindConsumer[T](b Binding, h Handler[T]) (Subscription, error)` / `BindProducer[T](b Binding) (Producer[T], error)` pair.
2. **`DLQPolicy`** struct on the binding — `Enabled`, `Name` (defaults to `<dest>.<group>.dlq`), `Partitions`, `MaxAttempts`. Drivers materialise: Kafka driver creates the DLT topic; AMQP driver wires `x-dead-letter-exchange`. Canonical `DLQHeaders` struct with `OriginalTopic`, `ExceptionMessage`, `ExceptionStacktrace`, `DeliveryAttempt`.
3. **`RetryPolicy`** struct on the binding — `MaxAttempts`, `InitialInterval`, `Multiplier`, `MaxInterval`, `Retryable func(error) bool`. Reuses `common/resilience/` ([YA-0076](https://github.com/guidomantilla/yarumo/issues/76)). After exhaustion, route to DLQ if `DLQPolicy.Enabled`, else to `OnError`.
4. **`PartitionPolicy`** struct on the binding — `KeyFn func(Message[T]) string`, `Count int`, `Index int`. Drivers implement: Kafka uses native partitions; AMQP simulates with `<dest>-<partition>` queues. Opt-in only (default off, like Stream).
5. **`MessageConverter`** interface — `Marshal(T) ([]byte, error)` / `Unmarshal([]byte) (T, error)` + a `ContentType()` method. JSON default; pluggable for Avro/Protobuf (the latter wires through `messaging/schema/`).
6. **`RequiredGroups []string` on producer bindings** — driver materialises consumer groups before the producer starts. Solves publish-before-subscribe loss.
7. **Driver responsibility doc** — written contract: implement `Channel` **and** honour `Binding` (DLQ / retry / partition / content-type). This is a docs change, not code.

### 4.c Anti-patterns to avoid

- **No bean-based extension points** (`@StreamListener`, `KafkaBindingRebalanceListener`, `ListenerContainerCustomizer`, `DeclarableCustomizer`). Replace with explicit `Option` funcs on the driver constructor.
- **No SpEL routing** (`spring.cloud.function.routing-expression=headers['type']`). Replace with `RoutingFn func(Message) string`.
- **No reflective binding discovery** (the magic `fn1|fn2|fn3` string). Compose functions in code.
- **No reactive wrapper** (`Flux<I> → Flux<O>`). Go has goroutines + channels.
- **No god-bag of properties.** A `Binding` is a typed struct; mis-named properties become compile errors, not runtime surprises.
- **No actuator-level binding pause/resume** at runtime (`BindingsLifecycleController`) — niche; defer until a concrete demand. The planned `managed.Lifecycle` Start/Stop already covers process-level lifecycle.

## 5. Recommendation

**PARTIAL** — adopt the **Binding + DLQ + Retry + Partition + consumer-group + RequiredGroups + content-type negotiation** normalisation shape (~30 % of Spring Cloud Stream's surface) directly into the already-planned `modules/messaging/` core. Reject the rest:

- **Kafka Streams binder** — re-implementation is enormous and Go has no good base (goka is limited). Confirmed rejection.
- **Reactive functional model** — JVM-only. N/A.
- **SpEL routing** — replace with Go closures.
- **Bean-based extension points / channel interceptors / customizer beans** — replace with explicit construction and `Option` funcs.
- **Cloud Function definition strings (`fn1|fn2|fn3`)** — only needed to disambiguate auto-discovery among multiple beans; with explicit construction in Go, the problem doesn't exist.
- **Pipeline orchestration (Spring Cloud Data Flow)** — out of scope; not even part of Stream proper.

The single net architectural contribution worth porting is the **Binding struct** as a normalisation layer above broker drivers. The current `modules/messaging/` plan (roadmap § 1.3) anticipates this in spirit — this analysis turns "spirit" into a concrete shape and a checklist of properties to surface.

## 6. Proposed yarumo placement

No new module. **All adoptable concepts land in the already-planned `modules/messaging/`** (roadmap § 1.3). Concretely:

```
modules/messaging/
  message.go        existing plan: Message[T], Headers, Builder
  channel.go        existing plan: Channel interface
  binding.go        NEW: Binding struct (Destination, Group, ContentType,
                         DLQ, Retry, Partition, RequiredGroups, OnError)
  converter.go      NEW: MessageConverter interface + JSON default
  dlq.go            NEW: DLQPolicy struct + DLQHeaders canonical fields
  retry.go          NEW: RetryPolicy struct (wraps common/resilience)
  partition.go      NEW: PartitionPolicy struct (KeyFn, Count, Index)
  channels/         existing plan: direct, queue, pubsub, priority, rendezvous
  endpoints/        existing plan: Transformer, Filter, Router, Splitter,
                                   Aggregator, Activator
  schema/           existing plan: Schema Registry (Confluent / Glue / Apicurio)
  events/           existing plan: nominal-typed pub/sub façade
  rabbitmq/
    amqp/           existing plan: AMQP driver — implements Binding via
                                   TopicExchange + queue + DLX
    streams/        existing plan: Streams driver
  kafka/
    cdc/            existing plan: Debezium CDC envelope parsing
                    NEW: Kafka driver — implements Binding via topic +
                         consumer-group + DLT
  nats/             existing plan: future
```

The driver contract becomes "implement `Channel` **and** honour a `Binding`". DLQ, retry, partition and content-type live on the `Binding`, not on the `Channel`, because they are **deployment policies, not transport semantics**. The transport-level EIP layer (channels + endpoints + adapters) is documented separately in [`spring-integration.md`](spring-integration.md).

No new ticket prerequisites — `common/resilience/` is already shipped, and `messaging/` is still **Planned** in § 1.3 so the binding shape can fold into the initial design without rework.

## 7. Open questions

1. **`Binding` as a generic struct?** Lean toward `Binding[T any]` parameterised by payload type, so the `MessageConverter` resolves into a concrete Go type instead of `any`. Cost: drivers carry a type parameter through `BindConsumer[T]` / `BindProducer[T]` — acceptable.
2. **Content-type negotiation — match Spring's MIME-type discipline?** Spring uses `application/json;type=com.example.Person`, carrying Java FQCN. Go has no FQCN. Drop the `type=` parameter; rely on the registered handler's Go type.
3. **DLQ policy materialisation strategy** — keep `DLQPolicy{Enabled, Name, Partitions, MaxAttempts}` on the binding (broker-neutral); each driver translates to its native topology. Same split Spring uses.
4. **Schema-registry coupling level** — should the driver consult the registry pre-publish, or should `MessageConverter` do it? Lean toward `MessageConverter` so the registry stays orthogonal to the driver, and Avro/Protobuf becomes "pick a different converter, not a different driver".
5. **Partition simulation on RabbitMQ** — opt-in (only when `PartitionPolicy.Count > 0`), matching Stream's default-off behaviour.
6. **Per-binding observability** — target [OTel messaging semantic conventions](https://opentelemetry.io/docs/specs/semconv/messaging/) (`messaging.system`, `messaging.destination.name`, `messaging.operation`, `messaging.consumer.group`). File a follow-up under Phase 3 Telemetry milestone (#9) when `messaging/` is promoted from Planned.
7. **Error-handler shape** — single `OnError func(ctx, msg, err)` callback on the binding, or full pluggable `ErrorHandler` interface? Lean toward the callback; the interface escalation can wait for a real use case (mirrors Stream's own `Consumer<ErrorMessage>` simplicity).
8. **`requiredGroups` semantics in AMQP** — straightforward (pre-declare named durable queues bound to the exchange). In Kafka, "group" is consumer-side state — does pre-declaring mean creating the topic + zero-offset commit, or just topic + partitions? Decide when the Kafka driver is implemented.

## 8. ROADMAP delta proposed (NOT applied)

In `docs/ROADMAP_NEW_MODULES.md` § 1.3 (`modules/messaging/`):

1. **Add to the proposed layout** the four new files at module root: `binding.go`, `converter.go`, `dlq.go`, `retry.go`, `partition.go`.
2. **Add a "Driver contract" bullet** stating: drivers implement `Channel` **and** materialise a `Binding` (DLQ / retry / partition / content-type / required groups).
3. **Cross-reference** [`docs/spring/spring-cloud-stream.md`](spring/spring-cloud-stream.md) and [`docs/spring/spring-integration.md`](spring/spring-integration.md) for design rationale, replacing any lingering references to the deleted Annex A in section bodies.
4. **No new tickets to file yet** — `messaging/` is still **Planned** (un-ticketed); the binding design folds into the first ticket when the module is promoted. The OTel-conventions follow-up (open question #6) gets filed under Phase 3 Telemetry only when `messaging/` work begins, to avoid premature scope.
