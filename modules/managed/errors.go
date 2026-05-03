package managed

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for managed component errors.
const (
	ManagedType = "managed"
)

var _ error = (*Error)(nil)

// Error is the domain error for managed component operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for managed component failure modes.
var (
	ErrListenFailed    = errors.New("listen failed")
	ErrServeFailed     = errors.New("serve failed")
	ErrShutdownFailed  = errors.New("shutdown failed")
	ErrShutdownTimeout = errors.New("shutdown timeout")
	ErrStartFailed     = errors.New("start failed")
	ErrNotImplemented  = errors.New("not implemented")
)

// ErrListen creates a managed domain error joining the given causes with ErrListenFailed.
func ErrListen(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ManagedType,
			Err:  errors.Join(append(errs, ErrListenFailed)...),
		},
	}
}

// ErrServe creates a managed domain error joining the given causes with ErrServeFailed.
func ErrServe(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ManagedType,
			Err:  errors.Join(append(errs, ErrServeFailed)...),
		},
	}
}

// ErrShutdown creates a managed domain error joining the given causes with ErrShutdownFailed.
func ErrShutdown(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ManagedType,
			Err:  errors.Join(append(errs, ErrShutdownFailed)...),
		},
	}
}

// ErrStart creates a managed domain error joining the given causes with ErrStartFailed.
func ErrStart(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ManagedType,
			Err:  errors.Join(append(errs, ErrStartFailed)...),
		},
	}
}
