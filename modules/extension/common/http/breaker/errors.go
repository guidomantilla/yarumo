package breaker

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// BreakerTransportType is the domain type tag attached to every Error
// produced by this package.
const BreakerTransportType = "http-breaker"

// Error is the domain error for breaker transport operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for breaker transport failure modes.
var (
	// ErrBreakerRejectedFailed indicates the breaker rejected (or
	// otherwise failed) the request — it wraps the underlying domain
	// error from the resilience.Breaker so callers can use errors.Is to
	// match the specific cause (ErrBreakerOpen / ErrBreakerTooManyRequests
	// from rbreaker).
	ErrBreakerRejectedFailed = errors.New("breaker rejected the request")
)

// ErrBreakerRejected wraps a breaker domain error returned by
// Breaker.Execute into the http-breaker domain error.
func ErrBreakerRejected(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BreakerTransportType,
			Err:  errors.Join(append(causes, ErrBreakerRejectedFailed)...),
		},
	}
}

// StatusCodeError represents an HTTP response that was treated as a
// failure by the breaker transport. The transport synthesizes this error
// when FailOnResponseFn returns true so the underlying breaker (which
// only counts errors) observes the response as a failure. The response
// is stored alongside so callers can reconstitute it after the breaker
// chain unwinds.
type StatusCodeError struct {
	StatusCode int
}

// Error returns a description of the status code.
func (e *StatusCodeError) Error() string {
	cassert.NotNil(e, "status code error is nil")
	return fmt.Sprintf("http status %d reported as breaker failure", e.StatusCode)
}
