package retry

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RetryType is the domain type tag attached to every Error produced by this package.
const RetryType = "retry"

// Error is the domain error for retry policy failures.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for retry policy failure modes.
var (
	// ErrRetryFailed indicates the retry policy gave up — either the
	// attempt budget was exhausted or the configured RetryIf predicate
	// rejected the last error.
	ErrRetryFailed = errors.New("retry policy gave up")
	// ErrContextNil indicates Do received a nil context.Context.
	ErrContextNil = errors.New("context is nil")
	// ErrFnNil indicates Do received a nil function to retry.
	ErrFnNil = errors.New("retry fn is nil")
)

// ErrRetry creates a retry domain error joining the given causes with
// ErrRetryFailed.
func ErrRetry(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RetryType,
			Err:  errors.Join(append(causes, ErrRetryFailed)...),
		},
	}
}
