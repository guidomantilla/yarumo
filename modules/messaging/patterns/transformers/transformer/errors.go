package transformer

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// TransformerType is the error domain identifier for transformer
// operations.
const TransformerType = "transformer"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for transformer operations.
var (
	// ErrTransformerFailed is the top-level sentinel embedded in every
	// transformer-domain Error returned by ErrTransformer.
	ErrTransformerFailed = errors.New("transformer failed")
	// ErrTransformFailed indicates that TransformFn returned a non-nil
	// error. The original error is joined alongside this sentinel.
	ErrTransformFailed = errors.New("transform function returned error")
	// ErrTransformerPanic indicates that TransformFn panicked during
	// dispatch. The recovered value is embedded via fmt-formatting.
	ErrTransformerPanic = errors.New("transform function panicked")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for transformer operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("transformer %s error: %s", e.Type, e.Err)
}

// ErrTransformer wraps the given causes into a domain Error joined with
// ErrTransformerFailed.
func ErrTransformer(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: TransformerType,
			Err:  errors.Join(append(causes, ErrTransformerFailed)...),
		},
	}
}
