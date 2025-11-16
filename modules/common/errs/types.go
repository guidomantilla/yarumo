package errs

import (
	"fmt"

	"github.com/guidomantilla/yarumo/common/assert"
)

type TypedError struct {
	Type string
	Err  error
}

func (e *TypedError) Error() string {
	assert.NotEmpty(e, "error is nil")
	assert.NotEmpty(e.Err, "internal error is nil")
	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

func (e *TypedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

//

type ErrorInfo struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}
