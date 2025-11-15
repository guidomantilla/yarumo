package uids

import (
	"fmt"

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
	if e == nil || e.Err == nil {
		return "<nil>"
	}
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
