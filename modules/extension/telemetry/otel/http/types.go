// Package otelhttp provides http.RoundTripper decorators that add
// OpenTelemetry metrics and tracing instrumentation to any http.Client.
//
// The package exposes two stackable transports:
//
//   - NewMetricsTransport: counter + histogram per request, scoped under
//     `http.client.request.count` / `http.client.request.duration`.
//   - NewTracingTransport: client-kind span per request, W3C trace
//     context propagation, configurable span name and header redaction.
//
// Compose with the rest of the chain through any base RoundTripper. The
// recommended outermost-to-innermost order is tracing → metrics →
// behaviour (retry/limiter) → base, so the trace and metric attributes
// describe the user-visible request including retries:
//
//	import (
//	    chttp        "github.com/guidomantilla/yarumo/core/common/http"
//	    chttplimiter "github.com/guidomantilla/yarumo/extension/common/http/limiter"
//	    chttpretry   "github.com/guidomantilla/yarumo/extension/common/http/retry"
//	    otelhttp     "github.com/guidomantilla/yarumo/extension/telemetry/otel/http"
//	)
//
//	client := chttp.NewClient(chttp.WithTransport(
//	    otelhttp.NewTracingTransport(
//	        otelhttp.NewMetricsTransport(
//	            chttpretry.NewRetryTransport(
//	                chttplimiter.NewLimiterTransport(http.DefaultTransport, lim),
//	            ),
//	        ),
//	    ),
//	))
//
// Defaults pull the global meter and tracer providers
// (`otel.GetMeterProvider()`, `otel.GetTracerProvider()`) so the
// decorators start emitting as soon as the application bootstraps OTel.
// Tests and alternative pipelines override the providers via
// WithMeterProvider / WithTracerProvider.
package otelhttp

import "net/http"

var (
	_ http.RoundTripper = (*metricsTransport)(nil)
	_ http.RoundTripper = (*tracingTransport)(nil)
)

// SpanNameFn produces the span name for a given outgoing request. Used by
// the tracing transport via WithSpanNameFn. Implementations should avoid
// embedding the full URL when high-cardinality span names are a concern.
type SpanNameFn func(req *http.Request) string
