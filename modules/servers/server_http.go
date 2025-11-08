package servers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"
)

type httpServer struct {
	ctx      context.Context
	name     string
	internal *http.Server
}

func BuildHttpServer(server *http.Server) (string, Server) {
	return "http-server", NewHttpServer(server)
}

func NewHttpServer(server *http.Server) lifecycle.Server {
	assert.NotNil(server, fmt.Sprintf("%s - error starting up: server is nil", "http-server"))
	return &httpServer{
		name:     "http-server",
		internal: server,
	}
}

func (server *httpServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s - error starting up: context is nil", server.name))

	log.Info().Str("stage", "startup").Str("component", server.name).Msg("starting up")

	server.ctx = ctx
	err := server.internal.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Str("stage", "startup").Str("component", server.name).Err(err).Msg("failed to listen or serve")
		return ErrServerFailedToStart(server.name, err)
	}

	return nil
}

func (server *httpServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", server.name))

	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopping")
	defer log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopped")

	err := server.internal.Shutdown(ctx)
	if err != nil {
		log.Error().Str("stage", "shut down").Str("component", server.name).Err(err).Msg("failed to stop")
		return ErrServerFailedToStop(server.name, err)
	}
	return nil
}
