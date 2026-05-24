package otelhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// errRoundTripper always returns a transport error. Used to exercise the
// "http.error" attribute path on the metrics transport.
type errRoundTripper struct {
	err error
}

func (e *errRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, e.err
}

// collectMetrics returns the freshly-collected ResourceMetrics from reader.
// Tests use this to inspect what the metrics transport recorded.
func collectMetrics(t *testing.T, reader *metric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()

	var rm metricdata.ResourceMetrics
	err := reader.Collect(context.Background(), &rm)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	return rm
}

// findMetric scans the ResourceMetrics for a metric matching name.
// Returns nil when not found so tests can assert presence explicitly.
func findMetric(rm metricdata.ResourceMetrics, name string) *metricdata.Metrics {
	for _, sm := range rm.ScopeMetrics {
		for i := range sm.Metrics {
			if sm.Metrics[i].Name == name {
				return &sm.Metrics[i]
			}
		}
	}
	return nil
}

// findAttr returns the string value of the attribute matching key, or ""
// when missing. Tests fail explicitly when a required attribute is absent.
func findAttr(attrs []attribute.KeyValue, key string) string {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return kv.Value.Emit()
		}
	}
	return ""
}

func TestNewMetricsTransport(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil transport", func(t *testing.T) {
		t.Parallel()

		rt := NewMetricsTransport(http.DefaultTransport)
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("falls back to http.DefaultTransport when base is nil", func(t *testing.T) {
		t.Parallel()

		rt := NewMetricsTransport(nil)
		if rt == nil {
			t.Fatal("expected non-nil transport from nil base")
		}
	})
}

func TestMetricsTransport_RecordsCounterAndHistogram(t *testing.T) {
	t.Parallel()

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewMetricsTransport(http.DefaultTransport, WithMeterProvider(provider)),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/api/v1/widgets", http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	rm := collectMetrics(t, reader)

	counter := findMetric(rm, "http.client.request.count")
	if counter == nil {
		t.Fatal("expected http.client.request.count metric to be recorded")
	}

	sum, ok := counter.Data.(metricdata.Sum[int64])
	if !ok {
		t.Fatalf("expected counter Data to be Sum[int64], got %T", counter.Data)
	}
	if len(sum.DataPoints) != 1 {
		t.Fatalf("expected 1 counter data point, got %d", len(sum.DataPoints))
	}
	if sum.DataPoints[0].Value != 1 {
		t.Fatalf("counter value = %d, want 1", sum.DataPoints[0].Value)
	}

	attrs := sum.DataPoints[0].Attributes.ToSlice()
	if findAttr(attrs, "http.method") != http.MethodGet {
		t.Fatalf("http.method = %q, want %q", findAttr(attrs, "http.method"), http.MethodGet)
	}
	if findAttr(attrs, "http.status") != "200" {
		t.Fatalf("http.status = %q, want %q", findAttr(attrs, "http.status"), "200")
	}
	if findAttr(attrs, "http.path") != "/api/v1/widgets" {
		t.Fatalf("http.path = %q, want %q", findAttr(attrs, "http.path"), "/api/v1/widgets")
	}

	histogram := findMetric(rm, "http.client.request.duration")
	if histogram == nil {
		t.Fatal("expected http.client.request.duration metric to be recorded")
	}

	hist, ok := histogram.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatalf("expected histogram Data to be Histogram[float64], got %T", histogram.Data)
	}
	if len(hist.DataPoints) != 1 {
		t.Fatalf("expected 1 histogram data point, got %d", len(hist.DataPoints))
	}
	if hist.DataPoints[0].Count != 1 {
		t.Fatalf("histogram count = %d, want 1", hist.DataPoints[0].Count)
	}
}

func TestMetricsTransport_RecordsErrorAttribute(t *testing.T) {
	t.Parallel()

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	base := &errRoundTripper{err: errors.New("dial tcp: connection refused")}

	client := &http.Client{
		Transport: NewMetricsTransport(base, WithMeterProvider(provider)),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://unreachable.test/path", http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	_, doErr := client.Do(req)
	if doErr == nil {
		t.Fatal("expected transport error")
	}

	rm := collectMetrics(t, reader)

	counter := findMetric(rm, "http.client.request.count")
	if counter == nil {
		t.Fatal("expected counter to be recorded even on transport failure")
	}

	sum := counter.Data.(metricdata.Sum[int64])
	attrs := sum.DataPoints[0].Attributes.ToSlice()

	if findAttr(attrs, "http.error") != "true" {
		t.Fatalf("expected http.error=true on transport failure, attrs=%v", attrs)
	}
	if findAttr(attrs, "http.status") != "" {
		t.Fatalf("expected http.status to be absent on transport failure, got %q", findAttr(attrs, "http.status"))
	}
}

