package boot

import (
	"context"
	"syscall"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/server"
)

func Run[T any](ctx context.Context, name string, version string, wireFn WireFn[T], opts ...Option) {
	assert.NotNil(ctx, "server - error running: ctx is nil")
	assert.NotEmpty(name, "server - error running: name is empty")
	assert.NotEmpty(version, "server - error running: version is empty")
	assert.NotNil(wireFn, "server - error running: wireFn is nil")

	wctx := NewWireContext(name, version, opts...)
	defer wctx.Stop(ctx)

	app := lifecycle.NewApp(
		lifecycle.WithName(name), lifecycle.WithVersion(version),
		lifecycle.WithSignal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
	)
	app.Attach(server.BuildBaseServer())

	config := pointer.Zero[T]()
	if utils.NotNil(wctx.Config) {
		config = (wctx.Config).(T)
	}
	err := wireFn(ctx, config, app)
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "application").Err(err).Msg("error wiring the application")
	}

	err = app.Run()
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "application").Err(err).Msg("error running the application")
	}
}
