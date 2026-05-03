package rest

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// RequestType is the error domain identifier for REST request errors.
const RequestType = "rest-request"

var (
	_ error = (*Error)(nil)
	_ error = (*HTTPError)(nil)
	_ error = (*DecodeResponseError[any])(nil)
)

// Error is a domain error type for REST operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("rest request %s error: %s", e.Type, e.Err)
}

// HTTPError represents an HTTP error response with a non-2xx status code.
type HTTPError struct {
	StatusCode int
	Status     string
	Body       ctypes.Bytes
}

// Error returns the formatted error string.
func (e *HTTPError) Error() string {
	cassert.NotNil(e, "error is nil")

	return fmt.Sprintf("unexpected status code %d: %s", e.StatusCode, e.Status)
}

// DecodeResponseError indicates that the response content-type cannot be decoded into the target type.
type DecodeResponseError[T any] struct {
	ContentType string
	T           T
}

// Error returns the formatted error string.
func (e *DecodeResponseError[T]) Error() string {
	cassert.NotNil(e, "error is nil")

	return fmt.Sprintf("content type %s not supported for type %T", e.ContentType, e.T)
}

// Sentinel errors for common REST call failure modes.
var (
	ErrRestCallFailed    = errors.New("rest call failed")
	ErrContextNil        = errors.New("context is nil")
	ErrRequestSpecNil    = errors.New("request spec is nil")
	ErrResponseTooLarge  = errors.New("response body exceeds maximum allowed size")
	ErrReadBodyFailed    = errors.New("reading response body failed")
	ErrUnmarshalFailed   = errors.New("unmarshalling response body failed")
	ErrMarshalBodyFailed = errors.New("marshaling request body failed")
	ErrURLParseFailed    = errors.New("parsing URL failed")
)

// ErrCall wraps one or more errors into a domain Error.
func ErrCall(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RequestType,
			Err:  errors.Join(append(errs, ErrRestCallFailed)...),
		},
	}
}
