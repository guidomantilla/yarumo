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
	"google.golang.org/grpc"
)

type StopFn func(ctx context.Context)

func noopStop(ctx context.Context) {}

func Observe(ctx context.Context, conn *grpc.ClientConn, options ...Option) (StopFn, error) {

	stopTracer, err := Trace(ctx, conn, options...)
	if err != nil {
		return noopStop, err
	}

	stopMetrics, err := Measure(ctx, conn, options...)
	if err != nil {
		return noopStop, err
	}

	stopProfiler, err := Profile(ctx, conn, options...)
	if err != nil {
		return noopStop, err
	}

	stopLogger, err := Log(ctx, conn, options...)
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

func Trace(ctx context.Context, conn *grpc.ClientConn, options ...Option) (StopFn, error) {
	opts := NewOptions(options...)

	opts.tracePropagators = append(opts.tracePropagators, propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(opts.tracePropagators...))

	opts.traceExporterOptions = append(opts.traceExporterOptions, otlptracegrpc.WithGRPCConn(conn))
	exporter, err := otlptracegrpc.New(ctx, opts.traceExporterOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel tracer").Msg("error starting tracer")
		return noopStop, fmt.Errorf("error starting otel tracer: %w", err)
	}

	opts.traceProviderOptions = append(opts.traceProviderOptions, sdktrace.WithBatcher(exporter))
	tracerProvider := sdktrace.NewTracerProvider(opts.traceProviderOptions...)

	otel.SetTracerProvider(tracerProvider)

	stopFn := func(ctx context.Context) {
		err := tracerProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel tracer").Msg("error shutting down tracer")
		}
	}

	return stopFn, nil
}

func Profile(_ context.Context, _ *grpc.ClientConn, options ...Option) (StopFn, error) {
	opts := NewOptions(options...)

	opts.profileOptions = append(opts.profileOptions, runtimemetrics.WithMinimumReadMemStatsInterval(opts.profileInternal))
	err := runtimemetrics.Start(opts.profileOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel profiler").Msg("error starting err")
		return noopStop, fmt.Errorf("error starting otel profiler: %w", err)
	}

	// No-op stop: runtime metrics no exponen Stop()
	return noopStop, nil
}

func Measure(ctx context.Context, conn *grpc.ClientConn, options ...Option) (StopFn, error) {
	opts := NewOptions(options...)

	opts.metricExporterOptions = append(opts.metricExporterOptions, otlpmetricgrpc.WithGRPCConn(conn))
	exporter, err := otlpmetricgrpc.New(ctx, opts.metricExporterOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel metrics").Msg("error starting metrics")
		return noopStop, fmt.Errorf("error starting otel metrics: %w", err)
	}

	opts.metricProviderOptions = append(opts.metricProviderOptions, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(opts.metricInterval))))
	meterProvider := sdkmetric.NewMeterProvider(opts.metricProviderOptions...)

	otel.SetMeterProvider(meterProvider)

	stopFn := func(ctx context.Context) {
		err := meterProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel metrics").Msg("error shutting down metrics")
		}
	}

	return stopFn, nil
}

func Log(ctx context.Context, conn *grpc.ClientConn, options ...Option) (StopFn, error) {
	opts := NewOptions(options...)

	opts.logExporterOptions = append(opts.logExporterOptions, otlploggrpc.WithGRPCConn(conn))
	exporter, err := otlploggrpc.New(ctx, opts.logExporterOptions...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "otel logger").Msg("error starting logger")
		return noopStop, fmt.Errorf("error starting otel logger: %w", err)
	}

	/*
	 *	sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)), 	// for dev
	 *	sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),  		// for prod
	 */
	opts.logProviderOptions = append(opts.logProviderOptions, sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))
	loggerProvider := sdklog.NewLoggerProvider(opts.logProviderOptions...)

	global.SetLoggerProvider(loggerProvider)

	stopFn := func(ctx context.Context) {
		err := loggerProvider.Shutdown(ctx)
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "otel logger").Msg("error shutting down logger")
		}
	}

	return stopFn, nil
}
