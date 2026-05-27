// Package retry defines the contract for a retry policy: a component
// that re-invokes a fallible operation under a configurable schedule
// until it succeeds, the predicate gives up, or the attempt budget is
// exhausted.
//
// The package exposes a single Retry interface with one method:
//
//   - Do(ctx, fn): invokes fn until it returns nil or the policy gives up.
//     Returns an error wrapping ErrRetryFailed when ctx is nil, when ctx is
//     canceled, or when the last attempt's error survives the policy.
//
// This package is implementation-free. The concrete implementation
// (backed by github.com/avast/retry-go/v4) lives in
// modules/extension/common/resilience/retry/ and depends on this
// package for the contract.
//
// Concurrency: every method on Retry is safe for concurrent use by
// multiple goroutines.
package retry

import (
	"context"
)

var (
	_ RetryIfFn = AlwaysRetry
	_ OnRetryFn = NoopOnRetry

	_ error      = (*Error)(nil)
	_ ErrRetryFn = ErrRetry
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

// RetryIfFn decides whether an error returned by fn should trigger a
// retry. Typical use: retry on transient errors (network, timeout) and
// give up on permanent errors (4xx, validation, malformed input).
// Implementations must be safe for concurrent use.
type RetryIfFn func(err error) bool

// OnRetryFn is the hook invoked before each retry attempt with the
// attempt index (zero-based: 0 is the first retry, i.e. the second total
// attempt) and the error that triggered the retry. Implementations must
// be safe for concurrent use.
type OnRetryFn func(attempt uint, err error)

// ErrRetryFn is the function type for ErrRetry.
type ErrRetryFn func(causes ...error) error

// Retry is the interface for a configured retry policy.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type Retry interface {
	// Do invokes fn until it returns nil, the configured retry-if
	// predicate rejects the error (giving up), the attempt budget is
	// exhausted, or ctx is canceled. Returns nil on success; otherwise
	// returns an error wrapping ErrRetryFailed and the last underlying
	// error. fn MUST be non-nil; passing nil yields ErrRetry wrapping
	// ErrFnNil without any invocation.
	Do(ctx context.Context, fn func() error) error
}
