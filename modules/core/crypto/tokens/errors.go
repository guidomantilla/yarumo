package tokens

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
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
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("token %s error: %s", e.Type, e.Err)
}

// Sentinel errors for token operations.
var (
	ErrMethodIsNil          = errors.New("method is nil")
	ErrSubjectEmpty         = errors.New("subject is empty")
	ErrPayloadNil           = errors.New("payload is nil")
	ErrTokenEmpty           = errors.New("token is empty")
	ErrSigningKeyNil        = errors.New("signing key is nil")
	ErrVerifyingKeyNil      = errors.New("verifying key is nil")
	ErrSigningMethodNil     = errors.New("signing method is nil")
	ErrTokenSignFailed      = errors.New("token signing failed")
	ErrTokenParseFailed     = errors.New("token parse failed")
	ErrTokenPayloadEmpty    = errors.New("token payload is empty")
	ErrGenerationFailed     = errors.New("generation failed")
	ErrValidationFailed     = errors.New("validation failed")
	ErrAlgorithmUnknown     = errors.New("algorithm is unknown")
	ErrCipherNil            = errors.New("cipher is nil")
	ErrTokenExpired         = errors.New("token is expired")
	ErrTokenNotYetValid     = errors.New("token is not yet valid")
	ErrTokenIssuerMismatch  = errors.New("token issuer mismatch")
	ErrTokenDecodeFailed    = errors.New("token decode failed")
	ErrTokenDecryptFailed   = errors.New("token decrypt failed")
	ErrTokenMarshalFailed   = errors.New("token marshal failed")
	ErrTokenUnmarshalFailed = errors.New("token unmarshal failed")
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

// ErrAlgorithmInvalid creates an error for unknown Algorithm enum values
// passed to NewMethod. The cause chain always contains ErrAlgorithmUnknown.
func ErrAlgorithmInvalid(algorithm Algorithm) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TokenMethod,
			Err:  errors.Join(fmt.Errorf("algorithm %q is not a recognized tokens.Algorithm value", string(algorithm)), ErrAlgorithmUnknown),
		},
	}
}
