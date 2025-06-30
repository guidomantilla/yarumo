package main

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/boot"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/server"
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

	withConfig := boot.WithConfig(func(wctx *boot.WireContext) any {
		viper.AutomaticEnv()

		config := Config{
			DebugMode: utils.Ternary(viper.IsSet("DEBUG_MODE"),
				viper.GetBool("DEBUG_MODE"), false),
		}
		clogOpts := clog.Chain().
			WithCaller(config.DebugMode).
			WithGlobalLevel(utils.Ternary(config.DebugMode, zerolog.DebugLevel, wctx.LogLevel)).
			Build()
		clog.Configure(wctx.AppName, wctx.AppVersion, clogOpts)

		return config
	})

	name, version := "yarumo-app", "1.0.0"
	ctx := context.Background()
	boot.Run[Config](ctx, name, version, func(ctx context.Context, config Config, app server.Application) error {

		if config.DebugMode {
			fmt.Println("Debug mode is enabled")
		} else {
			fmt.Println("Debug mode is disabled")
		}

		return nil
	}, withConfig)
}
