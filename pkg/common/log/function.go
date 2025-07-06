package log

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

func Configure(name string, version string, opts ...Option) zerolog.Logger {
	options := NewOptions(opts...)
	logger := zerolog.New(os.Stderr).With()

	if utils.NotEmpty(name) {
		logger = logger.Str("name", name)
	}
	if utils.NotEmpty(version) {
		logger = logger.Str("version", version)
	}

	if options.Caller {
		logger = logger.Caller()
	}

	log.Logger = logger.Timestamp().Logger()
	return log.Logger
}
