package errs

import "fmt"

type TypedError struct {
	Type string
	Err  error
}

func (e *TypedError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
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
