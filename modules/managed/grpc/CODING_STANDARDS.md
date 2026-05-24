# Coding Standards — modules/managed/grpc/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Server` interface + `server` private impl; `ErrServerFn`, `UnaryInterceptorFn`, `StreamInterceptorFn` aliases in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Server` is public, `server` is private |
| 4 | Constructor returns interface | Yes | `NewServer` returns `Server` |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewServer` call owns its inner `*grpc.Server` |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`google.golang.org/grpc` plus `google.golang.org/protobuf` and
`google.golang.org/genproto/googleapis/rpc` account for the largest share
of `modules/common/`'s transitive footprint by source size. Most consumers
of `common` never use gRPC. For that reason the gRPC wrapper lives at the
top-level module layer alongside `modules/managed/`, `modules/managed/cache/`,
`modules/managed/telemetry/` and `modules/config/`, never inside `modules/common/`.

### Override: Shape B (canonical Shape B layout)

This is a Shape B package — it wraps a third-party library
(`google.golang.org/grpc`) with managed-style server builders and
interceptors. Package layout follows the canonical Shape B template:
`types.go` (interfaces + aliases), `server.go` (private impl + constructor),
`options.go` (`Option`/`Options`/`With*`), `errors.go` (domain error type),
`functions.go` (free interceptor factories).
