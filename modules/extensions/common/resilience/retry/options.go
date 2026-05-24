package retry

import (
	"time"
)

// Backoff names the delay schedule applied between retry attempts.
type Backoff int

const (
	// BackoffFixed waits the configured Delay between every attempt.
	BackoffFixed Backoff = iota
	// BackoffExponential doubles the delay on each attempt, capped at
	// MaxDelay. This is the default.
	BackoffExponential
	// BackoffRandom waits a uniformly random duration between 0 and Delay
	// before each attempt.
	BackoffRandom
)

// Default values for retry configuration. The defaults yield 3 attempts
// (1 original + 2 retries), starting at 100ms and doubling up to 5s.
const (
	// DefaultAttempts is the total attempt count when the caller does not
	// configure it explicitly (1 original + N-1 retries).
	DefaultAttempts uint = 3
	// DefaultDelay is the base delay between attempts. With
	// BackoffExponential this is the delay before the first retry.
	DefaultDelay = 100 * time.Millisecond
	// DefaultMaxDelay is the cap on the exponential backoff growth.
	DefaultMaxDelay = 5 * time.Second
	// DefaultBackoff is the default delay schedule (exponential).
	DefaultBackoff = BackoffExponential
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied to a Retry at construction
// time. Fields are unexported; callers configure them through With*.
type Options struct {
	attempts uint
	delay    time.Duration
	maxDelay time.Duration
	backoff  Backoff
	retryIf  RetryIfFn
	onRetry  OnRetryFn
}

// NewOptions creates Options with safe defaults and applies the given
// functional options. Defaults: 3 attempts, 100ms base delay, 5s max
// delay, exponential backoff, AlwaysRetry predicate, NoopOnRetry hook.
func NewOptions(opts ...Option) *Options {
	options := &Options{
		attempts: DefaultAttempts,
		delay:    DefaultDelay,
		maxDelay: DefaultMaxDelay,
		backoff:  DefaultBackoff,
		retryIf:  AlwaysRetry,
		onRetry:  NoopOnRetry,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithAttempts sets the total number of attempts (1 original + N-1
// retries). Values less than 2 are ignored, preserving the default. A
// caller that wants to disable retry should not wrap the call in a Retry
// in the first place.
func WithAttempts(attempts uint) Option {
	return func(opts *Options) {
		if attempts > 1 {
			opts.attempts = attempts
		}
	}
}

// WithDelay sets the base delay between attempts. For BackoffExponential
// this is the delay before the first retry. Non-positive values are
// ignored, preserving the default.
func WithDelay(delay time.Duration) Option {
	return func(opts *Options) {
		if delay > 0 {
			opts.delay = delay
		}
	}
}

// WithMaxDelay sets the cap on backoff growth. Non-positive values are
// ignored, preserving the default. Only meaningful for BackoffExponential.
func WithMaxDelay(maxDelay time.Duration) Option {
	return func(opts *Options) {
		if maxDelay > 0 {
			opts.maxDelay = maxDelay
		}
	}
}

// WithBackoff sets the delay schedule between attempts. Invalid values
// are ignored, preserving the default.
func WithBackoff(backoff Backoff) Option {
	return func(opts *Options) {
		switch backoff {
		case BackoffFixed, BackoffExponential, BackoffRandom:
			opts.backoff = backoff
		}
	}
}

// WithRetryIf sets the predicate that decides whether an error should
// trigger a retry. Nil values are ignored, preserving the default
// (AlwaysRetry).
func WithRetryIf(predicate RetryIfFn) Option {
	return func(opts *Options) {
		if predicate != nil {
			opts.retryIf = predicate
		}
	}
}

// WithOnRetry sets the hook invoked before each retry attempt. Nil values
// are ignored, preserving the default (NoopOnRetry).
func WithOnRetry(hook OnRetryFn) Option {
	return func(opts *Options) {
		if hook != nil {
			opts.onRetry = hook
		}
	}
}

