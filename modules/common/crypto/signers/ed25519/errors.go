package ed25519

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	Ed25519Method = "ed25519_method"
)

var (
	_ error = (*Error)(nil)
)

type Error struct {
	cerrs.TypedError
}

func (e *Error) Error() string {
	assert.NotNil(e, "error is nil")
	assert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("ed25519 %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodIsNil            = errors.New("method is nil")
	ErrKeyIsNil               = errors.New("key is nil")
	ErrKeyLengthIsInvalid     = errors.New("key length is invalid")
	ErrSignatureLengthInvalid = errors.New("signature length is invalid")
	ErrKeyGenerationFailed    = errors.New("key generation failed")
	ErrSigningFailed          = errors.New("signing failed")
	ErrVerificationFailed     = errors.New("verification failed")
)

func ErrAlgorithmNotSupported(name string) error {
	return fmt.Errorf("ed25519 function %s not found", name)
}

func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: Ed25519Method,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}
