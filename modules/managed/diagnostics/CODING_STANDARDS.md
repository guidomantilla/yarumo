# Coding Standards — modules/managed/diagnostics/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `TraceFlightRecorder` + `BlockProfiling` interfaces in `types.go`; `tracefr`, `PluggableTraceFlightRecorder`, `blockprof` impls with compliance vars |
| 3 | Public Interface, Private Implementation | Yes | `TraceFlightRecorder`/`BlockProfiling` public, `tracefr`/`blockprof` private (`PluggableTraceFlightRecorder` is the documented Shape B variant) |
| 4 | Constructor returns interface | Yes | `NewTraceFlightRecorder`, `NewBlockProfiling` return interfaces |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewTraceFlightRecorder`/`NewBlockProfiling` call owns its inner state |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`modules/core/common/` is a pure library with no lifecycle opinions. The
`TraceFlightRecorder` and `BlockProfiling` types returned here own
lifecycle: `Start(ctx)` enables a runtime hook (flight recorder buffer
or block-profile sampling) that must be released via `Stop(ctx)`. That
violates the "no lifecycle" clause of `modules/core/common/`. The diagnostics
module lives at the top-level module layer alongside `modules/managed/cron/`,
`modules/managed/grpc/`, `modules/managed/http/`, `modules/managed/`, `modules/extension/common/cache/`,
`modules/managed/telemetry/` and `modules/config/`.

### Override: Shape B (canonical Shape B layout)

`TraceFlightRecorder` and `BlockProfiling` are Shape B types — each
wraps a runtime hook with the managed-component idiom. Layout:
`types.go` (interfaces + `Fn` aliases + compliance vars), `tracefr.go`
(private impl + constructor), `tracefr_pluggable.go` (pluggable variant
documented under Shape B R1), `blockprof.go` (private impl + constructor),
`handlers.go` (stateless pprof handler), `options.go` (`Option`/`Options`/
`With*`), `errors.go` (domain error type), `functions.go` (one-shot
capture helpers + `BuildTraceFlightRecorder` + `BuildBlockProfiling`
builders), `internals.go` (`captureNamedProfile` shared helper).

### Override: Lifecycle integration

`TraceFlightRecorder` and `BlockProfiling` both implement
`common/lifecycle.Component` with worker-style semantics: `Start(ctx)`
enables the underlying runtime hook and returns immediately; `Done`
closes when `Stop(ctx)` completes. `BuildTraceFlightRecorder` and
`BuildBlockProfiling` wrap construction with the standard goroutine
+ `CloseFn` pattern (mirrors `cron.BuildScheduler`, `http.BuildServer`,
`grpc.BuildServer`).

### Override: stateless capture helpers in `functions.go`

The four one-shot capture helpers (`CaptureCPUProfile`,
`CaptureHeapProfile`, `CaptureGoroutineProfile`, `CaptureBlockProfile`)
live alongside the lifecycle builders in `functions.go`. They are
stateless and have no lifecycle; the file groups all free functions of
the package as Shape B R1 prescribes.
