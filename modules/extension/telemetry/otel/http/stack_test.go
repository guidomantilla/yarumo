package otelhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestStack_TracingOutsideMetrics verifies that both decorators record
// their signals when stacked tracing-outside-metrics (the recommended
// order).
func TestStack_TracingOutsideMetrics(t *testing.T) {
	t.Parallel()

	metricReader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))

	traceExporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traceExporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	transport := NewTracingTransport(
		NewMetricsTransport(http.DefaultTransport,
			WithMeterProvider(meterProvider),
		),
		WithTracerProvider(tracerProvider),
		WithPropagator(propagation.TraceContext{}),
	)

	client := &http.Client{Transport: transport}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	rm := collectMetrics(t, metricReader)
	if findMetric(rm, "http.client.request.count") == nil {
		t.Fatal("expected metrics to be recorded under stacked transport")
	}

	if len(traceExporter.GetSpans().Snapshots()) != 1 {
		t.Fatal("expected one span recorded under stacked transport")
	}
}

// TestStack_MetricsOutsideTracing verifies that both decorators still
// record their signals when stacked in the opposite order (metrics on
// the outside). Order independence is part of the contract.
func TestStack_MetricsOutsideTracing(t *testing.T) {
	t.Parallel()

	metricReader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))

	traceExporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traceExporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	transport := NewMetricsTransport(
		NewTracingTransport(http.DefaultTransport,
			WithTracerProvider(tracerProvider),
			WithPropagator(propagation.TraceContext{}),
		),
		WithMeterProvider(meterProvider),
	)

	client := &http.Client{Transport: transport}

	req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
	if reqErr != nil {
		t.Fatalf("NewRequest: %v", reqErr)
	}

	res, doErr := client.Do(req)
	if doErr != nil {
		t.Fatalf("Do: %v", doErr)
	}
	_ = res.Body.Close()

	rm := collectMetrics(t, metricReader)
	counter := findMetric(rm, "http.client.request.count")
	if counter == nil {
		t.Fatal("expected metrics to be recorded under stacked transport")
	}

	sum := counter.Data.(metricdata.Sum[int64])
	if sum.DataPoints[0].Value != 1 {
		t.Fatalf("counter value = %d, want 1", sum.DataPoints[0].Value)
	}

	if len(traceExporter.GetSpans().Snapshots()) != 1 {
		t.Fatal("expected one span recorded under stacked transport")
	}
}

// TestStack_AccumulatesAcrossCalls verifies that repeated calls accumulate
// counter/histogram correctly and emit one span per call (no shared state
// leaking across requests).
func TestStack_AccumulatesAcrossCalls(t *testing.T) {
	t.Parallel()

	metricReader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))

	traceExporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(traceExporter))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	transport := NewTracingTransport(
		NewMetricsTransport(http.DefaultTransport,
			WithMeterProvider(meterProvider),
		),
		WithTracerProvider(tracerProvider),
		WithPropagator(propagation.TraceContext{}),
	)
	client := &http.Client{Transport: transport}

	const calls = 3
	for range calls {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
		if reqErr != nil {
			t.Fatalf("NewRequest: %v", reqErr)
		}

		res, doErr := client.Do(req)
		if doErr != nil {
			t.Fatalf("Do: %v", doErr)
		}
		_ = res.Body.Close()
	}

	rm := collectMetrics(t, metricReader)
	counter := findMetric(rm, "http.client.request.count")
	sum := counter.Data.(metricdata.Sum[int64])
	if sum.DataPoints[0].Value != calls {
		t.Fatalf("counter value = %d, want %d", sum.DataPoints[0].Value, calls)
	}

	if len(traceExporter.GetSpans().Snapshots()) != calls {
		t.Fatalf("span count = %d, want %d", len(traceExporter.GetSpans().Snapshots()), calls)
	}
}
