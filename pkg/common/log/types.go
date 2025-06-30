package log

import "github.com/rs/zerolog"

var (
	_ ConfigureFn = Configure
)

type ConfigureFn func(name string, version string, opts ...Option) zerolog.Logger
