// Package fuzzy provides types shared across the fuzzy inference sub-packages.
package fuzzy

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for fuzzy inference errors.
const (
	FuzzyType = "inference-fuzzy"
)

var _ error = (*Error)(nil)

// Error is the domain error for fuzzy inference operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for fuzzy inference failure modes.
var (
	ErrVariableNotFound = errors.New("variable not found")
	ErrNoRules          = errors.New("no rules provided")
	ErrFuzzyFailed      = errors.New("fuzzy inference failed")
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
