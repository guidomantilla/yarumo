package resequencer

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// ResequencerType is the error domain identifier for resequencer
// operations.
const ResequencerType = "resequencer"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for resequencer operations.
var (
	// ErrResequencerFailed is the top-level sentinel embedded in
	// every resequencer-domain Error returned by ErrResequencer.
	ErrResequencerFailed = errors.New("resequencer failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error during emit.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for resequencer operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("resequencer %s error: %s", e.Type, e.Err)
}

// ErrResequencer wraps the given causes into a domain Error joined
// with ErrResequencerFailed.
func ErrResequencer(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ResequencerType,
			Err:  errors.Join(append(causes, ErrResequencerFailed)...),
		},
	}
}
