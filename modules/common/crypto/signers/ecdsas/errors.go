package ecdsas

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// EcdsaMethod is the error type constant for the ecdsas package.
const (
	EcdsaMethod = "ecdsa_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the ecdsas package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("ecdsa %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the ecdsas package.
var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrKeyIsNil            = errors.New("key is nil")
	ErrKeyCurveIsInvalid   = errors.New("key curve is invalid")
	ErrSignatureInvalid    = errors.New("signature is invalid")
	ErrSignFailed          = errors.New("sign failed")
	ErrFormatUnsupported   = errors.New("format unsupported")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrSigningFailed       = errors.New("signing failed")
	ErrVerificationFailed  = errors.New("verification failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named ECDSA algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  fmt.Errorf("ecdsa function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrSigning wraps the given errors into a domain Error for signing failures.
func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

// ErrVerification wraps the given errors into a domain Error for verification failures.
func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}
