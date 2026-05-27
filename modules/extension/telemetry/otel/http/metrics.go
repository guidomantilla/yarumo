package otelhttp

import (
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// metricsTransport wraps a base RoundTripper and records a counter and a
// duration histogram per outgoing request. Attributes attached to each
// record: http.method, http.host, http.path, http.status (omitted on
// transport failure), http.error=true (only on transport failure).
type metricsTransport struct {
	base      http.RoundTripper
	counter   metric.Int64Counter
	histogram metric.Float64Histogram
}

// NewMetricsTransport wraps base with a RoundTripper that records a
// per-request counter (http.client.request.count) and duration histogram
// (http.client.request.duration) via the meter resolved from the shared
// Options. Only the meterProvider + name fields of Options are
// consulted; the tracing-only fields are ignored.
//
// A nil base falls back to http.DefaultTransport. Instrument construction
// failures are silently swallowed: the returned transport degrades to a
// pass-through that delegates to base without recording. This keeps the
// client functional when the meter provider is misconfigured at bootstrap.
//
// The returned RoundTripper is safe for concurrent use as long as base is.
func NewMetricsTransport(base http.RoundTripper, opts ...Option) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	options := NewOptions(opts...)
	meter := options.meterProvider.Meter(options.name)

	counter, _ := meter.Int64Counter(
		"http.client.request.count",
		metric.WithDescription("Total HTTP client requests made."),
		metric.WithUnit("{request}"),
	)
	histogram, _ := meter.Float64Histogram(
		"http.client.request.duration",
		metric.WithDescription("HTTP client request duration."),
		metric.WithUnit("s"),
	)

	return &metricsTransport{
		base:      base,
		counter:   counter,
		histogram: histogram,
	}
}

// RoundTrip delegates to base and records counter + histogram with
// attributes describing the request and its outcome.
func (t *metricsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cassert.NotNil(t, "metrics transport receiver is nil")
	cassert.NotNil(req, "request is nil")

	start := time.Now()
	res, err := t.base.RoundTrip(req)
	elapsed := time.Since(start).Seconds()

	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.host", req.URL.Host),
		attribute.String("http.path", req.URL.Path),
	}

	if err != nil {
		attrs = append(attrs, attribute.Bool("http.error", true))
	} else if res != nil {
		attrs = append(attrs, attribute.String("http.status", strconv.Itoa(res.StatusCode)))
	}

	t.counter.Add(req.Context(), 1, metric.WithAttributes(attrs...))
	t.histogram.Record(req.Context(), elapsed, metric.WithAttributes(attrs...))

	return res, err
}
