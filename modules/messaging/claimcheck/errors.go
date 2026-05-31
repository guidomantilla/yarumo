package claimcheck

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// ClaimCheckType is the error domain identifier for claim check
// operations.
const ClaimCheckType = "claimcheck"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for claim check operations.
var (
	// ErrClaimCheckFailed is the top-level sentinel embedded in every
	// claimcheck-domain Error returned by ErrClaimCheck.
	ErrClaimCheckFailed = errors.New("claim check failed")
	// ErrStorePut indicates that MessageStore.Put returned a non-nil
	// error during ClaimCheckIn. The original store error is joined
	// alongside this sentinel. Put failures are fail-CLOSED: the
	// reference is NOT forwarded.
	ErrStorePut = errors.New("message store put failed")
	// ErrStoreGet indicates that MessageStore.Get returned a non-nil
	// error during ClaimCheckOut (including the "key not found" case
	// when the store reports it). The original store error is joined
	// alongside this sentinel. Get failures are fail-CLOSED: the
	// original is NOT forwarded.
	ErrStoreGet = errors.New("message store get failed")
	// ErrStoreDelete indicates that MessageStore.Delete returned a
	// non-nil error during ClaimCheckOut. The original store error is
	// joined alongside this sentinel. Delete failures are fail-OPEN:
	// the original IS forwarded — losing the chance to clean up the
	// store is preferable to losing the message.
	ErrStoreDelete = errors.New("message store delete failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for claim check operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("claimcheck %s error: %s", e.Type, e.Err)
}

// ErrClaimCheck wraps the given causes into a domain Error joined with
// ErrClaimCheckFailed.
func ErrClaimCheck(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ClaimCheckType,
			Err:  errors.Join(append(causes, ErrClaimCheckFailed)...),
		},
	}
}
