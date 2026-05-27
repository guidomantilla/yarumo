package limiter

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// LimiterType is the domain type tag attached to every Error produced by this package.
const LimiterType = "http-limiter"

// Error is the domain error for limiter transport operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for limiter transport failure modes.
var (
	ErrRateLimiterFailed = errors.New("rate limiter exceeded")
)

// ErrRateLimiterExceeded creates a limiter domain error joining the given causes with ErrRateLimiterFailed.
func ErrRateLimiterExceeded(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: LimiterType,
			Err:  errors.Join(append(causes, ErrRateLimiterFailed)...),
		},
	}
}
