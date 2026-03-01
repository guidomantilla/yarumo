package hashes

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// HashNotFound is the error type constant for hash algorithm lookup failures.
const (
	HashNotFound = "hash_function_not_found"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the hashes package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("hash %s error: %s", e.Type, e.Err)
}

// ErrAlgorithmNotSupported returns an error indicating the named hash algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HashNotFound,
			Err:  fmt.Errorf("hash function %s not found", name),
		},
	}
}
