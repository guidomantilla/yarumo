# Coding Standards — modules/extension/common/resilience/limiter/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../../../core/common/CODING_STANDARDS.md).

This module is the **implementation** of the contract defined in
[`modules/core/common/resilience/limiter/`](../../../../core/common/resilience/limiter/).
The contract package owns the `Limiter` interface and the domain
error type + sentinels. This package owns the concrete implementation,
the `Option`/`Options` configuration surface, and the
`golang.org/x/time/rate` adapter.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ climiter.Limiter = (*limiter)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `climiter.Limiter`; impl `*limiter` is private |
| 4 | Constructor returns interface | Yes | `NewLimiter(opts ...Option) climiter.Limiter` |
| 5 | Options | Yes | `Options` + `WithRate` / `WithBurst`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewLimiter(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Contract split

The `Limiter` interface, domain `Error` type, and sentinels live in
[`core/common/resilience/limiter`](../../../../core/common/resilience/limiter/).
This package imports them under the short alias `climiter` so the
local impl type `limiter` does not clash with the contract package
name.

### Override: No registry, no pluggable

Earlier versions of `extension/common/resilience/` shipped a
`RateLimiterRegistry` (lazy-create-by-name + `Get`/`Use`) and used the
pluggable-struct pattern (function fields populated at construction). This
module deliberately drops both:

- **No registry.** Consumers that need multiple limiters construct multiple
  `Limiter`s directly via `NewLimiter`. The registry indirection was
  over-engineered for the actual call sites in the workspace.
- **No pluggable function fields.** The private `limiter` struct holds the
  underlying `*rate.Limiter` directly; `Allow`/`Wait` are plain methods
  that delegate. No closures captured at construction.

### Override: Wraps `golang.org/x/time/rate.Limiter`

The underlying token-bucket comes from the standard `golang.org/x/time/rate`
library. Layout: `types.go` (package doc + compliance var), `limiter.go`
(private impl + `NewLimiter`), `options.go` (`Option`/`Options`/`With*`).
Contract-level files (errors) live in the core package.
