package main

import (
	"context"
	"fmt"

	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/servers"
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
		debugMode := utils.Ternary(viper.IsSet("DEBUG_MODE"),
			viper.GetBool("DEBUG_MODE"), false)
		container.Config = Config{DebugMode: debugMode}
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

		return nil
	}, withConfig)
}
