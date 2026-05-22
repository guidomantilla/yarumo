package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
)

// loggerHolder wraps a Logger inside an atomic.Value cell so that Use can
// swap the package-level current logger without locks.
type loggerHolder struct {
	logger Logger
}

// Package-level singleton storage: current holds the active *loggerHolder
// (lazy-initialised on first read), and internal is the default Logger used
// until the caller invokes Use. The default is a noopLogger so that
// modules/common/log does not depend on any concrete implementation; the
// consumer wires a real logger (typically modules/log/slog) by calling Use
// at startup.
var (
	current  atomic.Value
	internal Logger = noopLogger{}
)

// osExit is an indirection to os.Exit to allow error-path testing of the
// noopLogger's Fatal behaviour.
var osExit = os.Exit

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

// noopLogger is the default Logger when no implementation has been
// registered via Use. It discards Trace/Debug/Info/Warn/Error and writes a
// minimal message to stderr before exiting on Fatal so that "no logger
// configured" cannot hide a fatal condition triggered by assert or other
// callers.
type noopLogger struct{}

func (noopLogger) Trace(_ context.Context, _ string, _ ...any) {}
func (noopLogger) Debug(_ context.Context, _ string, _ ...any) {}
func (noopLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (noopLogger) Warn(_ context.Context, _ string, _ ...any)  {}
func (noopLogger) Error(_ context.Context, _ string, _ ...any) {}
func (noopLogger) Fatal(_ context.Context, msg string, _ ...any) {
	fmt.Fprintln(os.Stderr, msg)
	osExit(1)
}
