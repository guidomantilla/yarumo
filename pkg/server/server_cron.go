package server

import (
	"context"
	"fmt"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type cronServer struct {
	ctx          context.Context
	name         string
	internal     CronServer
	closeChannel chan struct{}
}

func BuildCronServer(cron CronServer) (string, Server) {
	return "cron-server", NewCronServer(cron)
}

func NewCronServer(cron CronServer) lifecycle.Server {
	assert.NotNil(cron, fmt.Sprintf("%s - error starting up: cron is nil", "cron-server"))

	return &cronServer{
		name:         "cron-server",
		internal:     cron,
		closeChannel: make(chan struct{}),
	}
}

func (server *cronServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s - error starting up: context is nil", server.name))

	server.ctx = ctx
	log.Info().Str("stage", "startup").Str("component", server.name).Msg("starting up")
	server.internal.Start()
	<-server.closeChannel
	return nil
}

func (server *cronServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", server.name))

	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopping")
	close(server.closeChannel)
	server.internal.Stop()
	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopped")
	return nil
}
