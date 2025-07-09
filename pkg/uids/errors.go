package uids

import (
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/pkg/common/errs"
)

const (
	UIDNotFound = "uid_function_not_found"
)

type UIDError struct {
	cerrs.TypedError
}

func (e *UIDError) Error() string {
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
