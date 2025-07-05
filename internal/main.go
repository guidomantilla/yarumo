package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/comm"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
	"github.com/guidomantilla/yarumo/pkg/servers"
)

type Config struct {
	DebugMode    bool   `mapstructure:"DEBUG_MODE"`
	Host         string `mapstructure:"HOST"`
	HttpPort     string `mapstructure:"HTTP_PORT"`
	GrpcPort     string `mapstructure:"GRPC_PORT"`
	CipherKey    string `mapstructure:"CIPHER_KEY"`
	TokenKey     string `mapstructure:"TOKEN_KEY"`
	TokenTimeout string `mapstructure:"TOKEN_TIMEOUT"`
}

func main() {
	withConfig := boot.WithConfig(func(container *boot.Container) {
		config := container.Config.(Config)

		debugMode := utils.Ternary(viper.IsSet("DEBUG_MODE"),
			viper.GetBool("DEBUG_MODE"), false)
		config.DebugMode = debugMode

		container.Config = config
	})

	withTokenGenerator := boot.WithTokenGenerator(func(container *boot.Container) {
		if !viper.IsSet("TOKEN_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "token generator").Msg("TOKEN_KEY is not set in the configuration")
		}

		config := container.Config.(Config)

		issuer := tokens.WithJwtIssuer(container.AppName)
		signingKey := tokens.WithJwtSigningKey([]byte(viper.GetString("TOKEN_KEY")))
		verifyingKey := tokens.WithJwtVerifyingKey([]byte(viper.GetString("TOKEN_KEY")))

		timeout := tokens.WithJwtTimeout(
			utils.Ternary(viper.IsSet("TOKEN_TIMEOUT"),
				viper.GetDuration("TOKEN_TIMEOUT"), 15*time.Minute),
		)

		config.TokenKey = viper.GetString("TOKEN_KEY")
		config.TokenTimeout = viper.GetString("TOKEN_TIMEOUT")

		container.TokenGenerator = tokens.NewJwtGenerator(issuer, signingKey, verifyingKey, timeout)
		container.Config = config
	})

	withCipher := boot.WithCipher(func(container *boot.Container) {
		if !viper.IsSet("CIPHER_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "cipher").Msg("CIPHER_KEY is not set in the configuration")
		}

		config := container.Config.(Config)

		key := cryptos.WithAesCipherKey(viper.GetString("CIPHER_KEY"))

		config.CipherKey = viper.GetString("CIPHER_KEY")

		container.Cipher = cryptos.NewAesCipher(key)
		container.Config = config
	})

	/**/
	name, version := "yarumo-app", "1.0.0"
	ctx := context.Background()
	boot.Run[Config](ctx, name, version, func(ctx context.Context, app servers.Application) error {
		wctx, err := boot.Context[Config]()
		if err != nil {
			return fmt.Errorf("error getting context: %w", err)
		}

		fmt.Println("Configuration:", fmt.Sprintf("%+v", wctx.Config))

		principal := tokens.Principal{
			"username": "test-user",
			"email":    "guido.mantilla@yahoo.com",
			"roles":    []string{"admin", "user"},
		}
		token, err := wctx.TokenGenerator.Generate("test-subject", principal)
		if err != nil {
			return err
		}
		fmt.Println("Generated token:", *token)

		principal, err = wctx.TokenGenerator.Validate(*token)
		if err != nil {
			return err
		}
		fmt.Println("Validated principal:", principal)

		encrypt, err := wctx.Cipher.Encrypt([]byte("encrypted-data"))
		if err != nil {
			return err
		}

		fmt.Println("Encrypted data:", string(encrypt))

		decrypt, err := wctx.Cipher.Decrypt(encrypt)
		if err != nil {
			return err
		}

		fmt.Println("Decrypted data:", string(decrypt))

		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		resMap, err := comm.RESTCall[any](timeoutCtx, http.MethodGet, "https://8f28c446-6960-481c-9ff6-2d9562f1f4c0.mock.pstmn.io", nil, http.Header{}, comm.WithHTTPClient(wctx.HttpClient))
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}

		fmt.Println("Response map:", resMap)

		return nil
	}, withConfig, withTokenGenerator, withCipher)
}
