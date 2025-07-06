package boot

import (
	"context"
	"syscall"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	"github.com/guidomantilla/yarumo/pkg/servers"
)

var (
	_ RunFn = Run[any]
)

type WireFn func(ctx context.Context, application servers.Application) error

type RunFn func(ctx context.Context, name string, version string, wireFn WireFn, opts ...WireContextOption)

//

func Run[C any](ctx context.Context, name string, version string, wireFn WireFn, opts ...WireContextOption) {
	assert.NotNil(ctx, "server - error running: ctx is nil")
	assert.NotEmpty(name, "server - error running: name is empty")
	assert.NotEmpty(version, "server - error running: version is empty")
	assert.NotNil(wireFn, "server - error running: wireFn is nil")

	wctx := NewWireContext[C](name, version, opts...)
	defer wctx.Stop(ctx)

	app := lifecycle.NewApp(
		lifecycle.WithName(name), lifecycle.WithVersion(version),
		lifecycle.WithSignal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
	)
	app.Attach(servers.BuildBaseServer())

	err := wireFn(ctx, app)
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "context").Err(err).Msg("error wiring the application")
	}

	err = app.Run()
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "context").Err(err).Msg("error running the application")
	}
}
