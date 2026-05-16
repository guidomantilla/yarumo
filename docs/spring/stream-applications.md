# Spring Cloud Stream Applications — Yarumo Analysis

> **Source**: https://docs.spring.io/stream-applications/
> **Analyzed**: 2026-05-16 (re-analysis after roadmap cleanup — § 3 brainstorm + Annex A deleted)
> **Recommendation**: REJECT (function-shape vocabulary implicitly absorbed via `messaging/endpoints/` — no new ticket)

## 1. Project summary

Spring Cloud Stream Applications is **a catalog of pre-built, deployable Spring Boot apps** — Sources, Processors, and Sinks — that plug into Spring Cloud Stream (binders) and are orchestrated by Spring Cloud Data Flow (SCDF) into pipelines. It is not a framework or a library API; it is the **output** of a framework: ~58 ready-to-run JAR/Docker artifacts (`time-source-kafka`, `jdbc-sink-rabbit`, `filter-processor-kafka`, …) consumers pull from a Maven repo or Docker Hub and wire together via configuration.

**Status confirmed**: archived 2026-02-26 on GitHub (read-only). Last OSS release is `2025.0.1` (2025-10-18). Version `2025.0.2` and beyond ship only through Broadcom's commercial Tanzu Spring artifact repository. From Broadcom's notice: *"Spring Cloud Stream Applications are no longer maintained as an open-source project by Broadcom, Inc."* The upstream is therefore frozen for any consumer not on Tanzu — even Java teams.

The catalog (counts from current ref guide):

- **Sources (23)**: Debezium (CDC), File, FTP, HTTP, JDBC, JMS, Kafka, Load Generator, Mail, MongoDB, MQTT, RabbitMQ, S3, SFTP, Syslog, TCP, Time, Twitter (Stream / Search / Messages), Websocket, XMPP, ZeroMQ.
- **Processors (10)**: Aggregator, Bridge, Filter (SpEL), Groovy, Header Enricher, Header Filter, HTTP Request, Script (JS/Python/Ruby/Groovy), Splitter, Transform (SpEL), Twitter Trend.
- **Sinks (25)**: Analytics, Cassandra, Elasticsearch, File, FTP, JDBC, Kafka, Log, MongoDB, MQTT, PGCOPY, RabbitMQ, Redis, Router, RSocket, S3, SFTP, TCP, Throughput, Twitter (Update / DM), Wavefront, Websocket, XMPP, ZeroMQ.

**Internal design** — the only structural takeaway: every app is a thin Spring Boot wrapper around a function from the **Spring Functions Catalog** (separate library):

```
Source    = Supplier<T>
Processor = Function<T, R>
Sink      = Consumer<T>
```

Apps are emitted by the `spring-cloud-stream-app-maven-plugin`, which combines one function + one binder (Kafka or RabbitMQ) + Spring Boot autoconfig into an uber-JAR. Naming convention: `<func>-<role>-<binder>` (`jdbc-sink-kafka`, `http-source-rabbit`).

## 2. Pareto features (top-20%)

There are no "features" to adopt in the conventional sense — the project is a **distribution** of pre-built artifacts, not a library API. The 20% that has any conceptual interest:

1. **Functional shape (`Supplier` / `Function` / `Consumer`)** as the unit of stream computation. A first-class, transport-independent vocabulary for "this piece produces / transforms / consumes messages" — the same abstraction Spring Cloud Function exposes, surfaced here as a packaging convention.
2. **Function-Catalog separation**: function logic (`spring-functions-catalog`) lives independently of the deployable app. Reusable as a plain library by any Boot consumer; the app artifacts are just `functionCatalog + binder + boot`.
3. **Binder-pluggable packaging**: the same function shipped as 2–3 binder variants (kafka, rabbit, kinesis) without changing function code — a deployment-time concern, not a code concern.

## 3. Long-tail features (skip)

Essentially everything else:

- **The catalog itself (~58 apps)**: concrete deployables. Yarumo does not ship pre-built microservices — consumers compose their own.
- **SCDF integration**: Spring Cloud Data Flow is a pipeline orchestrator. Yarumo's lean roadmap has no orchestrator track (§ 1 modules are libraries; § 2 is a code generator; § 4 is migration accounting).
- **SpEL filter/transform processors**: yarumo has `modules/common/expressions/` for expression evaluation. Consumer-driven, never wrapped in a pre-packaged app.
- **Groovy / JavaScript / Python / Ruby script processors**: out of scope (Go workspace, no embedded scripting).
- **Maven / Docker Hub distribution + app-generator Maven plugin**: JVM packaging stack. Go binaries do not ship by Maven coordinates and do not need a per-binder fan-out plugin.
- **Twitter / XMPP / ZeroMQ / Wavefront specific apps**: long-tail integrations with no DaaS / Aluna demand.
- **Function-composition pipelining (`source | proc1 | proc2 | sink`) via SCDF DSL**: orchestrator concern.
- **Archived OSS**: tracking a project whose next release is commercial-only buys nothing.

## 4. Mapping to Yarumo

### 4.1. Categorical mismatch

Stream Applications is a **catalog of deployable microservices**. Yarumo is a **catalog of libraries** that user microservices import. These are different layers of the stack:

```
Spring world:    [Stream Apps (deployables)] -> [Stream (binders)] -> [Functions Catalog (libs)]
Yarumo analog:   [user's microservice]       -> [modules/messaging/] -> [function-shaped helpers?]
```

Yarumo would never ship `http-source-kafka` as a binary because:

1. Each yarumo consumer **writes its own microservice** and imports the libraries it needs. The "source / processor / sink" wrappers add zero value when the consumer already has a `main.go` and a `modules/boot/` container.
2. There is no SCDF-equivalent in scope — the lean roadmap (§ 1) lists no orchestrator module, and the deleted § 3 brainstorm is gone for good.
3. Go's deployable unit is one statically linked binary. Per-binder × per-function fan-out (`jdbc-sink-kafka`, `jdbc-sink-rabbit`, …) makes no sense when the consumer already chooses their binder at import time (`modules/messaging/kafka/` vs `modules/messaging/rabbitmq/`).

### 4.2. What about the function shape?

The one transferable idea is the `Supplier` / `Function` / `Consumer` triplet as the shape for stream-handling code. In Go:

```go
type Source[T any]       func(ctx context.Context) (T, error)            // Supplier
type Processor[I, O any] func(ctx context.Context, in I) (O, error)      // Function
type Sink[T any]         func(ctx context.Context, in T) error           // Consumer
```

This is conceptually clean, and `modules/messaging/endpoints/` (planned in § 1.3 of `ROADMAP_NEW_MODULES.md` — Transformer / Filter / Router / Splitter / Aggregator / Service Activator) is already going to express EIP endpoints using exactly these shapes. The function vocabulary is **implicitly absorbed** via the Spring Integration design — see `docs/spring/spring-integration.md`. No additional input is contributed by Stream Applications.

### 4.3. Existing yarumo coverage of adjacent concerns

| Stream Applications concern | Where it lives in yarumo |
|---|---|
| HTTP source | Consumer writes a `managed/HttpServer` route directly. |
| JDBC sink | `modules/datasource/gorm/` (§ 1.1) + a `messaging` consumer (§ 1.3). |
| Kafka source/sink | `modules/messaging/kafka/` (§ 1.3, Planned, low prio in § 4.1). |
| RabbitMQ source/sink | `modules/messaging/rabbitmq/{amqp,streams}/` (§ 1.3). |
| S3 / file / FTP / SFTP | Not on the lean roadmap. Would land as a `datasource/` driver if real demand appears — none today. |
| Debezium CDC source | `modules/messaging/kafka/cdc/` (§ 1.3 — explicitly planned as a Kafka sub-module). |
| Filter / Transform / Aggregator processors | `modules/messaging/endpoints/` (§ 1.3). |
| Schema validation (Avro / Protobuf / Confluent SR) | `modules/messaging/schema/` (§ 1.3 — already planned). |
| SCDF orchestration / app composition DSL | **Not in scope** (lean roadmap has no orchestrator track). |
| Pre-built deployable apps | **Not in scope** — yarumo is a library workspace, not an app catalog. |

Every cell that yarumo cares about is already covered by a planned module under `messaging/`. Stream Applications adds **no design input** that Spring Messaging (`docs/spring/spring-messaging.md`), Spring Integration (`docs/spring/spring-integration.md`), and Spring Cloud Stream (`docs/spring/spring-cloud-stream.md`) have not already provided.

### 4.4. Anti-patterns to avoid

1. **Pre-built artifact catalog**: shipping `kafka-sink-yarumo` binaries duplicates `main.go` machinery the consumer already writes. Never do this.
2. **Per-binder fan-out at the library level**: each binder picks itself by `import` path. No `modules/messaging/jdbc-sink-kafka/` cross-products.
3. **DSL-based pipeline composition** (à la SCDF): pipeline orchestration is an *application-level* concern. If it ever becomes needed (DaaS workflows etc.), it goes through `sdks/processes/` (out-of-scope here), not `modules/`.
4. **String-based SpEL expressions inside packaged apps**: yarumo's `common/expressions/` exists, but it is consumed in user code, not baked into pre-built artifacts.

## 5. Recommendation

**REJECT**. The previous analysis's REJECT verdict stands and strengthens after roadmap cleanup. Rationale, updated:

1. **Wrong artifact type**: a catalog of *deployable apps*, not a library. Yarumo is exclusively a library workspace (modules + tools + SDKs + apps consumers build). No app catalog track exists or is planned.
2. **No orchestrator track in lean roadmap**: § 3 brainstorm (which once flirted with pipeline orchestrators) is deleted. There is now no roadmap surface that Stream Applications could plug into even speculatively.
3. **Upstream archived**: `2025.0.1` is the last OSS release (2025-10-18). Future versions are Broadcom Tanzu-only. The project is no longer a moving target the way Spring Messaging / Spring Integration still are.
4. **Per-binder packaging fan-out**: Java/Maven-specific concern. Go binaries do not multiply by binder — consumers pick at import time.
5. **All useful design ideas live upstream** — in Spring Cloud Function (functional shape), Spring Messaging (channels, headers, `Message<T>`), Spring Integration (EIP endpoints). Cross-references now point at `docs/spring/spring-integration.md`, `docs/spring/spring-messaging.md`, and `docs/spring/spring-cloud-stream.md` (the deleted Annex A's replacements).

The single conceptual idea worth carrying — the `Supplier` / `Function` / `Consumer` shape — is already inbound through `messaging/endpoints/` (§ 1.3) via Spring Integration's `MessageEndpoint` / `MessageHandler` model. No new ticket needed.

## 6. Proposed yarumo placement (if applicable)

None. No new module, no new ticket, no roadmap delta.

If the function-shape vocabulary ever wants to be made explicit in `modules/messaging/`, it lands as small generic aliases (`Source[T]`, `Processor[I, O]`, `Sink[T]`) inside `modules/messaging/endpoints/` — but that is a § 1.3 design decision, not a Stream-Applications-driven one.

## 7. Open questions

1. Should `modules/messaging/endpoints/` publish `Source` / `Processor` / `Sink` generic aliases as **first-class exported types** (one-liner each) or just document the convention in comments alongside `MessageHandler`? — Defer to § 1.3 design work.
2. Once `modules/messaging/kafka/cdc/` lands (§ 1.3), is there value in a `modules/messaging/kafka/connect/` for Kafka Connect interop, mirroring how Stream Applications shipped Debezium and JDBC as opinionated sources? — Probably no: Kafka Connect is itself a deployable runtime, not a library. Same rejection class as SCDF. Park indefinitely.
3. Should yarumo ever publish a `modules/messaging/functions/` bundle of reusable generic transformers (JSON parse, header enrich, JSON-to-struct, etc.) — a "Functions Catalog" lite? — Only if more than one yarumo consumer ends up reimplementing the same transformer. No demand today. Track as a brainstorm if/when demand appears.

## 8. ROADMAP delta proposed (NOT applied)

**None.** No additions or modifications to `ROADMAP_NEW_MODULES.md`.

The lean roadmap stands as-is for this analysis:
- § 1.3 `modules/messaging/` already covers the (very few) ideas worth importing.
- § 4.1 already tracks the relevant go-feather-lib pieces (`integration/messaging/`, `messaging/rabbitmq/{amqp,streams}/`) at low priority.
- Annex A's deletion does not create any gap that Stream Applications could fill — the replacements (`docs/spring/spring-integration.md`, `spring-messaging.md`, `spring-cloud-stream.md`) cover the design space.
