package ecdsas

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	EcdsaMethod = "ecdsa_method"
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
	return fmt.Sprintf("ecdsa %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrKeyIsNil            = errors.New("key is nil")
	ErrKeyCurveIsInvalid   = errors.New("key curve is invalid")
	ErrSignatureInvalid    = errors.New("signature is invalid")
	ErrSignFailed          = errors.New("sign failed")
	ErrFormatUnsupported   = errors.New("format unsupported")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrSigningFailed       = errors.New("signing failed")
	ErrVerificationFailed  = errors.New("verification failed")
)

func ErrAlgorithmNotSupported(name string) error {
	return fmt.Errorf("ecdsa function %s not found", name)
}

func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

func ErrSigning(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrSigningFailed)...),
		},
	}
}

func ErrVerification(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EcdsaMethod,
			Err:  errors.Join(append(errs, ErrVerificationFailed)...),
		},
	}
}
