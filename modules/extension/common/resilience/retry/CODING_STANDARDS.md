# Coding Standards — modules/extension/common/resilience/retry/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ Retry = (*retry)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `Retry`; impl `*retry` is private |
| 4 | Constructor returns interface | Yes | `NewRetry(opts ...Option) Retry` |
| 5 | Options | Yes | `Options` + `WithAttempts` / `WithDelay` / `WithBackoff` / `WithRetryIf` / `WithOnRetry`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewRetry(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

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

Per PACKAGES.md L68 (R2 Excepción adicional), the constructor
`NewRetry(opts ...Option) Retry` does NOT declare an Fn alias or
compliance var — the contract is fixed by the `Option` type at the entry
and by `Retry` + its compliance at the output.

### Override: Wraps `github.com/avast/retry-go/v4`

The retry loop (attempts, delay, backoff, predicate, hook) comes from the
upstream `avast/retry-go/v4` library. Layout follows the canonical Shape B
template: `types.go` (interface + Fn aliases + compliance), `retry.go`
(private impl + `NewRetry`), `options.go` (`Option`/`Options`/`With*`),
`errors.go` (domain error type + `ErrRetry`).

### Override: Sibling to `extension/common/http/retry/`

`extension/common/http/retry/` wraps an `http.RoundTripper` with retry
semantics — HTTP-transport-specific. This module is the **generic**
counterpart: `Retry.Do(ctx, fn)` retries an arbitrary function. The two
modules cover different layers and may coexist in the same app.
