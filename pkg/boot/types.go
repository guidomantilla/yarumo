package boot

import (
	"context"

	"github.com/guidomantilla/yarumo/pkg/server"
)

var (
	_ RunFn = Run
)

type WireFn func(ctx context.Context, application server.Application) error

type RunFn func(ctx context.Context, name string, version string, wireFn WireFn)
