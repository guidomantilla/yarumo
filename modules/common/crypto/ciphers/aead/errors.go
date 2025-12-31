package aead

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	AeadMethod = "aead_method"
)

var (
	_ error = (*Error)(nil)
)

type Error struct {
	cerrs.TypedError
}

func (e *Error) Error() string {
	assert.NotNil(e, "error is nil")
	assert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("aes %s error: %s", e.Type, e.Err)
}

//

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

func ErrAlgorithmNotSupported(name string) error {
	return fmt.Errorf("aead function %s not found", name)
}

func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

func ErrEncryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrEncryptFailed)...),
		},
	}
}

func ErrDecryption(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AeadMethod,
			Err:  errors.Join(append(errs, ErrDecryptFailed)...),
		},
	}
}
