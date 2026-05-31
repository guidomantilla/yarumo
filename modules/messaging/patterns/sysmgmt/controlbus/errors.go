package controlbus

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// ControlBusType is the error domain identifier for control-bus
// operations.
const ControlBusType = "controlbus"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for control-bus operations.
var (
	// ErrControlBusFailed is the top-level sentinel embedded in every
	// control-bus-domain Error returned by ErrControlBus.
	ErrControlBusFailed = errors.New("controlbus failed")
	// ErrHandlerPanic indicates that a Handler panicked during
	// dispatch. The recovered value is embedded via fmt-formatting.
	ErrHandlerPanic = errors.New("handler panicked")
	// ErrForwardFailed indicates that the reply Channel.Send returned
	// a non-nil error.
	ErrForwardFailed = errors.New("forward to reply channel failed")
)

// Error is the domain error type for control-bus operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("controlbus %s error: %s", e.Type, e.Err)
}

// ErrControlBus wraps the given causes into a domain Error joined with
// ErrControlBusFailed.
func ErrControlBus(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ControlBusType,
			Err:  errors.Join(append(causes, ErrControlBusFailed)...),
		},
	}
}
