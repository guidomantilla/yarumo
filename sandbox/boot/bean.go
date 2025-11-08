package boot

import (
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/security/passwords"
)

var (
	_ BeanFn = Hasher
	_ BeanFn = UIDGen
	_ BeanFn = Config
	_ BeanFn = Validator
	_ BeanFn = PasswordEncoder
	_ BeanFn = PasswordGenerator
	_ BeanFn = PasswordManager
	_ BeanFn = TokenGenerator
	_ BeanFn = Cipher
	_ BeanFn = RateLimiterRegistry
	_ BeanFn = BreakerRegistry
	_ BeanFn = HttpClient
)

type BeanFn func(container *Container)

//

func Hasher(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "hasher").Msg("hasher function not implemented. using BLAKE2b-512 hasher")
}

func UIDGen(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "uid-generator").Msg("uid generator function not implemented. using UUIDv7 uid generator")
}

func Config(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "configuration").Msg("config function not implemented. using default configuration")
}

func Validator(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "validation").Msg("validator function not implemented. using default validator")
}

func PasswordEncoder(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "password-encoder").Msg("password encoder function not implemented. using bcrypt password encoder")
}

func PasswordGenerator(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "password-generator").Msg("password generator function not implemented. using default password generator")
}

func PasswordManager(container *Container) {
	log.Warn().Str("stage", "startup").Str("component", "password-manager").Msg("password manager function not implemented. using default password manager")
	container.PasswordManager = passwords.NewManager(container.PasswordEncoder, container.PasswordGenerator)
}

func TokenGenerator(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "token-generator").Msg("token generator function not implemented. using jwt token generator")
}

func Cipher(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "cipher").Msg("cipher function not implemented. using default cipher")
}

func RateLimiterRegistry(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "rate-limiter-registry").Msg("rate limiter registry function not implemented. using default rate limiter registry")
}

func BreakerRegistry(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "circuit-breaker-registry").Msg("circuit breaker registry function not implemented. using default circuit breaker registry")
}

func HttpClient(_ *Container) {
	log.Warn().Str("stage", "startup").Str("component", "http-client").Msg("http client function not implemented. using zero global timeout http client")
}
