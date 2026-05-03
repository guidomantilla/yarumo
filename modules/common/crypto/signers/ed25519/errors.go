package ed25519

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Ed25519Method is the error type constant for the ed25519 package.
const (
	Ed25519Method = "ed25519_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the ed25519 package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("ed25519 %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the ed25519 package.
var (
	ErrMethodIsNil            = errors.New("method is nil")
	ErrKeyIsNil               = errors.New("key is nil")
	ErrKeyLengthIsInvalid     = errors.New("key length is invalid")
	ErrSignatureLengthInvalid = errors.New("signature length is invalid")
	ErrKeyGenerationFailed    = errors.New("key generation failed")
	ErrSigningFailed          = errors.New("signing failed")
	ErrVerificationFailed     = errors.New("verification failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named Ed25519 algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  fmt.Errorf("ed25519 function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrSigning wraps the given errors into a domain Error for signing failures.
func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

// ErrVerification wraps the given errors into a domain Error for verification failures.
func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}
