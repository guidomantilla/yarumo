package fsm

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for fsm errors.
const (
	FSMType = "math-fsm"
)

var _ error = (*Error)(nil)

// Error is the domain error for fsm operations.
type Error struct {
	cerrs.TypedError
}

// Error sentinels for fsm operations.
var (
	ErrStateNotFound       = errors.New("state not found")
	ErrTransitionNotFound  = errors.New("no transition for event from current state")
	ErrDuplicateState      = errors.New("duplicate state")
	ErrDuplicateTransition = errors.New("duplicate transition")
	ErrGuardRejected       = errors.New("guard rejected transition")
	ErrInvalidTransition   = errors.New("transition references unknown state")
	ErrNoInitialState      = errors.New("initial state not found")
	ErrInvalidEvent        = errors.New("event must not be empty")
	ErrFSMFailed           = errors.New("fsm operation failed")
)

// ErrFSM creates an fsm domain error joining the given causes.
func ErrFSM(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FSMType,
			Err:  errors.Join(append(errs, ErrFSMFailed)...),
		},
	}
}
