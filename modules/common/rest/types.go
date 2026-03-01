// Package rest provides a high-level REST client for executing HTTP requests
// with structured request and response specifications.
//
// Requests are described by RequestSpec and executed via Call (decoded response) or
// CallStream (raw streaming body). Configuration is provided through functional Options
// including the HTTP execution function (WithDoFn) and response size limits (WithMaxResponseSize).
//
// Error contract: operations wrap errors into a domain Error type with RequestType.
// Callers should prefer errors.Is/As instead of relying on string messages.
// Non-2xx responses are returned as HTTPError with status code, status text, and body.
//
// Concurrency: Call and CallStream are safe for concurrent use by multiple goroutines.
package rest

import "context"

var (
	_ CallFn[any]            = Call
	_ CallStreamFn           = CallStream
	_ ErrCallFn              = ErrCall
	_ DecodeHTTPErrorFn[any] = DecodeHTTPError[any]
)

// CallFn is the function type for Call.
type CallFn[T any] func(ctx context.Context, spec *RequestSpec, options ...Option) (*ResponseSpec[T], error)

// CallStreamFn is the function type for CallStream.
type CallStreamFn func(ctx context.Context, spec *RequestSpec, options ...Option) (*StreamResponseSpec, error)

// ErrCallFn is the function type for ErrCall.
type ErrCallFn func(errs ...error) error

// DecodeHTTPErrorFn is the function type for DecodeHTTPError.
type DecodeHTTPErrorFn[E any] func(err error) (E, bool)
