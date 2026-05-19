package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newRecordingTracerProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	return tp, exp
}

func TestNewTracingRoundTripper_NilBaseFallsBackToDefault(t *testing.T) {
	t.Parallel()

	rt := NewTracingRoundTripper(nil)
	if rt == nil {
		t.Fatalf("expected non-nil RoundTripper")
	}
}

func TestTracingRoundTripper_EmitsSpanWithHTTPAttributes(t *testing.T) {
	t.Parallel()

	tp, exp := newRecordingTracerProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewTracingRoundTripper(http.DefaultTransport, WithTracingTracerProvider(tp)),
	}

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/widgets", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	spans := exp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	s := spans[0]
	if s.Name() != "HTTP POST" {
		t.Fatalf("span name = %q, want HTTP POST", s.Name())
	}
	if s.SpanKind() != trace.SpanKindClient {
		t.Fatalf("span kind = %v, want client", s.SpanKind())
	}
	want := map[string]string{
		"http.method": "POST",
		"http.path":   "/widgets",
	}
	for _, a := range s.Attributes() {
		key := string(a.Key)
		if exp, ok := want[key]; ok {
			if got := a.Value.Emit(); got != exp {
				t.Fatalf("attr %s = %q, want %q", key, got, exp)
			}
			delete(want, key)
		}
		if key == "http.status" && a.Value.AsInt64() != int64(http.StatusCreated) {
			t.Fatalf("http.status = %v, want %d", a.Value.AsInt64(), http.StatusCreated)
		}
	}
	if len(want) > 0 {
		t.Fatalf("missing attributes: %v", want)
	}
}

func TestTracingRoundTripper_InjectsTraceparent(t *testing.T) {
	t.Parallel()

	tp, _ := newRecordingTracerProvider(t)

	var seenHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenHeader = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewTracingRoundTripper(http.DefaultTransport,
			WithTracingTracerProvider(tp),
			WithTracingPropagator(propagation.TraceContext{}),
		),
	}

	resp, err := client.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	if seenHeader == "" {
		t.Fatalf("expected traceparent header on the outgoing request")
	}
}

func TestTracingRoundTripper_FailureSetsErrorStatus(t *testing.T) {
	t.Parallel()

	tp, exp := newRecordingTracerProvider(t)

	base := &failingRoundTripper{err: errors.New("transport boom")}
	rt := NewTracingRoundTripper(base, WithTracingTracerProvider(tp))

	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/x", nil)
	resp, err := rt.RoundTrip(req)
	if err == nil || resp != nil {
		t.Fatalf("expected transport error, got resp=%v err=%v", resp, err)
	}

	spans := exp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("span status = %v, want Error", spans[0].Status().Code)
	}
}

func TestTracingRoundTripper_4xxResponseSetsErrorStatus(t *testing.T) {
	t.Parallel()

	tp, exp := newRecordingTracerProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewTracingRoundTripper(http.DefaultTransport, WithTracingTracerProvider(tp)),
	}
	resp, err := client.Get(srv.URL + "/missing")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	spans := exp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("span status = %v on 404, want Error", spans[0].Status().Code)
	}
}

func TestTracingRoundTripper_HeaderRedaction(t *testing.T) {
	t.Parallel()

	tp, exp := newRecordingTracerProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewTracingRoundTripper(http.DefaultTransport,
			WithTracingTracerProvider(tp),
			WithTracingHeaderRedaction("Authorization"),
		),
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	spans := exp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	found := false
	for _, a := range spans[0].Attributes() {
		if string(a.Key) == "http.request.header.Authorization" {
			found = true
			if a.Value.Emit() != "<redacted>" {
				t.Fatalf("Authorization not redacted: %v", a.Value.Emit())
			}
		}
	}
	if !found {
		t.Fatalf("expected http.request.header.Authorization attribute")
	}
}

func TestTracingRoundTripper_CustomSpanNameFn(t *testing.T) {
	t.Parallel()

	tp, exp := newRecordingTracerProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: NewTracingRoundTripper(http.DefaultTransport,
			WithTracingTracerProvider(tp),
			WithTracingSpanNameFn(func(method, host, path string) string {
				return method + " " + path
			}),
		),
	}

	resp, err := client.Get(srv.URL + "/foo/bar")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	spans := exp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span")
	}
	if spans[0].Name() != "GET /foo/bar" {
		t.Fatalf("span name = %q, want GET /foo/bar", spans[0].Name())
	}
}

func TestRoundTrippers_StackTogether(t *testing.T) {
	t.Parallel()

	tp, traceExp := newRecordingTracerProvider(t)
	mp, reader := newManualReaderProvider(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Stack: tracing → metrics → base. The two decorators emit independent
	// signals (one span, one set of metrics); stacking does not duplicate.
	transport := NewTracingRoundTripper(
		NewMetricsRoundTripper(http.DefaultTransport, WithMetricsMeterProvider(mp)),
		WithTracingTracerProvider(tp),
	)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL + "/x")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	resp.Body.Close()

	spans := traceExp.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	var rm metricdata.ResourceMetrics
	err = reader.Collect(t.Context(), &rm)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	sawCounter := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.client.request.count" {
				sawCounter = true
			}
		}
	}
	if !sawCounter {
		t.Fatalf("expected http.client.request.count emitted exactly once by the stack")
	}
}
