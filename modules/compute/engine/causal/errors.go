// Package causal provides types shared across the causal inference sub-packages.
package causal

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for causal inference errors.
const (
	CausalType = "inference-causal"
)

var _ error = (*Error)(nil)

// Error is the domain error for causal inference operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for causal inference failure modes.
var (
	ErrCyclicModel       = errors.New("model contains a cycle")
	ErrVariableNotFound  = errors.New("variable not found in model")
	ErrDuplicateVariable = errors.New("duplicate variable")
	ErrNilEquation       = errors.New("equation function is nil")
	ErrParentNotFound    = errors.New("parent variable not found")
	ErrCausalFailed      = errors.New("causal inference failed")
)

// ErrCausal creates a causal domain error joining the given causes.
func ErrCausal(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CausalType,
			Err:  errors.Join(append(errs, ErrCausalFailed)...),
		},
	}
}
