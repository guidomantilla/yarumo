package aead

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	AeadNotFound = "aead_function_not_found"
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
	ErrCipherInitFailed   = errors.New("cipher initialization failed")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
	ErrDecryptFailed      = errors.New("decrypt failed")
)

func ErrAlgorithmNotSupported(name string) error {
	return fmt.Errorf("aead function %s not found", name)
}
