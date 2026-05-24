package ristretto

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// CacheRistrettoType is the domain type tag attached to every Error produced for the ristretto backend.
const CacheRistrettoType = "cache-ristretto"

// Error is the domain error for ristretto cache backend operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for ristretto backend failure modes.
var (
	ErrRistrettoInitFailed  = errors.New("ristretto cache init failed")
	ErrRistrettoSetRejected = errors.New("ristretto cache rejected the set")
)

// ErrInit creates a ristretto cache domain error joining the given causes with ErrRistrettoInitFailed.
func ErrInit(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheRistrettoType,
			Err:  errors.Join(append(causes, ErrRistrettoInitFailed)...),
		},
	}
}

// ErrSet creates a ristretto cache domain error joining the given causes with ErrRistrettoSetRejected.
func ErrSet(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheRistrettoType,
			Err:  errors.Join(append(causes, ErrRistrettoSetRejected)...),
		},
	}
}
