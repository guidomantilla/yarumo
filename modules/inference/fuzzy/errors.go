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
	ErrTermNotFound     = errors.New("term not found")
	ErrNoRules          = errors.New("no rules provided")
	ErrNoInputs         = errors.New("no inputs provided")
	ErrInputOutOfRange  = errors.New("input value out of variable range")
)

// ErrInfer creates a fuzzy domain error joining the given causes with ErrNoRules.
func ErrInfer(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FuzzyType,
			Err:  errors.Join(append(errs, ErrNoRules)...),
		},
	}
}

// ErrValidation creates a fuzzy domain error joining the given causes with ErrVariableNotFound.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FuzzyType,
			Err:  errors.Join(append(errs, ErrVariableNotFound)...),
		},
	}
}
