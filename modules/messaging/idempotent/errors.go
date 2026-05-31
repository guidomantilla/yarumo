package idempotent

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// IdempotentType is the error domain identifier for idempotent
// receiver operations.
const IdempotentType = "idempotent"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for idempotent receiver operations.
var (
	// ErrIdempotentFailed is the top-level sentinel embedded in every
	// idempotent-domain Error returned by ErrIdempotent.
	ErrIdempotentFailed = errors.New("idempotent receiver failed")
	// ErrStoreCheck indicates that the MetadataStore.Has lookup
	// returned a non-nil error. The original store error is joined
	// alongside this sentinel. Has failures are fail-closed: the
	// message is NOT forwarded.
	ErrStoreCheck = errors.New("metadata store check failed")
	// ErrStoreAdd indicates that the MetadataStore.Add record returned
	// a non-nil error. The original store error is joined alongside
	// this sentinel. Add failures are fail-open: the message IS
	// forwarded despite the recording failure.
	ErrStoreAdd = errors.New("metadata store add failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for idempotent receiver operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("idempotent %s error: %s", e.Type, e.Err)
}

// ErrIdempotent wraps the given causes into a domain Error joined with
// ErrIdempotentFailed.
func ErrIdempotent(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: IdempotentType,
			Err:  errors.Join(append(causes, ErrIdempotentFailed)...),
		},
	}
}
