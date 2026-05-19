package grpc

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
)

type testService any

type testServiceImpl struct{}

func newTestDesc() *grpc.ServiceDesc {
	return &grpc.ServiceDesc{
		ServiceName: "test.Service",
		HandlerType: (*testService)(nil),
	}
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	t.Run("returns server with correct address", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("localhost", "50051", WithService(&testServiceImpl{}, newTestDesc()))
		if srv == nil {
			t.Fatal("expected non-nil server")
		}

		expected := net.JoinHostPort("localhost", "50051")
		if srv.Address() != expected {
			t.Fatalf("expected address %q, got %q", expected, srv.Address())
		}
	})

	t.Run("passes server options to grpc", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("0.0.0.0", "9090",
			WithService(&testServiceImpl{}, newTestDesc()),
			WithServerOption(grpc.MaxRecvMsgSize(1024)),
		)
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})

	t.Run("stop does not panic", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("0.0.0.0", "9090", WithService(&testServiceImpl{}, newTestDesc()))
		srv.Stop()
	})

	t.Run("graceful stop does not panic", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("0.0.0.0", "9091", WithService(&testServiceImpl{}, newTestDesc()))
		srv.GracefulStop()
	})

	t.Run("serve accepts listener", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("127.0.0.1", "0", WithService(&testServiceImpl{}, newTestDesc()))

		var lc net.ListenConfig

		lis, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}

		go func() {
			_ = srv.Serve(lis)
		}()

		srv.GracefulStop()
	})

	t.Run("registers multiple services", func(t *testing.T) {
		t.Parallel()

		desc1 := &grpc.ServiceDesc{
			ServiceName: "test.Service1",
			HandlerType: (*testService)(nil),
		}

		desc2 := &grpc.ServiceDesc{
			ServiceName: "test.Service2",
			HandlerType: (*testService)(nil),
		}

		srv := NewServer("0.0.0.0", "9092",
			WithService(&testServiceImpl{}, desc1),
			WithService(&testServiceImpl{}, desc2),
		)
		if srv == nil {
			t.Fatal("expected non-nil server")
		}

		srv.Stop()
	})

	t.Run("works with no options", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("0.0.0.0", "9093")
		if srv == nil {
			t.Fatal("expected non-nil server")
		}

		srv.Stop()
	})
}

func TestRegisterService(t *testing.T) {
	t.Parallel()

	t.Run("registers service after construction", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("0.0.0.0", "9094")

		srv.RegisterService(newTestDesc(), &testServiceImpl{})

		srv.Stop()
	})
}

func TestAddress(t *testing.T) {
	t.Parallel()

	t.Run("returns host port combination", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("127.0.0.1", "8080", WithService(&testServiceImpl{}, newTestDesc()))

		expected := "127.0.0.1:8080"
		if srv.Address() != expected {
			t.Fatalf("expected %q, got %q", expected, srv.Address())
		}
	})

	t.Run("returns ipv6 address", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("::1", "443", WithService(&testServiceImpl{}, newTestDesc()))

		expected := net.JoinHostPort("::1", "443")
		if srv.Address() != expected {
			t.Fatalf("expected %q, got %q", expected, srv.Address())
		}
	})
}
