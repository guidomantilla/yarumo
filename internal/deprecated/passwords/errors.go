package passwords

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

var (
	_ error = (*Error)(nil)
)

// Error is the domain error for password operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	return fmt.Sprintf("password validation error: %s", e.Err)
}

// Sentinel errors for password validation.
var (
	ErrValidationFailed       = errors.New("validation failed")
	ErrPasswordLength         = errors.New("password length is too short")
	ErrPasswordSpecialChars   = errors.New("password does not have enough special characters")
	ErrPasswordNumbers        = errors.New("password does not have enough numbers")
	ErrPasswordUppercaseChars = errors.New("password does not have enough uppercase characters")
)

// ErrValidation creates a password validation error with optional cause errors.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: "password_validation",
			Err:  errors.Join(append(errs, ErrValidationFailed)...),
		},
	}
}
