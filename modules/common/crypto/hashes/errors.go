package hashes

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error type constants for the hashes package.
const (
	HashNotFound    = "hash_function_not_found"
	HashUnavailable = "hash_function_unavailable"
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

// Sentinel errors for the hashes package.
var (
	ErrHashFunctionUnavailable = errors.New("hash function not available — call crypto.RegisterHash")
)

// ErrAlgorithmNotSupported returns an error indicating the named hash algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HashNotFound,
			Err:  fmt.Errorf("hash function %s not found", name),
		},
	}
}

// ErrDigest wraps the given errors into a domain Error for hash digest failures
// caused by an unavailable crypto.Hash driver.
func ErrDigest(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HashUnavailable,
			Err:  errors.Join(append(errs, ErrHashFunctionUnavailable)...),
		},
	}
}
