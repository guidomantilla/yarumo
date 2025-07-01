package boot

import (
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type Container struct {
	opts       []Option
	AppName    string
	AppVersion string
	Config     any
	Logger     zerolog.Logger
	Validator  *validator.Validate
}

func Logger(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "logger").Msg("logger function not implemented. using default logger")
	debugMode := utils.Ternary(viper.IsSet("DEBUG_MODE"),
		viper.GetBool("DEBUG_MODE"), false)
	clogOpts := clog.Chain().
		WithCaller(debugMode).
		WithGlobalLevel(utils.Ternary(debugMode, zerolog.DebugLevel, zerolog.InfoLevel)).
		Build()
	container.Logger = clog.Configure(container.AppName, container.AppVersion, clogOpts)
}

func Config(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "configuration").Msg("config function not implemented. using default configuration")
}

func Validator(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "validation").Msg("validator function not implemented. using default validator")
	container.Validator = validator.New()
}
