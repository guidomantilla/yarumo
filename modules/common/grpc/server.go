package grpc

import (
	"net"

	"google.golang.org/grpc"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

type server struct {
	inner *grpc.Server
	addr  string
}

// NewServer creates a new gRPC Server using the provided options.
func NewServer(host string, port string, options ...Option) Server {
	cassert.NotEmpty(host, "host is empty")
	cassert.NotEmpty(port, "port is empty")

	opts := NewOptions(options...)

	internal := grpc.NewServer(opts.serverOptions...)

	for _, reg := range opts.services {
		internal.RegisterService(reg.descriptor, reg.service)
	}

	return &server{
		inner: internal,
		addr:  net.JoinHostPort(host, port),
	}
}

// Address returns the network address the server is configured to listen on.
func (s *server) Address() string {
	cassert.NotNil(s, "server is nil")

	return s.addr
}

// RegisterService registers a service and its implementation to the gRPC server.
func (s *server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	cassert.NotNil(s, "server is nil")

	s.inner.RegisterService(desc, impl)
}

// Stop stops the gRPC server immediately.
func (s *server) Stop() {
	cassert.NotNil(s, "server is nil")

	s.inner.Stop()
}

// GracefulStop stops the gRPC server gracefully.
func (s *server) GracefulStop() {
	cassert.NotNil(s, "server is nil")

	s.inner.GracefulStop()
}

// Serve accepts incoming connections on the listener.
func (s *server) Serve(lis net.Listener) error {
	cassert.NotNil(s, "server is nil")

	return s.inner.Serve(lis)
}
