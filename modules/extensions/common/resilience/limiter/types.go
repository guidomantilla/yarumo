// Package limiter provides a token-bucket rate limiter backed by
// golang.org/x/time/rate.
//
// The package exposes a single Limiter interface with two methods:
//
//   - Allow(): non-blocking; returns true when a token is available.
//   - Wait(ctx): blocks until a token is available or ctx is canceled.
//
// Construct an instance via NewLimiter(opts ...Option). Each call returns
// an independent Limiter; there is no registry, no global singleton, and
// no pluggable function fields. Callers that need multiple limiters
// construct multiple instances.
//
// Concurrency: every method on Limiter is safe for concurrent use by
// multiple goroutines.
//
// Goroutine-free: NewLimiter does not spawn goroutines. Rate replenishment
// happens lazily on each Allow/Wait call.
package limiter

import (
	"context"
)

var (
	_ Limiter = (*limiter)(nil)

	_ error     = (*Error)(nil)
	_ ErrWaitFn = ErrWait
)

// ErrWaitFn is the function type for ErrWait.
type ErrWaitFn func(causes ...error) error

// Limiter is the interface for a token-bucket rate limiter.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type Limiter interface {
	// Allow reports whether a token is available right now without
	// blocking. Returns false when the bucket is empty; callers can
	// fail-fast or fall back to Wait when blocking is acceptable.
	Allow() bool
	// Wait blocks until a token is available or ctx is canceled. Returns
	// an error wrapping ErrWaitFailed when ctx is nil, when ctx expires,
	// or when the underlying limiter reports an error (e.g. the burst is
	// smaller than the requested 1 token).
	Wait(ctx context.Context) error
}
