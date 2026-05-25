// Package zerolog provides a structured logger implementation built on top of
// github.com/rs/zerolog.
//
// The package exposes a single constructor, NewLogger, which returns a
// clog.Logger implementation backed by zerolog. Construct an instance via
// NewLogger(opts ...Option) and wire it as the process logger via clog.Use.
// There is no registry, no global singleton, and no pluggable function
// fields — callers configure each instance explicitly through Options.
//
// Concurrency: every method on the returned Logger is safe for concurrent
// use by multiple goroutines.
package zerolog

import (
	clog "github.com/guidomantilla/yarumo/common/log"
)

// Interface compliance: this package's private logger satisfies the Logger
// interface declared by modules/common/log.
var (
	_ clog.Logger = (*logger)(nil)
)
