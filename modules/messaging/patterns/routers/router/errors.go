package router

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// RouterType is the error domain identifier for router operations.
const RouterType = "router"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for router operations.
var (
	// ErrRouteFailed is the top-level sentinel embedded in every
	// router-domain Error returned by ErrRoute.
	ErrRouteFailed = errors.New("route failed")
	// ErrNoRoute indicates that RouteFn returned a key that is not
	// present in the routes map and no default channel was configured.
	ErrNoRoute = errors.New("no route matches")
	// ErrRouteFnFailed indicates that RouteFn returned a non-nil error.
	// The original error is joined alongside this sentinel.
	ErrRouteFnFailed = errors.New("route function returned error")
	// ErrRoutePanic indicates that RouteFn panicked during dispatch.
	// The recovered value is embedded via fmt-formatting.
	ErrRoutePanic = errors.New("route function panicked")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for router operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("router %s error: %s", e.Type, e.Err)
}

// ErrRoute wraps the given causes into a domain Error joined with
// ErrRouteFailed.
func ErrRoute(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RouterType,
			Err:  errors.Join(append(causes, ErrRouteFailed)...),
		},
	}
}
