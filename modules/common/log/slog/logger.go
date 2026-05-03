package slog

import (
	"context"
	"log/slog"
	"os"
)

// osExit is an indirection to os.Exit to allow error-path testing.
// Tests may override this variable within the package to simulate failures.
var osExit = os.Exit

// Logger is a thin wrapper around slog.Logger.
type Logger struct {
	internal *slog.Logger
}

// NewLogger returns a new Logger.
func NewLogger(options ...Option) *Logger {
	opts := NewOptions(options...)

	handlers := opts.handlers
	if len(handlers) == 0 {
		handlerOpts := &slog.HandlerOptions{
			Level:       opts.level.toSlog(),
			ReplaceAttr: ReplaceLevel,
		}
		handlers = append(handlers, slog.NewJSONHandler(opts.writer, handlerOpts))
	}

	return &Logger{
		internal: slog.New(NewFanoutHandler(handlers...)),
	}
}

// Trace logs a message at the trace level, including optional key-value pairs for additional context.
func (l *Logger) Trace(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelTrace.toSlog(), msg, args...)
}

// Debug logs a message at the debug level, including optional key-value pairs for additional context.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelDebug.toSlog(), msg, args...)
}

// Info logs a message at the info level, including optional key-value pairs for additional context.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelInfo.toSlog(), msg, args...)
}

// Warn logs a message at the warn level, including optional key-value pairs for additional context.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelWarn.toSlog(), msg, args...)
}

// Error logs a message at the error level, including optional key-value pairs for additional context.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelError.toSlog(), msg, args...)
}

// Fatal logs a message at the fatal level, including optional key-value pairs for additional context, and exits the program immediately.
func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.internal.Log(ctx, LevelFatal.toSlog(), msg, args...)
	osExit(1)
}
