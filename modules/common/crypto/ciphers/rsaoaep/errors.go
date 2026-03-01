package rsaoaep

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RsaOaepMethod is the error type constant for the rsaoaep package.
const (
	RsaOaepMethod = "rsa_oaep_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the rsaoaep package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("rsa_oaep %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the rsaoaep package.
var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrKeyIsNil            = errors.New("key is nil")
	ErrKeyLengthIsInvalid  = errors.New("key length is invalid")
	ErrHashNotAvailable    = errors.New("hash function not available")
	ErrKeySizeNotAllowed   = errors.New("key size not allowed")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrDecryptionFailed    = errors.New("decryption failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named RSA-OAEP algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaOaepMethod,
			Err:  fmt.Errorf("rsa_oaep function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaOaepMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrEncryption wraps the given errors into a domain Error for encryption failures.
func ErrEncryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaOaepMethod,
			Err:  errors.Join(append(errs, ErrEncryptionFailed)...),
		},
	}
}

// ErrDecryption wraps the given errors into a domain Error for decryption failures.
func ErrDecryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaOaepMethod,
			Err:  errors.Join(append(errs, ErrDecryptionFailed)...),
		},
	}
}
