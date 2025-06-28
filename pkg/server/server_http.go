package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type httpServer struct {
	ctx      context.Context
	internal *http.Server
}

func NewHttpServer(server *http.Server) lifecycle.Server {
	assert.NotNil(server, "starting up - error setting up http server: server is nil")

	return &httpServer{
		internal: server,
	}
}

func (server *httpServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, "http server - error starting: context is nil")

	server.ctx = ctx
	log.Info().Msg(fmt.Sprintf("starting up - starting http server: %s", server.internal.Addr))

	if err := server.internal.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Info().Msg(fmt.Sprintf("starting up - starting http server error: %s", err.Error()))
		return err
	}
	return nil
}

func (server *httpServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, "http server - error shutting down: context is nil")

	log.Info().Msg("shutting down - stopping http server")
	if err := server.internal.Shutdown(ctx); err != nil {
		log.Info().Msg(fmt.Sprintf("shutting down - forced shutdown: %s", err.Error()))
		return err
	}
	log.Info().Msg("shutting down - http server stopped")
	return nil
}
