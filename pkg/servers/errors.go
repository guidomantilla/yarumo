package servers

import (
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/modules/common/errs"
)

const (
	ServerStartType = "start"
	ServerStopType  = "stop"
)

var (
	_ error = (*ServerError)(nil)
)

type ServerError struct {
	cerrs.TypedError
}

func ErrServerFailedToStart(name string, err error) error {
	return &ServerError{
		TypedError: cerrs.TypedError{
			Type: ServerStartType,
			Err:  fmt.Errorf("server %s failed to start: %w", name, err),
		},
	}
}

func ErrServerFailedToStop(name string, err error) error {
	return &ServerError{
		TypedError: cerrs.TypedError{
			Type: ServerStopType,
			Err:  fmt.Errorf("server %s failed to stop: %w", name, err),
		},
	}
}
