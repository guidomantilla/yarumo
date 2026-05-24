package http

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// HTTPType is the domain type tag attached to every Error produced by this package.
const HTTPType = "http"

// Error is the domain error type for http transport operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for http transport failure modes.
var (
	ErrTransportFailed = errors.New("transport failed")
)

// ErrTransport creates a transport domain error joining the given causes with ErrTransportFailed.
func ErrTransport(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HTTPType,
			Err:  errors.Join(append(causes, ErrTransportFailed)...),
		},
	}
}
