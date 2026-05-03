package managed

import (
	"context"
	"net"
	"sync"

	cgrpc "github.com/guidomantilla/yarumo/common/grpc"
)

type grpcAdapter struct {
	g        cgrpc.Server
	network  string
	listener net.Listener
	mutex    sync.Mutex
}

// NewGrpcServer creates a new managed gRPC server wrapping the given server and network.
func NewGrpcServer(g cgrpc.Server, network string) GrpcServer {
	return &grpcAdapter{
		g:       g,
		network: network,
	}
}

func (g *grpcAdapter) ListenAndServe(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, g.network, g.g.Address())
	if err != nil {
		return ErrListen(err)
	}

	g.mutex.Lock()
	g.listener = listener
	g.mutex.Unlock()

	err = g.g.Serve(listener)
	if err != nil {
		return ErrServe(err)
	}

	return nil
}

func (g *grpcAdapter) ListenAndServeTLS(_ context.Context, _ string, _ string) error {
	return ErrServe(ErrNotImplemented)
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
		return ErrShutdown(ErrShutdownTimeout, ctx.Err())
	}
}
