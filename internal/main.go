package main

import (
	"context"

	"github.com/samber/lo"

	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/server"
)

func main() {
	options := clog.Chain().WithCaller(false).Build()
	server.Run("yarumo-app", "1.0.0", func(ctx context.Context, app server.Application) error {

		_ = lo.Empty[int]()

		return nil
	}, options)
}
