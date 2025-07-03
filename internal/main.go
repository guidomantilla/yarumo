package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/servers"
	"github.com/guidomantilla/yarumo/pkg/tokens"
)

type Config struct {
	DebugMode            bool   `mapstructure:"DEBUG_MODE"`
	Host                 string `mapstructure:"HOST"`
	HttpPort             string `mapstructure:"HTTP_PORT"`
	GrpcPort             string `mapstructure:"GRPC_PORT"`
	TokenSignatureKey    string `mapstructure:"TOKEN_SIGNATURE_KEY"`
	TokenVerificationKey string `mapstructure:"TOKEN_VERIFICATION_KEY"`
	TokenTimeout         string `mapstructure:"TOKEN_TIMEOUT"`
	DatasourceDriver     string `mapstructure:"DATASOURCE_DRIVER"`
	DatasourceUsername   string `mapstructure:"DATASOURCE_USERNAME"`
	DatasourcePassword   string `mapstructure:"DATASOURCE_PASSWORD"`
	DatasourceServer     string `mapstructure:"DATASOURCE_SERVER"`
	DatasourceService    string `mapstructure:"DATASOURCE_SERVICE"`
	DatasourceUrl        string `mapstructure:"DATASOURCE_URL"`
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
		if !viper.IsSet("TOKEN_SIGNATURE_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "token generator").Msg("TOKEN_SIGNATURE_KEY is not set in the configuration")
		}

		if !viper.IsSet("TOKEN_VERIFICATION_KEY") {
			log.Fatal().Str("stage", "startup").Str("component", "token generator").Msg("TOKEN_VERIFICATION_KEY is not set in the configuration")
		}

		config := container.Config.(Config)
		issuer := tokens.WithJwtIssuer(container.AppName)
		signingKey := tokens.WithJwtSigningKey([]byte(viper.GetString("TOKEN_SIGNATURE_KEY")))
		verifyingKey := tokens.WithJwtVerifyingKey([]byte(viper.GetString("TOKEN_VERIFICATION_KEY")))

		timeout := tokens.WithJwtTimeout(
			utils.Ternary(viper.IsSet("TOKEN_TIMEOUT"),
				viper.GetDuration("TOKEN_TIMEOUT"), 15*time.Minute),
		)

		container.TokenGenerator = tokens.NewJwtGenerator(issuer, signingKey, verifyingKey, timeout)
		config.TokenVerificationKey = viper.GetString("TOKEN_VERIFICATION_KEY")
		config.TokenSignatureKey = viper.GetString("TOKEN_SIGNATURE_KEY")
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

		if wctx.Config.DebugMode {
			fmt.Println("Debug mode is enabled")
		} else {
			fmt.Println("Debug mode is disabled")
		}

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
		return nil
	}, withConfig, withTokenGenerator)
}
