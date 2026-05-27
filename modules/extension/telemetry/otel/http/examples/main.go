// Demo that exercises the public API of the telemetry/otel/http
// transport decorators:
//
//  1. NewMetricsTransport instruments outgoing requests with a counter +
//     duration histogram via a manually-installed MeterProvider that
//     writes to a manual reader. The demo flushes the reader and prints
//     the recorded data points.
//  2. NewTracingTransport opens a client span per request via a
//     TracerProvider with an in-memory span recorder. The demo dumps the
//     captured span attributes.
//  3. The two transports stack: tracing wraps metrics wraps default
//     transport.
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/guidomantilla/yarumo/config"
	otelhttp "github.com/guidomantilla/yarumo/extension/telemetry/otel/http"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/telemetry/otel/http/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"NewMetricsTransport records counter + histogram", demoMetrics},
		{"NewTracingTransport opens a client span", demoTracing},
		{"Tracing + metrics stacked", demoStacked},
	}

	for _, d := range demos {
		fmt.Printf("=== Demo: %s ===\n", d.title)
		err := d.fn(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", d.title, err)
		}
		fmt.Println()
	}

	return nil
}

// demoMetrics installs a private MeterProvider with a manual reader,
// fires one request through NewMetricsTransport, and dumps the metric
// data points.
func demoMetrics(ctx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	defer func() { _ = provider.Shutdown(ctx) }()

	transport := otelhttp.NewMetricsTransport(
		http.DefaultTransport,
		otelhttp.WithMeterProvider(provider),
		otelhttp.WithName("demo.metrics"),
	)
	client := &http.Client{Transport: transport}

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request %d: %w", i, err)
		}
		resp.Body.Close()
	}

	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	for _, scope := range rm.ScopeMetrics {
		fmt.Printf("  scope=%s\n", scope.Scope.Name)
		for _, m := range scope.Metrics {
			fmt.Printf("    metric=%s (%s)\n", m.Name, m.Description)
		}
	}

	return nil
}

// demoTracing installs a private TracerProvider with an in-memory span
// recorder, fires one request through NewTracingTransport, and dumps
// the captured span(s).
func demoTracing(ctx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = provider.Shutdown(ctx) }()

	transport := otelhttp.NewTracingTransport(
		http.DefaultTransport,
		otelhttp.WithTracerProvider(provider),
		otelhttp.WithName("demo.tracing"),
	)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	resp.Body.Close()

	spans := recorder.Ended()
	fmt.Printf("  recorded %d span(s)\n", len(spans))
	for _, s := range spans {
		fmt.Printf("    name=%q kind=%s status=%s\n", s.Name(), s.SpanKind(), s.Status().Code)
		for _, attr := range s.Attributes() {
			fmt.Printf("      attr %s = %v\n", attr.Key, attr.Value.AsInterface())
		}
	}

	return nil
}

// demoStacked composes tracing on top of metrics on top of the default
// transport, fires a request, and confirms both sides recorded.
func demoStacked(ctx context.Context) error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	reader := metric.NewManualReader()
	meterProvider := metric.NewMeterProvider(metric.WithReader(reader))
	defer func() { _ = meterProvider.Shutdown(ctx) }()

	recorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer func() { _ = tracerProvider.Shutdown(ctx) }()

	stacked := otelhttp.NewTracingTransport(
		otelhttp.NewMetricsTransport(http.DefaultTransport,
			otelhttp.WithMeterProvider(meterProvider),
			otelhttp.WithName("demo.stacked"),
		),
		otelhttp.WithTracerProvider(tracerProvider),
		otelhttp.WithName("demo.stacked"),
	)
	client := &http.Client{Transport: stacked}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	resp.Body.Close()

	var rm metricdata.ResourceMetrics
	_ = reader.Collect(ctx, &rm)

	fmt.Printf("  metrics scopes=%d, spans=%d\n", len(rm.ScopeMetrics), len(recorder.Ended()))
	return nil
}
