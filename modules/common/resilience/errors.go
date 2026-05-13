package resilience

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type constants for resilience errors.
const (
	// CircuitBreakerType identifies circuit-breaker related errors.
	CircuitBreakerType = "circuit-breaker"
	// RateLimiterType identifies rate-limiter related errors.
	RateLimiterType = "rate-limiter"
	// RegistryType identifies registry related errors.
	RegistryType = "resilience-registry"
)

var _ error = (*Error)(nil)

// Error is the domain error for resilience operations.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error message including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("resilience %s error: %s", e.Type, e.Err)
}

// Sentinel errors for resilience failure modes.
var (
	// ErrCircuitBreakerOpen indicates the breaker rejected the call because it
	// is in the open state.
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrCircuitBreakerTooManyRequests indicates the breaker rejected the call
	// in half-open state because the in-flight probe budget is exhausted.
	ErrCircuitBreakerTooManyRequests = errors.New("circuit breaker too many requests")
	// ErrCircuitBreakerExecuteFnNil indicates Execute was called with a nil fn.
	ErrCircuitBreakerExecuteFnNil = errors.New("circuit breaker execute fn is nil")
	// ErrRateLimiterWaitFailed indicates the limiter Wait operation failed.
	ErrRateLimiterWaitFailed = errors.New("rate limiter wait failed")
	// ErrContextNil indicates a required context.Context was nil.
	ErrContextNil = errors.New("context is nil")
	// ErrRegistryNameEmpty indicates a registry operation received an empty name.
	ErrRegistryNameEmpty = errors.New("registry name is empty")
)

// ErrCircuitBreakerExecute wraps Execute failures into the breaker domain error.
func ErrCircuitBreakerExecute(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CircuitBreakerType,
			Err:  errors.Join(errs...),
		},
	}
}

// ErrRateLimiterWait wraps limiter Wait failures into the limiter domain error.
func ErrRateLimiterWait(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RateLimiterType,
			Err:  errors.Join(append(errs, ErrRateLimiterWaitFailed)...),
		},
	}
}

// ErrRegistryUse wraps registry Use failures into the registry domain error.
func ErrRegistryUse(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RegistryType,
			Err:  errors.Join(errs...),
		},
	}
}
