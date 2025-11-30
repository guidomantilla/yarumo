package ed25519

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	Ed25519NotFound = "ed25519_function_not_found"
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
	return fmt.Sprintf("ed25519 %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodInvalid    = errors.New("method is invalid")
	ErrKeyInvalid       = errors.New("key is invalid")
	ErrSignatureInvalid = errors.New("signature is invalid")
)

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519NotFound,
			Err:  fmt.Errorf("ed25519 function %s not found", name),
		},
	}
}
