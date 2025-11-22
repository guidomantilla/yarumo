package servers

import (
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

const (
	ServerStartType = "start"
	ServerStopType  = "stop"
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
	return e.Err.Error()
}

func ErrServerFailedToStart(name string, err error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ServerStartType,
			Err:  fmt.Errorf("server %s failed to start: %w", name, err),
		},
	}
}

func ErrServerFailedToStop(name string, err error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ServerStopType,
			Err:  fmt.Errorf("server %s failed to stop: %w", name, err),
		},
	}
}
