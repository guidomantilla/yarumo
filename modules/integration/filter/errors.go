package filter

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// FilterType is the error domain identifier for filter operations.
const FilterType = "filter"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for filter operations.
var (
	// ErrFilterFailed is the top-level sentinel embedded in every
	// filter-domain Error returned by ErrFilter.
	ErrFilterFailed = errors.New("filter failed")
	// ErrPredicateFailed indicates that PredicateFn returned a non-nil
	// error. The original error is joined alongside this sentinel.
	ErrPredicateFailed = errors.New("predicate returned error")
	// ErrPredicatePanic indicates that PredicateFn panicked during
	// dispatch. The recovered value is embedded via fmt-formatting.
	ErrPredicatePanic = errors.New("predicate panicked")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for filter operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("filter %s error: %s", e.Type, e.Err)
}

// ErrFilter wraps the given causes into a domain Error joined with
// ErrFilterFailed.
func ErrFilter(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: FilterType,
			Err:  errors.Join(append(causes, ErrFilterFailed)...),
		},
	}
}
