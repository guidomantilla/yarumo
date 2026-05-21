package lifecycle

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// LifecycleType is the error domain identifier for lifecycle component errors.
const LifecycleType = "lifecycle"

var (
	_ error = (*Error)(nil)
)

// Error is the domain error type for lifecycle component operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for lifecycle component failure modes.
var (
	ErrShutdownFailed  = errors.New("shutdown failed")
	ErrShutdownTimeout = errors.New("shutdown timeout")
	ErrStartFailed     = errors.New("start failed")
)

// ErrStart creates a lifecycle domain error joining the given causes with ErrStartFailed.
func ErrStart(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: LifecycleType,
			Err:  errors.Join(append(errs, ErrStartFailed)...),
		},
	}
}

// ErrShutdown creates a lifecycle domain error joining the given causes with ErrShutdownFailed.
func ErrShutdown(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: LifecycleType,
			Err:  errors.Join(append(errs, ErrShutdownFailed)...),
		},
	}
}
