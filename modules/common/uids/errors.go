package uids

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error type constants for UID operations.
const (
	UidNotFound = "uid_function_not_found"
)

// Error is a domain error type for UID operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("uid %s error: %s", e.Type, e.Err)
}

// ErrAlgorithmNotSupported creates an error indicating the requested UID algorithm is not registered.
func ErrAlgorithmNotSupported(name string) error {
	cassert.NotEmpty(name, "name is empty")

	return &Error{
		TypedError: cerrs.TypedError{
			Type: UidNotFound,
			Err:  fmt.Errorf("uid algorithm %s not found", name),
		},
	}
}
