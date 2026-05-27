// Demo that exercises the public API of the extension/common/log/zerolog
// package:
//
//  1. NewLogger with default options (silent) — LevelOff is the default.
//  2. NewLogger(WithLevel(LevelInfo), WithWriter(stdout)) emits JSON
//     records at info+.
//  3. WithConsole(true) flips on the human-readable zerolog console
//     writer (great for local dev).
//  4. WithSampling samples 1-in-N records — useful for very chatty logs.
package main

import (
	"context"
	"fmt"
	"os"

	czerolog "github.com/guidomantilla/yarumo/extension/common/log/zerolog"

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
	name, version, env := "modules/extension/common/log/zerolog/examples/main.go", "1.0", "examples"
	ctx := config.Default(context.Background(), name, version, env)

	demos := []struct {
		title string
		fn    func(context.Context) error
	}{
		{"Default logger (silent)", demoSilent},
		{"Info-level JSON logger", demoInfo},
		{"Console writer (human-readable)", demoConsole},
		{"Sampling 1-in-N", demoSampling},
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
	logger := czerolog.NewLogger(czerolog.WithWriter(os.Stdout))
	logger.Info(ctx, "this should not appear (LevelOff is the default)")
	fmt.Println("  (no record emitted)")
	return nil
}

// demoInfo wires LevelInfo + stdout and emits one record per severity.
func demoInfo(ctx context.Context) error {
	logger := czerolog.NewLogger(
		czerolog.WithLevel(czerolog.LevelInfo),
		czerolog.WithWriter(os.Stdout),
	)

	logger.Debug(ctx, "filtered (below info)")
	logger.Info(ctx, "user signed in", "user_id", "u-123", "ip", "10.0.0.1")
	logger.Warn(ctx, "quota near limit", "used", 95, "limit", 100)
	logger.Error(ctx, "downstream failed", "service", "billing", "code", 503)

	return nil
}

// demoConsole flips the console writer on top of stdout for a
// human-readable rendering.
func demoConsole(ctx context.Context) error {
	logger := czerolog.NewLogger(
		czerolog.WithLevel(czerolog.LevelInfo),
		czerolog.WithWriter(os.Stdout),
		czerolog.WithConsole(true),
	)

	logger.Info(ctx, "human-friendly", "request_id", "req-42")

	return nil
}

// demoSampling configures 1-in-5 sampling. Only ~2 out of 10 calls
// produce output.
func demoSampling(ctx context.Context) error {
	logger := czerolog.NewLogger(
		czerolog.WithLevel(czerolog.LevelInfo),
		czerolog.WithWriter(os.Stdout),
		czerolog.WithSampling(5),
	)

	fmt.Println("  emitting 10 sampled records (expect ~2 to appear):")
	for i := 1; i <= 10; i++ {
		logger.Info(ctx, "tick", "n", i)
	}

	return nil
}
