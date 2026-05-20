# Coding Standards — modules/http/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Server` interface + `server` private impl; `ErrServerFn` alias in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Server` is public, `server` is private |
| 4 | Constructor returns interface | Yes | `NewServer` returns `Server` |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewServer` call owns its inner `*http.Server` |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

The HTTP server has lifecycle (Start/Stop, listener goroutines) and therefore
violates `modules/common/`'s no-lifecycle clause. The HTTP client wrapper
(stateless, no lifecycle) remains in `modules/common/http/`; this module
hosts only the server side. Splitting it out keeps `common` lifecycle-free
and lets consumers depend on the server independently of `common/http`.

### Override: Shape B (canonical Shape B layout)

This is a Shape B package — it wraps the stdlib `net/http.Server` with
managed-style timeout defaults and shutdown semantics. Package layout follows
the canonical Shape B template: `types.go` (interface + aliases), `server.go`
(private impl + constructor), `options.go` (`Option`/`Options`/`With*`),
`errors.go` (domain error type).
