package managed

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
)

type mockGRPCServer struct {
	addressVal      string
	serveErr        error
	gracefulStopped bool
	stopped         bool
}

func (m *mockGRPCServer) RegisterService(_ *grpc.ServiceDesc, _ any) {}

func (m *mockGRPCServer) Address() string { return m.addressVal }

func (m *mockGRPCServer) Stop() { m.stopped = true }

func (m *mockGRPCServer) GracefulStop() { m.gracefulStopped = true }

func (m *mockGRPCServer) Serve(_ net.Listener) error { return m.serveErr }

func TestNewGrpcServer(t *testing.T) {
	t.Parallel()

	srv := &mockGRPCServer{}
	adapter := NewGrpcServer(srv, "tcp")
	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
}

func Test_grpcAdapter_ListenAndServe(t *testing.T) {
	t.Parallel()

	t.Run("listen fails on invalid network", func(t *testing.T) {
		t.Parallel()

		srv := &mockGRPCServer{addressVal: "invalid-addr"}
		adapter := NewGrpcServer(srv, "invalid-network")

		err := adapter.ListenAndServe(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("listen and serve succeeds", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		addr := listener.Addr().String()
		listener.Close()

		srv := &mockGRPCServer{addressVal: addr}
		adapter := NewGrpcServer(srv, "tcp")

		err = adapter.ListenAndServe(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("serve returns error", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		addr := listener.Addr().String()
		listener.Close()

		srv := &mockGRPCServer{
			addressVal: addr,
			serveErr:   net.ErrClosed,
		}
		adapter := NewGrpcServer(srv, "tcp")

		err = adapter.ListenAndServe(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_grpcAdapter_ListenAndServeTLS(t *testing.T) {
	t.Parallel()

	srv := &mockGRPCServer{}
	adapter := NewGrpcServer(srv, "tcp")

	err := adapter.ListenAndServeTLS(context.Background(), "cert.pem", "key.pem")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_grpcAdapter_Stop(t *testing.T) {
	t.Parallel()

	t.Run("graceful stop succeeds", func(t *testing.T) {
		t.Parallel()

		srv := &mockGRPCServer{}
		adapter := NewGrpcServer(srv, "tcp")

		err := adapter.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("stop with canceled context force stops", func(t *testing.T) {
		t.Parallel()

		blockCh := make(chan struct{})
		srv := &mockGRPCServer{}
		origGracefulStop := srv.GracefulStop

		_ = origGracefulStop

		blockingSrv := &blockingGRPCServer{blockCh: blockCh}
		adapter := NewGrpcServer(blockingSrv, "tcp")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := adapter.Stop(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		close(blockCh)
	})
}

type blockingGRPCServer struct {
	blockCh chan struct{}
}

func (m *blockingGRPCServer) RegisterService(_ *grpc.ServiceDesc, _ any) {}

func (m *blockingGRPCServer) Address() string { return "" }

func (m *blockingGRPCServer) Stop() {}

func (m *blockingGRPCServer) GracefulStop() { <-m.blockCh }

func (m *blockingGRPCServer) Serve(_ net.Listener) error { return nil }
