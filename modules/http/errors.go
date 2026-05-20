package http

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ServerType is the error domain identifier for HTTP server errors.
const ServerType = "http-server"

var (
	_ error = (*Error)(nil)
)

// Error is a domain error type for HTTP server operations.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error message including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("http server %s error: %s", e.Type, e.Err)
}

// Sentinel errors for common HTTP server failure modes.
var (
	ErrHttpServerFailed = errors.New("http server failed")
)

// ErrServer wraps one or more errors into a domain Error.
func ErrServer(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ServerType,
			Err:  errors.Join(append(errs, ErrHttpServerFailed)...),
		},
	}
}
