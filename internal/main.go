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

	withConfig := boot.WithConfig(func(wctx *boot.WireContext) {
		viper.AutomaticEnv()

		wctx.DebugMode = utils.Ternary(viper.IsSet("DEBUG_MODE"),
			viper.GetBool("DEBUG_MODE"), false)
		wctx.Config = Config{DebugMode: wctx.DebugMode}

		clogOpts := clog.Chain().
			WithCaller(wctx.DebugMode).
			WithGlobalLevel(utils.Ternary(wctx.DebugMode, zerolog.DebugLevel, wctx.LogLevel)).
			Build()
		wctx.Logger = clog.Configure(wctx.AppName, wctx.AppVersion, clogOpts)
	})
	/**/
	name, version := "yarumo-app", "1.0.0"
	ctx := context.Background()
	boot.Run[Config](ctx, name, version, func(ctx context.Context, wctx *boot.WireContext, app server.Application) error {

		if wctx.DebugMode {
			fmt.Println("Debug mode is enabled")
		} else {
			fmt.Println("Debug mode is disabled")
		}

		return nil
	}, withConfig)
}
