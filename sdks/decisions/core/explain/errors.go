package explain

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// ExplainType is the error type for explain errors.
const ExplainType = "explain"

var _ error = (*Error)(nil)

// Error is the domain error type for the explain package.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for explain operations.
var (
	ErrRenderFailed = errors.New("template render failed")
)

// ErrRender creates a render error from the given causes.
func ErrRender(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ExplainType,
			Err:  errors.Join(append(errs, ErrRenderFailed)...),
		},
	}
}
