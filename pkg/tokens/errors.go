package tokens

import (
	"errors"
	"fmt"
)

var (
	ErrTokenFailedParsing         = errors.New("token failed to parse")
	ErrTokenInvalid               = errors.New("token is invalid")
	ErrTokenEmptyClaims           = errors.New("token claims is empty")
	ErrTokenEmptyUsernameClaim    = errors.New("token username claim is empty")
	ErrTokenEmptyRoleClaim        = errors.New("token role claim is empty")
	ErrTokenEmptyResourcesClaim   = errors.New("token resources claim is empty")
	ErrTokenInvalidResourcesClaim = errors.New("token resources claim is invalid")
)

func ErrTokenGenerationFailed(errs ...error) error {
	return fmt.Errorf("token generation failed: %s", errors.Join(errs...).Error())
}

func ErrTokenValidationFailed(errs ...error) error {
	return fmt.Errorf("token validation failed: %s", errors.Join(errs...).Error())
}
