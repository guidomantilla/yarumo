package tokens

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/pkg/common/errs"
)

const (
	TokenGenerationType = "generation"
	TokenValidationType = "validation"
)

type TokenError struct {
	cerrs.TypedError
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("token %s error: %s", e.Type, e.Err)
}

//

var (
	ErrTokenFailedParsing    = errors.New("token failed to parse")
	ErrTokenInvalid          = errors.New("token is invalid")
	ErrTokenEmptyClaims      = errors.New("token claims is empty")
	ErrTokenEmptyPrincipal   = errors.New("token principal is empty")
	ErrTokenGenerationFailed = errors.New("token generation failed")
	ErrTokenValidationFailed = errors.New("token validation failed")
)

func ErrTokenGeneration(errs ...error) error {
	return &TokenError{
		TypedError: cerrs.TypedError{
			Type: TokenGenerationType,
			Err:  errors.Join(append(errs, ErrTokenGenerationFailed)...),
		},
	}
}

func ErrTokenValidation(errs ...error) error {
	return &TokenError{
		TypedError: cerrs.TypedError{
			Type: TokenValidationType,
			Err:  errors.Join(append(errs, ErrTokenValidationFailed)...),
		},
	}
}
