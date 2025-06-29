package boot

import (
	"context"
	"github.com/guidomantilla/yarumo/pkg/server"
	"github.com/rs/zerolog/log"
	"syscall"

	"github.com/qmdx00/lifecycle"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

func Run(ctx context.Context, name string, version string, wireFn WireFn) {
	assert.NotNil(ctx, "server - error running: ctx is nil")
	assert.NotEmpty(name, "server - error running: name is empty")
	assert.NotEmpty(version, "server - error running: version is empty")
	assert.NotNil(wireFn, "server - error running: wireFn is nil")

	app := lifecycle.NewApp(
		lifecycle.WithName(name), lifecycle.WithVersion(version),
		lifecycle.WithSignal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
	)

	app.Attach(server.BuildBaseServer())

	err := wireFn(ctx, app)
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "main").Err(err).Msg("error wiring the application")
	}

	err = app.Run()
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "main").Err(err).Msg("error running the application")
	}
}
