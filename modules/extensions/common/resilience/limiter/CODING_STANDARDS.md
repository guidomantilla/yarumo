# Coding Standards — modules/extensions/common/resilience/limiter/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ Limiter = (*limiter)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `Limiter`; impl `*limiter` is private |
| 4 | Constructor returns interface | Yes | `NewLimiter(opts ...Option) Limiter` |
| 5 | Options | Yes | `Options` + `WithRate` / `WithBurst`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewLimiter(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: No registry, no pluggable

Earlier versions of `extensions/common/resilience/` shipped a
`RateLimiterRegistry` (lazy-create-by-name + `Get`/`Use`) and used the
pluggable-struct pattern (function fields populated at construction). This
module deliberately drops both:

- **No registry.** Consumers that need multiple limiters construct multiple
  `Limiter`s directly via `NewLimiter`. The registry indirection was
  over-engineered for the actual call sites in the workspace.
- **No pluggable function fields.** The private `limiter` struct holds the
  underlying `*rate.Limiter` directly; `Allow`/`Wait` are plain methods
  that delegate. No closures captured at construction.

If a future consumer genuinely needs lazy-by-name discovery, build a
registry on top of `Limiter`; do not push the indirection into this module.

### Override: Wraps `golang.org/x/time/rate.Limiter`

The underlying token-bucket comes from the standard `golang.org/x/time/rate`
library. Layout follows the canonical Shape B template: `types.go`
(interface + Fn aliases + compliance), `limiter.go` (private impl +
`NewLimiter`), `options.go` (`Option`/`Options`/`With*`), `errors.go`
(domain error type + `ErrWait`). Per PACKAGES.md L68 (R2 Excepción
adicional), the constructor `NewLimiter(opts ...Option) Limiter` does
NOT declare an Fn alias or compliance var — the contract is already
fixed by the `Option` type at the entry and by `Limiter` + its
compliance at the output. The `Limiter` interface is owned by this module
and matches the same shape as
`extensions/common/http/limiter.NewLimiterTransport` (which takes a
preconstructed `*rate.Limiter`); the two modules cover different layers
(generic vs HTTP-transport).
