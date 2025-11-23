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
	_ error = (*Error)(nil)
)

type Error struct {
	cerrs.TypedError
}

func (e *Error) Error() string {
	assert.NotEmpty(e, "error is nil")
	assert.NotEmpty(e.Err, "internal error is nil")
	return fmt.Sprintf("uid %s error: %s", e.Type, e.Err)
}

func ErrUIDFunctionNotFound(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: UIDNotFound,
			Err:  fmt.Errorf("uid function %s not found", name),
		},
	}
}
