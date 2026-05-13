package validation

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ValidationType is the error type classification for validation failures.
const ValidationType = "validation"

var _ error = (*Error)(nil)

// Error is the domain error for validation failures. It embeds errs.TypedError
// so AsErrorInfo groups violations under the "validation" type.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted validation error message.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("validation %s error: %s", e.Type, e.Err)
}

// Sentinel errors for the validation leaves and reflection helpers.
var (
	ErrValidationFailed   = errors.New("validation failed")
	ErrFieldRequired      = errors.New("field is required")
	ErrFieldMustBeUndefined = errors.New("field must be undefined")
	ErrMinLen             = errors.New("value is shorter than the minimum length")
	ErrMaxLen             = errors.New("value is longer than the maximum length")
	ErrRegexInvalid       = errors.New("regex pattern is invalid")
	ErrRegexMismatch      = errors.New("value does not match the required pattern")
	ErrEmailInvalid       = errors.New("value is not a valid email")
	ErrURLInvalid         = errors.New("value is not a valid URL")
	ErrMinValue           = errors.New("value is below the minimum")
	ErrMaxValue           = errors.New("value is above the maximum")
	ErrOutOfRange         = errors.New("value is out of the allowed range")
	ErrInvalidRange       = errors.New("min must be less than or equal to max")
	ErrUUIDInvalid        = errors.New("value is not a valid UUID")
	ErrULIDInvalid        = errors.New("value is not a valid ULID")
	ErrCollectionEmpty    = errors.New("collection is empty")
	ErrEachFailed         = errors.New("collection element failed validation")
	ErrPathInvalid        = errors.New("field path is invalid")
	ErrPathNotFound       = errors.New("field path not found")
	ErrPathTypeMismatch   = errors.New("field path traversal hit incompatible type")
	ErrIndexOutOfRange    = errors.New("slice index is out of range")
	ErrObjectNil          = errors.New("target object is nil")
)

// ErrValidation creates a validation domain error joining the given causes with
// ErrValidationFailed.
func ErrValidation(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ValidationType,
			Err:  errors.Join(append(causes, ErrValidationFailed)...),
		},
	}
}
