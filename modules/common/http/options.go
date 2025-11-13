package http

import (
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/guidomantilla/yarumo/common/utils"
	"golang.org/x/time/rate"
)

type Option func(opts *Options)

type Options struct {
	timeout      time.Duration
	transport    http.RoundTripper
	attempts     uint
	retryIf      retry.RetryIfFunc
	retryHook    retry.OnRetryFunc
	limiterRate  rate.Limit
	limiterBurst int
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		timeout:   30 * time.Second,
		transport: http.DefaultTransport,
		attempts:  1,
		retryIf: func(err error) bool {
			return false // no retry by default
		},
		retryHook: func(_ uint, err error) {
			// do nothing
		},
		limiterRate:  rate.Inf, // unlimited - same as not having a limiter
		limiterBurst: 0,        // unlimited - same as not having a limiter
	}

	for _, opt := range opts {
		opt(options)
	}

	// Hardening: if limiter is enabled (finite rate) and burst <= 0, normalize to a minimal safe burst of 1 to avoid over-restrictive behavior.
	mustHarden := options.limiterRate != rate.Inf && options.limiterBurst <= 0
	if mustHarden {
		options.limiterBurst = 1
	}

	// Timeout alignment: cap selected transport timeouts so they do not exceed the client-level timeout.
	// We only cap non-zero values (0 means no timeout for that hop), and we do not mutate the original transport instance.
	if options.timeout > 0 {
		t, ok := options.transport.(*http.Transport)
		if ok {
			clone := t.Clone()
			clone.TLSHandshakeTimeout = utils.Ternary(clone.TLSHandshakeTimeout > 0 && clone.TLSHandshakeTimeout > options.timeout, options.timeout, clone.TLSHandshakeTimeout)
			clone.ResponseHeaderTimeout = utils.Ternary(clone.ResponseHeaderTimeout > 0 && clone.ResponseHeaderTimeout > options.timeout, options.timeout, clone.ResponseHeaderTimeout)
			clone.ExpectContinueTimeout = utils.Ternary(clone.ExpectContinueTimeout > 0 && clone.ExpectContinueTimeout > options.timeout, options.timeout, clone.ExpectContinueTimeout)

			// Note: DialContext timeout cannot be reliably capped here without replacing the dialer/function. We intentionally avoid overriding
			// DialContext to preserve custom transport.
			options.transport = clone
		}
	}

	return options
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.timeout = timeout
		}
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(opts *Options) {
		if transport != nil {
			opts.transport = transport
		}
	}
}

func WithAttempts(attempts uint) Option {
	return func(opts *Options) {
		if attempts > 1 {
			opts.attempts = attempts
		}
	}
}

func WithRetryIf(retryIf retry.RetryIfFunc) Option {
	return func(opts *Options) {
		if retryIf != nil {
			opts.retryIf = retryIf
		}
	}
}

func WithRetryHook(retryHook retry.OnRetryFunc) Option {
	return func(opts *Options) {
		if retryHook != nil {
			opts.retryHook = retryHook
		}
	}
}

func WithLimiterRate(limiterRate float64) Option {
	return func(opts *Options) {
		if limiterRate != float64(rate.Inf) {
			opts.limiterRate = rate.Limit(limiterRate)
		}
	}
}

func WithLimiterBurst(burst int) Option {
	return func(opts *Options) {
		if burst > 0 {
			opts.limiterBurst = burst
		}
	}
}
