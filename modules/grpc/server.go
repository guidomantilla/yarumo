package grpc

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/lifecycle"
)

// server implements Server. It wraps a *grpc.Server and exposes the configured
// listen address. Start is blocking (server-style lifecycle): it opens the
// listener and calls Serve, returning only when the server has been shut
// down or the listener fails. Done closes when Start returns.
type server struct {
	*grpc.Server

	name     string
	network  string
	addr     string
	listener net.Listener
	mutex    sync.Mutex

	done chan struct{}
	once sync.Once
}

// NewServer creates a new gRPC Server with the given name, network (e.g.,
// "tcp"), host and port, applying the provided options.
func NewServer(name string, network string, host string, port string, options ...Option) Server {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotEmpty(network, "network is empty")
	cassert.NotEmpty(host, "host is empty")
	cassert.NotEmpty(port, "port is empty")

	opts := NewOptions(options...)

	internal := grpc.NewServer(opts.serverOptions...)

	for _, reg := range opts.services {
		internal.RegisterService(reg.descriptor, reg.service)
	}

	return &server{
		Server:  internal,
		name:    name,
		network: network,
		addr:    net.JoinHostPort(host, port),
		done:    make(chan struct{}),
	}
}

// RegisterService registers a service and its implementation to the gRPC server.
func (s *server) RegisterService(desc *grpc.ServiceDesc, impl any) {
	cassert.NotNil(s, "server is nil")

	s.Server.RegisterService(desc, impl)
}

// Name returns the server's identity used in logs.
func (s *server) Name() string {
	cassert.NotNil(s, "server is nil")

	return s.name
}

// Start opens the configured listener and serves gRPC requests on it. It
// blocks until the server is gracefully or forcibly stopped, or until the
// listener fails. Done is closed when Start returns, regardless of the exit
// path (success, listen error, or serve error).
func (s *server) Start(ctx context.Context) error {
	cassert.NotNil(s, "server is nil")

	defer s.once.Do(func() { close(s.done) })

	lc := net.ListenConfig{}

	listener, err := lc.Listen(ctx, s.network, s.addr)
	if err != nil {
		return lifecycle.ErrStart(err)
	}

	s.mutex.Lock()
	s.listener = listener
	s.mutex.Unlock()

	err = s.Server.Serve(listener)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return lifecycle.ErrStart(err)
	}

	return nil
}

// Stop gracefully stops the gRPC server bounded by ctx's deadline. If ctx
// expires before GracefulStop drains, the server is forcibly stopped via
// Stop and any open listener is closed. Done is closed by the deferred
// once.Do here as a safety net for callers that invoke Stop without Start
// (the normal close path is Start's defer).
func (s *server) Stop(ctx context.Context) error {
	cassert.NotNil(s, "server is nil")

	defer s.once.Do(func() { close(s.done) })

	graceful := make(chan struct{})

	go func() {
		s.Server.GracefulStop()
		close(graceful)
	}()

	select {
	case <-graceful:
		return nil
	case <-ctx.Done():
		s.mutex.Lock()
		if s.listener != nil {
			_ = s.listener.Close()
		}
		s.mutex.Unlock()

		s.Server.Stop()

		<-graceful

		return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, ctx.Err())
	}
}

// Done returns the channel that is closed when Start has returned (server
// shut down) or when Stop has been invoked, whichever comes first.
func (s *server) Done() <-chan struct{} {
	cassert.NotNil(s, "server is nil")

	return s.done
}
