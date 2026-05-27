package otelhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	otrace "go.opentelemetry.io/otel/trace"
)

// findSpanAttr returns the string value of the attribute matching key on
// the given snapshot, or "" when missing.
func findSpanAttr(span trace.ReadOnlySpan, key string) string {
	for _, kv := range span.Attributes() {
		if string(kv.Key) == key {
			return kv.Value.Emit()
		}
	}
	return ""
}

// findSpanAttrKV returns the attribute KeyValue matching key, or zero
// value when missing. Lets tests assert on attribute presence + type.
func findSpanAttrKV(span trace.ReadOnlySpan, key string) attribute.KeyValue {
	for _, kv := range span.Attributes() {
		if string(kv.Key) == key {
			return kv
		}
	}
	return attribute.KeyValue{}
}

func TestNewTracingTransport(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil transport", func(t *testing.T) {
		t.Parallel()

		rt := NewTracingTransport(http.DefaultTransport)
		if rt == nil {
			t.Fatal("expected non-nil transport")
		}
	})

	t.Run("falls back to http.DefaultTransport when base is nil", func(t *testing.T) {
		t.Parallel()

		rt := NewTracingTransport(nil)
		if rt == nil {
			t.Fatal("expected non-nil transport from nil base")
		}
	})
}

func TestTracingTransport_OpensClientSpan(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/widgets", http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]

	if span.SpanKind() != otrace.SpanKindClient {
		t.Fatalf("span kind = %v, want client", span.SpanKind())
	}
	if span.Name() != "HTTP GET" {
		t.Fatalf("span name = %q, want %q", span.Name(), "HTTP GET")
	}
	if findSpanAttr(span, "http.method") != http.MethodGet {
		t.Fatalf("http.method = %q, want %q", findSpanAttr(span, "http.method"), http.MethodGet)
	}
	if findSpanAttr(span, "http.path") != "/widgets" {
		t.Fatalf("http.path = %q, want %q", findSpanAttr(span, "http.path"), "/widgets")
	}
	if findSpanAttr(span, "http.status") != "200" {
		t.Fatalf("http.status = %q, want %q", findSpanAttr(span, "http.status"), "200")
	}
	if span.Status().Code != codes.Unset {
		t.Fatalf("expected unset status for 2xx, got %v", span.Status().Code)
	}
}

func TestTracingTransport_InjectsTraceContext(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	var capturedTraceparent string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedTraceparent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	if capturedTraceparent == "" {
		t.Fatal("expected traceparent header to be injected")
	}
	if !strings.HasPrefix(capturedTraceparent, "00-") {
		t.Fatalf("traceparent format unexpected: %q", capturedTraceparent)
	}
}

func TestTracingTransport_StatusErrorOn5xx(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("expected span status Error for 5xx, got %v", spans[0].Status().Code)
	}
}

func TestTracingTransport_StatusErrorOnTransportFailure(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	base := &errRoundTripper{err: errors.New("dial tcp: connection refused")}

	client := &http.Client{
		Transport: NewTracingTransport(base,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://unreachable.test/path", http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	_, doErr := client.Do(req)
	if doErr == nil {
		t.Fatal("expected transport error")
	}

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	if span.Status().Code != codes.Error {
		t.Fatalf("expected span status Error on transport failure, got %v", span.Status().Code)
	}
	if findSpanAttr(span, "http.error") != "true" {
		t.Fatalf("expected http.error=true on transport failure, attrs=%v", span.Attributes())
	}
}

func TestTracingTransport_HeaderRedaction(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
			WithHeaderRedaction("authorization"),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}
	req.Header.Set("Authorization", "Bearer super-secret-token")

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	spans := exporter.GetSpans().Snapshots()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	kv := findSpanAttrKV(spans[0], "http.request.header.Authorization")
	if string(kv.Key) == "" {
		t.Fatal("expected redacted authorization header attr on span")
	}
	if kv.Value.AsString() != "<redacted>" {
		t.Fatalf("expected redacted placeholder, got %q", kv.Value.AsString())
	}
}

func TestTracingTransport_CustomSpanName(t *testing.T) {
	t.Parallel()

	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := &http.Client{
		Transport: NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(provider),
			WithPropagator(propagation.TraceContext{}),
			WithSpanNameFn(func(r *http.Request) string {
				return "outbound:" + r.Method
			}),
		),
	}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodDelete, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	spans := exporter.GetSpans().Snapshots()
	if spans[0].Name() != "outbound:DELETE" {
		t.Fatalf("span name = %q, want %q", spans[0].Name(), "outbound:DELETE")
	}
}

