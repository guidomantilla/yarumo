package aesgcm

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	AesNotFound = "aes_function_not_found"
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
	ErrMethodInvalid      = errors.New("cipher method is invalid")
	ErrKeyInvalid         = errors.New("cipher key is invalid")
	ErrNonceInvalid       = errors.New("nonce generation failed")
	ErrCipherInitFailed   = errors.New("cipher initialization failed")
	ErrNonceMissing       = errors.New("nonce missing")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
	ErrDecryptFailed      = errors.New("decrypt failed")
)

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AesNotFound,
			Err:  fmt.Errorf("hmac function %s not found", name),
		},
	}
}
