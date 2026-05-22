package uids

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error type constants for UID operations.
const (
	UidNotFound        = "uid_function_not_found"
	UidGenerationError = "uid_generation_error"
)

// ErrGenerationFailed is returned by Generate (and the underlying UIDFn) when
// the entropy source backing the algorithm fails. Callers can match it with
// errors.Is to distinguish generator failures from other errors. The typed
// wrapper returned by ErrGeneration also wraps this sentinel.
var ErrGenerationFailed = errors.New("uid generation failed")

// Error is a domain error type for UID operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("uid %s error: %s", e.Type, e.Err)
}

// ErrAlgorithmNotSupported creates an error indicating the requested UID algorithm is not registered.
func ErrAlgorithmNotSupported(name string) error {
	cassert.NotEmpty(name, "name is empty")

	return &Error{
		TypedError: cerrs.TypedError{
			Type: UidNotFound,
			Err:  fmt.Errorf("uid algorithm %s not found", name),
		},
	}
}

// ErrGeneration wraps one or more underlying provider errors as a typed
// uid_generation_error so that errs.AsErrorInfo can classify them.
// ErrGenerationFailed is joined into the chain so callers can still match
// the sentinel via errors.Is. Provider sub-modules call this when a
// generator's entropy source fails.
func ErrGeneration(errs ...error) error {
	joined := cerrs.Wrap(append([]error{ErrGenerationFailed}, errs...)...)

	return &Error{
		TypedError: cerrs.TypedError{
			Type: UidGenerationError,
			Err:  joined,
		},
	}
}
