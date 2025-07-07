package log

import "github.com/rs/zerolog"

var (
	_ ConfigureFn = Configure
)

type EventFn func(e *zerolog.Event)

type ConfigureFn func(name string, version string) zerolog.Logger
