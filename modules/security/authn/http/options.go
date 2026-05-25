package http

import (
	"net/http"
)

// Default settings for the Bearer middleware. The header name and
// scheme are case-insensitive per RFC 7235, but we publish the
// conventional spelling for readability.
const (
	defaultHeaderName = "Authorization"
	defaultScheme     = "Bearer"
)

// defaultErrorHandler writes 401 Unauthorized with an empty body. It is
// installed by NewOptions when the caller does not override via
// WithErrorHandler.
func defaultErrorHandler(w http.ResponseWriter, _ *http.Request, _ error) {
	w.WriteHeader(http.StatusUnauthorized)
}

// Option is a functional option for configuring http middleware
// Options.
type Option func(opts *Options)

// Options holds the configuration for the Bearer middleware.
type Options struct {
	headerName   string
	scheme       string
	errorHandler ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		headerName:   defaultHeaderName,
		scheme:       defaultScheme,
		errorHandler: defaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithHeaderName overrides the request header read by the middleware.
// Empty values are ignored (the default "Authorization" is preserved).
func WithHeaderName(name string) Option {
	return func(opts *Options) {
		if name != "" {
			opts.headerName = name
		}
	}
}

// WithScheme overrides the credential scheme expected as the first
// whitespace-delimited token of the header value. Empty values are
// ignored (the default "Bearer" is preserved). Comparisons are
// case-insensitive.
func WithScheme(scheme string) Option {
	return func(opts *Options) {
		if scheme != "" {
			opts.scheme = scheme
		}
	}
}

// WithErrorHandler installs a custom failure-response handler. Nil
// values are ignored (the default 401-empty handler is preserved).
func WithErrorHandler(handler ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}
