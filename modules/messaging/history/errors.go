package history

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// HistoryType is the error domain identifier for history operations.
const HistoryType = "history"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for history operations.
var (
	// ErrHistoryFailed is the top-level sentinel embedded in every
	// history-domain Error returned by ErrHistory.
	ErrHistoryFailed = errors.New("history failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for history operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("history %s error: %s", e.Type, e.Err)
}

// ErrHistory wraps the given causes into a domain Error joined with
// ErrHistoryFailed.
func ErrHistory(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HistoryType,
			Err:  errors.Join(append(causes, ErrHistoryFailed)...),
		},
	}
}
