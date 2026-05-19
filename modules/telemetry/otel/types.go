// Package otel provides OpenTelemetry observability setup for tracing, metrics, and logging.
package otel

import (
	"context"

	"github.com/guidomantilla/yarumo/managed"
	"go.opentelemetry.io/otel/sdk/resource"
)

// ObserveFn is the function type for setting up full observability (tracing, metrics, logging).
type ObserveFn func(ctx context.Context, serviceName string, serviceVersion string, env string, options ...Option) (context.Context, managed.StopFn, error)

// ResourcesFn is the function type for creating an OpenTelemetry resource.
type ResourcesFn func(ctx context.Context, serviceName string, serviceVersion string, env string) (*resource.Resource, error)

// ProviderFn is the function type for setting up an OpenTelemetry provider.
type ProviderFn func(ctx context.Context, options ...Option) (managed.StopFn, error)

var (
	_ ObserveFn   = Observe
	_ ResourcesFn = Resources
	_ ProviderFn  = Tracer
	_ ProviderFn  = Meter
	_ ProviderFn  = Logger
)
