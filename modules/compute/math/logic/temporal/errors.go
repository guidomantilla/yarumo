package temporal

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for temporal assertion errors.
const (
	TemporalType = "math-temporal"
)

var _ error = (*Error)(nil)

// Error is the domain error for temporal assertion operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for temporal assertions.
var (
	// ErrEventNotFound indicates that a required event label is not present in the trace.
	ErrEventNotFound = errors.New("event not found in trace")
	// ErrTemporalFailed is the generic fallback wrapped by ErrTemporal.
	ErrTemporalFailed = errors.New("temporal operation failed")
)

// ErrTemporal creates a temporal domain error joining the given causes.
func ErrTemporal(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TemporalType,
			Err:  errors.Join(append(errs, ErrTemporalFailed)...),
		},
	}
}
