// Package mcdm provides types shared across the MCDM sub-packages.
package mcdm

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for MCDM errors.
const (
	MCDMType = "inference-mcdm"
)

var _ error = (*Error)(nil)

// Error is the domain error for MCDM operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for MCDM failure modes.
var (
	ErrInvalidMatrix      = errors.New("invalid matrix dimensions")
	ErrNotSquareMatrix    = errors.New("matrix must be square")
	ErrInconsistentMatrix = errors.New("pairwise matrix is inconsistent")
	ErrEmptyMatrix        = errors.New("matrix is empty")
	ErrInvalidWeight      = errors.New("weights must be positive")
	ErrDimensionMismatch  = errors.New("dimensions do not match")
	ErrEmptyInput         = errors.New("empty input")
	ErrMCDMFailed         = errors.New("mcdm operation failed")
)

// ErrMCDM creates an MCDM domain error joining the given causes.
func ErrMCDM(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MCDMType,
			Err:  errors.Join(append(errs, ErrMCDMFailed)...),
		},
	}
}
