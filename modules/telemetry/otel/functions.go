package otel

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/guidomantilla/yarumo/managed"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func noopStop(ctx context.Context, timeout time.Duration) {}

type LoggerHookFn func(ctx context.Context) (context.Context, error)

func Observe(ctx context.Context, serviceName string, serviceVersion string, env string, hookFn LoggerHookFn, options ...Option) (context.Context, managed.StopFn, error) {

	res, err := Resources(ctx, serviceName, serviceVersion, env)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up resources: %w", err)
	}

	options = append(options, WithResource(res))

	stopLogger, err := Logger(ctx, options...)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up logger: %w", err)
	}

	hookedCtx, err := hookFn(ctx)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up logger hook: %w", err)
	}

	stopTracer, err := Tracer(ctx, options...)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up tracer: %w", err)
	}

	stopMetrics, err := Meter(ctx, options...)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up meter: %w", err)
	}

	stopProfiler, err := Profiler(ctx, options...)
	if err != nil {
		return ctx, noopStop, fmt.Errorf("error setting up profiler: %w", err)
	}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		stopProfiler(ctx, timeout)
		stopMetrics(ctx, timeout)
		stopTracer(ctx, timeout)
		stopLogger(ctx, timeout)
	}

	return hookedCtx, stopFn, nil
}

func Resources(ctx context.Context, serviceName string, serviceVersion string, env string) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithFromEnv(), resource.WithTelemetrySDK(), resource.WithHost(),
		resource.WithAttributes(semconv.ServiceName(serviceName), semconv.ServiceVersion(serviceVersion), semconv.DeploymentEnvironment(env)),
	)

	if err != nil {
		return nil, fmt.Errorf("error creating resource: %w", err)
	}

	return res, nil
}

func Tracer(ctx context.Context, options ...Option) (managed.StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", "otel tracer").Msg("starting up")

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
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel tracer").Msg("error starting tracer")
		return noopStop, fmt.Errorf("error starting otel tracer: %w", err)
	}

	opts.tracerProviderOptions = append(opts.tracerProviderOptions, sdktrace.WithResource(opts.resource), sdktrace.WithBatcher(exporter))
	tracerProvider := sdktrace.NewTracerProvider(opts.tracerProviderOptions...)

	otel.SetTracerProvider(tracerProvider)

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel tracer").Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel tracer").Msg("stopped")

		timeoutCtx, cancelTimeoutFn := context.WithTimeout(ctx, timeout)
		defer cancelTimeoutFn()

		err = tracerProvider.Shutdown(timeoutCtx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel tracer").Msg("error shutting down tracer")
		}
	}

	return stopFn, nil
}

func Profiler(_ context.Context, _ ...Option) (managed.StopFn, error) {
	return noopStop, nil
}

func Meter(ctx context.Context, options ...Option) (managed.StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", "otel meter").Msg("starting up")

	opts := NewOptions(options...)

	if opts.secure {
		opts.meterExporterOptions = append(opts.meterExporterOptions, otlpmetricgrpc.WithEndpoint(opts.endpoint))
	} else {
		opts.meterExporterOptions = append(opts.meterExporterOptions, otlpmetricgrpc.WithEndpoint(opts.endpoint), otlpmetricgrpc.WithInsecure())
	}
	exporter, err := otlpmetricgrpc.New(ctx, opts.meterExporterOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel meter").Msg("error starting meter")
		return noopStop, fmt.Errorf("error starting otel meter: %w", err)
	}

	opts.meterProviderOptions = append(opts.meterProviderOptions, sdkmetric.WithResource(opts.resource), sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(opts.meterInterval))))
	meterProvider := sdkmetric.NewMeterProvider(opts.meterProviderOptions...)

	otel.SetMeterProvider(meterProvider)

	if opts.meterRuntimeMetricsEnabled {
		opts.meterRuntimeMetricsOptions = append(opts.meterRuntimeMetricsOptions, runtimemetrics.WithMinimumReadMemStatsInterval(opts.meterRuntimeMetricsInterval))
		err = runtimemetrics.Start(opts.meterRuntimeMetricsOptions...)
		if err != nil {
			log.Error().Err(err).Str("stage", "startup").Str("component", "otel meter").Msg("error starting meter")
			return noopStop, fmt.Errorf("error starting otel meter: %w", err)
		}
	}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel meter").Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel meter").Msg("stopped")

		timeoutCtx, cancelTimeoutFn := context.WithTimeout(ctx, timeout)
		defer cancelTimeoutFn()

		err = meterProvider.Shutdown(timeoutCtx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel meter").Msg("error shutting down meter")
		}
	}

	return stopFn, nil
}

func Logger(ctx context.Context, options ...Option) (managed.StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", "otel logger").Msg("starting up")

	opts := NewOptions(options...)

	if opts.secure {
		opts.loggerExporterOptions = append(opts.loggerExporterOptions, otlploggrpc.WithEndpoint(opts.endpoint))
	} else {
		opts.loggerExporterOptions = append(opts.loggerExporterOptions, otlploggrpc.WithEndpoint(opts.endpoint), otlploggrpc.WithInsecure())
	}
	exporter, err := otlploggrpc.New(ctx, opts.loggerExporterOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel logger").Msg("error starting logger")
		return noopStop, fmt.Errorf("error starting otel logger: %w", err)
	}

	/*
	 *	sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)), 	// for dev
	 *	sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),  		// for prod
	 */
	opts.loggerProviderOptions = append(opts.loggerProviderOptions, sdklog.WithResource(opts.resource), sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	loggerProvider := sdklog.NewLoggerProvider(opts.loggerProviderOptions...)

	global.SetLoggerProvider(loggerProvider)

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel logger").Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", "otel logger").Msg("stopped")

		timeoutCtx, cancelTimeoutFn := context.WithTimeout(ctx, timeout)
		defer cancelTimeoutFn()

		err = loggerProvider.Shutdown(timeoutCtx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel logger").Msg("error shutting down logger")
		}
	}

	return stopFn, nil
}
