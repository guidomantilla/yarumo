package hmacs

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	HmacMethod = "hmac_method"
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
	return fmt.Sprintf("hmac %s error: %s", e.Type, e.Err)
}

//

var (
	ErrMethodIsNil         = errors.New("method is nil")
	ErrHashNotAvailable    = errors.New("hash not available")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrDigestFailed        = errors.New("digest failed")
	ErrValidationFailed    = errors.New("validation failed")
)

func ErrAlgorithmNotSupported(name string) error {
	return fmt.Errorf("hmac function %s not found", name)
}

func ErrKeyGeneration(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrKeyGenerationFailed)...),
		},
	}
}

func ErrDigest(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrDigestFailed)...),
		},
	}
}

func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: HmacMethod,
			Err:  errors.Join(append(errs, ErrValidationFailed)...),
		},
	}
}
