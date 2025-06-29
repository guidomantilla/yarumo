package main

import (
	"context"
	"github.com/guidomantilla/yarumo/pkg/server"
	"github.com/samber/lo"
)

func main() {
	server.Run("yarumo-app", "1.0.0", func(ctx context.Context, app server.Application) error {

		_ = lo.Empty[int]()

		return nil
	})
}
