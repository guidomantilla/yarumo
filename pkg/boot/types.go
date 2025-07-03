package boot

import (
	"context"

	"github.com/guidomantilla/yarumo/pkg/servers"
)

var (
	_ BeanFn = Logger
	_ BeanFn = Config
	_ BeanFn = Validator
	_ RunFn  = Run[any]
)

type BeanFn func(concontainer *Container)

type WireFn func(ctx context.Context, application servers.Application) error

type RunFn func(ctx context.Context, name string, version string, wireFn WireFn, opts ...Option)
