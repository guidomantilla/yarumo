package log

import (
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
// swap the package-level current logger without locks. Has no methods; its
// only purpose is to be the typed value stored in current.
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
