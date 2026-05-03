// Package log provides a structured logging abstraction with support for multiple log levels.
package log

import (
	"context"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
)

var (
	_ Logger = (*cslog.Logger)(nil)
)

// LogFn is the function type for package-level logging functions.
type LogFn func(ctx context.Context, msg string, args ...any)

// UseFn is the function type for Use.
type UseFn func(l Logger)

var (
	_ UseFn = Use
	_ LogFn = Trace
	_ LogFn = Debug
	_ LogFn = Info
	_ LogFn = Warn
	_ LogFn = Error
	_ LogFn = Fatal
)

// Logger defines the interface for structured logging with six severity levels.
type Logger interface {
	// Trace logs a message at trace level.
	Trace(ctx context.Context, msg string, args ...any)
	// Debug logs a message at debug level.
	Debug(ctx context.Context, msg string, args ...any)
	// Info logs a message at info level.
	Info(ctx context.Context, msg string, args ...any)
	// Warn logs a message at warn level.
	Warn(ctx context.Context, msg string, args ...any)
	// Error logs a message at error level.
	Error(ctx context.Context, msg string, args ...any)
	// Fatal logs a message at fatal level and terminates the process.
	Fatal(ctx context.Context, msg string, args ...any)
}
