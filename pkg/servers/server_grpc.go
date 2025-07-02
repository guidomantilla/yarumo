package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/qmdx00/lifecycle"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type grpcServer struct {
	ctx      context.Context
	name     string
	address  string
	internal GrpcServer
}

func BuildGrpcServer(address string, server GrpcServer) (string, Server) {
	return "grpc-server", NewGrpcServer(address, server)
}

func NewGrpcServer(address string, server GrpcServer) lifecycle.Server {
	assert.NotEmpty(address, fmt.Sprintf("%s - error starting up: address is nil", "grpc-server"))
	assert.NotNil(server, fmt.Sprintf("%s - error starting up: server is nil", "grpc-server"))

	return &grpcServer{
		name:     "grpc-server",
		address:  address,
		internal: server,
	}
}

func (server *grpcServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s - error starting up: context is nil", server.name))

	log.Info().Str("stage", "startup").Str("component", server.name).Msg("starting up")

	server.ctx = ctx
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		log.Error().Str("stage", "startup").Str("component", server.name).Err(err).Msg("failed to listen")
		return err
	}

	err = server.internal.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Str("stage", "startup").Str("component", server.name).Err(err).Msg("failed to serve")
		return err
	}

	return nil
}

func (server *grpcServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", server.name))

	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopping")
	log.Info().Str("stage", "shut down").Str("component", server.name).Msg("stopped")

	server.internal.GracefulStop()
	return nil
}
