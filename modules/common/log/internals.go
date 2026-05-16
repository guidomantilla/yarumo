package log

import (
	"sync/atomic"

	cslog "github.com/guidomantilla/yarumo/common/log/slog"
)

// loggerHolder wraps a Logger inside an atomic.Value cell so that Use can swap
// the package-level current logger without locks.
type loggerHolder struct {
	logger Logger
}

// Package-level singleton storage: current holds the active *loggerHolder
// (lazy-initialised on first read), and internal is the default Logger used
// until the caller invokes Use.
var (
	current  atomic.Value
	internal = cslog.NewLogger()
)

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
