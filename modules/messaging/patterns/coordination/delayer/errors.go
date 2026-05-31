package delayer

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// DelayerType is the error domain identifier for delayer operations.
const DelayerType = "delayer"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for delayer operations.
var (
	// ErrDelayerFailed is the top-level sentinel embedded in every
	// delayer-domain Error returned by ErrDelayer.
	ErrDelayerFailed = errors.New("delayer failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error after the delay elapsed.
	ErrForwardFailed = errors.New("forward to destination failed")
	// ErrScheduleFailed indicates that the internal scheduled channel
	// rejected the deferred enqueue (typically because the scheduler is
	// stopped or its ctx expired).
	ErrScheduleFailed = errors.New("schedule deferred delivery failed")
	// ErrMaxPendingExceeded indicates that the number of in-flight
	// pending messages reached the configured WithMaxPending bound and
	// the new message was dropped instead of being scheduled.
	ErrMaxPendingExceeded = errors.New("max pending messages exceeded")
)

// Error is the domain error type for delayer operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("delayer %s error: %s", e.Type, e.Err)
}

// ErrDelayer wraps the given causes into a domain Error joined with
// ErrDelayerFailed.
func ErrDelayer(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: DelayerType,
			Err:  errors.Join(append(causes, ErrDelayerFailed)...),
		},
	}
}
