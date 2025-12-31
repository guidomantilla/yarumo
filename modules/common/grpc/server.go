package grpc

import (
	"net"

	"google.golang.org/grpc"
)

type server struct {
	*grpc.Server
	Addr string
}

func NewServer(host string, port string, service any, descriptor *grpc.ServiceDesc, options ...grpc.ServerOption) Server {
	internal := grpc.NewServer(options...)
	internal.RegisterService(descriptor, service)
	return &server{
		Server: internal,
		Addr:   net.JoinHostPort(host, port),
	}
}

func (s *server) Address() string {
	return s.Addr
}
