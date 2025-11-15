package rest

import (
	"errors"
	"fmt"

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
	if e == nil || e.Err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("rest request %s error: %s", e.Type, e.Err)
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("unexpected status code %d: %s", e.StatusCode, e.Status)
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
