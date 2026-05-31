package headerfilter

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// HeaderFilterType is the error domain identifier for header filter
// operations.
const HeaderFilterType = "headerfilter"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for header filter operations.
var (
	// ErrHeaderFilterFailed is the top-level sentinel embedded in every
	// headerfilter-domain Error returned by ErrHeaderFilter.
	ErrHeaderFilterFailed = errors.New("header filter failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for header filter operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("headerfilter %s error: %s", e.Type, e.Err)
}

// ErrHeaderFilter wraps the given causes into a domain Error joined
// with ErrHeaderFilterFailed.
func ErrHeaderFilter(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HeaderFilterType,
			Err:  errors.Join(append(causes, ErrHeaderFilterFailed)...),
		},
	}
}
