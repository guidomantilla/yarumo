package http

import (
	"crypto/tls"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"golang.org/x/time/rate"
)

// Option is a functional option for configuring HTTP Client and Server Options.
type Option func(opts *Options)

// Options holds the configuration for HTTP Client and Server construction.
type Options struct {
	clientTimeout         time.Duration
	clientTransport       http.RoundTripper
	clientAttempts        uint
	clientRetryIf         retry.RetryIfFunc
	clientRetryHook       retry.OnRetryFunc
	clientRetryOnResponse RetryOnResponseFn
	clientLimiterRate     rate.Limit
	clientLimiterBurst    uint

	clientRetryIfSet         bool
	clientRetryOnResponseSet bool

	serverReadHeaderTimeout time.Duration
	serverReadTimeout       time.Duration
	serverWriteTimeout      time.Duration
	serverIdleTimeout       time.Duration
	serverMaxHeaderBytes    int
	serverTLSConfig         *tls.Config
}

// NewOptions creates Options with secure defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		clientTimeout:         30 * time.Second,
		clientTransport:       http.DefaultTransport,
		clientAttempts:        1,
		clientRetryIf:         NoopRetryIf,
		clientRetryHook:       NoopRetryHook,
		clientRetryOnResponse: NoopRetryOnResponse,
		clientLimiterRate:     rate.Inf, // unlimited - same as not having a limiter
		clientLimiterBurst:    0,        // unlimited - same as not having a limiter

		serverReadHeaderTimeout: 5 * time.Second,
		serverReadTimeout:       15 * time.Second,
		serverWriteTimeout:      15 * time.Second,
		serverIdleTimeout:       60 * time.Second,
		serverMaxHeaderBytes:    1 << 20, // 1 MiB
		serverTLSConfig:         nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	// client configuration
	{
		// Auto-wire: if retryOnResponse was explicitly configured but retryIf was not,
		// default retryIf to RetryIfHttpError so response-based retries actually trigger retries.
		if options.clientRetryOnResponseSet && !options.clientRetryIfSet {
			options.clientRetryIf = RetryIfHttpError
		}

		// Hardening: if limiter is enabled (finite rate) and burst <= 0, normalize to a minimal safe burst of 1 to avoid over-restrictive behavior.
		mustHarden := options.clientLimiterRate != rate.Inf && options.clientLimiterBurst <= 0
		if mustHarden {
			options.clientLimiterBurst = 1
		}

		// Timeout alignment: cap selected transport timeouts so they do not exceed the client-level timeout.
		// We only cap non-zero values (0 means no timeout for that hop), and we do not mutate the original transport instance.
		t, ok := options.clientTransport.(*http.Transport)
		if options.clientTimeout > 0 && ok {
			clone := t.Clone()
			if clone.TLSHandshakeTimeout > 0 {
				clone.TLSHandshakeTimeout = min(clone.TLSHandshakeTimeout, options.clientTimeout)
			}

			if clone.ResponseHeaderTimeout > 0 {
				clone.ResponseHeaderTimeout = min(clone.ResponseHeaderTimeout, options.clientTimeout)
			}

			if clone.ExpectContinueTimeout > 0 {
				clone.ExpectContinueTimeout = min(clone.ExpectContinueTimeout, options.clientTimeout)
			}
			// Note: DialContext timeout cannot be reliably capped here without replacing the dialer/function. We intentionally avoid overriding
			// DialContext to preserve custom transport.
			options.clientTransport = clone
		}
	}

	// server configuration
	{
		// Nothing to do here yet
	}

	return options
}

/*
 * Client configuration
 */

// WithClientTimeout sets the overall per-request timeout for the HTTP client.
func WithClientTimeout(clientTimeout time.Duration) Option {
	return func(opts *Options) {
		if clientTimeout > 0 {
			opts.clientTimeout = clientTimeout
		}
	}
}

// WithClientTransport sets the HTTP transport (RoundTripper) for the client.
func WithClientTransport(clientTransport http.RoundTripper) Option {
	return func(opts *Options) {
		if clientTransport != nil {
			opts.clientTransport = clientTransport
		}
	}
}

// WithClientAttempts sets the maximum number of attempts for retryable requests.
func WithClientAttempts(clientAttempts uint) Option {
	return func(opts *Options) {
		if clientAttempts > 1 {
			opts.clientAttempts = clientAttempts
		}
	}
}

// WithClientRetryIf sets the function that decides whether an error is retryable.
func WithClientRetryIf(clientRetryIf retry.RetryIfFunc) Option {
	return func(opts *Options) {
		if clientRetryIf != nil {
			opts.clientRetryIf = clientRetryIf
			opts.clientRetryIfSet = true
		}
	}
}

// WithClientRetryHook sets the hook function called on each retry attempt.
func WithClientRetryHook(clientRetryHook retry.OnRetryFunc) Option {
	return func(opts *Options) {
		if clientRetryHook != nil {
			opts.clientRetryHook = clientRetryHook
		}
	}
}

// WithClientRetryOnResponse sets the function that decides whether a response should trigger a retry.
func WithClientRetryOnResponse(clientRetryOnResponse RetryOnResponseFn) Option {
	return func(opts *Options) {
		if clientRetryOnResponse != nil {
			opts.clientRetryOnResponse = clientRetryOnResponse
			opts.clientRetryOnResponseSet = true
		}
	}
}

// WithClientLimiterRate sets the token bucket rate for client-side rate limiting.
func WithClientLimiterRate(clientLimiterRate float64) Option {
	return func(opts *Options) {
		if clientLimiterRate > 0 && clientLimiterRate != float64(rate.Inf) {
			opts.clientLimiterRate = rate.Limit(clientLimiterRate)
		}
	}
}

// WithClientLimiterBurst sets the token bucket burst size for client-side rate limiting.
func WithClientLimiterBurst(clientLimiterBurst uint) Option {
	return func(opts *Options) {
		if clientLimiterBurst > 0 {
			opts.clientLimiterBurst = clientLimiterBurst
		}
	}
}

/*
 * Server configuration
 */

// WithServerReadHeaderTimeout sets the maximum duration for reading request headers.
func WithServerReadHeaderTimeout(serverReadHeaderTimeout time.Duration) Option {
	return func(opts *Options) {
		if serverReadHeaderTimeout > 0 {
			opts.serverReadHeaderTimeout = serverReadHeaderTimeout
		}
	}
}

// WithServerReadTimeout sets the maximum duration for reading the entire request.
func WithServerReadTimeout(serverReadTimeout time.Duration) Option {
	return func(opts *Options) {
		if serverReadTimeout > 0 {
			opts.serverReadTimeout = serverReadTimeout
		}
	}
}

// WithServerWriteTimeout sets the maximum duration for writing the response.
func WithServerWriteTimeout(serverWriteTimeout time.Duration) Option {
	return func(opts *Options) {
		if serverWriteTimeout > 0 {
			opts.serverWriteTimeout = serverWriteTimeout
		}
	}
}

// WithServerIdleTimeout sets the maximum duration to wait for the next request on a keep-alive connection.
func WithServerIdleTimeout(serverIdleTimeout time.Duration) Option {
	return func(opts *Options) {
		if serverIdleTimeout > 0 {
			opts.serverIdleTimeout = serverIdleTimeout
		}
	}
}

// WithServerMaxHeaderBytes sets the maximum number of bytes the server reads parsing request headers.
func WithServerMaxHeaderBytes(serverMaxHeaderBytes int) Option {
	return func(opts *Options) {
		if serverMaxHeaderBytes > 0 {
			opts.serverMaxHeaderBytes = serverMaxHeaderBytes
		}
	}
}

// WithServerTLSConfig sets the TLS configuration for the server.
func WithServerTLSConfig(serverTLSConfig *tls.Config) Option {
	return func(opts *Options) {
		if serverTLSConfig != nil {
			opts.serverTLSConfig = serverTLSConfig
		}
	}
}
