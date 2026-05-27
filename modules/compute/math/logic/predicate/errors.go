package predicate

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for predicate quantification errors.
const (
	PredicateType = "math-predicate"
)

var _ error = (*Error)(nil)

// Error is the domain error for predicate quantification operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for predicate quantification.
var (
	ErrEmptyCollection = errors.New("collection is empty")
	ErrNilPredicate    = errors.New("predicate is nil")
	ErrPredicateFailed = errors.New("predicate operation failed")
)

// ErrPredicate creates a predicate domain error joining the given causes.
func ErrPredicate(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PredicateType,
			Err:  errors.Join(append(errs, ErrPredicateFailed)...),
		},
	}
}
