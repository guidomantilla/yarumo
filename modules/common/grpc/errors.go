package grpc

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ServerType is the error domain identifier for gRPC server errors.
const ServerType = "grpc-server"

var (
	_ error = (*Error)(nil)
)

// Error is a domain error type for gRPC server operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("grpc server %s error: %s", e.Type, e.Err)
}

// Sentinel errors for common gRPC server failure modes.
var (
	ErrGrpcServerFailed = errors.New("grpc server failed")
)

// ErrServer wraps one or more errors into a domain Error.
func ErrServer(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ServerType,
			Err:  errors.Join(append(errs, ErrGrpcServerFailed)...),
		},
	}
}
