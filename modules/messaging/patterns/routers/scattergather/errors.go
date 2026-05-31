package scattergather

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// ScatterGatherType is the error domain identifier for scatter-gather
// operations.
const ScatterGatherType = "scattergather"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for scatter-gather operations.
var (
	// ErrScatterGatherFailed is the top-level sentinel embedded in
	// every scattergather-domain Error returned by ErrScatterGather.
	ErrScatterGatherFailed = errors.New("scatter-gather failed")
	// ErrMaxScattersExceeded indicates that a new request arrived
	// while the scatter-gather already tracked
	// WithMaxConcurrentScatters in-flight gathers. The new request is
	// rejected without scattering and reported through the
	// ErrorHandler; in-flight gathers are untouched.
	ErrMaxScattersExceeded = errors.New("max concurrent scatters exceeded")
	// ErrScatterFailed indicates that the internal Recipient List
	// reported an error while scattering the request to workers
	// (missing worker key, forward Send failed, selector error or
	// panic). The original error is joined alongside this sentinel.
	ErrScatterFailed = errors.New("scatter to workers failed")
	// ErrGatherFailed indicates that the internal Aggregator reported
	// an error while gathering or aggregating replies. The original
	// error is joined alongside this sentinel.
	ErrGatherFailed = errors.New("gather of replies failed")
)

// Error is the domain error type for scatter-gather operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("scattergather %s error: %s", e.Type, e.Err)
}

// ErrScatterGather wraps the given causes into a domain Error joined
// with ErrScatterGatherFailed.
func ErrScatterGather(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: ScatterGatherType,
			Err:  errors.Join(append(causes, ErrScatterGatherFailed)...),
		},
	}
}
