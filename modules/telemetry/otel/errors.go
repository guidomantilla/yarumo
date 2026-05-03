package otel

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for OpenTelemetry setup errors.
const (
	OtelType = "otel"
)

var _ error = (*Error)(nil)

// Error is the domain error for OpenTelemetry setup operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for OpenTelemetry setup failure modes.
var (
	ErrResourceFailed = errors.New("resource creation failed")
	ErrTracerFailed   = errors.New("tracer setup failed")
	ErrMeterFailed    = errors.New("meter setup failed")
	ErrLoggerFailed   = errors.New("logger setup failed")
	ErrHookFailed     = errors.New("logger hook failed")
	ErrObserveFailed  = errors.New("observe setup failed")
	ErrProfilerFailed = errors.New("profiler setup failed")
)

// ErrResource creates an otel domain error joining the given causes with ErrResourceFailed.
func ErrResource(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: OtelType,
			Err:  errors.Join(append(errs, ErrResourceFailed)...),
		},
	}
}

// ErrTracer creates an otel domain error joining the given causes with ErrTracerFailed.
func ErrTracer(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: OtelType,
			Err:  errors.Join(append(errs, ErrTracerFailed)...),
		},
	}
}

// ErrMeter creates an otel domain error joining the given causes with ErrMeterFailed.
func ErrMeter(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: OtelType,
			Err:  errors.Join(append(errs, ErrMeterFailed)...),
		},
	}
}

// ErrLogger creates an otel domain error joining the given causes with ErrLoggerFailed.
func ErrLogger(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: OtelType,
			Err:  errors.Join(append(errs, ErrLoggerFailed)...),
		},
	}
}

// ErrObserve creates an otel domain error joining the given causes with ErrObserveFailed.
func ErrObserve(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: OtelType,
			Err:  errors.Join(append(errs, ErrObserveFailed)...),
		},
	}
}
