# Spring Cloud Circuit Breaker — Yarumo Analysis

> **Source**: https://docs.spring.io/spring-cloud-circuitbreaker
> **Analyzed**: 2026-05-16
> **Re-analyzed**: 2026-05-16 (editorial cleanup after roadmap trim)
> **Recommendation**: PARTIAL — validates current `common/resilience/` direction; adopts a small, surgical subset (Retrier + Time Limiter + Bulkhead) as extensions, **not** a wrapper of Spring or of any Java port.

## 1. Project summary

Spring Cloud Circuit Breaker is a **thin abstraction** in Spring Cloud Commons over two concrete backends — primarily **Resilience4j** (sync + reactive variants) and **Spring Retry**. Its surface is small: a `CircuitBreakerFactory` that builds named `CircuitBreaker` instances, a `Customizer` pattern for per-instance config, and property-based defaults via YAML. The real value lives one layer down in Resilience4j: circuit breaker (sliding window count- or time-based, failure-rate + slow-call thresholds, half-open probe budget), **Bulkhead** (semaphore vs fixed thread-pool), **Time Limiter** (timeout with cancellation), **Retry**, **Rate Limiter**, fallback handlers, Micrometer metrics, Spring Boot health indicators, and event publishers (`onError`, `onSuccess`, `onStateTransition`).

The Spring abstraction itself adds: auto-configuration, property hierarchy (instance → group → default), and the ability to swap Spring Retry in for Resilience4j without code changes. For Go services none of that "wrapper value" survives — there is no Spring container, no auto-config, no annotations. The Pareto value sits in **the resilience patterns themselves**, which `common/resilience/` already partially covers and where gobreaker / `x/time/rate` already provide solid Go primitives.

## 2. Pareto features (top-20%)

| # | Feature | Description | Why it matters for Go microservices |
|---|---|---|---|
| 1 | Circuit breaker (state machine: closed / open / half-open) | Fail-fast under sustained failure; auto-recovery via probe budget | Already covered by `common/resilience/CircuitBreaker` via `sony/gobreaker`. The pattern itself is the #1 outbound-call resilience tool. |
| 2 | Rate limiter (token bucket / client-side throttle) | Cap outbound RPS to a dependency | Already covered by `common/resilience/RateLimiter` via `golang.org/x/time/rate`. Required to be a polite client. |
| 3 | **Retry with exponential backoff + jitter** | Re-invoke a failed operation N times with growing delays and randomized jitter | **Gap in `common/resilience/`** — tracked in [#165](https://github.com/guidomantilla/yarumo/issues/165). The single most common resilience pattern in Go microservices that yarumo currently does not standardize. |
| 4 | **Time Limiter (per-call timeout with cancellation)** | Bound the duration of a single call independent of the caller's `ctx.Deadline()` | **Gap.** Per-dependency or per-operation timeouts that defaults to a reasonable value rather than the caller's full request budget. Often implemented inline with `context.WithTimeout` and ad-hoc. A named, configurable primitive removes the duplication. |
| 5 | **Bulkhead (concurrency cap per dependency)** | Cap concurrent in-flight calls to one downstream so it cannot consume the whole goroutine / connection budget | **Gap.** In Go the natural form is a **semaphore bulkhead** (`chan struct{}` or `golang.org/x/sync/semaphore`). The Resilience4j "fixed thread-pool bulkhead" is a JVM artefact and **does not transfer**. |
| 6 | Fallback hook | Provide an alternate value (cache / default / degraded mode) when the primary call is rejected or fails | In Go, idiomatic to just inspect the error and route. A `Fallback[T]` helper that wraps `func() (T, error) + func(error) (T, error)` is a small ergonomic win, optional. |
| 7 | Sliding-window failure-rate threshold | Trip the breaker on **% failures over last N calls** rather than only **K consecutive failures** | gobreaker uses consecutive failures by default. Failure-rate over a window is more accurate for high-RPS bursty services. Worth offering as an opt-in `Option`, not a redesign. |
| 8 | Slow-call rate threshold | Count "too slow but successful" calls as failures for tripping purposes | Useful for latency-sensitive consumers; couples with time limiter. Lower priority than #3/#4/#5. |
| 9 | Named instances + registry | Look up the breaker / limiter / retrier by string key; lazy-create on first access | Already covered by `common/resilience/CircuitBreakerRegistry` and `RateLimiterRegistry`. Apply the same shape to retriers, time-limiters, bulkheads. |
| 10 | Event publisher (on-open / on-close / on-half-open / on-rejected) | Hook for logs, metrics, traces, alerts | gobreaker has a single `OnStateChange` callback — sufficient if exposed in the registry. Wire to OTel via `modules/telemetry/otel/` rather than coupling here. |
| 11 | Metrics integration (Micrometer / Spring Boot health) | Built-in counters and gauges for breaker state, rejections, latency | Belongs in `modules/telemetry/otel/` — `common/resilience/` should expose hooks (state-change callback, rejection counter), not import an OTel SDK. |

## 3. Long-tail features (skip)

- **Reactive variants** (`ReactiveResilience4JCircuitBreakerFactory`, `Mono`/`Flux` decorators). Go has no Reactor — `context.Context` is the substrate. Skip.
- **Fixed thread-pool bulkhead** (`ThreadPoolBulkheadConfig`). A JVM mitigation for blocking-thread exhaustion. Goroutines are cheap; the bulkhead pattern in Go is a counting semaphore. Skip the thread-pool flavor entirely.
- **Spring Retry alternative backend**. The "swap implementations via properties" trick only matters when you are already on Spring. Skip.
- **Auto-configuration / starters** (`spring-cloud-starter-circuitbreaker-resilience4j`). No equivalent in Go; wiring lives in `modules/boot/` (§ 1.5) when needed. Skip.
- **YAML property hierarchy** (`resilience4j.circuitbreaker.configs.default` + `.instances.backendA`). yarumo binds via `modules/config/` → `Options` structs. The "named config templates" idea is nice but premature — `registry.Use(name, opts...)` is enough until a real consumer asks for templates.
- **Spring Boot Actuator health indicators**. Replaced by `modules/common/health/` (leaves, closed) + `modules/health/` (runtime, planned § 1.4). Resilience primitives should publish hooks; the health module aggregates.
- **`@CircuitBreaker` / `@TimeLimiter` annotations** (Resilience4j-Spring AOP). No AOP in Go. Composition via wrapper / decorator funcs.
- **Adaptive / Netflix concurrency-limits style limiters** (offered by `slok/goresilience`, `failsafe-go`). Real value but **not Spring** — track as a separate brainstorm item if a consumer needs it.
- **Hedge / parallel duplicate request** (in `failsafe-go`). Not in Spring Cloud Circuit Breaker scope; skip from this analysis.
- **Cache as a resilience policy** (in `failsafe-go`). yarumo has `modules/cache/`; conflating cache with resilience would muddle layering. Skip.

## 4. Mapping to Yarumo

**Existing modules with overlap**:

- **`modules/common/resilience/`** ([YA-0076](https://github.com/guidomantilla/yarumo/issues/76), closed 2026-05-13) — already ships:
  - `CircuitBreaker` interface + `CircuitBreakerRegistry` (lazy, goroutine-free) backed by `sony/gobreaker`.
  - `RateLimiter` interface + `RateLimiterRegistry` (lazy, goroutine-free) backed by `golang.org/x/time/rate`.
  - Options pattern with `WithCircuitBreakerMaxRequests`, `WithCircuitBreakerInterval`, `WithCircuitBreakerTimeout`, `WithCircuitBreakerConsecutiveFailures`, `WithRateLimiterInterval`, `WithRateLimiterBurst`.
  - Sentinel-translated errors (`ErrCircuitBreakerOpen`, `ErrCircuitBreakerTooManyRequests`, `ErrRateLimiterWait`, …).
  - Pluggable-struct pattern (criterion 4 Exception 3), single-call `Execute(ctx, fn)` surface, package-level `DefaultCircuitBreakerRegistry` / `DefaultRateLimiterRegistry`.
- **[#165](https://github.com/guidomantilla/yarumo/issues/165)** — open follow-up: revisit implementations + **add Retrier pattern**. Explicitly scoped as an extension of `common/resilience/`, **not** a new module.
- **Top-level `modules/resilience/`** — **explicitly decided NOT to create**. Trigger condition ("if `common/resilience/` cannot stay goroutine-free") never materialised; the planned extensions below stay inside `common/resilience/`.
- **`modules/common/http/`** — outbound HTTP client features ([YA-0042](https://github.com/guidomantilla/yarumo/issues/42)) will compose `CircuitBreaker` + `RateLimiter` + (future) `Retrier` + `TimeLimiter` as middlewares.

**Gaps this could fill**:

| Pattern | Current state | Where it lands |
|---|---|---|
| **Retrier** (exponential backoff + jitter + max-attempts + retry-if predicate) | Missing — #165 already files it | `common/resilience/` extension: `Retrier` interface + `RetrierRegistry`. |
| **TimeLimiter** (per-call deadline with `context.WithTimeout`) | Missing — every caller writes it inline | `common/resilience/` extension: `TimeLimiter` interface or a `WithTimeout(d, fn)` helper. Most likely an `Option` on `Retrier` and `CircuitBreaker.Execute` rather than a separate primitive — simpler. |
| **Bulkhead** (semaphore-based concurrency cap) | Missing | `common/resilience/` extension: `Bulkhead` interface + `BulkheadRegistry`, backed by `golang.org/x/sync/semaphore` or a hand-rolled `chan struct{}`. Goroutine-free. |
| **Sliding-window failure rate** | Missing (gobreaker uses consecutive-failure ReadyToTrip) | New `Option`: `WithCircuitBreakerFailureRate(window int, threshold float64)`. Configures a different `ReadyToTrip` over `gobreaker.Counts`. No backend change needed. |
| **Slow-call rate** | Missing | Lower priority — defer until a consumer asks. |
| **State-change observer** (open/close/half-open events) | Hidden — `gobreaker.Settings.OnStateChange` is set but not surfaced through `Options` | Expose `WithCircuitBreakerStateChange(fn)` so callers (and `modules/telemetry/otel/`) can hook in. Tiny API addition. |
| **Fallback helper** | Missing | Optional: `func Fallback[T any](primary func() (T, error), onErr func(error) (T, error)) (T, error)`. Trivial — may not even need to live in `common/resilience/`. |
| **Composable pipeline** (CB → Retry → Timeout → Bulkhead) | Missing | Compose at the call site (CB.Execute wrapping a Retrier wrapping a TimeLimiter). Don't build a "Pipeline" type until the wrap-noise becomes real. Spring's `Customizer` and failsafe-go's `failsafe.Get(executor, fn)` both effectively assemble this — keep it explicit and Go-idiomatic in yarumo. |

**Anti-patterns to avoid**:

1. **Wrapping `failsafe-go` or `slok/goresilience` wholesale.** Both are reasonable Go libraries, but they bring their own pipeline / runner abstraction that fights the lazy-goroutine-free, registry-by-name shape `common/resilience/` already has. We borrow ideas, not the surface.
2. **Re-creating Spring's `Customizer` pattern** with anonymous lambdas mutating factories. yarumo uses `Option` functions returning fully-constructed primitives via the registry's `Use(name, opts...)`. Don't fold a second config style on top.
3. **Annotations / AOP.** No `@CircuitBreaker`. Wrap explicitly: `registry.Get("x").Execute(ctx, fn)`.
4. **Thread-pool bulkhead.** Pure JVM concept (decouple blocking I/O from the request thread). In Go, goroutines are cheap and `context.Context` carries cancellation — a semaphore cap is enough. Do not import that mental model.
5. **Reactive types in the public API.** No `Mono`, no callbacks-of-callbacks. Plain `func() (any, error)` (or generic `func() (T, error)` if/when we generify the interface).
6. **OTel SDK import inside `common/resilience/`.** The package must stay free of telemetry SDK deps. Expose hooks (`OnStateChange`, `OnReject`); wire to OTel in `modules/telemetry/otel/` or in the consumer.
7. **Background goroutines** (timer-driven half-open transitions, decay sweeps, metric exports). `common/resilience/` is explicitly goroutine-free — every transition evaluated synchronously on the next `Execute`. Any new primitive (Retrier, TimeLimiter, Bulkhead) must respect this invariant.
8. **YAML property-template hierarchy.** Premature. Stick to `NewOptions(opts...)` + `registry.Use(name, opts...)`. Revisit if a consumer needs profile-style "config templates".
9. **Spring Retry as a "swappable backend".** Single backend (gobreaker) is fine; replaceability is achieved by the `CircuitBreaker` interface, not by maintaining N implementations.

## 5. Recommendation

**PARTIAL** — keep `common/resilience/` exactly where it is and extend it with three new primitives plus two small additions, all under the same package, all goroutine-free, all registry-shaped.

### What is validated by this analysis

- The placement decision is correct: **library, no lifecycle** → belongs in `modules/common/`. The decision to keep resilience inside `common/` (rather than promote to a top-level `modules/resilience/`) stands.
- The goroutine-free design is the right invariant. Spring Cloud Circuit Breaker delegates this to Resilience4j's `ScheduledThreadPoolExecutor`; yarumo's "evaluate on access" is the simpler, cheaper Go-native equivalent. Keep it.
- The two-primitive split (`CircuitBreaker`, `RateLimiter`) is the right shape. Spring's `CircuitBreakerFactory` + separate `BulkheadProvider` + `TimeLimiterConfig` validates the "one interface per pattern, registry per family" approach.
- Backing `CircuitBreaker` with `sony/gobreaker` and `RateLimiter` with `golang.org/x/time/rate` is fine. Both are mature, popular, single-purpose, goroutine-free.

### What to add (small, surgical)

1. **`Retrier`** — fulfills [#165](https://github.com/guidomantilla/yarumo/issues/165). Exponential backoff + jitter + max-attempts + retry-if predicate (`func(error) bool`). Backed by hand-rolled math (the algorithm is tiny — no library dep needed) or `cenkalti/backoff/v4` (depends on whether we want jitter strategies out of the box). Goroutine-free: backoff is computed; the sleep happens inline in the caller's goroutine via `time.NewTimer` + `select { ctx.Done() / timer.C }`.
2. **`Bulkhead`** — semaphore-based concurrency cap. Single field: `maxConcurrent int`. Backed by `golang.org/x/sync/semaphore` (already in the workspace's dependency surface via other modules) or a `chan struct{}`. Goroutine-free.
3. **`TimeLimiter`** — most likely **not a standalone primitive**, but a `WithTimeout(d)` option that wraps `context.WithTimeout` inside `Execute` / `Retrier.Run`. Spring makes it a separate config block because of JVM `Future` cancellation semantics; Go has cancellation built into `context.Context`, so the value-add is just "don't make every caller write `ctx, cancel := context.WithTimeout(ctx, d); defer cancel()`". Decide during YA-#165 implementation whether it deserves its own type.
4. **Failure-rate `ReadyToTrip`** — opt-in `Option` (`WithCircuitBreakerFailureRate(window int, threshold float64)`) that produces an alternate `ReadyToTrip` closure consuming `gobreaker.Counts`. No backend swap.
5. **Observer hooks** — `WithCircuitBreakerStateChange(fn)`, optionally `WithRateLimiterReject(fn)`. Forwards to `gobreaker.Settings.OnStateChange`. Enables consumers and the OTel module to attach metrics/logs without coupling.

### What to explicitly say no to

- A `Pipeline` / `Runner` type that composes all primitives. Wait for real demand. Composition via nested `Execute` calls is fine until it isn't.
- A new `modules/resilience/` module. The trigger never fired; do not revisit.
- Wrapping `failsafe-go` or `slok/goresilience` end-to-end. Use them as **inspiration** only.
- Reactive interfaces.
- Auto-config / DI / annotations.

## 6. Proposed yarumo placement

**Module**: `modules/common/resilience/` (extension — same package, no submodule split).

**Subpackages**: none. All new primitives live in the same package (`circuit_breaker.go`, `rate_limiter.go`, plus new `retrier.go`, `bulkhead.go`, possibly `timer.go`). Same pattern as today: one file per primitive + `registry_<primitive>.go` + tests. Matches `internals.go` / `options.go` layout the package already uses.

**Internal deps**:

- `modules/common/assert` (existing).
- `modules/common/errs` (existing — sentinel errors + typed-error pattern).
- `modules/common/log/slog/slogctx` only if the observer hooks need a default logger; prefer to keep `common/resilience/` log-free and let consumers wire logging via the hooks.

**Go libraries to wrap or reference for inspiration**:

| Library | Role | Verdict |
|---|---|---|
| `github.com/sony/gobreaker` | Already backs `CircuitBreaker` | **Keep.** Mature, no goroutines, single-purpose. |
| `golang.org/x/time/rate` | Already backs `RateLimiter` | **Keep.** Stdlib-adjacent, no goroutines. |
| `golang.org/x/sync/semaphore` | Backing for new `Bulkhead` | **Adopt** (or use a plain `chan struct{}` — decide during YA-#165 follow-up work; semaphore is more idiomatic and supports `Acquire(ctx, n)`). |
| `github.com/cenkalti/backoff/v4` | Optional backing for `Retrier` | **Reference.** Has retry orchestration, but hand-rolling exponential-backoff-with-jitter is ~30 lines of Go. Adopt only if we want pluggable backoff strategies (constant / exponential / decorrelated jitter) out of the box. |
| `github.com/failsafe-go/failsafe-go` | Full resilience kit (CB, retry, RL, bulkhead, time limiter, hedge, fallback, cache, adaptive limiter) | **Reference only.** Strong inspiration for the Retrier API surface and the "compose policies" idea. Do not wrap — its `Executor[T]` builder pulls a different abstraction shape than the registry-by-name approach yarumo has committed to. |
| `github.com/slok/goresilience` | Runner-based middleware chain (CB, retry, bulkhead, timeout, chaos, adaptive concurrency) | **Reference only.** Same as above — borrow ideas (chaos injection is interesting for a future proposed NEW `modules/testing/`), do not wrap. |

**Out of scope for v1**:

- Thread-pool bulkhead variant.
- Reactive / streaming variants.
- Hedged requests (parallel duplicate calls).
- Adaptive concurrency limiting (Netflix concurrency-limits style).
- Cache-as-resilience-policy.
- Slow-call rate threshold (defer until a consumer asks).
- Property/YAML-template hierarchy for named configs.
- Built-in OTel metrics export (lives in `modules/telemetry/otel/` consuming the observer hooks).
- A composing `Pipeline` / `Runner` type. Compose by nesting `Execute` calls until that proves painful.

## 7. Open questions

1. **Retrier API shape**: does `Retrier` mirror `CircuitBreaker.Execute(ctx, fn)` exactly (single shot, internally loops), or does it expose `Do(ctx, fn) iter.Seq[Attempt]` so callers can observe individual attempts? Lean toward the former for parity; decide in YA-#165.
2. **TimeLimiter as primitive or option**: a separate `TimeLimiter` interface with its own registry feels over-engineered when `context.WithTimeout` is one line. Recommend **option** (`WithTimeout(d)` on `Retrier` and `CircuitBreaker.Execute`), unless a consumer needs to share / reconfigure a timeout policy across call sites by name. Defer the decision.
3. **Failure-rate threshold default**: should the new opt-in failure-rate `ReadyToTrip` ship with sane defaults (e.g. `window=50, threshold=0.5`), or stay opt-in only? Defaults risk silently changing trip behavior; stay strictly opt-in.
4. **Generic `Execute[T]`**: today `Execute(ctx, fn) (any, error)`. Should we offer a typed `Execute[T any](cb CircuitBreaker, ctx, fn func() (T, error)) (T, error)` helper to remove the `any` cast at call sites? Mirrors the cache / validation pattern (criterion 4). Probably yes — file as a small follow-up alongside YA-#165.
5. **Observer hook fan-out**: single callback per breaker, or registered slice (`AddOnStateChange(fn)`)? gobreaker exposes a single callback. Start with single; promote to slice only when a second consumer (e.g. OTel + audit) needs both simultaneously.
6. **Where does Retrier's randomness come from?** `math/rand/v2` directly, or `common/crypto/random`? Probably `math/rand/v2` — jitter does not need crypto-grade randomness, and the crypto subpackage carries different guarantees.
7. **Should `common/http/` (YA-0042) be the one to ship the composition pipeline?** Outbound HTTP is by far the most common composer of CB+Retry+RL+Timeout+Bulkhead. May be more natural to keep `common/resilience/` strictly as primitives and put any composition convenience inside `common/http/` where it has a concrete use case.
8. **Adaptive concurrency limiting** (Netflix concurrency-limits) — not in Spring Cloud Circuit Breaker, but offered by `failsafe-go` and `slok/goresilience`. If a real consumer asks (likely DaaS or Aluna under load), where does it live? Probably a new `common/resilience/` primitive (`AdaptiveLimiter`); file when the use case appears, not preemptively.

## 8. ROADMAP delta proposed (NOT applied)

Nothing is proposed for ROADMAP_NEW_MODULES.md itself — the resilience surface lives entirely inside `modules/common/resilience/`, which is already closed (YA-0076). The follow-up work (Retrier, Bulkhead, TimeLimiter option, failure-rate `ReadyToTrip`, observer hooks) is tracked on issue [#165](https://github.com/guidomantilla/yarumo/issues/165) and lands as additions to the existing package — no roadmap entry needed.
