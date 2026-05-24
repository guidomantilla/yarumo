# Coding Standards — modules/extensions/common/http/breaker/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ http.RoundTripper = (*breakerTransport)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Returns `http.RoundTripper`; impl `*breakerTransport` is private |
| 4 | Constructor returns interface | Yes | `NewBreakerTransport(base, rbreaker.Breaker, opts...) http.RoundTripper` |
| 5 | Options | Yes | `Options` + `WithFailOnResponse`, defaults via `NewOptions` |
| 6 | Preconfigured Default Singletons | No | Callers construct instances directly via `NewBreakerTransport`. |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Thin adapter on resilience.Breaker

This module is a **thin adapter**: it accepts a pre-configured
`resilience.Breaker` and wraps each `RoundTrip` call in `breaker.Execute`.
The state machine (Closed → Open → Half-Open → Closed) and all the
configuration (consecutive failures, timeout, half-open probe budget,
state-change hook) live in the resilience module. This module owns only:

- Synthesizing a `*StatusCodeError` when `FailOnResponseFn` returns true,
  so the breaker observes 5xx/429 responses as failures.
- Translating the breaker domain errors back into HTTP-friendly behavior.

### Override: No registry, no pluggable

Callers construct multiple `Breaker` instances via `rbreaker.NewBreaker`
and pass them by DI. The HTTP transport stays small (one struct, one
method) per PACKAGES.md L60 Shape B.

Per PACKAGES.md L68 (R2 Excepción adicional) the constructor
`NewBreakerTransport(base http.RoundTripper, b rbreaker.Breaker, opts ...Option) http.RoundTripper`
does NOT declare an Fn alias or compliance var — the contract is fixed by
`http.RoundTripper` + its compliance var.

### Override: Sibling to `extensions/common/http/{limiter,retry}/`

Same module layout, same shape, same no-registry / no-pluggable rule.
Each of the three covers one resilience primitive at the HTTP transport
layer; the generic primitives live in
`extensions/common/resilience/{breaker,limiter,retry}/` and these modules
are thin adapters over them.
