package server

import (
	"context"
	"net/http"
	"syscall"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
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

func Run(name string, version string, fn func(ctx context.Context, application Application) error, opts ...clog.Option) {
	assert.NotEmpty(name, "server - error running: name is empty")
	assert.NotEmpty(version, "server - error running: version is empty")
	assert.NotNil(fn, "server - error running: function is nil")

	clog.Configure(name, version, opts...)

	//environment.Load()

	app := lifecycle.NewApp(
		lifecycle.WithName(name), lifecycle.WithVersion(version),
		lifecycle.WithSignal(syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL),
	)

	app.Attach(BuildBaseServer())

	err := fn(context.Background(), app)
	if err != nil {
		log.Info().Msg(err.Error())
	}

	if err := app.Run(); err != nil {
		log.Info().Msg(err.Error())
	}
}
