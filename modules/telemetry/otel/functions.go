package otel

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/common/log"
	"github.com/guidomantilla/yarumo/managed"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func noopStop(_ context.Context, _ time.Duration) {}

// Observe sets up full OpenTelemetry observability (logging, tracing, metrics, profiling) and returns a combined stop function.
func Observe(ctx context.Context, serviceName string, serviceVersion string, env string, hookFn LoggerHookFn, options ...Option) (context.Context, managed.StopFn, error) {

	res, err := Resources(ctx, serviceName, serviceVersion, env)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrResourceFailed, err)
	}

	options = append(options, WithResource(res))

	stopLogger, err := Logger(ctx, options...)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrLoggerFailed, err)
	}

	hookedCtx, err := hookFn(ctx)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrHookFailed, err)
	}

	stopTracer, err := Tracer(ctx, options...)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrTracerFailed, err)
	}

	stopMetrics, err := Meter(ctx, options...)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrMeterFailed, err)
	}

	stopProfiler, err := Profiler(ctx, options...)
	if err != nil {
		return ctx, noopStop, ErrObserve(ErrProfilerFailed, err)
	}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		stopProfiler(ctx, timeout)
		stopMetrics(ctx, timeout)
		stopTracer(ctx, timeout)
		stopLogger(ctx, timeout)
	}

	return hookedCtx, stopFn, nil
}

// Resources creates an OpenTelemetry resource with the given service name, version, and environment.
func Resources(ctx context.Context, serviceName string, serviceVersion string, env string) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(), resource.WithTelemetrySDK(), resource.WithHost(),
		resource.WithAttributes(semconv.ServiceName(serviceName), semconv.ServiceVersion(serviceVersion), semconv.DeploymentEnvironment(env)),
	)

	if err != nil {
		return nil, ErrResource(err)
	}

	return res, nil
}

// Tracer sets up an OpenTelemetry trace provider with OTLP gRPC exporter.
func Tracer(ctx context.Context, options ...Option) (managed.StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", "otel tracer")

	opts := NewOptions(options...)

	opts.tracerPropagators = append(opts.tracerPropagators, propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(opts.tracerPropagators...))

	if opts.secure {
		opts.tracerExporterOptions = append(opts.tracerExporterOptions, otlptracegrpc.WithEndpoint(opts.endpoint))
	} else {
		opts.tracerExporterOptions = append(opts.tracerExporterOptions, otlptracegrpc.WithEndpoint(opts.endpoint), otlptracegrpc.WithInsecure())
	}
	exporter, err := otlptracegrpc.New(ctx, opts.tracerExporterOptions...)
	if err != nil {
		clog.Error(ctx, "error starting tracer", "stage", "startup", "component", "otel tracer", "error", err)
		return noopStop, ErrTracer(err)
	}

	opts.tracerProviderOptions = append(opts.tracerProviderOptions, sdktrace.WithResource(opts.resource), sdktrace.WithBatcher(exporter))
	tracerProvider := sdktrace.NewTracerProvider(opts.tracerProviderOptions...)

	otel.SetTracerProvider(tracerProvider)

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", "otel tracer")
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", "otel tracer")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := tracerProvider.Shutdown(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "error shutting down tracer", "stage", "shutdown", "component", "otel tracer", "error", err)
		}
	}

	return stopFn, nil
}

// Profiler is a placeholder for future profiling provider setup.
func Profiler(_ context.Context, _ ...Option) (managed.StopFn, error) {
	return noopStop, nil
}

// Meter sets up an OpenTelemetry meter provider with OTLP gRPC exporter.
func Meter(ctx context.Context, options ...Option) (managed.StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", "otel meter")

	opts := NewOptions(options...)

	if opts.secure {
		opts.meterExporterOptions = append(opts.meterExporterOptions, otlpmetricgrpc.WithEndpoint(opts.endpoint))
	} else {
		opts.meterExporterOptions = append(opts.meterExporterOptions, otlpmetricgrpc.WithEndpoint(opts.endpoint), otlpmetricgrpc.WithInsecure())
	}
	exporter, err := otlpmetricgrpc.New(ctx, opts.meterExporterOptions...)
	if err != nil {
		clog.Error(ctx, "error starting meter", "stage", "startup", "component", "otel meter", "error", err)
		return noopStop, ErrMeter(err)
	}

	opts.meterProviderOptions = append(opts.meterProviderOptions, sdkmetric.WithResource(opts.resource), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(opts.meterInterval))))
	meterProvider := sdkmetric.NewMeterProvider(opts.meterProviderOptions...)

	otel.SetMeterProvider(meterProvider)

	if opts.meterRuntimeMetricsEnabled {
		opts.meterRuntimeMetricsOptions = append(opts.meterRuntimeMetricsOptions, runtimemetrics.WithMinimumReadMemStatsInterval(opts.meterRuntimeMetricsInterval))
		err = runtimemetrics.Start(opts.meterRuntimeMetricsOptions...)
		if err != nil {
			clog.Error(ctx, "error starting runtime metrics", "stage", "startup", "component", "otel meter", "error", err)
			return noopStop, ErrMeter(err)
		}
	}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", "otel meter")
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", "otel meter")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := meterProvider.Shutdown(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "error shutting down meter", "stage", "shutdown", "component", "otel meter", "error", err)
		}
	}

	return stopFn, nil
}

// Logger sets up an OpenTelemetry logger provider with OTLP gRPC exporter.
func Logger(ctx context.Context, options ...Option) (managed.StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", "otel logger")

	opts := NewOptions(options...)

	if opts.secure {
		opts.loggerExporterOptions = append(opts.loggerExporterOptions, otlploggrpc.WithEndpoint(opts.endpoint))
	} else {
		opts.loggerExporterOptions = append(opts.loggerExporterOptions, otlploggrpc.WithEndpoint(opts.endpoint), otlploggrpc.WithInsecure())
	}
	exporter, err := otlploggrpc.New(ctx, opts.loggerExporterOptions...)
	if err != nil {
		clog.Error(ctx, "error starting logger", "stage", "startup", "component", "otel logger", "error", err)
		return noopStop, ErrLogger(err)
	}

	opts.loggerProviderOptions = append(opts.loggerProviderOptions, sdklog.WithResource(opts.resource), sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	loggerProvider := sdklog.NewLoggerProvider(opts.loggerProviderOptions...)

	global.SetLoggerProvider(loggerProvider)

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", "otel logger")
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", "otel logger")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := loggerProvider.Shutdown(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "error shutting down logger", "stage", "shutdown", "component", "otel logger", "error", err)
		}
	}

	return stopFn, nil
}
