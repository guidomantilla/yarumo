package boot

import (
	"context"

	"github.com/guidomantilla/yarumo/pkg/server"
)

var (
	_ RunFn[any] = Run
)

type ConfigFn func(wctx *WireContext) any

type WireFn[T any] func(ctx context.Context, config T, application server.Application) error

type RunFn[T any] func(ctx context.Context, name string, version string, wireFn WireFn[T], opts ...Option)
