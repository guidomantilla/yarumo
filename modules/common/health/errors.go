package health

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for health-check operations.
const (
	HealthType = "health"
)

var _ error = (*Error)(nil)

// Error is the domain error for health-check operations.
type Error struct {
	cerrs.TypedError
}

// Error returns a formatted error message including the error type and cause.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("health %s error: %s", e.Type, e.Err)
}

// Sentinel errors for health-check failure modes.
var (
	// ErrHealthFailed is the umbrella sentinel for any health-check failure.
	ErrHealthFailed = errors.New("health check failed")
	// ErrCheckNil is returned when a nil [Check] is registered.
	ErrCheckNil = errors.New("check is nil")
	// ErrContextNil is returned when a nil context is passed to Status.
	ErrContextNil = errors.New("context is nil")
)

// ErrHealth creates a health domain error joining the given causes with ErrHealthFailed.
func ErrHealth(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HealthType,
			Err:  errors.Join(append(errs, ErrHealthFailed)...),
		},
	}
}
