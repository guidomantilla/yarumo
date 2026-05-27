// Demo that exercises the public API of the extensions/common/log/slog
// package:
//
//  1. NewLogger with default options (silent) — proves the off-by-default
//     posture.
//  2. NewLogger(WithLevel(LevelInfo), WithWriter(stdout)) emits a JSON
//     record. Demonstrates Trace/Debug/Info/Warn/Error.
//  3. Custom slog.Handler injected via WithHandlers.
//  4. WithContextExtractors enriches every record with attrs derived
//     from ctx.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	cslog "github.com/guidomantilla/yarumo/extensions/common/log/slog"

	"github.com/guidomantilla/yarumo/config"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	name, version, env := "modules/extensions/common/log/slog/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Default logger (silent)", demoSilent},
		{"Info-level JSON logger", demoInfo},
		{"Custom slog.Handler via WithHandlers", demoCustomHandler},
		{"WithContextExtractors", demoContextExtractor},
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

// demoSilent constructs a logger with default options. Default level is
// LevelOff, so no record should be written.
func demoSilent(ctx context.Context) error {
	logger := cslog.NewLogger(cslog.WithWriter(os.Stdout))
	logger.Info(ctx, "this should not appear (LevelOff is the default)")
	fmt.Println("  (no record emitted)")
	return nil
}

// demoInfo wires LevelInfo + stdout writer and emits one record per
// severity. Trace/Debug are filtered out by the level threshold.
func demoInfo(ctx context.Context) error {
	logger := cslog.NewLogger(
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithWriter(os.Stdout),
	)

	logger.Debug(ctx, "filtered (below info)")
	logger.Info(ctx, "user signed in", "user_id", "u-123", "ip", "10.0.0.1")
	logger.Warn(ctx, "quota near limit", "used", 95, "limit", 100)
	logger.Error(ctx, "downstream failed", "service", "billing", "code", 503)

	return nil
}

// demoCustomHandler injects a stdlib slog.TextHandler. The bundled
// fanout sends each record to every registered handler — useful when a
// service writes to multiple sinks (e.g. JSON to stdout + text to file).
func demoCustomHandler(ctx context.Context) error {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := cslog.NewLogger(
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithHandlers(handler),
	)

	logger.Info(ctx, "text-format record", "key", "value")

	return nil
}

// demoContextExtractor registers an AttrExtractor that reads a value
// stored on ctx and surfaces it on every record.
func demoContextExtractor(ctx context.Context) error {
	type requestIDKey struct{}

	extractor := func(c context.Context) []slog.Attr {
		v, ok := c.Value(requestIDKey{}).(string)
		if !ok {
			return nil
		}
		return []slog.Attr{slog.String("request_id", v)}
	}

	logger := cslog.NewLogger(
		cslog.WithLevel(cslog.LevelInfo),
		cslog.WithWriter(os.Stdout),
		cslog.WithContextExtractors(extractor),
	)

	enrichedCtx := context.WithValue(ctx, requestIDKey{}, "req-42")
	logger.Info(enrichedCtx, "operation complete", "op", "create_user")

	return nil
}
