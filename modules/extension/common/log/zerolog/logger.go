package zerolog

import (
	"context"
	"os"

	"github.com/rs/zerolog"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// osExit is an indirection to os.Exit to allow error-path testing. Tests
// may override this variable within the package to simulate failures.
var osExit = os.Exit

// logger is a thin wrapper around zerolog.Logger that satisfies the
// clog.Logger contract.
type logger struct {
	internal zerolog.Logger
}

// NewLogger returns a clog.Logger backed by this package's zerolog
// implementation. Defaults: silent (LevelOff), JSON writer to os.Stderr,
// RFC3339Nano timestamps, no sampling.
func NewLogger(options ...Option) clog.Logger {
	opts := NewOptions(options...)

	zerolog.TimeFieldFormat = opts.timeFormat

	zl := zerolog.New(opts.effectiveWriter()).
		Level(opts.level.toZerolog()).
		With().
		Timestamp().
		Logger()

	if opts.sampling > 1 {
		zl = zl.Sample(&zerolog.BasicSampler{N: opts.sampling})
	}

	return &logger{internal: zl}
}

// Trace logs a message at the trace level, including optional key-value
// pairs for additional context.
func (l *logger) Trace(ctx context.Context, msg string, args ...any) {
	emit(l.internal.Trace(), ctx, msg, args...)
}

// Debug logs a message at the debug level, including optional key-value
// pairs for additional context.
func (l *logger) Debug(ctx context.Context, msg string, args ...any) {
	emit(l.internal.Debug(), ctx, msg, args...)
}

// Info logs a message at the info level, including optional key-value
// pairs for additional context.
func (l *logger) Info(ctx context.Context, msg string, args ...any) {
	emit(l.internal.Info(), ctx, msg, args...)
}

// Warn logs a message at the warn level, including optional key-value
// pairs for additional context.
func (l *logger) Warn(ctx context.Context, msg string, args ...any) {
	emit(l.internal.Warn(), ctx, msg, args...)
}

// Error logs a message at the error level, including optional key-value
// pairs for additional context.
func (l *logger) Error(ctx context.Context, msg string, args ...any) {
	emit(l.internal.Error(), ctx, msg, args...)
}

// Fatal logs a message at the fatal level, including optional key-value
// pairs for additional context, and exits the program immediately.
func (l *logger) Fatal(ctx context.Context, msg string, args ...any) {
	// We do not call zerolog's WithLevel(FatalLevel) `.Msg()` path because
	// zerolog would invoke os.Exit itself, bypassing the test seam. Emit
	// via a Fatal-level event and call osExit afterwards.
	emit(l.internal.WithLevel(zerolog.FatalLevel), ctx, msg, args...)
	osExit(1)
}
