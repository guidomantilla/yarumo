package errs

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// TypedError is a base error type that carries a type classification string.
// Designed to be embedded in domain-specific error structs.
type TypedError struct {
	Type string
	Err  error
}

// NewTypedError creates a new TypedError with the given type classification and inner error.
func NewTypedError(typ string, err error) error {
	cassert.NotEmpty(typ, "type is empty")
	cassert.NotNil(err, "error is nil")

	return &TypedError{Type: typ, Err: err}
}

// Error returns the formatted error string including the type classification.
func (e *TypedError) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Unwrap returns the inner wrapped error.
func (e *TypedError) Unwrap() error {
	cassert.NotNil(e, "error is nil")
	return e.Err
}

// ErrorType returns the type classification string.
func (e *TypedError) ErrorType() string {
	cassert.NotNil(e, "error is nil")
	return e.Type
}

// ErrorInfo is a JSON-serializable representation of errors grouped by type classification.
type ErrorInfo struct {
	Type     string   `json:"type,omitempty"`
	Messages []string `json:"messages,omitempty"`
}
