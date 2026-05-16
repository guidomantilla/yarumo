# Spring Cloud Bus — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-bus
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: DEFER (revisit when `modules/messaging/` + a feature-flags module + a real cluster-config use case land simultaneously)

## 1. Project summary

Spring Cloud Bus links nodes of a distributed system with a lightweight message broker so management instructions and state changes can be broadcast in-band, alongside the application's own traffic. Concretely it is **two things glued together**:

1. **A typed event envelope (`RemoteApplicationEvent`)** carried over Spring Cloud Stream onto an AMQP or Kafka destination. Each event has `id`, `origin` (service-id of sender), `destination` (path-matched service-id pattern, e.g. `customers:**`), plus typed payload. Wire format is JSON; the deserializer must know the concrete subtypes at build time (subpackages of `org.springframework.cloud.bus.event` or registered via `@RemoteApplicationEventScan`).
2. **Two actuator endpoints (`/actuator/busrefresh`, `/actuator/busenv`)** that publish the two stock events (`RefreshRemoteApplicationEvent`, `EnvironmentChangeRemoteApplicationEvent`) so every instance listening on the same `spring.cloud.bus.destination` topic reloads its `@ConfigurationProperties` / mutates its `Environment` in lockstep with Spring Cloud Config server.

It is **not** a generic message bus — it is a *control plane* sidecar for Spring Cloud Config and the actuator subsystem. Custom events are supported but the design center is "POST to one node, every node sees it". Tracing (`spring.cloud.bus.trace.enabled`) records `spring.cloud.bus.sent` / `.ack` signals per event so the operator can see propagation.

Self-loop suppression is built in: each instance has a `spring.cloud.bus.id` (default `${spring.application.name}:${server.port}`); when an event arrives whose origin matches the local id, it is dropped. Addressing uses Spring's `PathMatcher` colon-separated patterns (`service`, `service:port`, `service:**`).

**Stock event taxonomy** (everything Spring ships out of the box):

| Event class | Triggered by | Payload | What it does on receivers |
|---|---|---|---|
| `RefreshRemoteApplicationEvent` | `POST /actuator/busrefresh` | none | clear `RefreshScope` cache, rebind `@ConfigurationProperties` |
| `EnvironmentChangeRemoteApplicationEvent` | `POST /actuator/busenv` | `{name, value}` | mutate runtime `Environment` with key/value pairs |
| `AckRemoteApplicationEvent` | every receiver after handling | original event id + origin | bookkeeping; consumed only by the `bus.trace` ring buffer |
| `SentApplicationEvent` | sender after publish | event id + destination | bookkeeping; observable for tracing |
| `UnknownRemoteApplicationEvent` | deserializer fallback | raw payload | seen when a node receives a custom event type it has not registered |

That is the entire surface. The reason this list is short and the recommendation later is "thin layer" is that Spring Cloud Bus is itself a thin layer — almost all weight lives in Spring Cloud Stream (transport) and Spring Cloud Config (the consumer of `busrefresh`). The bus owns the envelope and the two actuator endpoints; nothing else.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | **`RemoteApplicationEvent` envelope** | JSON message with `id`, `type`, `origin`, `destination`, payload — typed at both ends | Any cluster control-plane wants this. The envelope + addressing pattern is broker-agnostic and reusable. Equivalent in Go is a 30-line struct + a registry of `type` → factory. |
| 2 | **Broadcast config refresh (`busrefresh`)** | One node receives HTTP POST, every node re-reads its config | The single most-cited Spring Cloud Bus use case. Yarumo has `modules/config/` (one-shot bootstrap) — a refresh story doesn't exist yet but is the precondition for any future feature-flags module and live ruleset updates in DaaS. |
| 3 | **Addressing patterns (`service:**`, `service:port`)** | Path-matched selectors over a logical service-id | Surgical rollouts ("refresh only `daas-decisions:**`") matter when one bus topic carries traffic for many services. Cheap to implement. |
| 4 | **Self-loop suppression via origin-id** | Sender's own id stamped on event; receivers drop if `origin == local-id` | Trivially correct, prevents the obvious infinite-broadcast bug. Any Go re-implementation needs the same trick. |
| 5 | **Pluggable transport (AMQP / Kafka) via Spring Cloud Stream** | Same event code, two brokers selectable via starter dependency | Mirrors what `modules/messaging/` aims for: one `Channel` interface, broker-specific drivers. The bus is essentially a thin overlay on a pub/sub channel. |
| 6 | **Tracing endpoint (`sent` / `ack` signals)** | Optional `bus.trace` ring buffer records propagation per event id | Operators *will* ask "did my refresh reach prod-pod-3?". A simple in-memory trace + an HTTP endpoint covers 80% of debug needs. |

That is **the whole thing**. Six features. The rest is Spring plumbing — `@RemoteApplicationEventScan`, Spring Cloud Stream binders, actuator integration, Reactor variants — that does not survive translation to Go.

## 3. Long-tail features (skip)

- **`@RemoteApplicationEventScan` classpath scanning** — Go has no classpath; type registration is explicit (a `Register[T]()` call). Annotation magic is not a feature, it's a workaround for Java's reflection cost.
- **Spring Cloud Stream binders** — `modules/messaging/` (§ 1.3) replaces this entire layer. The bus on top of it is ~200 lines.
- **`/actuator/busenv` (live environment mutation)** — explicitly an anti-pattern. Mutating env at runtime is hard to audit, hard to roll back, and conflicts with twelve-factor config. If config must change at runtime, version the config and refresh-pull, not push-mutate.
- **Eureka-coupled service discovery for addressing** — yarumo targets K8s; service-id is `${app}:${pod-name}` or `${app}:${k8s-namespace}`. No registry dependency needed.
- **Spring Cloud Config server integration** — that's a *server* product yarumo has no equivalent of (and has no plan to build — `modules/config/` is one-shot bootstrap, not a remote config server). Without a Config server the `busrefresh` story degrades to "every node re-reads its own files", which is just SIGHUP.
- **AckRemoteApplicationEvent reflection back to origin** — useful in theory; in practice nobody wires it up. A trace endpoint covers the same diagnostic need without N² traffic.
- **WebFlux / Reactor variants** — N/A in Go.

**Concretely**, here is what each Spring concept maps to in a Go implementation:

| Spring Cloud Bus concept | Go translation | Lines of code (rough) |
|---|---|---|
| `RemoteApplicationEvent` base class | `RemoteEvent[T]` generic struct with `ID string`, `Type string`, `Origin string`, `Destination string`, `Payload T` | ~30 |
| `@RemoteApplicationEventScan` | `bus.Register[T](typeName string)` called from `BeanFn` at startup; populates a `map[string]func([]byte) (any, error)` registry | ~40 |
| Spring `PathMatcher` for destination | `Match(pattern, serviceID string) bool` — split on `:`, support `**` tail wildcard and `?` single-char | ~40 |
| `spring.cloud.bus.id` self-loop drop | one-line check in `Subscriber.handle` against the registered service-id | ~5 |
| `/actuator/busrefresh` | HTTP handler that constructs a `RefreshEvent` and calls `Publisher.Publish`; subscriber walks `config.Reloadable` registry | ~30 |
| Spring Cloud Stream binder | already covered by `modules/messaging/rabbitmq/` and `modules/messaging/kafka/` | 0 (reuse) |
| `bus.trace` ring buffer | bounded slice + RWMutex + HTTP GET `/admin/bus/trace` | ~30 |

Roughly **175 lines of code** for the entire bus, on top of `modules/messaging/`. That's the entire reason this evaluation lands on DEFER-with-clear-promotion-path rather than a separate-module plan.

The cost asymmetry matters: a separate `modules/bus/` would carry its own go.mod, its own linter config, its own coverage gate, its own `graph.go`, its own CODING_STANDARDS — boilerplate that is **larger than the feature itself**. A sub-package adds none of that overhead.

## 4. Mapping to Yarumo

**Existing/planned modules with overlap**:

- **`modules/messaging/` (§ 1.3)** — the bus *is* a thin overlay on `PubSubChannel`. Every transport concern (AMQP, Kafka, serialization, retries) belongs there, not in a bus module.
- **`modules/messaging/events/` (§ 1.3)** — already proposed as the in-process nominal-typed pub/sub façade. A cross-process variant addressed by `origin`/`destination` patterns is the same shape, one layer up. The envelope schema (`id` / `type` / `origin` / `destination` / payload) is essentially what a remote `events/` would need.
- **`modules/config/`** — one-shot today. A "refresh" story needs a `Reload()` hook + a subscriber that triggers it. Without that, there is nothing for a bus to broadcast to.
- **Proposed NEW `modules/featureflags/`** — not on the canonical roadmap; the canonical secondary consumer. Flag changes pushed via bus are the textbook Spring Cloud Bus use case after config refresh.
- **`modules/health/` (§ 1.4)** — could plausibly subscribe to a `DrainEvent` to flip readiness for blue-green deployments. Real but not urgent.
- **`modules/managed/`** — the bus subscriber is a lifecycle component (Start = subscribe, Stop = unsubscribe). Standard managed component, no surprises.

**Gaps this could fill**:

- A *cluster control plane* primitive is missing from yarumo entirely. Right now there is no answer to "I changed a ruleset in DaaS, propagate it to every pod". The current answer is "redeploy" or "poll". A bus event would close that gap.
- The `RemoteApplicationEvent` envelope + addressing pattern is the **right abstraction** for cross-process domain events. If/when `messaging/events/` extends to a cross-process variant, this is exactly the design.

**Anti-patterns to avoid**:

1. **Don't build `modules/bus/` as a separate module.** The bus is 200 lines of code on top of `messaging/`. A sibling module would duplicate the broker plumbing and force consumers to import two modules to get one feature. Fold it in as `modules/messaging/bus/` (sub-package), the same way `messaging/cdc/` and `messaging/schema/` are.
2. **Don't ship `busenv` (runtime env mutation).** Twelve-factor; auditability; rollback. Provide `busrefresh` only — re-read configured sources. If a value must change without redeploy, ship it through a future feature-flags module, not through environment-variable mutation.
3. **Don't couple to a service registry.** Service-id should be a string the operator provides (env var `BUS_SERVICE_ID`, default `${HOSTNAME}` or `${app}:${pod}`). K8s DNS / labels do discovery; the bus does broadcast.
4. **Don't require `RemoteApplicationEvent`-style classpath scanning.** Explicit `bus.Register[T](typeName)` at startup. No reflection registry magic, no `init()` side effects.
5. **Don't promise total ordering or exactly-once.** The bus is best-effort broadcast. If a node misses a refresh, it picks up on the next periodic reconciliation. Don't try to make it a state-machine replication protocol — that's a different concern (transactional outbox) that would warrant a proposed NEW `modules/outbox/` of its own.
6. **Don't tie the bus to a single broker.** Use the `Channel` abstraction from `messaging/`, not the AMQP or Kafka driver directly.
7. **Don't conflate control plane with data plane.** The bus is a control-plane primitive — config refresh, kill switches, drain signals. Domain events (`OrderPlaced`, `UserCreated`) belong in `messaging/events/` and downstream consumers, not on the bus topic. Mixing the two on one destination makes both noisier and harder to reason about.
8. **Don't expose `busrefresh` to the public internet.** It is an admin endpoint; ship it on the management port behind `modules/auth/` middleware, never on the user-facing HTTP listener. Spring's default is to lump it onto `/actuator/*` and hope operators configure auth correctly; yarumo should separate the two by default.

## 5. Recommendation

**DEFER.** The pattern is sound and the design center is small (six features, ~300 LOC in Go). But three preconditions must be satisfied before this is worth implementing:

1. **`modules/messaging/` exists** (currently Planned, § 1.3) — the bus is parasitic on its `PubSubChannel`. Without messaging, the bus has nothing to ride on.
2. **`modules/config/` grows a `Reload()` hook** — there must be *something to refresh*. Today config is one-shot; broadcasting "go refresh" to nodes that can't refresh is a no-op.
3. **A real consumer materializes** — either a future feature-flags module or DaaS ruleset live-updates. Without a concrete user, this is a speculative module that will rot. The whole point of yarumo's filter (criterion 1: "real pain every consumer reimplements") is to avoid that.

When those three line up — most plausibly when DaaS hits production and needs live ruleset reloads without redeploy — promote this to **PARTIAL** (`busrefresh` only, no `busenv`, no `ack` reflection, no Spring Cloud Stream emulation). Until then the pattern is documented here and the design is encoded in this file as a forward reference.

**Why not ADOPT now**: the *interesting* part of Spring Cloud Bus — the Config-server-pushes-to-every-node loop — has no yarumo analogue today and won't exist for at least two phases (Phase 3 closes the `config/` module; the `Reload()` API is not on its ticket list). Shipping a bus without consumers is the textbook case of "build the framework before the use case".

**Why not REJECT**: the pattern is not Spring-specific. The `RemoteApplicationEvent` envelope + addressing pattern is exactly what cross-process domain events need, and `messaging/events/` (§ 1.3) is on the planned list. When `messaging/events/` gets a cross-process variant, the design described in § 6 is what it should look like. Recording this evaluation now avoids re-deriving it later.

**Re-evaluation trigger**: file the ticket the day any one of these lands — (a) DaaS needs live ruleset reload, (b) a feature-flags module gets filed as Planned, (c) `modules/config/` ships a `Reload()` API. Whichever comes first flips this DEFER to PARTIAL with the placement in § 6.

## 6. Proposed yarumo placement (if/when promoted)

**`modules/messaging/bus/`** — sub-package of `messaging/`, **not** a sibling module.

```
modules/messaging/
  bus/
    event.go          RemoteEvent[T]{ID, Type, Origin, Destination, Payload}
    registry.go       Register[T](typeName), Decode(rawJSON) → typed event
    publisher.go      Publish(ctx, event) — wraps a PubSubChannel
    subscriber.go     Subscribe(handler) — origin-id self-loop drop, destination match
    addressing.go     Match(pattern, serviceID) — colon-separated, ** wildcard
    refresh/
      handler.go      RefreshHandler — calls registered config.Reloader callbacks
      endpoint.go     HTTP handler POST /admin/bus/refresh → publish RefreshEvent
    trace.go          In-memory ring buffer of sent/received events + HTTP GET endpoint
```

**Why a sub-package, not `modules/bus/`**:

- The bus has zero infrastructure of its own. It is a typed wrapper on `PubSubChannel`.
- A standalone module would re-export `messaging.Channel`, `messaging.Publisher`, `messaging.Subscriber` — pure indirection.
- Spring made it a separate project because Spring Cloud Bus predated Spring Cloud Stream as a unification effort; yarumo does not have that historical baggage.
- The companion model already in use (`messaging/cdc/`, `messaging/schema/`, `messaging/events/`) is the same shape and the right precedent.

**Internal deps**: `modules/messaging/` (channels + brokers), `modules/config/` (for `Reload()` callback registration), `modules/managed/` (lifecycle).

**Out of scope for v1**: `busenv`, `AckRemoteApplicationEvent`, custom event auto-discovery, service registry integration, Reactor/streaming APIs.

## 7. Open questions

1. **Does `modules/config/` get a `Reload()` API?** Without it the bus has nothing to drive. This is a `modules/config/` design decision, not a bus one — but it's the gating question. Tracked separately when Phase 3 (milestone #9) opens this surface. A minimal answer: `config.Reloadable` interface (`Reload(ctx) error`) plus a registry the bus subscriber walks on each `RefreshEvent`. Each `BeanFn` opts in by registering its reloader. No global `RefreshScope` — that's a Spring AOP feature that doesn't translate.
2. **What's the service-id source of truth?** Options: (a) env var `BUS_SERVICE_ID`, (b) computed from `os.Hostname()` + `${app-name}`, (c) injected via `BeanFn` in `modules/boot/`. K8s-friendly answer is (a) with (b) as fallback. Spring's default (`${app}:${port}`) is wrong for containers where port is fixed but pod identity is what matters.
3. **Does the bus need an `ack` reflection?** Spring has it; nobody uses it. A trace endpoint on each node (just "what did I see and when") plus a coordinator-side aggregator (out of scope) covers the same need. Recommendation: skip `ack` in v1; if operators want propagation visibility, expose per-node `/admin/bus/trace` and let observability pipelines (Loki, Datadog) aggregate.
4. **One destination per service, or shared destination across the whole cluster?** Spring's default is one topic for everything, addressing via `destination` field. Pro: cheaper to operate. Con: noisy — every node deserializes every event. K8s answer: shared topic, filter in subscriber. Confirm when concrete load numbers exist. For multi-tenant DaaS the answer may differ — tenant-scoped topics may matter for isolation, but that's a 2027 problem.
5. **Should `messaging/events/` (in-process) and `messaging/bus/` (cross-process) share the envelope?** Probably yes — `Event[T]` and `RemoteEvent[T]` with the same payload shape, where `RemoteEvent` adds `Origin` / `Destination` / `ID`. Decide when both are implemented; the in-process variant lands first per § 1.3's ordering.
6. **Does Aluna want this?** Spring Cloud Bus is config-broadcast, but the same envelope could carry agent-coordination signals (model swap, prompt-template update, kill-switch). If Aluna is the first real consumer instead of DaaS, the addressing model may need to be more selective than `service:**` — likely agent-id / agent-role selectors. Defer until Aluna writes its first Go code.
7. **Transport fallback if neither RabbitMQ nor Kafka is available?** Spring's bus is broker-bound. A `messaging/channels/direct.go` in-process fallback for single-node deployments + tests is trivial; ship it from day one. Single-pod deployments and integration tests should not need to spin up RabbitMQ in CI.
8. **What about ordering guarantees on the bus?** A `RefreshEvent` followed quickly by a `FlagUpdateEvent` may arrive out of order on a Kafka partition-per-event-id scheme. For control-plane events this is usually fine (idempotent + last-writer-wins), but worth declaring explicitly: the bus is **causal-broadcast at best, no FIFO promise across event types**. Document this on the public interface.
9. **Where does authentication of bus messages live?** Spring assumes the broker auth is enough (network-level trust). For multi-tenant deployments where one broker carries multiple customers' control traffic, message-level signatures may be needed — HMAC over the envelope using `common/crypto/tokens`. Probably overkill for v1 but worth flagging before someone deploys to a shared broker.

## 8. ROADMAP delta proposed (NOT applied)

Nothing lands in ROADMAP until promotion. If/when the three preconditions in § 5 are met, the delta would be a short sub-section under `§ 1.3 modules/messaging/` adding `messaging/bus/` to the proposed layout — alongside the existing `events/`, `cdc/`, `schema/` sub-modules. No new top-level module, no annex, no separate `modules/bus/`. The exact wording would mirror how `messaging/schema/` is currently described in ROADMAP_NEW_MODULES.md.
