package breaker

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// BreakerType is the domain type tag attached to every Error produced by this package.
const BreakerType = "circuit-breaker"

// Error is the domain error for breaker failures.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for breaker failure modes.
var (
	// ErrBreakerFailed indicates the breaker rejected or recorded a failure.
	ErrBreakerFailed = errors.New("breaker failed")
	// ErrBreakerOpen indicates the breaker rejected the call because it is
	// in the open state.
	ErrBreakerOpen = errors.New("circuit breaker is open")
	// ErrBreakerTooManyRequests indicates the breaker rejected the call in
	// half-open state because the probe budget is exhausted.
	ErrBreakerTooManyRequests = errors.New("circuit breaker too many requests")
	// ErrContextNil indicates Execute received a nil context.Context.
	ErrContextNil = errors.New("context is nil")
	// ErrFnNil indicates Execute received a nil function to run.
	ErrFnNil = errors.New("breaker fn is nil")
)

// ErrBreaker creates a breaker domain error joining the given causes with
// ErrBreakerFailed.
func ErrBreaker(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BreakerType,
			Err:  errors.Join(append(causes, ErrBreakerFailed)...),
		},
	}
}
