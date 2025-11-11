package http

import (
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
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
