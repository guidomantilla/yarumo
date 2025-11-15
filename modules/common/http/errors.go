package http

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	RequestType = "http-request"
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
	return fmt.Sprintf("http request %s error: %s", e.Type, e.Err)
}

//

var (
	ErrRateLimiterExceeded = errors.New("rate limit exceeded")
	ErrHttpRequestFailed   = errors.New("request failed")
)

func ErrDo(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RequestType,
			Err:  errors.Join(errs...),
		},
	}
}
