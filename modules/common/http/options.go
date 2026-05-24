package http

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring HTTP client Options.
type Option func(opts *Options)

// Options holds the configuration applied at client construction time.
// Fields are unexported; callers configure them through the With* functions
// and the consumer reads the final values through NewClient.
type Options struct {
	timeout   time.Duration
	transport http.RoundTripper
}

// NewOptions creates Options with secure defaults and applies the given
// functional options. Defaults:
//
//   - timeout: 30s (stdlib's *http.Client zero value is 0, meaning no
//     timeout — every consumer reinvents this; we standardize.)
//   - transport: http.DefaultTransport
//
// When the configured transport is a *http.Transport, its internal
// timeouts (TLSHandshake, ResponseHeader, ExpectContinue) are capped to
// the overall client timeout so a stalled handshake cannot exceed the
// request budget. The original transport is cloned, not mutated.
//
// Rate limiting is intentionally NOT wired here. Compose the transport
// chain manually via NewLimiterTransport (and any retry/tracing/etc.
// transports) and inject the assembled RoundTripper through
// WithTransport. This keeps the order of middlewares explicit at the
// call site.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		timeout:   30 * time.Second,
		transport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(options)
	}

	t, ok := options.transport.(*http.Transport)
	if options.timeout > 0 && ok {
		clone := t.Clone()
		if clone.TLSHandshakeTimeout > 0 {
			clone.TLSHandshakeTimeout = min(clone.TLSHandshakeTimeout, options.timeout)
		}

		if clone.ResponseHeaderTimeout > 0 {
			clone.ResponseHeaderTimeout = min(clone.ResponseHeaderTimeout, options.timeout)
		}

		if clone.ExpectContinueTimeout > 0 {
			clone.ExpectContinueTimeout = min(clone.ExpectContinueTimeout, options.timeout)
		}

		options.transport = clone
	}

	return options
}

// WithTimeout sets the overall per-request timeout for the *http.Client
// that NewClient produces. Non-positive values are ignored.
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.timeout = timeout
		}
	}
}

// WithTransport sets the RoundTripper. Pass the assembled middleware
// chain (limiter, retry, tracing, etc.) here so the order of wrappers is
// explicit. Nil values are ignored.
func WithTransport(transport http.RoundTripper) Option {
	return func(opts *Options) {
		if transport != nil {
			opts.transport = transport
		}
	}
}
