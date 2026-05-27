# Coding Standards — modules/extension/common/cache/ristretto/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ ccache.Cache[string, any] = (*ristrettoCache[any])(nil)` and `var _ lifecycle.Component = (*ristrettoCache[any])(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `ccache.Cache[string, V]` from `common/cache`; impl `*ristrettoCache[V]` is private |
| 4 | Constructor returns interface | Yes | `NewRistrettoCache[V](name, opts...) ccache.Cache[string, V]` |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewRistrettoCache` call owns its own ristretto client |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

`github.com/dgraph-io/ristretto/v2` pulls non-trivial transitive
dependencies (`xxhash`, `go-humanize`). Most consumers of `common/` never
use a concrete cache backend. For that reason the ristretto wrapper lives
at the top-level module layer alongside `modules/extension/common/cache/redis/`,
`modules/managed/cron/`, `modules/managed/grpc/`, `modules/managed/http/`,
`modules/managed/diagnostics/`, `modules/managed/keep-alive/` and
`modules/managed/telemetry/`, never inside `modules/core/common/`.

### Override: Shape B (canonical Shape B layout)

This is a Shape B package — it wraps `github.com/dgraph-io/ristretto/v2`
with the managed-component idiom. Layout: `types.go` (compliance vars +
package doc), `cache.go` (private impl + `NewRistrettoCache` constructor),
`options.go` (`Option`/`Options`/`With*`), `errors.go` (domain error type
+ ristretto-specific sentinels: `ErrInit`, `ErrSet`).

### Override: Lifecycle integration

`ristrettoCache` implements `common/lifecycle.Component` with worker-style
semantics: `Start(ctx)` instantiates the underlying ristretto client and
returns immediately; `Done` closes when `Stop(ctx)` completes. Wire the
component into `lifecycle.Build(ctx, c, errChan)` to get the standard
goroutine + `CloseFn` pattern.

### Override: Sibling module for the redis backend

The redis backend lives in `modules/extension/common/cache/redis/`. Sibling
submodules (rather than one combined module) keep deps split: an app that
only needs ristretto does not pull `go-redis`. Shared concepts
(`Cache[K, V]` contract, `Codec`, `ResolveKeyPrefix`) live in
`modules/core/common/cache/`; the 5-sentinel error vocabulary is duplicated
intentionally per backend (no shared parent module).
