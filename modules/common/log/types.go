package log

import "context"

// Type compliance: the package-level functions match their declared Fn
// aliases. Concrete Logger implementations live in separate modules (e.g.
// modules/log/slog) and assert compliance against this interface there.
var (
	_ UseFn     = Use
	_ DefaultFn = Default
	_ LogFn     = Trace
	_ LogFn     = Debug
	_ LogFn     = Info
	_ LogFn     = Warn
	_ LogFn     = Error
	_ LogFn     = Fatal
)

// LogFn is the function type for package-level logging functions.
type LogFn func(ctx context.Context, msg string, args ...any)

// UseFn is the function type for Use.
type UseFn func(l Logger)

// DefaultFn is the function type for Default.
type DefaultFn func() Logger

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
