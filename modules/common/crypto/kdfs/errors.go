package kdfs

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// KdfMethod is the error type constant for the kdfs package.
const (
	KdfMethod = "kdf_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the kdfs package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("kdf %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the kdfs package.
var (
	ErrMethodIsNil      = errors.New("method is nil")
	ErrSecretIsNil      = errors.New("secret is nil")
	ErrSaltIsNil        = errors.New("salt is nil")
	ErrLengthInvalid    = errors.New("length is invalid")
	ErrHashNotAvailable = errors.New("hash not available")
	ErrDeriveFailed     = errors.New("derive failed")
	ErrParamsMissing    = errors.New("method has no algorithm configuration")
)

// ErrAlgorithmNotSupported returns an error indicating the named KDF algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: KdfMethod,
			Err:  fmt.Errorf("kdf function %s not found", name),
		},
	}
}

// ErrDerive wraps the given errors into a domain Error for key derivation failures.
func ErrDerive(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: KdfMethod,
			Err:  errors.Join(append(errs, ErrDeriveFailed)...),
		},
	}
}
