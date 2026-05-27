// Package main demonstrates common/log: the package-level Trace/Debug/
// Info/Warn/Error helpers, the Use / Default slot swap, and how a custom
// Logger implementation receives all six severity levels. Fatal is
// intentionally NOT demoed — the default noop logger calls os.Exit(1)
// from Fatal so that a missing Use call cannot hide a fatal condition.
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// captureLogger is a Logger that appends each entry to a slice rather
// than writing to stdout/stderr. It exists so the demo's output is
// deterministic and easy to read.
type captureLogger struct {
	mu      sync.Mutex
	entries []string
}

// log records one entry tagged with the given severity.
func (l *captureLogger) log(level, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	pairs := make([]string, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		pairs = append(pairs, fmt.Sprintf("%v=%v", args[i], args[i+1]))
	}

	entry := fmt.Sprintf("[%s] %s", level, msg)
	if len(pairs) > 0 {
		entry = entry + " " + strings.Join(pairs, " ")
	}

	l.entries = append(l.entries, entry)
}

// Trace records a trace-level entry.
func (l *captureLogger) Trace(_ context.Context, msg string, args ...any) {
	l.log("TRACE", msg, args...)
}

// Debug records a debug-level entry.
func (l *captureLogger) Debug(_ context.Context, msg string, args ...any) {
	l.log("DEBUG", msg, args...)
}

// Info records an info-level entry.
func (l *captureLogger) Info(_ context.Context, msg string, args ...any) {
	l.log("INFO", msg, args...)
}

// Warn records a warn-level entry.
func (l *captureLogger) Warn(_ context.Context, msg string, args ...any) {
	l.log("WARN", msg, args...)
}

// Error records an error-level entry.
func (l *captureLogger) Error(_ context.Context, msg string, args ...any) {
	l.log("ERROR", msg, args...)
}

// Fatal records a fatal-level entry without terminating the process —
// the demo never wants to os.Exit on a normal demo path.
func (l *captureLogger) Fatal(_ context.Context, msg string, args ...any) {
	l.log("FATAL", msg, args...)
}

func main() {
	demoDefaultSlot()
	demoCustomLogger()
}

// demoDefaultSlot shows that Default() returns the noop logger when no
// logger has been registered. Calls to Info/Debug/etc. are silently
// discarded — nothing observable.
func demoDefaultSlot() {
	fmt.Println("=== Default slot (noop) ===")

	defaultLogger := clog.Default()
	fmt.Printf("  Default() -> %T\n", defaultLogger)

	clog.Info(context.Background(), "this line is discarded by noop")
	fmt.Println("  Info call returned (output suppressed by noop)")
}

// demoCustomLogger registers a captureLogger and prints every entry
// recorded across all five non-fatal levels.
func demoCustomLogger() {
	fmt.Println("=== Use(captureLogger) ===")

	capture := &captureLogger{}
	clog.Use(capture)

	ctx := context.Background()

	clog.Trace(ctx, "tracing pipeline", "step", "alpha")
	clog.Debug(ctx, "loaded config", "keys", 3)
	clog.Info(ctx, "service started", "addr", "127.0.0.1:8080")
	clog.Warn(ctx, "slow downstream", "latency_ms", 750)
	clog.Error(ctx, "request failed", "code", 502)

	for _, entry := range capture.entries {
		fmt.Printf("  %s\n", entry)
	}
}
