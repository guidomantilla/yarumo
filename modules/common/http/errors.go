package http

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
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
	assert.NotEmpty(e, "error is nil")
	assert.NotEmpty(e.Err, "internal error is nil")
	return fmt.Sprintf("http request %s error: %s", e.Type, e.Err)
}

type StatusCodeError struct {
	StatusCode int
}

func (e *StatusCodeError) Error() string {
	assert.NotEmpty(e, "error is nil")
	return fmt.Sprintf("http retryable status code: %d", e.StatusCode)
}

//

var (
	ErrRateLimiterExceeded   = errors.New("rate limit exceeded")
	ErrHttpRequestFailed     = errors.New("request failed")
	ErrHttpNonReplayableBody = errors.New("http non-replayable request body")
	ErrHttpGetBodyFailed     = errors.New("http get body failed")
)

func ErrDo(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RequestType,
			Err:  errors.Join(errs...),
		},
	}
}
