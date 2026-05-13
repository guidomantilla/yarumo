// Package resilience provides goroutine-free registries for outbound-call
// resilience patterns. Two registries are exposed:
//
//   - CircuitBreakerRegistry backed by github.com/sony/gobreaker — fail-fast
//     state machine that trips after a configurable number of consecutive
//     failures.
//   - RateLimiterRegistry backed by golang.org/x/time/rate — token-bucket
//     limiter for client-side throttling.
//
// Both registries lazy-create their values by name on the first Get call and
// return the same instance on subsequent calls. Use(name, opts...) reconfigures
// (or creates) an entry with the supplied options.
//
// Concurrency: every method on every registry, breaker, and limiter is safe
// for concurrent use by multiple goroutines.
//
// Goroutine-free: the package does not spawn goroutines under any code path.
// Time-based transitions in the circuit breaker happen synchronously, observed
// on the next Execute call.
package resilience

import (
	"context"
)

var (
	_ CircuitBreakerRegistry = (*circuitBreakerRegistry)(nil)
	_ CircuitBreaker         = (*circuitBreaker)(nil)
	_ RateLimiterRegistry    = (*rateLimiterRegistry)(nil)
	_ RateLimiter            = (*rateLimiter)(nil)

	_ ExecuteFn = (*circuitBreaker)(nil).Execute
	_ StateFn   = (*circuitBreaker)(nil).State
	_ AllowFn   = (*rateLimiter)(nil).Allow
	_ WaitFn    = (*rateLimiter)(nil).Wait
)

// State represents the operating state of a CircuitBreaker.
type State int

// State values for a CircuitBreaker, mirroring github.com/sony/gobreaker.
const (
	// StateClosed indicates the breaker passes all calls through.
	StateClosed State = iota
	// StateHalfOpen indicates the breaker is probing a limited number of calls.
	StateHalfOpen
	// StateOpen indicates the breaker is failing fast without invoking the call.
	StateOpen
)

// String returns the human-readable name of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// ExecuteFn is the function type for CircuitBreaker.Execute.
type ExecuteFn func(ctx context.Context, fn func() (any, error)) (any, error)

// StateFn is the function type for CircuitBreaker.State.
type StateFn func() State

// AllowFn is the function type for RateLimiter.Allow.
type AllowFn func() bool

// WaitFn is the function type for RateLimiter.Wait.
type WaitFn func(ctx context.Context) error

// CircuitBreaker defines the interface for a single named circuit breaker.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type CircuitBreaker interface {
	// Execute invokes fn through the breaker. If the breaker is open the call
	// is rejected immediately. If fn returns a non-nil error the failure is
	// recorded and may trip the breaker. The first return value is the value
	// returned by fn on success.
	Execute(ctx context.Context, fn func() (any, error)) (any, error)
	// State returns the current operating state of the breaker.
	State() State
}

// CircuitBreakerRegistry defines the interface for a thread-safe, lazy
// registry of named circuit breakers.
//
// Get lazy-creates a breaker with default options when the name has no entry.
// Subsequent Get calls with the same name return the same instance.
// Use replaces (or creates) the named entry with a breaker built from the
// supplied options.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type CircuitBreakerRegistry interface {
	// Get returns the breaker registered under name, lazy-creating it with
	// default options if none exists.
	Get(name string) CircuitBreaker
	// Use replaces (or creates) the named entry with a breaker built from the
	// supplied options. It returns a domain error when name is empty.
	Use(name string, opts ...Option) error
}

// RateLimiter defines the interface for a single named token-bucket limiter.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type RateLimiter interface {
	// Allow reports whether a token is available right now without blocking.
	Allow() bool
	// Wait blocks until a token is available or ctx is canceled.
	Wait(ctx context.Context) error
}

// RateLimiterRegistry defines the interface for a thread-safe, lazy registry
// of named rate limiters.
//
// Get lazy-creates a limiter with default options when the name has no entry.
// Subsequent Get calls with the same name return the same instance.
// Use replaces (or creates) the named entry with a limiter built from the
// supplied options.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type RateLimiterRegistry interface {
	// Get returns the limiter registered under name, lazy-creating it with
	// default options if none exists.
	Get(name string) RateLimiter
	// Use replaces (or creates) the named entry with a limiter built from the
	// supplied options. It returns a domain error when name is empty.
	Use(name string, opts ...Option) error
}
