// Demo that exercises every public entry point of modules/managed/telemetry/otel:
//
//  1. Resources                 — builds an OpenTelemetry *resource.Resource
//                                 with service.name / service.version /
//                                 deployment.environment plus host + SDK +
//                                 OTEL_RESOURCE_ATTRIBUTES env attributes.
//                                 Pure setup, no network.
//  2. Observe                   — wires logger + tracer + meter providers
//                                 against an OTLP/gRPC collector and returns
//                                 a single CloseFn that flushes + closes all
//                                 three. Returns a *context.Context* that the
//                                 caller MUST propagate downstream.
//  3. Tracer / Meter / Logger   — same as Observe but installed individually,
//                                 useful for tests or for processes that only
//                                 need one signal.
//  4. Options                   — the With* knobs (WithEndpoint, WithInsecure,
//                                 WithMeterInterval, WithMeterRuntimeMetricsEnabled,
//                                 etc.) that compose the provider config.
//
// NOTE: Observe/Tracer/Meter/Logger talk to an OTLP/gRPC collector. The
// constructor returns without contacting the endpoint (the gRPC client is
// lazy), so this demo runs cleanly even without a collector — the CloseFn
// returns once the configured timeout elapses. To send real telemetry,
// point WithEndpoint at a live collector (default port: 4317).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/guidomantilla/yarumo/config"
	telemetry "github.com/guidomantilla/yarumo/telemetry/otel"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/managed/telemetry/otel/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Resources (standalone resource builder)", demoResources},
		{"Observe (full stack: logger + tracer + meter)", demoObserve},
		{"Tracer (single-signal installer)", demoTracer},
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

// demoResources builds an *resource.Resource. This is pure setup and does not
// contact the collector. Useful to inspect what gets attached to every span /
// metric / log line.
func demoResources(ctx context.Context) error {
	res, err := telemetry.Resources(ctx, "demo-service", "1.0.0", "examples")
	if err != nil {
		return fmt.Errorf("Resources: %w", err)
	}

	fmt.Printf("[resources] schema URL: %s\n", res.SchemaURL())
	fmt.Printf("[resources] %d attributes attached:\n", res.Len())
	for _, attr := range res.Attributes() {
		fmt.Printf("            %s = %v\n", attr.Key, attr.Value.Emit())
	}

	return nil
}

// demoObserve wires the full telemetry stack with WithInsecure and a default
// endpoint. The constructor does not block on the collector handshake (gRPC
// clients connect lazily on first export), so this demo runs without a live
// collector — the CloseFn drops un-flushed data after the configured timeout.
func demoObserve(ctx context.Context) error {
	_, closeFn, err := telemetry.Observe(ctx, "demo-service", "1.0.0", "examples",
		telemetry.WithInsecure(),
		telemetry.WithEndpoint("localhost:4317"),
		telemetry.WithMeterInterval(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("Observe: %w", err)
	}

	fmt.Printf("[observe] logger + tracer + meter providers installed\n")
	fmt.Printf("[observe] endpoint=localhost:4317 (insecure)\n")
	fmt.Printf("[observe] in real apps: defer closeFn(ctx, 15*time.Second)\n")

	// In a real service this lives next to main; we close immediately for the
	// one-shot demo. Without a live collector the close returns once the per-
	// provider timeout elapses (sub-second with the default settings).
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	closeFn(shutdownCtx, 2*time.Second)
	fmt.Printf("[observe] closeFn returned; providers torn down\n")

	return nil
}

// demoTracer installs only the tracer provider — same shape as Observe but
// scoped to a single signal. Useful when a process needs traces but no
// metrics / logs through OTel.
func demoTracer(ctx context.Context) error {
	stopFn, err := telemetry.Tracer(ctx,
		telemetry.WithInsecure(),
		telemetry.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return fmt.Errorf("Tracer: %w", err)
	}

	fmt.Printf("[tracer] tracer provider installed\n")
	fmt.Printf("[tracer] use otel.Tracer(\"name\").Start(ctx, \"span\") to emit\n")

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	stopFn(shutdownCtx, 2*time.Second)
	fmt.Printf("[tracer] stopFn returned; tracer torn down\n")

	return nil
}
