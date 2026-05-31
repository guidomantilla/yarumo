package recipientlist

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// RecipientListType is the error domain identifier for recipient list
// operations.
const RecipientListType = "recipientlist"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for recipient list operations.
var (
	// ErrRecipientListFailed is the top-level sentinel embedded in every
	// recipientlist-domain Error returned by ErrRecipientList.
	ErrRecipientListFailed = errors.New("recipient list failed")
	// ErrNoRoute indicates that SelectorFn returned a key that is not
	// present in the routes map. Reported once per missing key; other
	// recipients in the same call are still delivered.
	ErrNoRoute = errors.New("no route matches")
	// ErrSelectorFnFailed indicates that SelectorFn returned a non-nil
	// error. The original error is joined alongside this sentinel.
	ErrSelectorFnFailed = errors.New("selector function returned error")
	// ErrSelectorPanic indicates that SelectorFn panicked during
	// dispatch. The recovered value is embedded via fmt-formatting.
	ErrSelectorPanic = errors.New("selector function panicked")
	// ErrForwardFailed indicates that a destination Channel.Send
	// returned a non-nil error. Reported once per failing recipient;
	// other recipients in the same call are still delivered.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for recipient list operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("recipientlist %s error: %s", e.Type, e.Err)
}

// ErrRecipientList wraps the given causes into a domain Error joined
// with ErrRecipientListFailed.
func ErrRecipientList(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RecipientListType,
			Err:  errors.Join(append(causes, ErrRecipientListFailed)...),
		},
	}
}
