package slog

import "context"

// Logger defines the contract for the slog-backed logger.
// Methods are safe for concurrent use.
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

var _ Logger = (*logger)(nil)
