package http

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// metricsRoundTripper wraps an http.RoundTripper and records a per-request
// counter and duration histogram on every RoundTrip. Attributes are
// method / host / path / status.
type metricsRoundTripper struct {
	base       http.RoundTripper
	counter    metric.Int64Counter
	histogram  metric.Float64Histogram
}

// NewMetricsRoundTripper returns an http.RoundTripper that records OTel
// metrics for every request before delegating to base. If base is nil it
// falls back to http.DefaultTransport. The meter is resolved via the
// provided MeterProvider (default: otel.GetMeterProvider()).
//
// On instrument creation failure (e.g. an invalid meter provider) the
// constructor returns the base unchanged — instrumentation is best-effort
// and must not break the client's request path.
func NewMetricsRoundTripper(base http.RoundTripper, opts ...MetricsOption) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	o := NewMetricsOptions(opts...)
	meter := o.meterProvider.Meter(o.meterName)

	counter, err := meter.Int64Counter(
		"http.client.request.count",
		metric.WithDescription("Total number of HTTP client requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return base
	}

	histogram, err := meter.Float64Histogram(
		"http.client.request.duration",
		metric.WithDescription("HTTP client request duration."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return base
	}

	return &metricsRoundTripper{base: base, counter: counter, histogram: histogram}
}

// RoundTrip delegates to the base RoundTripper, recording a counter and
// duration histogram observation regardless of success or failure.
func (m *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := m.base.RoundTrip(req)
	duration := time.Since(start).Seconds()

	attrs := requestAttributes(req, resp, err)
	m.counter.Add(req.Context(), 1, metric.WithAttributes(attrs...))
	m.histogram.Record(req.Context(), duration, metric.WithAttributes(attrs...))

	return resp, err
}

// requestAttributes builds the standard attribute set used by both the
// counter and histogram. status defaults to 0 when the request failed
// before a response was received.
func requestAttributes(req *http.Request, resp *http.Response, err error) []attribute.KeyValue {
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	host := ""
	path := ""
	if req.URL != nil {
		host = req.URL.Host
		path = req.URL.Path
	}
	attrs := []attribute.KeyValue{
		attribute.String("http.method", req.Method),
		attribute.String("http.host", host),
		attribute.String("http.path", path),
		attribute.Int("http.status", status),
	}
	if err != nil {
		attrs = append(attrs, attribute.Bool("http.error", true))
	}
	return attrs
}
