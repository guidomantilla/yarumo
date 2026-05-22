package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
)

// Package-level singleton storage: current holds the active *loggerHolder
// (lazy-initialised on first read), internal is the default Logger used
// until the caller invokes Use, and osExit is an indirection to os.Exit so
// noopLogger.Fatal can be tested without terminating the test process.
var (
	current  atomic.Value
	internal Logger = noopLogger{}
	osExit         = os.Exit
)

// loggerHolder wraps a Logger inside an atomic.Value cell so that Use can
// swap the package-level current logger without locks.
type loggerHolder struct {
	logger Logger
}

// load returns the currently active Logger, initialising current with the
// internal default on first access.
func load() Logger {
	value := current.Load()
	if value == nil {
		current.Store(&loggerHolder{logger: internal})
		return internal
	}

	holder, _ := value.(*loggerHolder)
	return holder.logger
}

// Use sets the default logger.
func Use(logger Logger) {
	if logger == nil {
		return
	}

	current.Store(&loggerHolder{logger: logger})
}

// Default returns the currently active default logger, initialising the
// slot with the internal noop logger on first access. Callers that need to
// capture-and-restore the slot (typically tests) should pair Default with
// Use.
func Default() Logger {
	return load()
}

// Trace logs a message at trace level.
func Trace(ctx context.Context, msg string, args ...any) {
	load().Trace(ctx, msg, args...)
}

// Debug logs a message at debug level.
func Debug(ctx context.Context, msg string, args ...any) {
	load().Debug(ctx, msg, args...)
}

// Info logs a message at info level.
func Info(ctx context.Context, msg string, args ...any) {
	load().Info(ctx, msg, args...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, msg string, args ...any) {
	load().Warn(ctx, msg, args...)
}

// Error logs a message at error level.
func Error(ctx context.Context, msg string, args ...any) {
	load().Error(ctx, msg, args...)
}

// Fatal logs a message at fatal level and exits the program immediately.
func Fatal(ctx context.Context, msg string, args ...any) {
	load().Fatal(ctx, msg, args...)
}

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

// Fatal writes the message to stderr and exits the process with code 1.
func (noopLogger) Fatal(_ context.Context, msg string, _ ...any) {
	fmt.Fprintln(os.Stderr, msg)
	osExit(1)
}
