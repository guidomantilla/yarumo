// Package retry provides a generic retry helper backed by
// github.com/avast/retry-go/v4.
//
// The package exposes a single Retry interface with one method:
//
//   - Do(ctx, fn): invokes fn until it returns nil or the policy gives up.
//     Returns an error wrapping ErrRetryFailed when ctx is nil, when ctx is
//     canceled, or when the last attempt's error survives the policy.
//
// Construct an instance via NewRetry(opts ...Option). Each call returns an
// independent Retry; there is no registry, no global singleton, and no
// pluggable function fields. Callers that need multiple policies construct
// multiple instances.
//
// Concurrency: every method on Retry is safe for concurrent use by
// multiple goroutines.
//
// Goroutine-free: NewRetry does not spawn goroutines. Each Do call runs
// the retry loop on the caller's goroutine; delays between attempts are
// time.Sleep equivalents bounded by ctx.
package retry

import (
	"context"
)

var (
	_ Retry = (*retry)(nil)

	_ RetryIfFn = AlwaysRetry
	_ OnRetryFn = NoopOnRetry

	_ error      = (*Error)(nil)
	_ ErrRetryFn = ErrRetry
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
