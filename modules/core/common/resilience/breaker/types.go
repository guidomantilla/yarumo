// Package breaker defines the contract for a circuit breaker: a
// component that gates calls to an unreliable downstream, fails fast when
// failures accumulate, and probes recovery after a cool-off window.
//
// The package exposes a single Breaker interface with two methods:
//
//   - Execute(ctx, fn): invokes fn through the breaker. Open breakers
//     reject the call fast with ErrBreakerOpen; half-open breakers reject
//     beyond the probe budget with ErrBreakerTooManyRequests. Errors
//     returned by fn are recorded and may trip the breaker.
//   - State(): reports the current operating state (Closed / Half-Open /
//     Open).
//
// This package is implementation-free. The concrete implementation
// (backed by github.com/sony/gobreaker) lives in
// modules/extension/common/resilience/breaker/ and depends on this
// package for the contract.
//
// Concurrency: every method on Breaker is safe for concurrent use by
// multiple goroutines.
package breaker

import (
	"context"
)

var (
	_ OnStateChangeFn = NoopOnStateChange

	_ error        = (*Error)(nil)
	_ ErrBreakerFn = ErrBreaker
)

// OnStateChangeFn is the hook invoked when the breaker transitions
// between states (Closed ↔ Half-Open ↔ Open). The hook receives the
// previous and next states. Implementations must be safe for concurrent
// use; the hook runs inline on the goroutine that triggered the
// transition.
type OnStateChangeFn func(from State, to State)

// ErrBreakerFn is the function type for ErrBreaker.
type ErrBreakerFn func(causes ...error) error

// Breaker is the interface for a configured circuit breaker.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type Breaker interface {
	// Execute invokes fn through the breaker. Returns nil when fn returns
	// nil; otherwise returns an error wrapping ErrBreakerFailed and the
	// underlying error. Open breakers reject without invoking fn, wrapping
	// ErrBreakerOpen; half-open breakers beyond the probe budget reject
	// wrapping ErrBreakerTooManyRequests. ctx and fn MUST be non-nil;
	// passing nil yields ErrBreaker wrapping ErrContextNil or ErrFnNil
	// without any invocation.
	Execute(ctx context.Context, fn func() error) error
	// State returns the current operating state of the breaker.
	State() State
}
