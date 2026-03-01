package aead

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// AeadMethod is the error type constant for the aead package.
const (
	AeadMethod = "aead_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the aead package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("aead %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the aead package.
var (
	ErrMethodInvalid       = errors.New("cipher method is invalid")
	ErrKeyInvalid          = errors.New("cipher key is invalid")
	ErrCipherInitFailed    = errors.New("cipher initialization failed")
	ErrCiphertextTooShort  = errors.New("ciphertext too short")
	ErrKeySizeInvalid      = errors.New("key size is invalid")
	ErrNonceSizeInvalid    = errors.New("nonce size is invalid")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrEncryptFailed       = errors.New("encrypt failed")
	ErrDecryptFailed       = errors.New("decrypt failed")
)

// ErrAlgorithmNotSupported returns an error indicating the named AEAD algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  fmt.Errorf("aead function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrEncryption wraps the given errors into a domain Error for encryption failures.
func ErrEncryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrEncryptFailed)...),
		},
	}
}

// ErrDecryption wraps the given errors into a domain Error for decryption failures.
func ErrDecryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrDecryptFailed)...),
		},
	}
}
