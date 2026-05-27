package rsassas

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// RsassasMethod is the error type constant for the rsassas package.
const (
	RsassasMethod = "rsassas_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the rsassas package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("rsassas %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the rsassas package.
var (
	ErrMethodIsNil          = errors.New("method is nil")
	ErrKeyIsNil             = errors.New("key is nil")
	ErrKeyLengthIsInvalid   = errors.New("key length is invalid")
	ErrSignFailed           = errors.New("sign failed")
	ErrVerifyFailed         = errors.New("verify failed")
	ErrKeySizeNotAllowed    = errors.New("key size not allowed")
	ErrPaddingNotSupported  = errors.New("padding scheme not supported")
	ErrKeyGenerationFailed  = errors.New("key generation failed")
	ErrSigningFailed        = errors.New("signing failed")
	ErrVerificationFailed   = errors.New("verification failed")
	ErrPEMDecodeFailed      = errors.New("pem decode failed")
	ErrPEMBlockTypeMismatch = errors.New("pem block type mismatch")
	ErrKeyTypeMismatch      = errors.New("key type mismatch")
	ErrMarshalKeyFailed     = errors.New("marshal key failed")
	ErrParseKeyFailed       = errors.New("parse key failed")
	ErrPEMCodecFailed       = errors.New("pem codec failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named RSA algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsassasMethod,
			Err:  fmt.Errorf("rsassas function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsassasMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrSigning wraps the given errors into a domain Error for signing failures.
func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsassasMethod,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

// ErrVerification wraps the given errors into a domain Error for verification failures.
func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsassasMethod,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}

// ErrPEMCodec wraps the given errors into a domain Error for PEM marshal/parse failures.
func ErrPEMCodec(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsassasMethod,
			Err:  errors.Join(append(errs, ErrPEMCodecFailed)...),
		},
	}
}
