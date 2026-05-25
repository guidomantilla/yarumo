package messaging

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// MessagingType is the error domain identifier for messaging operations.
const MessagingType = "messaging"

var (
	_ error = (*Error)(nil)

	_ ErrSendFn      = ErrSend
	_ ErrSubscribeFn = ErrSubscribe
)

// ErrSendFn is the function type for ErrSend.
type ErrSendFn func(causes ...error) error

// ErrSubscribeFn is the function type for ErrSubscribe.
type ErrSubscribeFn func(causes ...error) error

// Sentinel errors for messaging operations.
var (
	// ErrSendFailed indicates that a Send operation failed.
	ErrSendFailed = errors.New("send failed")
	// ErrSubscribeFailed indicates that a Subscribe operation failed.
	ErrSubscribeFailed = errors.New("subscribe failed")
	// ErrClosed indicates that the channel has been closed and no
	// longer accepts Send or Subscribe.
	ErrClosed = errors.New("channel closed")
	// ErrHandlerNil indicates that a nil handler was passed to
	// Subscribe.
	ErrHandlerNil = errors.New("handler is nil")
	// ErrContextNil indicates that a nil context was passed to Send.
	ErrContextNil = errors.New("context is nil")
	// ErrTimeout indicates that an operation timed out (e.g. enqueue
	// blocked past the configured deadline).
	ErrTimeout = errors.New("operation timed out")
	// ErrDrainTimeout indicates that the queue did not finish draining
	// pending messages before Stop's context deadline expired.
	ErrDrainTimeout = errors.New("drain timeout")
)

// Error is the domain error type for messaging operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("messaging %s error: %s", e.Type, e.Err)
}

// ErrSend wraps the given causes into a domain Error for Send failures.
func ErrSend(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MessagingType,
			Err:  errors.Join(append(causes, ErrSendFailed)...),
		},
	}
}

// ErrSubscribe wraps the given causes into a domain Error for
// Subscribe failures.
func ErrSubscribe(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: MessagingType,
			Err:  errors.Join(append(causes, ErrSubscribeFailed)...),
		},
	}
}
