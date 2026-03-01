package http

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for HTTP request errors.
const (
	RequestType = "http-request"
)

var (
	_ error = (*Error)(nil)
	_ error = (*StatusCodeError)(nil)
)

// Error is the domain error for HTTP request operations.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error message including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("http request %s error: %s", e.Type, e.Err)
}

// StatusCodeError represents an HTTP response whose status code triggered a retry.
type StatusCodeError struct {
	StatusCode int
}

// Error returns a formatted message including the retryable status code.
func (e *StatusCodeError) Error() string {
	cassert.NotNil(e, "error is nil")
	return fmt.Sprintf("http retryable status code: %d", e.StatusCode)
}

// Sentinel errors for HTTP request failure modes.
var (
	ErrHttpRequestFailed     = errors.New("http request failed")
	ErrContextNil            = errors.New("context is nil")
	ErrHttpRequestNil        = errors.New("http request is nil")
	ErrRateLimiterExceeded   = errors.New("rate limit exceeded")
	ErrHttpNonReplayableBody = errors.New("http non-replayable request body")
	ErrHttpGetBodyFailed     = errors.New("http get body failed")
)

// ErrDo creates an HTTP request domain error joining the given causes with ErrHttpRequestFailed.
func ErrDo(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RequestType,
			Err:  errors.Join(append(errs, ErrHttpRequestFailed)...),
		},
	}
}
