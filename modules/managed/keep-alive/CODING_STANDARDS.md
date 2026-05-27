# Coding Standards — modules/managed/keep-alive/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ lifecycle.Component = (*keepAlive)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `lifecycle.Component` (no module-owned interface); impl `*keepAlive` is private |
| 4 | Constructor returns interface | Yes | `NewKeepAlive(name string) lifecycle.Component` |
| 5 | Options | No | The basic keep-alive carries only a name; no module-owned `Options` struct |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewKeepAlive` call owns its own state |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`modules/core/common/` is a pure library with no lifecycle opinions. The
`KeepAlive` constructor returns a `lifecycle.Component` — Start is a no-op
and Done closes when Stop is called. It exists to serve as the canonical
"basic component" building block for daemons that do not own a network
listener (heartbeats, long-running workers, application keep-alive loops).
For that reason it lives at the top-level module layer alongside
`modules/managed/cron/`, `modules/managed/grpc/`, `modules/managed/http/`,
`modules/managed/diagnostics/` and `modules/managed/telemetry/`, never
inside `modules/core/common/`.

### Override: Exception shape (thin component without owned interface)

This is an Exception package per `modules/PACKAGES.md` — its only purpose
is to provide a ready-made `lifecycle.Component` implementation whose
Start is a no-op. Layout is a minimal `types.go` (package doc + compliance
var) + `keep_alive.go` (the `NewKeepAlive` constructor and private impl).
No `functions.go`, no `errors.go`, no `options.go` — `common/lifecycle`
owns the `Component` contract and the `ErrShutdown` family.

### Override: Lifecycle integration

`keepAlive` implements `common/lifecycle.Component` with worker-style
semantics: `Start` returns nil immediately; `Done` closes when `Stop` is
called. There is no `BuildKeepAlive` wrapper — wire the component into
`lifecycle.Build(ctx, keepAlive, errChan)` directly when daemon-style
goroutine + `CloseFn` is needed.
