package ecdsas

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	EcdsaNotFound = "ecdsa_function_not_found"
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
	return fmt.Sprintf("ecdsa %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodInvalid     = errors.New("ecdsa method is invalid")
	ErrKeyInvalid        = errors.New("ecdsa key is invalid")
	ErrSignatureInvalid  = errors.New("ecdsa signature is invalid")
	ErrSignFailed        = errors.New("ecdsa sign failed")
	ErrFormatUnsupported = errors.New("ecdsa format unsupported")
)

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaNotFound,
			Err:  fmt.Errorf("ecdsa function %s not found", name),
		},
	}
}
