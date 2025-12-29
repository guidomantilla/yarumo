package v2

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
	"github.com/rs/zerolog/log"
)

type StopFn func(ctx context.Context)

func noopStop(ctx context.Context) {}

func Observe(ctx context.Context, options ...Option) (StopFn, error) {
	opts := NewOptions(options...)

	stopTracer, err := Trace(ctx, opts.tracerOptions...)
	if err != nil {
		return noopStop, err
	}

	stopProfiler, err := Profile(ctx, opts.profilerOptions...)
	if err != nil {
		return noopStop, err
	}

	stopMetrics, err := Metrics(ctx, opts.metricsAddr, opts.metricsOptions...)
	if err != nil {
		return noopStop, err
	}

	stopFn := func(ctx context.Context) {
		stopTracer(ctx)
		stopProfiler(ctx)
		stopMetrics(ctx)
	}
	return stopFn, nil
}

func Trace(ctx context.Context, options ...tracer.StartOption) (StopFn, error) {
	err := tracer.Start(options...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "datadog tracer").Msg("error starting tracer")
		return noopStop, fmt.Errorf("error starting datadog tracer: %w", err)
	}
	stopFn := func(ctx context.Context) {
		tracer.Stop()
	}
	return stopFn, nil
}

func Profile(ctx context.Context, options ...profiler.Option) (StopFn, error) {
	err := profiler.Start(options...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "datadog profiler").Msg("error starting profiler")
		return noopStop, fmt.Errorf("error starting datadog profiler: %w", err)
	}
	stopFn := func(ctx context.Context) {
		profiler.Stop()
	}
	return stopFn, nil
}

func Metrics(ctx context.Context, addr string, options ...statsd.Option) (StopFn, error) {
	statsd, err := statsd.New(addr, options...)
	if err != nil {
		log.Error().Err(err).Str("stage", "startup").Str("component", "datadog metrics").Msg("error starting metrics client")
		return noopStop, fmt.Errorf("error starting datadog metrics: %w", err)
	}
	stopFn := func(ctx context.Context) {
		err := statsd.Flush()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error shutting down metrics client")
		}
		err = statsd.Close()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error shutting down metrics client")
		}
	}
	instance.Store(statsd)
	return stopFn, nil
}

/*
 */
var instance atomic.Value

func GetMetricsClient() *statsd.Client {
	v := instance.Load()
	return v.(*statsd.Client)
}
