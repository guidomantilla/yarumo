package main

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	comm "github.com/guidomantilla/yarumo/pkg/comm"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
)

func GetOptions() []boot.WireContextOption {
	//return [][]boot.Option{}
	return []boot.WireContextOption{
		boot.WithConfig(Config()),
		boot.WithTokenGenerator(TokenGenerator()),
		boot.WithCipher(Cipher()),
		boot.With(RestClientToMockEndpoint()),
		boot.With(RestClientToFakeRestApiEndpoint()),
	}
}

func Config() boot.BeanFn {
	return func(container *boot.Container) {
		config := container.Config.(core.Config)

		config.DebugMode = utils.Coalesce(viper.GetBool("DEBUG_MODE"), false)
		//config.LogLevel = container.Logger.GetLevel().String()

		container.Config = config
	}
}

func TokenGenerator() boot.BeanFn {
	return func(container *boot.Container) {
		if !viper.IsSet("TOKEN_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "token generator").Msg("TOKEN_KEY is not set in the configuration")
		}

		config := container.Config.(core.Config)

		issuer := tokens.WithJwtIssuer(container.AppName)
		key := tokens.WithJwtKey([]byte(viper.GetString("TOKEN_KEY")))

		timeout := tokens.WithJwtTimeout(utils.Coalesce(viper.GetDuration("TOKEN_TIMEOUT"), 15*time.Minute))

		config.TokenKey = viper.GetString("TOKEN_KEY")
		config.TokenTimeout = viper.GetString("TOKEN_TIMEOUT")

		container.TokenGenerator = tokens.NewJwtGenerator(issuer, key, timeout)
		container.Config = config
	}
}

func Cipher() boot.BeanFn {
	return func(container *boot.Container) {
		if !viper.IsSet("CIPHER_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "cipher").Msg("CIPHER_KEY is not set in the configuration")
		}

		config := container.Config.(core.Config)

		key := cryptos.WithAesCipherKey(viper.GetString("CIPHER_KEY"))

		config.CipherKey = viper.GetString("CIPHER_KEY")

		container.Cipher = cryptos.NewAesCipher(key)
		container.Config = config
	}
}

func RestClientToMockEndpoint() boot.BeanFn {
	return func(container *boot.Container) {
		rest := comm.NewRESTClient("https://8f28c446-6960-481c-9ff6-2d9562f1f4c0.mock.pstmn.io", comm.WithHTTPClient(container.HttpClient))
		boot.Add(container, "RestClientToMockEndpoint", rest)

		log.Info().Str("stage", "startup").Str("component", "mock rest client").Msg("rest client to mock endpoint set up")
	}
}

func RestClientToFakeRestApiEndpoint() boot.BeanFn {
	return func(container *boot.Container) {
		rest := comm.NewRESTClient("https://fakerestapi.azurewebsites.net", comm.WithHTTPClient(container.HttpClient))
		boot.Add(container, "RestClientToFakeRestApiEndpoint", rest)

		log.Info().Str("stage", "startup").Str("component", "mock rest client").Msg("rest client to fake rest api endpoint set up")
	}
}
