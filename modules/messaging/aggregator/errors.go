package aggregator

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// AggregatorType is the error domain identifier for aggregator
// operations.
const AggregatorType = "aggregator"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for aggregator operations.
var (
	// ErrAggregateFailed is the top-level sentinel embedded in every
	// aggregator-domain Error returned by ErrAggregator.
	ErrAggregateFailed = errors.New("aggregate failed")
	// ErrAggregateFnFailed indicates that AggregateFn returned a non-nil
	// error or panicked. The original error (or the recovered value
	// formatted via fmt) is joined alongside this sentinel.
	ErrAggregateFnFailed = errors.New("aggregate function failed")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error after a successful AggregateFn.
	ErrForwardFailed = errors.New("forward to destination failed")
	// ErrMaxGroupsExceeded indicates that a new correlation key arrived
	// while the aggregator already tracked WithMaxGroups groups. The
	// new message is dropped and reported through the ErrorHandler;
	// existing groups remain untouched.
	ErrMaxGroupsExceeded = errors.New("max concurrent groups exceeded")
	// ErrGroupExpired indicates that a group was released by the
	// background sweeper after sitting idle past WithGroupTimeout. It
	// is joined with the error returned to the ErrorHandler when
	// aggregating the expired group fails, so observers can tell
	// timeout-driven aggregations apart from completion-driven ones.
	ErrGroupExpired = errors.New("group expired before completion")
)

// Error is the domain error type for aggregator operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("aggregator %s error: %s", e.Type, e.Err)
}

// ErrAggregator wraps the given causes into a domain Error joined with
// ErrAggregateFailed.
func ErrAggregator(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AggregatorType,
			Err:  errors.Join(append(causes, ErrAggregateFailed)...),
		},
	}
}
