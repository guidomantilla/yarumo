package boot

import (
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/common/comm"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/uids"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
)

var (
	_ BeanFn = Hasher
	_ BeanFn = UIDGen
	_ BeanFn = Logger
	_ BeanFn = Config
	_ BeanFn = Validator
	_ BeanFn = PasswordEncoder
	_ BeanFn = PasswordGenerator
	_ BeanFn = TokenGenerator
	_ BeanFn = Cipher
	_ BeanFn = HttpClient
)

type BeanFn func(container *Container)

//

func Hasher(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "hasher").Msg("hasher function not implemented. using BLAKE2b-512 hasher")
	container.Hasher = hashes.BLAKE2b_512
}

func UIDGen(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "uid generator").Msg("uid generator function not implemented. using UUIDv7 uid generator")
	container.UIDGen = uids.UUIDv7
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

	timeout := tokens.WithJwtTimeout(
		utils.Ternary(viper.IsSet("TOKEN_TIMEOUT"),
			viper.GetDuration("TOKEN_TIMEOUT"), 24*time.Hour),
	)

	container.TokenGenerator = tokens.NewJwtGenerator(issuer, timeout)
}

func Cipher(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "cipher").Msg("cipher function not implemented. using default cipher")
	container.Cipher = cryptos.NewAesCipher()
}

func HttpClient(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "http client").Msg("http client function not implemented. using zero global timeout http client")

	timeout := comm.WithTimeout(utils.Ternary(viper.IsSet("HTTP_CLIENT_TIMEOUT"),
		viper.GetDuration("HTTP_CLIENT_TIMEOUT"), 0))

	maxRetries := comm.WithMaxRetries(utils.Ternary(viper.IsSet("HTTP_CLIENT_MAX_RETRIES"),
		uint(viper.GetInt("HTTP_CLIENT_MAX_RETRIES")), 3)) //nolint:gosec

	container.HttpClient = comm.NewHTTPClient(timeout, maxRetries)
}
