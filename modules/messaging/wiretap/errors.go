package wiretap

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// WiretapType is the error domain identifier for wiretap operations.
const WiretapType = "wiretap"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for wiretap operations.
var (
	// ErrWiretapFailed is the top-level sentinel embedded in every
	// wiretap-domain Error returned by ErrWiretap.
	ErrWiretapFailed = errors.New("wiretap failed")
	// ErrForwardFailed indicates that the primary destination
	// Channel.Send returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
	// ErrTapSendFailed indicates that the tap-side Channel.Send returned
	// a non-nil error. This NEVER alters the primary flow — it only
	// fires the ErrorHandler so the "observability of the observability"
	// is visible to operators.
	ErrTapSendFailed = errors.New("forward to tap failed")
)

// Error is the domain error type for wiretap operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("wiretap %s error: %s", e.Type, e.Err)
}

// ErrWiretap wraps the given causes into a domain Error joined with
// ErrWiretapFailed.
func ErrWiretap(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: WiretapType,
			Err:  errors.Join(append(causes, ErrWiretapFailed)...),
		},
	}
}
