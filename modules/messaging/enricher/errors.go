package enricher

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// EnricherType is the error domain identifier for enricher operations.
const EnricherType = "enricher"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for enricher operations.
var (
	// ErrEnricherFailed is the top-level sentinel embedded in every
	// enricher-domain Error returned by ErrEnricher.
	ErrEnricherFailed = errors.New("enricher failed")
	// ErrEnrichFnFailed indicates that EnrichFn returned a non-nil
	// error. The original error is joined alongside this sentinel.
	ErrEnrichFnFailed = errors.New("enrich function returned error")
	// ErrEnrichPanic indicates that EnrichFn panicked during dispatch.
	// The recovered value is embedded via fmt-formatting.
	ErrEnrichPanic = errors.New("enrich function panicked")
	// ErrForwardFailed indicates that the destination Channel.Send
	// returned a non-nil error.
	ErrForwardFailed = errors.New("forward to destination failed")
)

// Error is the domain error type for enricher operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("enricher %s error: %s", e.Type, e.Err)
}

// ErrEnricher wraps the given causes into a domain Error joined with
// ErrEnricherFailed.
func ErrEnricher(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EnricherType,
			Err:  errors.Join(append(causes, ErrEnricherFailed)...),
		},
	}
}
