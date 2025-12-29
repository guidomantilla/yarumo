package otel

import (
	"context"
	"fmt"

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
)

type StopFn func(ctx context.Context)

func noopStop(ctx context.Context) {}

func Observe(ctx context.Context, options ...Option) (StopFn, error) {

	stopTracer, err := Tracer(ctx, options...)
	if err != nil {
		return noopStop, err
	}

	stopMetrics, err := Meter(ctx, options...)
	if err != nil {
		return noopStop, err
	}

	stopProfiler, err := Profiler(ctx, options...)
	if err != nil {
		return noopStop, err
	}

	stopLogger, err := Logger(ctx, options...)
	if err != nil {
		return noopStop, err
	}

	stopFn := func(ctx context.Context) {
		stopLogger(ctx)
		stopProfiler(ctx)
		stopMetrics(ctx)
		stopTracer(ctx)
	}
	return stopFn, nil
}

func Tracer(ctx context.Context, options ...Option) (StopFn, error) {
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

	opts.tracerProviderOptions = append(opts.tracerProviderOptions, sdktrace.WithBatcher(exporter))
	tracerProvider := sdktrace.NewTracerProvider(opts.tracerProviderOptions...)

	otel.SetTracerProvider(tracerProvider)

	stopFn := func(ctx context.Context) {
		err := tracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel tracer").Msg("error shutting down tracer")
		}
	}

	return stopFn, nil
}

func Profiler(_ context.Context, _ ...Option) (StopFn, error) {
	return noopStop, nil
}

func Meter(ctx context.Context, options ...Option) (StopFn, error) {
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

	opts.meterProviderOptions = append(opts.meterProviderOptions, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(opts.meterInterval))))
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

	stopFn := func(ctx context.Context) {
		err := meterProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel meter").Msg("error shutting down meter")
		}
	}

	return stopFn, nil
}

func Logger(ctx context.Context, options ...Option) (StopFn, error) {
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
	opts.loggerProviderOptions = append(opts.loggerProviderOptions, sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)))
	loggerProvider := sdklog.NewLoggerProvider(opts.loggerProviderOptions...)

	global.SetLoggerProvider(loggerProvider)

	stopFn := func(ctx context.Context) {
		err := loggerProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel logger").Msg("error shutting down logger")
		}
	}

	return stopFn, nil
}
