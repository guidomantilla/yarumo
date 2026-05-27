// Demo that exercises the public API of the telemetry/otel/slog
// bridge:
//
//  1. WithOtelTrace wires the trace<->log extractor into a slog logger.
//     When a span is active in ctx, every record carries trace_id,
//     span_id and trace_flags.
//  2. TraceExtractor() can be passed to WithContextExtractors directly
//     when a caller needs to mix it with other extractors.
//  3. NewBridge installs a slog.Handler that re-emits records through
//     the OTel Logs API (no-op until a real LoggerProvider is wired).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"

	"github.com/guidomantilla/yarumo/config"
	cslog "github.com/guidomantilla/yarumo/extension/common/log/slog"
	otelslog "github.com/guidomantilla/yarumo/extension/telemetry/otel/slog"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extension/telemetry/otel/slog/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"WithOtelTrace: trace IDs surface on every record", demoWithOtelTrace},
		{"TraceExtractor mixed with custom extractors", demoCustomMix},
		{"NewBridge: slog -> OTel Logs API (silently noop)", demoBridge},
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

// fakeSpanContext returns a deterministic span context for the demo
// (no real tracer SDK needed).
func fakeSpanContext() trace.SpanContext {
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("1112131415161718")
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
}

// demoWithOtelTrace builds a slog logger wired with WithOtelTrace,
// then logs once under a ctx that carries a span. The output should
// include trace_id / span_id / trace_flags attributes.
func demoWithOtelTrace(_ context.Context) error {
	logger := cslog.NewLogger(
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithWriter(os.Stdout),
		otelslog.WithOtelTrace(),
	)

	ctx := trace.ContextWithSpanContext(context.Background(), fakeSpanContext())

	logger.Info(ctx, "request processed", "op", "create_user")

	return nil
}

// demoCustomMix shows that TraceExtractor composes with caller-defined
// extractors via cslog.WithContextExtractors.
func demoCustomMix(_ context.Context) error {
	type requestIDKey struct{}

	requestExtractor := func(c context.Context) []slog.Attr {
		v, ok := c.Value(requestIDKey{}).(string)
		if !ok {
			return nil
		}
		return []slog.Attr{slog.String("request_id", v)}
	}

	logger := cslog.NewLogger(
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithWriter(os.Stdout),
		cslog.WithContextExtractors(otelslog.TraceExtractor(), requestExtractor),
	)

	ctx := trace.ContextWithSpanContext(context.Background(), fakeSpanContext())
	ctx = context.WithValue(ctx, requestIDKey{}, "req-42")

	logger.Info(ctx, "mixed extractors", "op", "list_items")

	return nil
}

// demoBridge constructs a NewBridge handler. Without a real OTel
// LoggerProvider configured globally, the bridge resolves the global
// noop provider — records are silently dropped. The demo prints how to
// wire one for inspection.
func demoBridge(ctx context.Context) error {
	handler := otelslog.NewBridge("demo.scope", slog.LevelInfo)

	std := slog.New(handler)
	std.InfoContext(ctx, "this record flows through OTel Logs API", "op", "demo")

	fmt.Println("  (no output: the global OTel LoggerProvider is noop until otel.SetLoggerProvider is called)")
	return nil
}
