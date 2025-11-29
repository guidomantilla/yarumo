package hashes

import (
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	HashNotFound = "hash_function_not_found"
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
	return fmt.Sprintf("hash %s error: %s", e.Type, e.Err)
}

//

func ErrAlgorithmNotSupported(name string) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HashNotFound,
			Err:  fmt.Errorf("hash function %s not found", name),
		},
	}
}
