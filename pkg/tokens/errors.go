package tokens

import (
	"errors"
	"fmt"
)

var (
	ErrTokenFailedParsing = errors.New("token failed to parse")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTokenEmptyClaims   = errors.New("token claims is empty")
)

func ErrTokenGenerationFailed(errs ...error) error {
	return fmt.Errorf("token generation failed: %s", errors.Join(errs...).Error())
}

func ErrTokenValidationFailed(errs ...error) error {
	return fmt.Errorf("token validation failed: %s", errors.Join(errs...).Error())
}
