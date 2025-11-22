package rest

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	RequestType = "rest-request"
)

var (
	_ error = (*Error)(nil)
)

type Error struct {
	cerrs.TypedError
}

func (e *Error) Error() string {
	assert.NotEmpty(e, "error is nil")
	assert.NotEmpty(e.Err, "internal error is nil")
	return fmt.Sprintf("rest request %s error: %s", e.Type, e.Err)
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *HTTPError) Error() string {
	assert.NotEmpty(e, "error is nil")
	return fmt.Sprintf("unexpected status code %d: %s", e.StatusCode, e.Status)
}

type DecodeResponseError[T any] struct {
	ContentType string
	T           T
}

func (e *DecodeResponseError[T]) Error() string {
	assert.NotEmpty(e, "error is nil")
	assert.NotNil(e.T, "type is nil")
	return fmt.Sprintf("content type %s not supported for type  %T", e.ContentType, e.T)
}

//

func ErrCall(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RequestType,
			Err:  errors.Join(errs...),
		},
	}
}
