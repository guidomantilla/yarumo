# Coding Standards — modules/messaging/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md).
The PACKAGES.md classification is **Shape B with multiple peers of a
single interface (no canonical impl)** — see the messaging section in
[`modules/PACKAGES.md`](../PACKAGES.md).

## Module-specific conventions

### File naming for Channel implementations

The four `Channel[T]` implementations follow R1 "Múltiples peers de
una interface sin canónica": file name `channel_<variant>.go`, struct
name `<variant>Channel`. Precedent: `extension/common/log/slog`'s
`handler_fanout.go` + `handler_context.go`.

| File | Struct |
|---|---|
| `channel_pipeline.go` | `pipelineChannel[T]` |
| `channel_broadcast.go` | `broadcastChannel[T]` |
| `channel_topic.go` | `topicChannel[T]` |
| `channel_queue.go` | `queueChannel[T]` |

### Constructors return the interface, not the struct

All `NewXxxChannel[T]` constructors return `Channel[T]`, not the
concrete `*xxxChannel[T]`. The struct types are unexported. This
keeps the public surface minimal and discourages callers from
depending on impl-specific methods.

### lifecycle.Component is opt-in via type assertion

`topicChannel` and `queueChannel` implement `lifecycle.Component`
(Name/Start/Stop/Done) in addition to `Channel[T]`. Callers that need
lifecycle wiring use Go's interface assertion:

```go
ch := messaging.NewTopicChannel[Event]("name", opts...)
component, _ := ch.(lifecycle.Component)
closeFn, err := lifecycle.Build(ctx, component, errChan)
```

The two sync channels (`pipelineChannel`, `broadcastChannel`) do NOT
implement `lifecycle.Component` because they own no goroutines.

### Dispatch semantics differ per channel

Each `Channel[T]` impl makes different trade-offs documented in its
file's package doc-comment:

- **Pipeline**: sync, sequential, fail-fast with `*ChainError` trace.
- **Broadcast**: sync, parallel with barrier, joined errors (no
  fail-fast).
- **Topic**: async, buffered, fan-out via single worker, errors via
  `WithErrorHandler` hook.
- **Queue**: async, buffered, point-to-point round-robin via worker
  pool (`WithWorkerCount`), errors via hook.

Tests must NOT make assumptions across impls — each variant has its
own test file with semantics-specific assertions.

### Error reporting policy

- **Sync impls (Pipeline, Broadcast)**: caller receives the error. Use
  `*ChainError` for Pipeline (trace per step) and `errors.Join` for
  Broadcast (no order, no skipped).
- **Async impls (Topic, Queue)**: caller cannot receive handler
  errors. Errors and recovered panics are routed through the
  `WithErrorHandler` hook. Default is no-op — production wiring should
  always install a hook.

### Panic recovery is mandatory per handler

Every Channel impl runs handlers under `defer/recover`. A panicking
handler must NOT crash the channel, kill the worker, or propagate to
the caller. Recovered panics surface as errors wrapping
`ErrHandlerPanic` (for `errors.Is` matching).

### Message envelope

`Message[T]` is the only payload shape consumed by `Channel[T]`. Its
`Headers` struct (NOT flat fields) groups routing/provenance metadata
end-to-end through every Channel and Handler.

Header extension (MessageID, ReplyTo, Priority, etc.) is tracked under
ticket YA-0166. Today's set covers the minimum for in-process
correlation.

## Examples

`examples/main.go` is a runnable demonstration of all four channels
plus the Cancel flow. It lives in its own go.mod (see
`examples/go.mod`) following the workspace convention for example
binaries.

Running `go run ./examples` prints labeled sections so the runtime
ordering of each dispatch model is observable.
