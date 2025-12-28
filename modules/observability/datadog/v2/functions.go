package v1

import (
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
	"github.com/rs/zerolog/log"
)

type StopFn func()

func Trace(opts ...tracer.StartOption) StopFn {
	err := tracer.Start(opts...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog tracer").Msg("error starting tracer")
	}
	stopFn := func() {
		tracer.Stop()
	}
	return stopFn
}

func Profile(opts ...profiler.Option) StopFn {
	err := profiler.Start(opts...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog profiler").Msg("error starting profiler")
	}
	stopFn := func() {
		profiler.Stop()
	}
	return stopFn
}

func Metrics(addr string, opts ...statsd.Option) (*statsd.Client, StopFn) {
	statsd, err := statsd.New(addr, opts...)
	if err != nil {
		log.Fatal().Err(err).Str("stage", "startup").Str("component", "datadog metrics").Msg("error starting metrics client")
	}
	stopFn := func() {
		err := statsd.Flush()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error flushing metrics client")
		}
		err = statsd.Close()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error closing metrics client")
		}
	}
	return statsd, stopFn
}
