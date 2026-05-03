package tokens

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	// TokenMethod is the error type for token operations.
	TokenMethod = "token_method"
)

var (
	_ error = (*Error)(nil)
)

// Error is the domain error for token operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	return fmt.Sprintf("token %s error: %s", e.Type, e.Err)
}

// Sentinel errors for token operations.
var (
	ErrSubjectEmpty      = errors.New("subject is empty")
	ErrPayloadNil        = errors.New("payload is nil")
	ErrTokenEmpty        = errors.New("token is empty")
	ErrSigningKeyNil     = errors.New("signing key is nil")
	ErrSigningMethodNil  = errors.New("signing method is nil")
	ErrTokenParseFailed  = errors.New("token parse failed")
	ErrTokenPayloadEmpty = errors.New("token payload is empty")
	ErrGenerationFailed  = errors.New("generation failed")
	ErrValidationFailed  = errors.New("validation failed")
)

// ErrGeneration creates a token generation error with optional cause errors.
func ErrGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TokenMethod,
			Err:  errors.Join(append(errs, ErrGenerationFailed)...),
		},
	}
}

// ErrValidation creates a token validation error with optional cause errors.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TokenMethod,
			Err:  errors.Join(append(errs, ErrValidationFailed)...),
		},
	}
}

// ErrAlgorithmNotSupported creates an error for unsupported algorithm lookups.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TokenMethod,
			Err:  fmt.Errorf("token method %s not found", name),
		},
	}
}
