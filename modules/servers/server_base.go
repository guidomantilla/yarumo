package servers

import (
	"context"
	"fmt"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/common/assert"
)

type baseServer struct {
	ctx          context.Context
	name         string
	closeChannel chan struct{}
}

func BuildBaseServer() (string, Server) {
	return "base-server", NewBaseServer()
}

func NewBaseServer() lifecycle.Server {
	return &baseServer{
		name:         "base-server",
		closeChannel: make(chan struct{}),
	}
}

func (server *baseServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s - error starting up: context is nil", server.name))

	log.Info().Str("stage", "startup").Str("component", server.name).Msg("starting up")

	server.ctx = ctx
	<-server.closeChannel
	return nil
}

func (server *baseServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", server.name))

	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopping")
	defer log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopped")

	close(server.closeChannel)
	return nil
}
