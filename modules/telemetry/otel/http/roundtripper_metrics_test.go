package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// failingRoundTripper returns the configured error on every call.
type failingRoundTripper struct{ err error }

func (f *failingRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, f.err
}

// newManualReaderProvider builds a MeterProvider with a manual reader so the
// test can synchronously collect the recorded metric data.
func newManualReaderProvider(t *testing.T) (*sdkmetric.MeterProvider, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	return sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)), reader
}

// findAttribute looks up key in attrs and returns the matching value as a
// string ("<missing>" when absent). Used by metric assertions.
func findAttribute(attrs []attribute.KeyValue, key string) string {
	for _, a := range attrs {
		if string(a.Key) == key {
			return a.Value.Emit()
		}
	}
	return "<missing>"
}

func TestNewMetricsRoundTripper_NilBaseFallsBackToDefault(t *testing.T) {
	t.Parallel()

	rt := NewMetricsRoundTripper(nil)
	if rt == nil {
		t.Fatalf("expected non-nil RoundTripper")
	}
}

func TestMetricsRoundTripper_RecordsCounterAndDurationOnSuccess(t *testing.T) {
	t.Parallel()

	provider, reader := newManualReaderProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewMetricsRoundTripper(http.DefaultTransport, WithMetricsMeterProvider(provider)),
	}

	resp, err := client.Get(srv.URL + "/ping")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	var sawCounter, sawHistogram bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch m.Name {
			case "http.client.request.count":
				sawCounter = true
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok || len(sum.DataPoints) == 0 {
					t.Fatalf("counter has no data points: %+v", m.Data)
				}
				if got := findAttribute(sum.DataPoints[0].Attributes.ToSlice(), "http.status"); got != "200" {
					t.Fatalf("http.status attr = %q, want 200", got)
				}
			case "http.client.request.duration":
				sawHistogram = true
			}
		}
	}
	if !sawCounter {
		t.Fatalf("expected http.client.request.count to be recorded")
	}
	if !sawHistogram {
		t.Fatalf("expected http.client.request.duration to be recorded")
	}
}

func TestMetricsRoundTripper_FailureSetsErrorAttribute(t *testing.T) {
	t.Parallel()

	provider, reader := newManualReaderProvider(t)

	base := &failingRoundTripper{err: errors.New("transport boom")}
	rt := NewMetricsRoundTripper(base, WithMetricsMeterProvider(provider))

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	resp, err := rt.RoundTrip(req)
	if err == nil || resp != nil {
		t.Fatalf("expected transport error, got resp=%v err=%v", resp, err)
	}

	var rm metricdata.ResourceMetrics
	collectErr := reader.Collect(t.Context(), &rm)
	if collectErr != nil {
		t.Fatalf("collect: %v", collectErr)
	}

	found := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.client.request.count" {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok || len(sum.DataPoints) == 0 {
				continue
			}
			if findAttribute(sum.DataPoints[0].Attributes.ToSlice(), "http.error") == "true" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected http.error=true on the counter attributes")
	}
}

func TestMetricsRoundTripper_AttributesIncludeMethodHostPath(t *testing.T) {
	t.Parallel()

	provider, reader := newManualReaderProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewMetricsRoundTripper(http.DefaultTransport, WithMetricsMeterProvider(provider)),
	}

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/things/42", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != "http.client.request.count" {
				continue
			}
			sum := m.Data.(metricdata.Sum[int64])
			attrs := sum.DataPoints[0].Attributes.ToSlice()
			if got := findAttribute(attrs, "http.method"); got != "PUT" {
				t.Fatalf("http.method = %q, want PUT", got)
			}
			if got := findAttribute(attrs, "http.path"); got != "/things/42" {
				t.Fatalf("http.path = %q, want /things/42", got)
			}
			if got := findAttribute(attrs, "http.status"); got != "202" {
				t.Fatalf("http.status = %q, want 202", got)
			}
		}
	}
}
