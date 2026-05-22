# Coding Standards

This module follows the conventions defined in [`modules/common/CODING_STANDARDS.md`](../common/CODING_STANDARDS.md) with the overrides documented below.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `Cache[K,V]` interface + `cache` private impl; `CacheFn` and `BuildCacheFn` function types in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | `Cache[K,V]` is public, `cache[K,V]` is private |
| 4 | Constructor returns interface | Yes | `NewCache` and `BuildCache` return `Cache[K,V]` |
| 5 | Options | Yes | `Options` + `With<Field>` functions, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No singleton; each `NewCache` call owns its backend |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Top-level module (not under common/)

The cache backends (`ristretto`, `bigcache`, `go-cache`) pull substantial transitive dependencies. For that reason the cache wrapper lives at the top-level module layer alongside `modules/managed/`, `modules/telemetry/`, and `modules/config/`, never inside `modules/common/`.

### Override: Lifecycle integration with managed

`BuildCache` is the lifecycle-aware constructor: it returns the cache together with a `managed.StopFn` so the cache participates in the managed shutdown chain. `NewCache` is the lifecycle-free constructor used by tests and direct consumers; in that case the caller must call `Cache.Stop` to release backend resources.

### Override: In-memory only

This release implements `ristretto`, `bigcache`, and `go-cache` only. Redis, memcached, chained caches and tag-based invalidation are deferred (see ticket YA-0079 for rationale).

### Override: Pluggable struct backend dispatch (criterion 4 Exception 3)

The private `cache[K, V]` implementation holds the backend behaviour in **function fields** (`getFn`, `setFn`, `deleteFn`, `hasFn`, `clearFn`, `stopFn`) populated by per-backend factories in `backends.go` (`newRistrettoCache`, `newBigcacheCache`, `newGoCacheCache`). This is a concrete instance of Exception 3 from criterion 4 of `modules/common/CODING_STANDARDS.md` — different "implementations" are different instances of the same struct, configured differently. The public `Cache[K, V]` interface stays unchanged.

There is intentionally **no internal `backendCache` interface**. The pattern mirrors crypto's `*Method` (11 packages under `modules/crypto/`).

## Reviewed Packages

- [x] cache
