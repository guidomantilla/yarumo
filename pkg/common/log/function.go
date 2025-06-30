package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Configure(name string, version string, opts ...Option) zerolog.Logger {
	options := NewOptions(opts...)
	logger := zerolog.New(os.Stdout).With().
		Str("name", name).Str("version", version).
		Timestamp()

	if options.Caller {
		logger = logger.Caller()
	}

	log.Logger = logger.Logger()
	return log.Logger
}
