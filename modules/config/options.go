package config

import (
	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// Option is a functional option for configuring Default's behavior.
type Option func(opts *Options)

// Options holds the configuration applied at Default() call time.
type Options struct {
	logger clog.Logger
}

// NewOptions creates Options with safe defaults and applies the given functional options.
func NewOptions(name string, version string, env string, opts ...Option) *Options {
	options := &Options{
		logger: SlogLogger(name, version, env),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithLogger overrides the default slog-backed logger built from LOG_LEVEL /
// DEBUG env vars. When provided, Default installs this logger as the
// process-global via clog.Use and skips the slog default factory entirely.
// Nil values are ignored.
func WithLogger(logger clog.Logger) Option {
	return func(opts *Options) {
		if logger != nil {
			opts.logger = logger
		}
	}
}
