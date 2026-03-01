package facts

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for fact errors.
const (
	FactType = "inference-fact"
)

var _ error = (*Error)(nil)

// Error is the domain error for fact operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for fact failure modes.
var (
	ErrVariableEmpty = errors.New("variable is empty")
	ErrNotFound      = errors.New("fact not found")
)

// ErrQuery creates a fact domain error joining the given causes with ErrNotFound.
func ErrQuery(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FactType,
			Err:  errors.Join(append(errs, ErrNotFound)...),
		},
	}
}
