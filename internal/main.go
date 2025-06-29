package main

import (
	"context"
	"fmt"
	"github.com/guidomantilla/yarumo/pkg/boot"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"

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
	viper.AutomaticEnv()

	name, version := "yarumo-app", "1.0.0"
	ctx := context.Background()

	conf := Config{}
	if viper.IsSet("DEBUG_MODE") {
		conf.DebugMode = viper.GetBool("DEBUG_MODE")
	}

	clogOpts := clog.Chain().
		WithCaller(conf.DebugMode).
		WithGlobalLevel(utils.Ternary(conf.DebugMode, zerolog.DebugLevel, zerolog.InfoLevel)).
		Build()
	clog.Configure(name, version, clogOpts)

	boot.Run(ctx, name, version, func(ctx context.Context, app server.Application) error {

		if conf.DebugMode {
			fmt.Println("Debug mode is enabled")
		} else {
			fmt.Println("Debug mode is disabled")
		}

		return nil
	})
}
