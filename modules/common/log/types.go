// Package log provides a structured logging abstraction with support for
// multiple log levels via a process-global default Logger swappable through
// Use. Concrete Logger implementations live in separate modules (e.g.
// modules/log/slog) and depend on this package, never the reverse.
//
// Until Use is called, the default slot serves a noopLogger that discards
// Trace/Debug/Info/Warn/Error and exits the process on Fatal (writing the
// message to stderr first) so that "no logger configured" cannot hide a
// fatal condition triggered by assert or any other caller.
//
// Tests in this package are intentionally serial (no t.Parallel()) because
// they mutate the slot; downstream packages and the modules/log subpackages
// own no global state and run fully parallel.
package log

import "context"

// Type compliance for the package-level logging functions and the bundled
// Logger interface. Concrete Logger implementations live in separate
// modules and assert compliance against this interface there.
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

// Logger defines the interface for structured logging with six severity
// levels. Implementations must be safe for concurrent use by multiple
// goroutines.
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
