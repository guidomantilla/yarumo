# Coding Standards — modules/extension/common/resilience/breaker/

This module follows the workspace-wide standards documented in
[`modules/core/common/CODING_STANDARDS.md`](../../../../core/common/CODING_STANDARDS.md).

This module is the **implementation** of the contract defined in
[`modules/core/common/resilience/breaker/`](../../../../core/common/resilience/breaker/).
The contract package owns the `Breaker` interface, the `State` enum,
the `OnStateChangeFn` hook signature, and the domain error type +
sentinels. This package owns the concrete implementation, the
`Option`/`Options` configuration surface, and the `gobreaker` adapter.

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ cbreaker.Breaker = (*breaker)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `cbreaker.Breaker`; impl `*breaker` is private |
| 4 | Constructor returns interface | Yes | `NewBreaker(opts ...Option) cbreaker.Breaker` |
| 5 | Options | Yes | `Options` + `WithMaxRequests` / `WithInterval` / `WithTimeout` / `WithConsecutiveFailures` / `WithOnStateChange`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewBreaker(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Contract split

The `Breaker` interface, `State` enum, `OnStateChangeFn` signature,
domain `Error` type, and sentinels live in
[`core/common/resilience/breaker`](../../../../core/common/resilience/breaker/).
This package imports them under the short alias `cbreaker` so the
local impl type `breaker` does not clash with the contract package
name.

### Override: No registry, no pluggable

Earlier `extension/common/resilience/` shipped registries and used the
pluggable-struct pattern (function fields populated at construction). This
module deliberately drops both:

- **No registry.** Consumers construct multiple `Breaker` instances via
  `NewBreaker`. Each protected resource gets its own breaker at bootstrap;
  the consumer's wiring code holds the references.
- **No pluggable function fields.** The private `breaker` struct holds the
  underlying `*gobreaker.CircuitBreaker` directly; `Execute`/`State`
  delegate. No closures captured at construction.

### Override: Wraps `github.com/sony/gobreaker`

The state machine (Closed → Open → Half-Open → Closed) and the failure
counting come from `sony/gobreaker`. Layout: `types.go` (package doc +
compliance var), `breaker.go` (private impl + `NewBreaker`),
`options.go` (`Option`/`Options`/`With*`), `internals.go`
(gobreaker-to-contract mappers). Contract-level files (errors,
predicates, states) live in the core package.

### Override: Sibling to `extension/common/resilience/{limiter,retry}`

Same parent directory, same Shape B clean pattern, same no-registry / no-
pluggable rule. The three together cover the common outbound-call
resilience patterns: throttle (limiter), retry (retry), fail-fast (breaker).
Consumers wire each one explicitly at bootstrap.
