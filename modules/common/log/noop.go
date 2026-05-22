package log

import (
	"context"
	"fmt"
	"os"
)

// noopLogger is the default Logger when no implementation has been
// registered via Use. It discards Trace/Debug/Info/Warn/Error and writes
// the message to stderr before exiting on Fatal so that a missing Use call
// cannot hide a fatal condition triggered by assert or any other caller.
type noopLogger struct{}

// Trace discards the message.
func (noopLogger) Trace(_ context.Context, _ string, _ ...any) {}

// Debug discards the message.
func (noopLogger) Debug(_ context.Context, _ string, _ ...any) {}

// Info discards the message.
func (noopLogger) Info(_ context.Context, _ string, _ ...any) {}

// Warn discards the message.
func (noopLogger) Warn(_ context.Context, _ string, _ ...any) {}

// Error discards the message.
func (noopLogger) Error(_ context.Context, _ string, _ ...any) {}

// Fatal writes the message to stderr and exits the process with code 1
// via osExit (declared in internals.go for test override).
func (noopLogger) Fatal(_ context.Context, msg string, _ ...any) {
	fmt.Fprintln(os.Stderr, msg)
	osExit(1)
}
