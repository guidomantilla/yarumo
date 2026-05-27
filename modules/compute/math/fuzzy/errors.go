package fuzzy

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for fuzzy math errors.
const (
	FuzzyType = "math-fuzzy"
)

var _ error = (*Error)(nil)

// Error is the domain error for fuzzy math operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for the fuzzy package.
var (
	ErrEmptySamples = errors.New("empty sample set")
	ErrInvalidRange = errors.New("invalid range")
	ErrFuzzyFailed  = errors.New("fuzzy operation failed")
)

// ErrFuzzy creates a fuzzy domain error joining the given causes.
func ErrFuzzy(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FuzzyType,
			Err:  errors.Join(append(errs, ErrFuzzyFailed)...),
		},
	}
}
