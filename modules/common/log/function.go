package log

import (
	"io"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/common/utils"
)

func Configure(name string, version string) zerolog.Logger {
	writers := []io.Writer{os.Stderr}

	if viper.IsSet("LOGSTASH_ADDRESS") {
		conn, err := net.Dial("tcp", viper.GetString("LOGSTASH_ADDRESS"))
		if err == nil {
			writers = append(writers, conn)
		}
	}

	logger := zerolog.New(zerolog.MultiLevelWriter(writers...)).With()
	if utils.NotEmpty(name) {
		logger = logger.Str("name", name)
	}

	if utils.NotEmpty(version) {
		logger = logger.Str("version", version)
	}

	debugMode := utils.Coalesce(viper.GetBool("DEBUG_MODE"), false)
	if debugMode {
		logger = logger.Caller()
	}

	level, err := zerolog.ParseLevel(utils.Coalesce(viper.GetString("LOG_LEVEL"), "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}

	if level >= zerolog.TraceLevel && level < zerolog.Disabled {
		zerolog.SetGlobalLevel(level)
	}

	// zerolog.DisableSampling(false)
	// zerolog.TimestampFieldName = name
	// zerolog.LevelFieldName = name
	// zerolog.MessageFieldName = name
	// zerolog.ErrorFieldName = name
	// zerolog.TimeFieldFormat = format
	// zerolog.DurationFieldUnit = unit
	// zerolog.DurationFieldInteger = integer
	// zerolog.ErrorHandler = handler
	// zerolog.FloatingPointPrecision = precision

	log.Logger = logger.Timestamp().Logger()

	return log.Logger
}
