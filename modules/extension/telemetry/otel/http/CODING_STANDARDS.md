# Coding Standards — modules/extension/telemetry/otel/http/

This module follows the workspace-wide standards documented in
[`modules/common/CODING_STANDARDS.md`](../../../../common/CODING_STANDARDS.md).

## Applicable Criteria

| # | Criterion | Applies | Notes |
|---|-----------|---------|-------|
| 1 | Bullet proof review | Yes | |
| 2 | Type Compliance | Yes | `var _ http.RoundTripper = (*metricsTransport)(nil)` and `var _ http.RoundTripper = (*tracingTransport)(nil)` in `types.go` |
| 3 | Public Interface, Private Implementation | Yes | Constructors return `http.RoundTripper`; impls (`*metricsTransport`, `*tracingTransport`) are private |
| 4 | Constructor returns interface | Yes | `NewMetricsTransport(base, opts...) http.RoundTripper`, `NewTracingTransport(base, opts...) http.RoundTripper` |
| 5 | Options | Yes | `MetricsOption`/`MetricsOptions` for metrics; `TracingOption`/`TracingOptions` for tracing — separate Options structs because fields are disjoint |
| 6 | Preconfigured Default Singletons | No | Defaults read from global OTel providers (`otel.GetMeterProvider()`, `otel.GetTracerProvider()`); override via With* |
| 7 | Linter | Yes | |
| 8 | Tests | Yes | |
| 9 | Documentation | Yes | |

## Overrides

### Override: Bridge module (extension/telemetry/otel/<X>/ pattern)

This module is a bridge between an HTTP client middleware and OTel
signals. It pairs `extension/telemetry/otel/slog/` (bridge between slog
and OTel logs) — same pattern: workspace concern ↔ OTel. It does NOT
belong under `modules/extension/common/http/` (where `limiter/` and
`retry/` live) because those middlewares carry orthogonal deps
(`golang.org/x/time`, `avast/retry-go`); this bridge's reason-to-exist is
the OTel SDK dependency, so it lives next to the other OTel bridges.

### Override: Two decorators, one module

`metricsTransport` and `tracingTransport` share the OTel SDK + contrib
dep tree. Splitting into two modules would duplicate `go.mod` /
`.golangci.yml` / boilerplate without saving any footprint. Naming uses
`Metrics`/`Tracing` prefixes (`MetricsOption`, `TracingOption`) to keep
the two pipelines disambiguated within the same package.

### Override: Compose-with-chain by default

Both transports follow the `extension/common/http/{limiter,retry}/`
shape: `NewXxxTransport(base http.RoundTripper, opts ...XxxOption) http.RoundTripper`.
Nil base falls back to `http.DefaultTransport`. The returned RoundTripper
is safe to embed in a `common/http.NewClient(WithTransport(...))` chain.
