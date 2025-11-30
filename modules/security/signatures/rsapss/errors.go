package rsapss

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	RsaPssNotFound = "rsa_pss_function_not_found"
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
	return fmt.Sprintf("rsa_pss %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodInvalid     = errors.New("method is invalid")
	ErrKeyInvalid        = errors.New("key is invalid")
	ErrSignFailed        = errors.New("sign failed")
	ErrVerifyFailed      = errors.New("verify failed")
	ErrKeySizeNotAllowed = errors.New("key size not allowed")
)

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RsaPssNotFound,
			Err:  fmt.Errorf("rsa_pss function %s not found", name),
		},
	}
}
