// Package http provides OpenTelemetry instrumentation for net/http clients
// as composable http.RoundTripper decorators. Two decorators are exposed:
//
//   - NewMetricsRoundTripper: emits per-request counter and duration histogram
//     on the configured meter, labelled by method/host/path/status.
//   - NewTracingRoundTripper: opens a client-kind span per request, injects
//     W3C trace context headers, records HTTP attributes and status.
//
// The decorators implement http.RoundTripper, accept any base RoundTripper,
// and can be stacked in either order. Each emits only its own signal type
// so combining them does not double-instrument.
//
// Recommended import alias by callers: otelhttp (to disambiguate from the
// stdlib net/http and from the OTel contrib package of the same name).
package http

import (
	"net/http"
)

var (
	_ http.RoundTripper = (*metricsRoundTripper)(nil)
	_ http.RoundTripper = (*tracingRoundTripper)(nil)

	_ NewMetricsRoundTripperFn = NewMetricsRoundTripper
	_ NewTracingRoundTripperFn = NewTracingRoundTripper
)

// NewMetricsRoundTripperFn is the function type for NewMetricsRoundTripper.
type NewMetricsRoundTripperFn func(base http.RoundTripper, opts ...MetricsOption) http.RoundTripper

// NewTracingRoundTripperFn is the function type for NewTracingRoundTripper.
type NewTracingRoundTripperFn func(base http.RoundTripper, opts ...TracingOption) http.RoundTripper
