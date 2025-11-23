package encoders

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	PasswordEncodingType  = "encoding"
	PasswordMatchingType  = "matching"
	PasswordUpgradingType = "upgrading"
)

var (
	_ error = (*TokenError)(nil)
)

type TokenError struct {
	cerrs.TypedError
}

func (e *TokenError) Error() string {
	assert.NotNil(e, "error is nil")
	assert.NotNil(e.Err, "internal error is nil")
	return fmt.Sprintf("password encoders %s error: %s", e.Type, e.Err)
}

//

var (
	ErrPasswordEncodingFailed    = errors.New("password encoding failed")
	ErrPasswordMatchingFailed    = errors.New("password matching failed")
	ErrPasswordUpgradingFailed   = errors.New("password upgrading failed")
	ErrPasswordLength            = errors.New("password length is too short")
	ErrPasswordSpecialChars      = errors.New("password must contain at least 2 special characters")
	ErrPasswordNumbers           = errors.New("password must contain at least 2 numbers")
	ErrPasswordUppercaseChars    = errors.New("password must contain at least 2 uppercase characters")
	ErrRawPasswordIsEmpty        = errors.New("rawPassword cannot be empty")
	ErrSaltIsNil                 = errors.New("salt cannot be nil")
	ErrSaltIsEmpty               = errors.New("salt cannot be empty")
	ErrHashFuncIsNil             = errors.New("hashFunc cannot be nil")
	ErrEncodedPasswordIsEmpty    = errors.New("encodedPassword cannot be empty")
	ErrEncodedPasswordNotAllowed = errors.New("encodedPassword format not allowed")
	ErrBcryptCostNotAllowed      = errors.New("bcryptCost not allowed")
)

func ErrPasswordEncoding(errs ...error) error {
	return &TokenError{
		TypedError: cerrs.TypedError{
			Type: PasswordEncodingType,
			Err:  errors.Join(append(errs, ErrPasswordEncodingFailed)...),
		},
	}
}

func ErrPasswordMatching(errs ...error) error {
	return &TokenError{
		TypedError: cerrs.TypedError{
			Type: PasswordMatchingType,
			Err:  errors.Join(append(errs, ErrPasswordMatchingFailed)...),
		},
	}
}

func ErrPasswordUpgrading(errs ...error) error {
	return &TokenError{
		TypedError: cerrs.TypedError{
			Type: PasswordMatchingType,
			Err:  errors.Join(append(errs, ErrPasswordUpgradingFailed)...),
		},
	}
}
