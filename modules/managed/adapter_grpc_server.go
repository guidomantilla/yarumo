package managed

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type grpcAdapter struct {
	g        *grpc.Server
	network  string
	address  string
	listener net.Listener
	mutex    sync.Mutex
}

func NewGrpcServer(g *grpc.Server, network string, address string) GrpcServer {
	return &grpcAdapter{
		g:       g,
		network: network,
		address: address,
	}
}

func (g *grpcAdapter) RegisterService(desc *grpc.ServiceDesc, impl any) {
	g.g.RegisterService(desc, impl)
}

func (g *grpcAdapter) ListenAndServe() error {
	listener, err := net.Listen(g.network, g.address)
	if err != nil {
		return err
	}

	g.mutex.Lock()
	g.listener = listener
	g.mutex.Unlock()

	err = g.g.Serve(listener)
	if err != nil {
		return err
	}

	return nil
}

func (g *grpcAdapter) ListenAndServeTLS(_ string, _ string) error {
	return errors.New("not implemented")
}

func (g *grpcAdapter) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		g.g.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		g.mutex.Lock()
		if g.listener != nil {
			_ = g.listener.Close()
		}
		g.mutex.Unlock()

		g.g.Stop()
		return ctx.Err()
	}
}
