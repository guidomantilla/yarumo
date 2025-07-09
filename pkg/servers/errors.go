package servers

import (
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/pkg/common/errs"
)

const (
	ServerStartType = "start"
	ServerStopType  = "stop"
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
