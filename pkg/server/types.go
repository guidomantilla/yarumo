package server

import (
	"net"

	"github.com/qmdx00/lifecycle"
)

var (
	_ Server = (*cronServer)(nil)
	_ Server = (*grpcServer)(nil)
	_ Server = (*httpServer)(nil)
	_ Server = (*MockServer)(nil)
)

type Server interface {
	lifecycle.Server
}

type CronServer interface {
	Start()
	Stop()
}

type GrpcServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
}
