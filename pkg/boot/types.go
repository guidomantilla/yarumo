package boot

import (
	"context"

	"github.com/guidomantilla/yarumo/pkg/server"
)

var (
	_ RunFn[any] = Run
)

type BeanFn func(wctx *WireContext)

type WireFn[T any] func(ctx context.Context, wctx *WireContext, application server.Application) error

type RunFn[T any] func(ctx context.Context, name string, version string, wireFn WireFn[T], opts ...Option)
