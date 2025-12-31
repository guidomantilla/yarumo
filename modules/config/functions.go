package config

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"
)

func Default(ctx context.Context, name string, version string, env string) context.Context {

	// Loading Environment Variables
	viper.AutomaticEnv()

	// Enable Asserts
	{
		enabled := func() bool {
			v := strings.ToLower(viper.GetString("ASSERTS_ENABLED"))
			return v == "1" || v == "true" || v == "yes"
		}()
		assert.Enable(enabled)
	}

	// Logger configuration
	{

		logger := zerolog.New(os.Stderr).With()
		if utils.NotEmpty(name) {
			logger = logger.Str("name", name)
		}

		if utils.NotEmpty(version) {
			logger = logger.Str("version", version)
		}

		if utils.NotEmpty(env) {
			logger = logger.Str("env", env)
		}

		debugMode := utils.Coalesce(viper.GetBool("DEBUG"), false)
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
		ctx = log.Logger.WithContext(ctx)
	}

	return ctx
}
