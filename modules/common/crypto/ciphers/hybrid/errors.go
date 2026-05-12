package hybrid

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// HybridMethod is the error type constant for the hybrid package.
const (
	HybridMethod = "hybrid_method"
)

// Type compliance.
var (
	_ error = (*Error)(nil)
)

// Error is the domain error for the hybrid package.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error string including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("hybrid %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the hybrid package.
var (
	ErrMethodIsNil          = errors.New("method is nil")
	ErrPublicKeyIsNil       = errors.New("public key is nil")
	ErrPrivateKeyIsNil      = errors.New("private key is nil")
	ErrSuiteSetupFailed     = errors.New("hpke suite setup failed")
	ErrEncapsulationFailed  = errors.New("hpke encapsulation failed")
	ErrDecapsulationFailed  = errors.New("hpke decapsulation failed")
	ErrCiphertextTooShort   = errors.New("ciphertext too short for encapsulated key")
	ErrKeyGenerationFailed  = errors.New("key generation failed")
	ErrEncryptionFailed     = errors.New("encryption failed")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrKeyTypeMismatch      = errors.New("key type does not match method KEM")
)

// ErrAlgorithmNotSupported returns an error indicating the named hybrid
// algorithm is not registered.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HybridMethod,
			Err:  fmt.Errorf("hybrid function %s not found", name),
		},
	}
}

// ErrKeyGeneration wraps the given errors into a domain Error for key
// generation failures.
func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HybridMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

// ErrEncrypt wraps the given errors into a domain Error for encryption
// failures.
func ErrEncrypt(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HybridMethod,
			Err:  errors.Join(append(errs, ErrEncryptionFailed)...),
		},
	}
}

// ErrDecrypt wraps the given errors into a domain Error for decryption
// failures.
func ErrDecrypt(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HybridMethod,
			Err:  errors.Join(append(errs, ErrDecryptionFailed)...),
		},
	}
}
