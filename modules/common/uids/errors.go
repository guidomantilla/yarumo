package uids

import (
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	UIDNotFound = "uid_function_not_found"
)

var (
	_ error = (*UIDError)(nil)
)

type UIDError struct {
	cerrs.TypedError
}

func (e *UIDError) Error() string {
	assert.NotEmpty(e, "error is nil")
	assert.NotEmpty(e.Err, "internal error is nil")
	return fmt.Sprintf("uid %s error: %s", e.Type, e.Err)
}

func ErrUIDFunctionNotFound(name string) error {
	return &UIDError{
		TypedError: cerrs.TypedError{
			Type: UIDNotFound,
			Err:  fmt.Errorf("uid function %s not found", name),
		},
	}
}
