package server

import (
	"context"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type baseServer struct {
	ctx          context.Context
	closeChannel chan struct{}
}

func NewBaseServer() lifecycle.Server {
	return &baseServer{
		closeChannel: make(chan struct{}),
	}
}

func (server *baseServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, "base server - error starting up: context is nil")

	server.ctx = ctx
	log.Info().Msg("starting up - starting base server")
	<-server.closeChannel
	return nil
}

func (server *baseServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, "base server - error shutting down: context is nil")

	log.Info().Msg("shutting down - stopping base server")
	close(server.closeChannel)
	log.Info().Msg("shutting down - default base stopped")
	return nil
}
