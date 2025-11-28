package tokens

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	TokenGenerationType = "generation"
	TokenValidationType = "validation"
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
	return fmt.Sprintf("token %s error: %s", e.Type, e.Err)
}

//

var (
	ErrAlgorithmNotSupported = errors.New("algorithm not supported")
	ErrSubjectCannotBeEmpty  = errors.New("subject cannot be empty")
	ErrPrincipalCannotBeNil  = errors.New("principal cannot be nil")
	ErrTokenExpired          = errors.New("token expired")
	ErrTokenCannotBeEmpty    = errors.New("token cannot be empty")
	ErrTokenFailedParsing    = errors.New("token failed to parse")
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
