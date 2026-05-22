package log

import "context"

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
