package headerfilter

import (
	"slices"

	"github.com/guidomantilla/yarumo/messaging"
)

// Option is a functional option for configuring headerfilter Options.
// Option is non-generic — headerfilter has no T-typed options.
type Option func(opts *Options)

// Options holds the configuration for a HeaderFilter.
type Options struct {
	headersToClear []string
	errorHandler   messaging.ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. The default ErrorHandler is
// messaging.DefaultErrorHandler (logs via common/log); the default
// headers-to-clear list is empty (no-op: forwards messages unchanged).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		headersToClear: nil,
		errorHandler:   messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithClearHeader appends a single header name to the clear list. The
// name may be either a recognised Headers struct field (zeroed) or an
// arbitrary string (deleted from Headers.Custom). Empty names are
// ignored. Calling WithClearHeader multiple times accumulates names;
// duplicates are deduplicated at registration time.
func WithClearHeader(name string) Option {
	return func(opts *Options) {
		if name == "" {
			return
		}

		if slices.Contains(opts.headersToClear, name) {
			return
		}

		opts.headersToClear = append(opts.headersToClear, name)
	}
}

// WithHeadersToClear is a variadic convenience for adding multiple
// header names in one call. Equivalent to calling WithClearHeader once
// per name. Empty names are ignored; duplicates are deduplicated.
func WithHeadersToClear(names ...string) Option {
	return func(opts *Options) {
		for _, name := range names {
			if name == "" {
				continue
			}

			if slices.Contains(opts.headersToClear, name) {
				continue
			}

			opts.headersToClear = append(opts.headersToClear, name)
		}
	}
}

// WithErrorHandler installs an observability hook fired once per
// forward Send failure. The default (when WithErrorHandler is not
// passed) is messaging.DefaultErrorHandler, which logs each failure via
// common/log so consumers that forget to wire observability still see
// forward failures. Pass messaging.SilentErrorHandler to opt out, or
// any custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
