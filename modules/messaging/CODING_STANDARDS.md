# Coding Standards — modules/integration/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../core/common/CODING_STANDARDS.md).

## Scope

`modules/integration/` hosts Enterprise Integration Pattern (EIP) packages
that compose over the primitives in `modules/core/common/messaging/`. The
boundary is firm:

- `core/common/messaging/` owns the **transport primitives** — `Channel[T]`,
  `Message[T]`, the four channel implementations (Pipeline / Broadcast /
  Topic / Queue) and `NullChannel`.
- `modules/integration/` owns the **composition patterns** that wire
  `Channel[T]` instances together to express routing, filtering,
  transformation, aggregation, splitting, gateways and adapters.

If a new sub-package is a transport (sync/async × fan-out/p2p), it belongs
in `messaging/`. If it composes existing channels into a new behavior, it
belongs here.

## Sub-package layout

Each pattern lives in its own sub-package under `modules/integration/`.
Sub-packages share this module's `go.mod` (no per-pattern modules) per the
workspace rule on
[Sub-package vs sub-module](../core/common/CODING_STANDARDS.md): no pattern
in this module pulls a heavy external dependency, so MVS isolation does not
justify a separate go-module.

Current sub-packages:

| Path | Pattern |
|------|---------|
| `bridge/` | One-to-one channel forwarder (identity transform, sync↔async decoupling) |
| `filter/` | Message Filter (predicate-gated forwarding with separate error/drop hooks) |
| `router/` | Content-Based Router (key → Channel[T]) |
| `transformer/` | Message Translator (T→U mapping via TransformFn) |
| `wiretap/` | Wire Tap (non-intrusive copy to a side channel) |
| `splitter/` | Splitter (1 msg → N msgs with SequenceNumber/Size headers populated) |

Future sub-packages (aggregator, delayer, endpoint, controlbus, ...) get
added one at a time when a real consumer asks for them. Do not pre-create
empty sub-packages.

## Applicable Criteria

| # | Criterion | Notes |
|---|-----------|-------|
| 1 | Bullet proof review | Yes |
| 2 | Type Compliance | `var _ lifecycle.Component = (*router[any])(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Constructors return `lifecycle.Component`; struct types are unexported |
| 4 | Constructor returns interface | Yes |
| 5 | Options | Functional options pattern; generic when the option carries `T`-typed values |
| 6 | Preconfigured Default Singletons | No |
| 7 | Linter | Yes |
| 8 | Tests | Yes — `t.Parallel()`, individual `t.Run` subtests, no testify, no table-driven |
| 9 | Documentation | Yes — package doc + exported symbol doc comments end with periods |

## Override: errors do not propagate to the source channel caller

Every pattern in this module subscribes to a source `Channel[T]` with a
handler. That handler **always returns `nil`** to the source channel:
routing/filtering/transformation failures are not failures of the source
channel itself, and the source channel's caller should not see them.

Pattern-level errors (no matching route, decision function panicked, forward
send failed, ...) flow through the pattern's own `WithErrorHandler` hook —
defaulting to `messaging.DefaultErrorHandler`, which logs via `common/log`
at Error level. Consumers that want silence opt in explicitly via
`WithErrorHandler(messaging.SilentErrorHandler)`.
