package http

import (
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"golang.org/x/time/rate"

	"github.com/guidomantilla/yarumo/common/utils"
)

type Option func(opts *Options)

type Options struct {
	clientTimeout         time.Duration
	clientTransport       http.RoundTripper
	clientAttempts        uint
	clientRetryIf         retry.RetryIfFunc
	clientRetryHook       retry.OnRetryFunc
	clientRetryOnResponse RetryOnResponseFn
	clientLimiterRate     rate.Limit
	clientLimiterBurst    uint
}

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
	}

	for _, opt := range opts {
		opt(options)
	}

	// Hardening: if limiter is enabled (finite rate) and burst <= 0, normalize to a minimal safe burst of 1 to avoid over-restrictive behavior.
	mustHarden := options.clientLimiterRate != rate.Inf && options.clientLimiterBurst <= 0
	if mustHarden {
		options.clientLimiterBurst = 1
	}

	// Timeout alignment: cap selected transport timeouts so they do not exceed the client-level timeout.
	// We only cap non-zero values (0 means no timeout for that hop), and we do not mutate the original transport instance.
	if utils.NotEmpty(options.clientTimeout) {
		t, ok := options.clientTransport.(*http.Transport)
		if ok {
			clone := t.Clone()
			clone.TLSHandshakeTimeout = utils.Ternary(clone.TLSHandshakeTimeout > 0 && clone.TLSHandshakeTimeout > options.clientTimeout, options.clientTimeout, clone.TLSHandshakeTimeout)
			clone.ResponseHeaderTimeout = utils.Ternary(clone.ResponseHeaderTimeout > 0 && clone.ResponseHeaderTimeout > options.clientTimeout, options.clientTimeout, clone.ResponseHeaderTimeout)
			clone.ExpectContinueTimeout = utils.Ternary(clone.ExpectContinueTimeout > 0 && clone.ExpectContinueTimeout > options.clientTimeout, options.clientTimeout, clone.ExpectContinueTimeout)

			// Note: DialContext timeout cannot be reliably capped here without replacing the dialer/function. We intentionally avoid overriding
			// DialContext to preserve custom transport.
			options.clientTransport = clone
		}
	}

	return options
}

func WithClientTimeout(clientTimeout time.Duration) Option {
	return func(opts *Options) {
		if clientTimeout > 0 {
			opts.clientTimeout = clientTimeout
		}
	}
}

func WithClientTransport(clientTransport http.RoundTripper) Option {
	return func(opts *Options) {
		if clientTransport != nil {
			opts.clientTransport = clientTransport
		}
	}
}

func WithClientAttempts(clientAttempts uint) Option {
	return func(opts *Options) {
		if clientAttempts > 1 {
			opts.clientAttempts = clientAttempts
		}
	}
}

func WithClientRetryIf(clientRetryIf retry.RetryIfFunc) Option {
	return func(opts *Options) {
		if clientRetryIf != nil {
			opts.clientRetryIf = clientRetryIf
		}
	}
}

func WithClientRetryHook(clientRetryHook retry.OnRetryFunc) Option {
	return func(opts *Options) {
		if clientRetryHook != nil {
			opts.clientRetryHook = clientRetryHook
		}
	}
}

func WithClientRetryOnResponse(clientRetryOnResponse RetryOnResponseFn) Option {
	return func(o *Options) {
		if clientRetryOnResponse != nil {
			o.clientRetryOnResponse = clientRetryOnResponse
		}
	}
}

func WithClientLimiterRate(clientLimiterRate float64) Option {
	return func(opts *Options) {
		if clientLimiterRate != float64(rate.Inf) {
			opts.clientLimiterRate = rate.Limit(clientLimiterRate)
		}
	}
}

func WithClientLimiterBurst(clientLimiterBurst uint) Option {
	return func(opts *Options) {
		if clientLimiterBurst > 0 {
			opts.clientLimiterBurst = clientLimiterBurst
		}
	}
}
