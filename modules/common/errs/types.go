package errs

import "fmt"

type TypedError struct {
	Type string
	Err  error
}

func (e *TypedError) Error() string {
	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

func (e *TypedError) Unwrap() error {
	return e.Err
}

//

type ErrorInfo struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}
