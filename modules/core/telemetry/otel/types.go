// Package otel provides OpenTelemetry observability setup for tracing, metrics, and logging.
package otel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
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
)

// ObserveFn is the function type for setting up full observability (tracing, metrics, logging).
type ObserveFn func(ctx context.Context, serviceName string, serviceVersion string, env string, options ...Option) (context.Context, lifecycle.CloseFn, error)

// ResourcesFn is the function type for creating an OpenTelemetry resource.
type ResourcesFn func(ctx context.Context, serviceName string, serviceVersion string, env string) (*resource.Resource, error)

// ProviderFn is the function type for setting up an OpenTelemetry provider.
type ProviderFn func(ctx context.Context, options ...Option) (lifecycle.CloseFn, error)

// ErrFn is the function type for the package's error factories (ErrResource,
// ErrTracer, ErrMeter, ErrLogger, ErrObserve). They all share the same
// signature: variadic causes joined with a sentinel wrapped in *Error.
type ErrFn func(causes ...error) error
