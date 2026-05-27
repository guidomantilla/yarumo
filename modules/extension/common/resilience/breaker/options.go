package breaker

import (
	"time"

	cbreaker "github.com/guidomantilla/yarumo/core/common/resilience/breaker"
)

// Default values for breaker configuration. The defaults yield: 5
// consecutive failures trip the breaker; once open it waits 15s before
// transitioning to half-open; in half-open it lets through 1 probe at a
// time; in closed it resets internal counters every 60s.
const (
	// DefaultMaxRequests is the max number of probes allowed in half-open
	// state before the breaker closes or re-opens.
	DefaultMaxRequests uint32 = 1
	// DefaultInterval is the cyclic period in closed state at which
	// internal failure counters are reset.
	DefaultInterval = 60 * time.Second
	// DefaultTimeout is the time the breaker stays open before
	// transitioning to half-open on the next call.
	DefaultTimeout = 15 * time.Second
	// DefaultConsecutiveFailures is the number of consecutive failures
	// that trips the breaker from closed to open.
	DefaultConsecutiveFailures uint32 = 5
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied to a Breaker at construction
// time. Fields are unexported; callers configure them through With*.
type Options struct {
	maxRequests         uint32
	interval            time.Duration
	timeout             time.Duration
	consecutiveFailures uint32
	onStateChange       cbreaker.OnStateChangeFn
}

// NewOptions creates Options with safe defaults and applies the given
// functional options. Defaults: name "breaker", 1 half-open probe, 60s
// interval, 15s timeout, 5 consecutive failures, NoopOnStateChange hook.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		maxRequests:         DefaultMaxRequests,
		interval:            DefaultInterval,
		timeout:             DefaultTimeout,
		consecutiveFailures: DefaultConsecutiveFailures,
		onStateChange:       cbreaker.NoopOnStateChange,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithMaxRequests sets the number of probes allowed in half-open state
// before the breaker closes or re-opens. Zero is ignored, preserving the
// default.
func WithMaxRequests(maxRequests uint32) Option {
	return func(opts *Options) {
		if maxRequests > 0 {
			opts.maxRequests = maxRequests
		}
	}
}

// WithInterval sets the cyclic period that resets the breaker's internal
// counters while it is in closed state. Non-positive values are ignored,
// preserving the default.
func WithInterval(interval time.Duration) Option {
	return func(opts *Options) {
		if interval > 0 {
			opts.interval = interval
		}
	}
}

// WithTimeout sets the time the breaker stays open before transitioning
// to half-open on the next call. Non-positive values are ignored,
// preserving the default.
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		if timeout > 0 {
			opts.timeout = timeout
		}
	}
}

// WithConsecutiveFailures sets the number of consecutive failures that
// trip the breaker from closed to open. Zero is ignored, preserving the
// default.
func WithConsecutiveFailures(failures uint32) Option {
	return func(opts *Options) {
		if failures > 0 {
			opts.consecutiveFailures = failures
		}
	}
}

// WithOnStateChange sets the hook invoked on every state transition. Nil
// values are ignored, preserving the default (NoopOnStateChange).
func WithOnStateChange(hook cbreaker.OnStateChangeFn) Option {
	return func(opts *Options) {
		if hook != nil {
			opts.onStateChange = hook
		}
	}
}
