package passwords

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	// PasswordMethod is the error type for password operations.
	PasswordMethod = "password_method"
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
	return fmt.Sprintf("password %s error: %s", e.Type, e.Err)
}

// Sentinel errors for password operations.
var (
	ErrRawPasswordEmpty      = errors.New("raw password is empty")
	ErrEncodedPasswordEmpty  = errors.New("encoded password is empty")
	ErrEncodedPasswordFormat = errors.New("encoded password format not allowed")
	ErrSaltGenerationFailed  = errors.New("salt generation failed")
	ErrEncodingFailed        = errors.New("encoding failed")
	ErrVerificationFailed    = errors.New("verification failed")
	ErrUpgradeCheckFailed    = errors.New("upgrade check failed")
	ErrMethodConfigMissing   = errors.New("method has no algorithm configuration")
	ErrBcryptCostNotAllowed  = errors.New("bcrypt cost not allowed")
)

// ErrEncoding creates a password encoding error with optional cause errors.
func ErrEncoding(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordMethod,
			Err:  errors.Join(append(errs, ErrEncodingFailed)...),
		},
	}
}

// ErrVerification creates a password verification error with optional cause errors.
func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordMethod,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}

// ErrUpgradeCheck creates a password upgrade check error with optional cause errors.
func ErrUpgradeCheck(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordMethod,
			Err:  errors.Join(append(errs, ErrUpgradeCheckFailed)...),
		},
	}
}

// ErrAlgorithmNotSupported creates an error for unsupported algorithm lookups.
func ErrAlgorithmNotSupported(name string) *Error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: PasswordMethod,
			Err:  fmt.Errorf("password method %s not found", name),
		},
	}
}
