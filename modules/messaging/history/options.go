package history

import (
	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring history Options.
// History has no T-typed options, so Option is non-generic — matching
// bridge/filter and diverging from router/Option[T] is intentional.
type Option func(opts *Options)

// Options holds the configuration for a History endpoint.
type Options struct {
	historyKey   string
	errorHandler messaging.ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default history key is DefaultHistoryKey
// ("History"); the default ErrorHandler is
// messaging.DefaultErrorHandler, which logs via common/log.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		historyKey:   DefaultHistoryKey,
		errorHandler: messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithHistoryKey overrides the Headers.Custom map key used to store
// the trail. Use this when DefaultHistoryKey ("History") would
// collide with another caller-defined entry in Custom. Empty strings
// are ignored (the previously configured key is preserved).
func WithHistoryKey(key string) Option {
	return func(opts *Options) {
		if key != "" {
			opts.historyKey = key
		}
	}
}

// WithErrorHandler installs an observability hook fired once per
// forward Send failure. The default (when WithErrorHandler is not
// passed) is messaging.DefaultErrorHandler, which logs each failure
// via common/log so consumers that forget to wire observability still
// see forward failures. Pass messaging.SilentErrorHandler to opt out,
// or any custom hook to redirect. Nil values are ignored (the
// previously installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
