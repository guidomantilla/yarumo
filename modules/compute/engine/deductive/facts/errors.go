package facts

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
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
	ErrNotFound        = errors.New("fact not found")
	ErrFactQueryFailed = errors.New("fact query failed")
)

// ErrQuery creates a fact domain error joining the given causes.
func ErrQuery(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FactType,
			Err:  errors.Join(append(errs, ErrFactQueryFailed)...),
		},
	}
}
