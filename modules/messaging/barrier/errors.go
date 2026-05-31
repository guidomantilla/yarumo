package barrier

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// BarrierType is the error domain identifier for barrier operations.
const BarrierType = "barrier"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for barrier operations.
var (
	// ErrBarrierFailed is the top-level sentinel embedded in every
	// barrier-domain Error returned by ErrBarrier.
	ErrBarrierFailed = errors.New("barrier failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error during release.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for barrier operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("barrier %s error: %s", e.Type, e.Err)
}

// ErrBarrier wraps the given causes into a domain Error joined with
// ErrBarrierFailed.
func ErrBarrier(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BarrierType,
			Err:  errors.Join(append(causes, ErrBarrierFailed)...),
		},
	}
}
