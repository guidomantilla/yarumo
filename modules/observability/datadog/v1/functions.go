package v1

import (
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/rs/zerolog/log"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

type StopFn func()

func Trace(opts ...tracer.StartOption) StopFn {
	tracer.Start(opts...)
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
		err := statsd.Close()
		if err != nil {
			log.Error().Err(err).Str("stage", "shut down").Str("component", "datadog metrics").Msg("error closing metrics client")
		}
	}
	return statsd, stopFn
}
