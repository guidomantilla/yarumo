# Coding Standards — modules/extension/common/resilience/retry/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../../../core/common/CODING_STANDARDS.md).

This module is the **implementation** of the contract defined in
[`modules/core/common/resilience/retry/`](../../../../core/common/resilience/retry/).
The contract package owns the `Retry` interface, the `Backoff` enum,
the `RetryIfFn` / `OnRetryFn` hook signatures with their default
implementations (`AlwaysRetry`, `NoopOnRetry`), and the domain error
type + sentinels. This package owns the concrete implementation, the
`Option`/`Options` configuration surface, and the `avast/retry-go/v4`
adapter.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ cretry.Retry = (*retry)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `cretry.Retry`; impl `*retry` is private |
| 4 | Constructor returns interface | Yes | `NewRetry(opts ...Option) cretry.Retry` |
| 5 | Options | Yes | `Options` + `WithAttempts` / `WithDelay` / `WithBackoff` / `WithRetryIf` / `WithOnRetry`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewRetry(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Contract split

The `Retry` interface, `Backoff` enum, `RetryIfFn` / `OnRetryFn`
signatures (and their defaults `AlwaysRetry` / `NoopOnRetry`), domain
`Error` type, and sentinels live in
[`core/common/resilience/retry`](../../../../core/common/resilience/retry/).
This package imports them under the short alias `cretry` so the
local impl type `retry` does not clash with the contract package
name.

### Override: No registry, no pluggable

Earlier `extension/common/resilience/` shipped registries and used the
pluggable-struct pattern (function fields populated at construction). This
module deliberately drops both:

- **No registry.** Consumers construct multiple `Retry` instances via
  `NewRetry`. Each protected operation gets its own retry policy at
  bootstrap; the consumer's wiring code holds the references.
- **No pluggable function fields.** The private `retry` struct holds the
  configured options directly; `Do` delegates to `github.com/avast/retry-go/v4`
  with those options. No closures captured at construction.

### Override: Wraps `github.com/avast/retry-go/v4`

The retry loop (attempts, delay, backoff, predicate, hook) comes from the
upstream `avast/retry-go/v4` library. Layout: `types.go` (package doc +
compliance var), `retry.go` (private impl + `NewRetry`), `options.go`
(`Option`/`Options`/`With*`), `internals.go` (`Backoff` → retry-go
adapter). Contract-level files (errors, predicates) live in the core
package.

### Override: Sibling to `extension/common/http/retry/`

`extension/common/http/retry/` wraps an `http.RoundTripper` with retry
semantics — HTTP-transport-specific. This module is the **generic**
counterpart: `Retry.Do(ctx, fn)` retries an arbitrary function. The two
modules cover different layers and may coexist in the same app.
