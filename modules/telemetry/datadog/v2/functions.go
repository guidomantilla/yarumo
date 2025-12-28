package v2

import (
	"context"
	"sync/atomic"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
	"github.com/rs/zerolog/log"
)

type StopFn func(ctx context.Context)

func Observe(ctx context.Context, options ...Option) StopFn {
	opts := NewOptions(options...)

	stopTracer := Trace(ctx, opts.tracerOptions...)
	stopProfiler := Profile(ctx, opts.profilerOptions...)
	stopMetrics := Metrics(ctx, opts.metricsAddr, opts.metricsOptions...)

	stopFn := func(ctx context.Context) {
		stopTracer(ctx)
		stopProfiler(ctx)
		stopMetrics(ctx)
	}
	return stopFn
}

func Trace(ctx context.Context, options ...tracer.StartOption) StopFn {
	err := tracer.Start(options...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog tracer").Msg("error starting tracer")
	}
	stopFn := func(ctx context.Context) {
		tracer.Stop()
	}
	return stopFn
}

func Profile(ctx context.Context, options ...profiler.Option) StopFn {
	err := profiler.Start(options...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog profiler").Msg("error starting profiler")
	}
	stopFn := func(ctx context.Context) {
		profiler.Stop()
	}
	return stopFn
}

func Metrics(ctx context.Context, addr string, options ...statsd.Option) StopFn {
	statsd, err := statsd.New(addr, options...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog metrics").Msg("error starting metrics client")
	}
	stopFn := func(ctx context.Context) {
		err := statsd.Flush()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error flushing metrics client")
		}
		err = statsd.Close()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error closing metrics client")
		}
	}
	instance.Store(statsd)
	return stopFn
}

/*
 */
var instance atomic.Value

func GetMetricsClient() *statsd.Client {
	v := instance.Load()
	return v.(*statsd.Client)
}
