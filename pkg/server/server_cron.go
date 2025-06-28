package server

import (
	"context"
	"github.com/guidomantilla/yarumo/pkg/common/assert"
	"github.com/rs/zerolog/log"

	"github.com/qmdx00/lifecycle"
)

type cronServer struct {
	ctx          context.Context
	internal     CronServer
	closeChannel chan struct{}
}

func NewCronServer(cron CronServer) lifecycle.Server {
	assert.NotNil(cron, "starting up - error setting up cron server: cron is nil")

	return &cronServer{
		internal:     cron,
		closeChannel: make(chan struct{}),
	}
}

func (server *cronServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, "cron server - error starting up: context is nil")

	server.ctx = ctx
	log.Info().Msg("starting up - starting cron server")
	server.internal.Start()
	<-server.closeChannel
	return nil
}

func (server *cronServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, "cron server - error shutting down: context is nil")

	log.Info().Msg("shutting down - stopping cron server")
	close(server.closeChannel)
	server.internal.Stop()
	log.Info().Msg("shutting down - cron server stopped")
	return nil
}
