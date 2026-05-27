package limiter

import (
	"time"

	"golang.org/x/time/rate"
)

// Default values for limiter configuration. The defaults yield ~10 rps
// (one token every 100ms) with a burst capacity of 10.
const (
	// DefaultInterval is the token-bucket refill period when the caller
	// does not configure a rate explicitly. One token per DefaultInterval.
	DefaultInterval = 100 * time.Millisecond
	// DefaultBurst is the token-bucket burst capacity when the caller does
	// not configure it explicitly.
	DefaultBurst = 10
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied to a Limiter at construction
// time. Fields are unexported; callers configure them through With*.
type Options struct {
	interval time.Duration
	burst    int
}

// NewOptions creates Options with safe defaults and applies the given
// functional options. Defaults: ~10 rps (one token every 100ms), burst 10.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		interval: DefaultInterval,
		burst:    DefaultBurst,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithRate configures the rate as `perInterval` tokens per `interval`.
// For example, WithRate(10, time.Second) yields 10 rps. Both arguments
// must be positive; otherwise the call is a no-op and defaults are
// preserved.
func WithRate(perInterval int, interval time.Duration) Option {
	return func(opts *Options) {
		if perInterval <= 0 || interval <= 0 {
			return
		}
		opts.interval = interval / time.Duration(perInterval)
	}
}

// WithBurst configures the token-bucket burst capacity. Non-positive
// values are ignored, preserving the default.
func WithBurst(burst int) Option {
	return func(opts *Options) {
		if burst > 0 {
			opts.burst = burst
		}
	}
}

// rateLimit converts the configured interval into a rate.Limit value
// suitable for rate.NewLimiter.
func (o *Options) rateLimit() rate.Limit {
	if o.interval <= 0 {
		return rate.Every(DefaultInterval)
	}

	return rate.Every(o.interval)
}
