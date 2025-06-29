package main

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/server"
)

func main() {
	server.Run("yarumo-app", "1.0.0", func(ctx context.Context, app server.Application) error {
		viper.AutomaticEnv()
		fmt.Println(viper.Get("LOCALSTACK_AUTH_TOKEN"))

		_ = lo.Empty[int]()

		return nil
	})
}
