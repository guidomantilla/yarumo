package splitter

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// SplitterType is the error domain identifier for splitter operations.
const SplitterType = "splitter"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for splitter operations.
var (
	// ErrSplitterFailed is the top-level sentinel embedded in every
	// splitter-domain Error returned by ErrSplitter.
	ErrSplitterFailed = errors.New("splitter failed")
	// ErrSplitFailed indicates that SplitFn returned a non-nil error.
	// The original error is joined alongside this sentinel.
	ErrSplitFailed = errors.New("split function returned error")
	// ErrSplitterPanic indicates that SplitFn panicked during dispatch.
	// The recovered value is embedded via fmt-formatting.
	ErrSplitterPanic = errors.New("split function panicked")
	// ErrForwardFailed indicates that the destination Channel.Send for
	// a child message returned a non-nil error. The splitter reports
	// the failure and continues emitting the remaining children.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for splitter operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("splitter %s error: %s", e.Type, e.Err)
}

// ErrSplitter wraps the given causes into a domain Error joined with
// ErrSplitterFailed.
func ErrSplitter(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: SplitterType,
			Err:  errors.Join(append(causes, ErrSplitterFailed)...),
		},
	}
}
