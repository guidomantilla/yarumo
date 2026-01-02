package managed

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	commongrpc "github.com/guidomantilla/yarumo/common/grpc"
)

type grpcAdapter struct {
	g        commongrpc.Server
	network  string
	listener net.Listener
	mutex    sync.Mutex
}

func NewGrpcServer(g commongrpc.Server, network string) GrpcServer {
	return &grpcAdapter{
		g:       g,
		network: network,
	}
}

func (g *grpcAdapter) ListenAndServe(_ context.Context) error {
	listener, err := net.Listen(g.network, g.g.Address())
	if err != nil {
		return fmt.Errorf("failed to listen on %s %s: %w", g.network, g.g.Address(), err)
	}

	g.mutex.Lock()
	g.listener = listener
	g.mutex.Unlock()

	err = g.g.Serve(listener)
	if err != nil {
		return fmt.Errorf("failed to serve on %s %s: %w", g.network, g.g.Address(), err)
	}

	return nil
}

func (g *grpcAdapter) ListenAndServeTLS(_ context.Context, _ string, _ string) error {
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
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}
