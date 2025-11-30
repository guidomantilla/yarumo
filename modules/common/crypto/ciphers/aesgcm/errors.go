package aesgcm

import (
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

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AesNotFound,
			Err:  fmt.Errorf("hmac function %s not found", name),
		},
	}
}
