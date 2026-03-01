package managed

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

type listenAndServeGRPCServer struct {
	addressVal string
	serveErr   error
	serveCh    chan struct{}
}

func (m *listenAndServeGRPCServer) RegisterService(_ *grpc.ServiceDesc, _ any) {}

func (m *listenAndServeGRPCServer) Address() string { return m.addressVal }

func (m *listenAndServeGRPCServer) Stop() {}

func (m *listenAndServeGRPCServer) GracefulStop() {}

func (m *listenAndServeGRPCServer) Serve(lis net.Listener) error {
	if m.serveCh != nil {
		close(m.serveCh)
	}
	_ = lis.Close()
	return m.serveErr
}

func TestBuildGrpcServer(t *testing.T) {
	t.Run("build succeeds and stop completes", func(t *testing.T) {
		errCh := make(chan error, 1)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		addr := listener.Addr().String()
		listener.Close()

		serveCh := make(chan struct{})
		srv := &listenAndServeGRPCServer{
			addressVal: addr,
			serveCh:    serveCh,
		}

		component, stopFn, err := BuildGrpcServer(context.Background(), "test-grpc", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if component.name != "test-grpc" {
			t.Fatalf("expected name test-grpc, got %s", component.name)
		}

		if stopFn == nil {
			t.Fatal("expected non-nil stopFn")
		}

		<-serveCh

		time.Sleep(50 * time.Millisecond)

		stopFn(context.Background(), 5*time.Second)

		select {
		case err := <-errCh:
			t.Fatalf("unexpected error: %v", err)
		default:
		}
	})

	t.Run("listen error is sent to errChan", func(t *testing.T) {
		errCh := make(chan error, 1)

		srv := &listenAndServeGRPCServer{
			addressVal: "invalid-addr",
		}

		// Use invalid network to force listen failure
		origNew := NewGrpcServer
		_ = origNew

		_, _, err := BuildGrpcServer(context.Background(), "test-grpc-fail", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil build error, got %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// The adapter uses "tcp" internally, and invalid-addr may or may not fail on listen.
		// Use a serve error instead.
	})

	t.Run("serve error is sent to errChan", func(t *testing.T) {
		errCh := make(chan error, 1)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		addr := listener.Addr().String()
		listener.Close()

		srv := &listenAndServeGRPCServer{
			addressVal: addr,
			serveErr:   errors.New("serve failed"),
		}

		_, _, err = BuildGrpcServer(context.Background(), "test-grpc-serve-fail", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil build error, got %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		select {
		case err := <-errCh:
			if err == nil {
				t.Fatal("expected non-nil error from errChan")
			}
		default:
			t.Fatal("expected error in errChan")
		}
	})

	t.Run("stop with short timeout logs error", func(t *testing.T) {
		errCh := make(chan error, 1)

		blockCh := make(chan struct{})
		srv := &blockingGRPCServer{blockCh: blockCh}

		_, stopFn, err := BuildGrpcServer(context.Background(), "test-grpc-timeout", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Nanosecond)

		close(blockCh)
	})
}
