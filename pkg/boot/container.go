package boot

import (
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/passwords"
	"github.com/guidomantilla/yarumo/pkg/tokens"
)

type Container struct {
	opts              []Option
	AppName           string
	AppVersion        string
	Config            any
	Logger            zerolog.Logger
	Validator         *validator.Validate
	PasswordEncoder   passwords.Encoder
	PasswordGenerator passwords.Generator
	TokenGenerator    tokens.Generator
}

func Logger(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "logger").Msg("logger function not implemented. using default logger")
	debugMode := utils.Ternary(viper.IsSet("DEBUG_MODE"),
		viper.GetBool("DEBUG_MODE"), false)
	container.Logger = clog.Configure(container.AppName, container.AppVersion, clog.WithCaller(debugMode), clog.WithGlobalLevel(utils.Ternary(debugMode, zerolog.DebugLevel, zerolog.InfoLevel)))
}

func Config(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "configuration").Msg("config function not implemented. using default configuration")
}

func Validator(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "validation").Msg("validator function not implemented. using default validator")
	container.Validator = validator.New()
}

func PasswordEncoder(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "password encoder").Msg("password encoder function not implemented. using bcrypt password encoder")
	container.PasswordEncoder = passwords.NewBcryptEncoder()
}

func PasswordGenerator(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "password generator").Msg("password generator function not implemented. using default password generator")
	container.PasswordGenerator = passwords.NewGenerator()
}

func TokenGenerator(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "token generator").Msg("token generator function not implemented. using jwt token generator")

	issuer := tokens.WithJwtIssuer(container.AppName)

	signingKey := tokens.WithJwtSigningKey(
		utils.Ternary(viper.IsSet("TOKEN_SIGNATURE_KEY"),
			viper.GetString("TOKEN_SIGNATURE_KEY"), "SecretYouShouldHide"),
	)

	verifyingKey := tokens.WithJwtVerifyingKey(
		utils.Ternary(viper.IsSet("TOKEN_VERIFICATION_KEY"),
			viper.GetString("TOKEN_VERIFICATION_KEY"), "SecretYouShouldHide"),
	)

	timeout := tokens.WithJwtTimeout(
		utils.Ternary(viper.IsSet("TOKEN_TIMEOUT"),
			viper.GetDuration("TOKEN_TIMEOUT"), 24*time.Hour),
	)

	container.TokenGenerator = tokens.NewJwtGenerator(issuer, signingKey, verifyingKey, timeout)
}
