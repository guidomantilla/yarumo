package resilience

import (
	"time"

	"golang.org/x/time/rate"
)

// Default values for circuit breaker and rate limiter configuration.
const (
	// DefaultCBMaxRequests is the max number of probes allowed in half-open
	// state before the breaker closes or re-opens.
	DefaultCBMaxRequests uint32 = 3
	// DefaultCBInterval is the cyclic period in closed state at which internal
	// failure counters are cleared.
	DefaultCBInterval = 60 * time.Second
	// DefaultCBTimeout is the time the breaker stays in open state before
	// transitioning to half-open on the next call.
	DefaultCBTimeout = 15 * time.Second
	// DefaultCBConsecutiveFailures is the number of consecutive failures that
	// trips the breaker from closed to open.
	DefaultCBConsecutiveFailures uint32 = 5
	// DefaultRateLimitInterval is the token-bucket refill period for the rate
	// limiter (one token every DefaultRateLimitInterval).
	DefaultRateLimitInterval = 100 * time.Millisecond
	// DefaultRateLimitBurst is the token-bucket burst size for the rate limiter.
	DefaultRateLimitBurst = 5
)

// Option is a functional option for configuring resilience Options.
type Option func(opts *Options)

// Options holds the configuration for a circuit breaker or rate limiter entry.
type Options struct {
	// circuit breaker configuration
	cbMaxRequests         uint32
	cbInterval            time.Duration
	cbTimeout             time.Duration
	cbConsecutiveFailures uint32

	// rate limiter configuration
	rateInterval time.Duration
	rateBurst    int
}

// NewOptions creates Options with defaults and applies the given functional options.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		cbMaxRequests:         DefaultCBMaxRequests,
		cbInterval:            DefaultCBInterval,
		cbTimeout:             DefaultCBTimeout,
		cbConsecutiveFailures: DefaultCBConsecutiveFailures,

		rateInterval: DefaultRateLimitInterval,
		rateBurst:    DefaultRateLimitBurst,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithCircuitBreakerMaxRequests sets the max number of probes allowed in
// half-open state.
func WithCircuitBreakerMaxRequests(maxRequests uint32) Option {
	return func(opts *Options) {
		if maxRequests > 0 {
			opts.cbMaxRequests = maxRequests
		}
	}
}

// WithCircuitBreakerInterval sets the cyclic period that resets the breaker's
// internal counters while it is in the closed state.
func WithCircuitBreakerInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.cbInterval = interval
		}
	}
}

// WithCircuitBreakerTimeout sets the time the breaker stays open before
// transitioning to half-open.
func WithCircuitBreakerTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.cbTimeout = timeout
		}
	}
}

// WithCircuitBreakerConsecutiveFailures sets the number of consecutive
// failures that trip the breaker from closed to open.
func WithCircuitBreakerConsecutiveFailures(failures uint32) Option {
	return func(opts *Options) {
		if failures > 0 {
			opts.cbConsecutiveFailures = failures
		}
	}
}

// WithRateLimiterInterval sets the token-bucket refill period; one token is
// produced every interval.
func WithRateLimiterInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.rateInterval = interval
		}
	}
}

// WithRateLimiterBurst sets the token-bucket burst size.
func WithRateLimiterBurst(burst int) Option {
	return func(opts *Options) {
		if burst > 0 {
			opts.rateBurst = burst
		}
	}
}

// rateLimit converts the configured interval into a rate.Limit value.
func (o *Options) rateLimit() rate.Limit {
	if o.rateInterval <= 0 {
		return rate.Every(DefaultRateLimitInterval)
	}

	return rate.Every(o.rateInterval)
}
