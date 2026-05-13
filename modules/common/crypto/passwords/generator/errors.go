package generator

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	// PasswordGenerator is the error type for password generator operations.
	PasswordGenerator = "password_generator"
)

var (
	_ error = (*Error)(nil)
)

// Error is the domain error for password generator operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("password %s error: %s", e.Type, e.Err)
}

// Sentinel errors for password generator operations.
var (
	ErrGeneratorIsNil          = errors.New("generator is nil")
	ErrPasswordEmpty           = errors.New("raw password is empty")
	ErrConstraintsExceedLength = errors.New("sum of minimum character constraints exceeds password length")
	ErrInvalidOption           = errors.New("invalid generator option")
	ErrShuffleFailed           = errors.New("crypto/rand-backed shuffle failed")
	ErrGenerationFailed        = errors.New("password generation failed")
	ErrValidationFailed        = errors.New("password validation failed")
	ErrPasswordLength          = errors.New("password length is too short")
	ErrPasswordSpecialChars    = errors.New("password does not have enough special characters")
	ErrPasswordNumbers         = errors.New("password does not have enough numeric characters")
	ErrPasswordUppercaseChars  = errors.New("password does not have enough uppercase characters")
	ErrPasswordLowercaseChars  = errors.New("password does not have enough lowercase characters")
)

// ErrConfiguration creates a generator configuration error with optional cause errors.
func ErrConfiguration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordGenerator,
			Err:  errors.Join(append(errs, ErrInvalidOption)...),
		},
	}
}

// ErrGeneration creates a password generation error with optional cause errors.
func ErrGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordGenerator,
			Err:  errors.Join(append(errs, ErrGenerationFailed)...),
		},
	}
}

// ErrValidation creates a password validation error with optional cause errors.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordGenerator,
			Err:  errors.Join(append(errs, ErrValidationFailed)...),
		},
	}
}
