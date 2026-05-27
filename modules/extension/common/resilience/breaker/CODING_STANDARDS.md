# Coding Standards — modules/extension/common/resilience/breaker/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ Breaker = (*breaker)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `Breaker`; impl `*breaker` is private |
| 4 | Constructor returns interface | Yes | `NewBreaker(opts ...Option) Breaker` |
| 5 | Options | Yes | `Options` + `WithName` / `WithMaxRequests` / `WithInterval` / `WithTimeout` / `WithConsecutiveFailures` / `WithOnStateChange`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | No registry, no facade, no singleton. Callers construct instances directly via `NewBreaker(opts...)`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

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

Per PACKAGES.md L68 (R2 Excepción adicional), the constructor
`NewBreaker(opts ...Option) Breaker` does NOT declare an Fn alias or
compliance var — the contract is fixed by the `Option` type at the entry
and by `Breaker` + its compliance at the output. Per L32 the methods
`Execute` and `State` do not get Fn aliases either.

### Override: Wraps `github.com/sony/gobreaker`

The state machine (Closed → Open → Half-Open → Closed) and the failure
counting come from `sony/gobreaker`. Layout follows the canonical Shape B
template: `types.go` (interface + Fn aliases + compliance), `states.go`
(State enum + String), `breaker.go` (private impl + `NewBreaker`),
`options.go` (`Option`/`Options`/`With*`), `errors.go` (domain error type
+ `ErrBreaker`), `predicates.go` (default `OnStateChangeFn` hook),
`internals.go` (gobreaker-to-domain mappers).

### Override: Sibling to `extension/common/resilience/{limiter,retry}`

Same parent directory, same Shape B clean pattern, same no-registry / no-
pluggable rule. The three together cover the common outbound-call
resilience patterns: throttle (limiter), retry (retry), fail-fast (breaker).
Consumers wire each one explicitly at bootstrap.
