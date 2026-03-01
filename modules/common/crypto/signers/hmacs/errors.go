package hmacs

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// HmacMethod is the error type constant for the hmacs package.
const (
	HmacMethod = "hmac_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the hmacs package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("hmac %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the hmacs package.
var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrKeyIsNil            = errors.New("key is nil")
	ErrHashNotAvailable    = errors.New("hash not available")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrDigestFailed        = errors.New("digest failed")
	ErrValidationFailed    = errors.New("validation failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named HMAC algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  fmt.Errorf("hmac function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrDigest wraps the given errors into a domain Error for digest computation failures.
func ErrDigest(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrDigestFailed)...),
		},
	}
}

// ErrValidation wraps the given errors into a domain Error for validation failures.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrValidationFailed)...),
		},
	}
}
