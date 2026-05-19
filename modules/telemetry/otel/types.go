// Package otel provides OpenTelemetry observability setup for tracing, metrics, and logging.
package otel

import (
	"context"
	"time"

	"github.com/guidomantilla/yarumo/managed"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	_ error = (*Error)(nil)

	_ ObserveFn   = Observe
	_ ResourcesFn = Resources
	_ ProviderFn  = Tracer
	_ ProviderFn  = Meter
	_ ProviderFn  = Logger

	_ ErrFn = ErrResource
	_ ErrFn = ErrTracer
	_ ErrFn = ErrMeter
	_ ErrFn = ErrLogger
	_ ErrFn = ErrObserve

	_ StatusFn        = TracerStatus
	_ StatusFn        = MeterStatus
	_ StatusFn        = LoggerStatus
	_ ObserveStatusFn = ObserveStatus
)

// ExporterStatus reports the current state of an OTLP exporter. It is updated
// by the per-provider recording wrappers installed during Tracer / Meter /
// Logger construction; callers obtain a snapshot via TracerStatus,
// MeterStatus, LoggerStatus, or the aggregate ObserveStatus.
//
// Connected is true when the last Export call succeeded (or no Export has
// run yet — exporters start optimistic). LastError / LastErrorTime carry
// the most recent failure for diagnostics. DroppedCount is the cumulative
// count of records lost in failed exports (one batch can hold many records);
// it never decreases. SuccessCount is the symmetric counter for successful
// exports.
type ExporterStatus struct {
	Connected     bool
	LastError     error
	LastErrorTime time.Time
	DroppedCount  int64
	SuccessCount  int64
}

// ObserveFn is the function type for setting up full observability (tracing, metrics, logging).
type ObserveFn func(ctx context.Context, serviceName string, serviceVersion string, env string, options ...Option) (context.Context, managed.StopFn, error)

// ResourcesFn is the function type for creating an OpenTelemetry resource.
type ResourcesFn func(ctx context.Context, serviceName string, serviceVersion string, env string) (*resource.Resource, error)

// ProviderFn is the function type for setting up an OpenTelemetry provider.
type ProviderFn func(ctx context.Context, options ...Option) (managed.StopFn, error)

// ErrFn is the function type for the package's error factories (ErrResource,
// ErrTracer, ErrMeter, ErrLogger, ErrObserve). They all share the same
// signature: variadic causes joined with a sentinel wrapped in *Error.
type ErrFn func(causes ...error) error

// StatusFn is the function type for the per-provider status accessors
// (TracerStatus, MeterStatus, LoggerStatus).
type StatusFn func() ExporterStatus

// ObserveStatusFn is the function type for ObserveStatus.
type ObserveStatusFn func() (tracer ExporterStatus, meter ExporterStatus, logger ExporterStatus)
