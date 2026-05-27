# Coding Standards — modules/managed/http/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Server` interface + `server` private impl; `BuildServerFn`, `ErrServerFn` aliases in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Server` is public, `server` is private |
| 4 | Constructor returns interface | Yes | `NewServer` returns `Server` |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewServer` call owns its inner `*http.Server` |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

The HTTP server has lifecycle: it launches a goroutine listening on a
socket and must be cleanly shut down via `Shutdown`/`Close`. That violates
the "no lifecycle" clause of `modules/core/common/`. The HTTP server wrapper
lives at the top-level module layer alongside `modules/managed/grpc/`,
`modules/managed/cron/`, `modules/managed/`, `modules/extension/common/cache/`, `modules/managed/telemetry/`
and `modules/config/`. The HTTP **client** (stateless) stays in
`modules/core/common/http/` — only the server moved out.

### Override: Shape B (canonical Shape B layout)

Server is a Shape B package — it wraps the stdlib `net/http.Server` with
the managed-component idiom. Layout: `types.go` (Server interface + Fn
aliases + compliance vars), `server.go` (private impl + constructor),
`options.go` (`Option`/`Options`/`With*`), `errors.go` (domain error type),
`functions.go` (BuildServer free function).

### Override: Lifecycle integration

`Server` implements `common/lifecycle.Component` with server-style
semantics: `Start(ctx)` blocks calling `Serve`/`ServeTLS` until shutdown;
`Done` closes when `Start` returns. `BuildServer` wraps construction with
the standard goroutine + `CloseFn` pattern.
