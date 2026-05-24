package limiter

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// LimiterType is the domain type tag attached to every Error produced by this package.
const LimiterType = "rate-limiter"

// Error is the domain error for limiter Wait failures.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for limiter Wait failure modes.
var (
	// ErrWaitFailed indicates the Wait call could not acquire a token.
	ErrWaitFailed = errors.New("rate limiter wait failed")
	// ErrContextNil indicates Wait received a nil context.Context.
	ErrContextNil = errors.New("context is nil")
)

// ErrWait creates a limiter domain error joining the given causes with
// ErrWaitFailed.
func ErrWait(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: LimiterType,
			Err:  errors.Join(append(causes, ErrWaitFailed)...),
		},
	}
}
