package server

import (
	"context"
	"net/http"
	"syscall"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

func BuildBaseServer() (string, Server) {
	return "base-server", NewBaseServer()
}

func BuildCronServer(cron CronServer) (string, Server) {
	return "cron-server", NewCronServer(cron)
}

func BuildHttpServer(server *http.Server) (string, Server) {
	return "http-server", NewHttpServer(server)
}

func BuildGrpcServer(address string, server GrpcServer) (string, Server) {
	return "grpc-server", NewGrpcServer(address, server)
}

func Run(ctx context.Context, name string, version string, wireFn WireFn) {
	assert.NotNil(ctx, "server - error running: ctx is nil")
	assert.NotEmpty(name, "server - error running: name is empty")
	assert.NotEmpty(version, "server - error running: version is empty")
	assert.NotNil(wireFn, "server - error running: wireFn is nil")

	app := lifecycle.NewApp(
		lifecycle.WithName(name), lifecycle.WithVersion(version),
		lifecycle.WithSignal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
	)

	app.Attach(BuildBaseServer())

	err := wireFn(ctx, app)
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "main").Err(err).Msg("error wiring the application")
	}

	err = app.Run()
	if err != nil {
		log.Fatal().Str("stage", "startup").Str("component", "main").Err(err).Msg("error running the application")
	}
}
