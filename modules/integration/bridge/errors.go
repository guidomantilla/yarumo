package bridge

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// BridgeType is the error domain identifier for bridge operations.
const BridgeType = "bridge"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for bridge operations.
var (
	// ErrBridgeFailed is the top-level sentinel embedded in every
	// bridge-domain Error returned by ErrBridge.
	ErrBridgeFailed = errors.New("bridge failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for bridge operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("bridge %s error: %s", e.Type, e.Err)
}

// ErrBridge wraps the given causes into a domain Error joined with
// ErrBridgeFailed.
func ErrBridge(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BridgeType,
			Err:  errors.Join(append(causes, ErrBridgeFailed)...),
		},
	}
}
