package hashes

import (
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/pkg/common/errs"
)

const (
	HashNotFound = "hash_function_not_found"
)

type HashError struct {
	cerrs.TypedError
}

func (e *HashError) Error() string {
	return fmt.Sprintf("hash %s error: %s", e.Type, e.Err)
}

func ErrHashFunctionNotFound(name string) error {
	return &HashError{
		TypedError: cerrs.TypedError{
			Type: HashNotFound,
			Err:  fmt.Errorf("hash function %s not found", name),
		},
	}
}
