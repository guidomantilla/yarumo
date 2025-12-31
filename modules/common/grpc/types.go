package grpc

import (
	"net"

	"google.golang.org/grpc"
)

type Server interface {
	grpc.ServiceRegistrar
	Address() string
	Stop()
	GracefulStop()
	Serve(net.Listener) error
}
