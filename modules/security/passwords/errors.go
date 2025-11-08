package passwords

import (
	"errors"
	"fmt"
)

var (
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

func ErrPasswordEncodingFailed(errs ...error) error {
	return fmt.Errorf("password encoding failed: %s", errors.Join(errs...).Error())
}

func ErrPasswordMatchingFailed(errs ...error) error {
	return fmt.Errorf("password matching failed: %s", errors.Join(errs...).Error())
}

func ErrPasswordUpgradeEncodingValidationFailed(errs ...error) error {
	return fmt.Errorf("password upgrade encoding validation failed: %s", errors.Join(errs...).Error())
}

func ErrPasswordValidationFailed(errs ...error) error {
	return fmt.Errorf("password validation failed: %s", errors.Join(errs...).Error())
}
