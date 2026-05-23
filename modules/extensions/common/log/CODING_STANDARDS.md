# Coding Standards — modules/extensions/common/log/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Logger` interface + `LogFn`/`UseFn` function types in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Logger` interface lives in `log/`; default impl is `slog.Logger` (concrete struct) |
| 4 | Constructor returns interface | Partial | `log.Use`/`log.Trace`/... operate on the `Logger` interface; `slog.NewLogger` returns the concrete `*slog.Logger` (per CODING_STANDARDS.md criterion 4 exception for pluggable struct impls). |
| 5 | Options | Yes | `slog.Options` + `slog.With*` constructors with validation |
| 6 | Preconfigured Default Singletons | Yes | `log/` carries a process-global default logger slot (`current`/`internal` in `internals.go`); see `doc.go` for the lifecycle contract |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`modules/common/` is a pure library and its packages must remain free of
process-level lifecycle state. The root `log/` package carries a
process-global default logger slot that callers `Use` once during startup.
For symmetry with `modules/telemetry/otel/` (its primary downstream
consumer) and to keep the swap-on-startup contract obvious from the
import path, `log/` lives at the top-level module layer.

### Override: Serial tests in `log/functions_test.go`

The root package tests mutate the process-global `current`/`internal`
logger slot via `Use` / `withLogger`. Running them with `t.Parallel()`
would race against any other test in the same package that observes the
slot. The subpackages (`slog/`, `slog/slogctx/`) own no global state and
their tests are fully parallel.

## Sub-packages

- `log/` — `Logger` interface + package-level helpers (`Use`, `Trace`,
  `Debug`, `Info`, `Warn`, `Error`, `Fatal`) backed by a process-global
  default slot.
- `log/slog/` — concrete `Logger` implementation on top of `log/slog`,
  with handler config (`NewFanoutHandler`, `NewContextHandler`),
  Options (`WithLevel`/`WithWriter`/`WithHandlers`/`WithContextExtractors`)
  and the `SlogctxExtractor` bridge.
- `log/slog/slogctx/` — context-bound attribute bag
  (`WithAttrs`/`SetAttrs`/`Attrs`) read by `SlogctxExtractor`.

## Reviewed Packages

- [x] log
- [x] log/slog
- [x] log/slog/slogctx
