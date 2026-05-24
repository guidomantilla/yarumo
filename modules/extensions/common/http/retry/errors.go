package retry

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RetryType is the domain type tag attached to every Error produced by this package.
const RetryType = "http-retry"

// Error is the domain error for retry transport operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for retry transport failure modes.
var (
	// ErrNonReplayableBodyFailed indicates a request with a non-replayable
	// body (Body != nil and GetBody == nil) cannot be retried because
	// subsequent attempts would silently send a consumed body.
	ErrNonReplayableBodyFailed = errors.New("request body cannot be replayed (no GetBody set)")
)

// ErrNonReplayableBody creates a retry domain error joining the given
// causes with ErrNonReplayableBodyFailed.
func ErrNonReplayableBody(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RetryType,
			Err:  errors.Join(append(causes, ErrNonReplayableBodyFailed)...),
		},
	}
}

// StatusCodeError represents an HTTP response that was treated as a retry
// trigger by the retry transport. The transport synthesizes this error
// when RetryOnResponseFn returns true so the underlying retrier (which
// only retries on errors) observes the response as a retryable failure.
// Use RetryIfHttpError (or errors.As) to recognize it.
type StatusCodeError struct {
	StatusCode int
}

// Error returns a description of the status code.
func (e *StatusCodeError) Error() string {
	cassert.NotNil(e, "status code error is nil")
	return fmt.Sprintf("http status %d treated as retryable", e.StatusCode)
}
