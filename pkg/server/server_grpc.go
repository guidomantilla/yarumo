package server

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
	address  string
	internal GrpcServer
}

func NewGrpcServer(address string, server GrpcServer) lifecycle.Server {
	assert.NotEmpty(address, "starting up - error setting up grpc server: address is empty")
	assert.NotNil(server, "starting up - error setting up grpc server: server is nil")

	return &grpcServer{
		address:  address,
		internal: server,
	}
}

func (server *grpcServer) Run(ctx context.Context) error {
	assert.NotNil(ctx, "grpc server - error starting up: context is nil")

	server.ctx = ctx
	log.Info().Msg(fmt.Sprintf("starting up - starting grpc server: %s", server.address))

	var err error
	var listener net.Listener
	if listener, err = net.Listen("tcp", server.address); err != nil {
		log.Info().Msg(fmt.Sprintf("starting up - starting grpc server error: %s", err.Error()))
		return err
	}

	if err = server.internal.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Info().Msg(fmt.Sprintf("starting up - starting grpc server error: %s", err.Error()))
		return err
	}
	return nil
}

func (server *grpcServer) Stop(ctx context.Context) error {
	assert.NotNil(ctx, "grpc server - error shutting down: context is nil")

	log.Info().Msg("shutting down - stopping grpc server")
	server.internal.GracefulStop()
	log.Info().Msg("shutting down - grpc server stopped")
	return nil
}
