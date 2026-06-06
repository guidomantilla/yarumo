package pollingconsumer

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// PollingConsumerType is the error domain identifier for polling
// consumer operations.
const PollingConsumerType = "pollingconsumer"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for polling consumer operations.
var (
	// ErrPollingConsumerFailed is the top-level sentinel embedded in
	// every polling-consumer-domain Error returned by
	// ErrPollingConsumer.
	ErrPollingConsumerFailed = errors.New("polling consumer failed")
	// ErrHandlerFailed indicates that the user-supplied Handler returned
	// a non-nil error during dispatch.
	ErrHandlerFailed = errors.New("handler returned error")
	// ErrHandlerPanic indicates that the user-supplied Handler panicked
	// during dispatch. The recovered value is embedded via fmt-formatting.
	ErrHandlerPanic = errors.New("handler panicked")
	// ErrPollFailed indicates that PollableChannel.Receive returned an
	// error other than the ones treated as a clean termination signal
	// (channel closed / ctx cancelled). The original error is joined
	// alongside this sentinel.
	ErrPollFailed = errors.New("poll receive failed")
)

// Error is the domain error type for polling consumer operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("pollingconsumer %s error: %s", e.Type, e.Err)
}

// ErrPollingConsumer wraps the given causes into a domain Error joined
// with ErrPollingConsumerFailed.
func ErrPollingConsumer(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PollingConsumerType,
			Err:  errors.Join(append(causes, ErrPollingConsumerFailed)...),
		},
	}
}
