package rsapss

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RsaPssMethod is the error type constant for the rsapss package.
const (
	RsaPssMethod = "rsa_pss_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the rsapss package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("rsa_pss %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the rsapss package.
var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrKeyIsNil            = errors.New("key is nil")
	ErrKeyLengthIsInvalid  = errors.New("key length is invalid")
	ErrSignFailed          = errors.New("sign failed")
	ErrVerifyFailed        = errors.New("verify failed")
	ErrKeySizeNotAllowed   = errors.New("key size not allowed")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrSigningFailed       = errors.New("signing failed")
	ErrVerificationFailed  = errors.New("verification failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named RSA-PSS algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaPssMethod,
			Err:  fmt.Errorf("rsa_pss function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaPssMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrSigning wraps the given errors into a domain Error for signing failures.
func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaPssMethod,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

// ErrVerification wraps the given errors into a domain Error for verification failures.
func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaPssMethod,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}
