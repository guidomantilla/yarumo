package engine

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for engine errors.
const (
	EngineType = "inference-engine"
)

var _ error = (*Error)(nil)

// Error is the domain error for engine operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for engine failure modes.
var (
	ErrMaxIterations = errors.New("maximum iterations reached")
	ErrNoRules       = errors.New("no rules provided")
)

// ErrForward creates an engine domain error joining the given causes with ErrMaxIterations.
func ErrForward(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EngineType,
			Err:  errors.Join(append(errs, ErrMaxIterations)...),
		},
	}
}

// ErrBackward creates an engine domain error joining the given causes with ErrNoRules.
func ErrBackward(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EngineType,
			Err:  errors.Join(append(errs, ErrNoRules)...),
		},
	}
}
