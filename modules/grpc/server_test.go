package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/guidomantilla/yarumo/common/lifecycle"
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

	t.Run("returns non-nil server", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("srv-1", "tcp", "127.0.0.1", "0",
			WithService(&testServiceImpl{}, newTestDesc()))
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})

	t.Run("carries the given name", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("srv-named", "tcp", "127.0.0.1", "0")
		if srv.Name() != "srv-named" {
			t.Fatalf("expected name %q, got %q", "srv-named", srv.Name())
		}
	})

	t.Run("done channel is open at construction", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("srv-open-done", "tcp", "127.0.0.1", "0")

		select {
		case <-srv.Done():
			t.Fatal("expected Done channel to be open before Start/Stop")
		default:
		}
	})

	t.Run("accepts server options", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("srv-opts", "tcp", "127.0.0.1", "0",
			WithService(&testServiceImpl{}, newTestDesc()),
			WithServerOption(grpc.MaxRecvMsgSize(1024)),
		)
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
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

		srv := NewServer("srv-multi", "tcp", "127.0.0.1", "0",
			WithService(&testServiceImpl{}, desc1),
			WithService(&testServiceImpl{}, desc2),
		)
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})

	t.Run("works with no options", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("srv-bare", "tcp", "127.0.0.1", "0")
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})
}

func TestServer_Name(t *testing.T) {
	t.Parallel()

	srv := NewServer("worker-name", "tcp", "127.0.0.1", "0")
	if srv.Name() != "worker-name" {
		t.Fatalf("expected name %q, got %q", "worker-name", srv.Name())
	}
}

func TestServer_Stop(t *testing.T) {
	t.Parallel()

	t.Run("closes the Done channel on first call", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("stop-1", "tcp", "127.0.0.1", "0")

		err := srv.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		select {
		case <-srv.Done():
		default:
			t.Fatal("expected Done channel closed after Stop")
		}
	})

	t.Run("is idempotent across multiple calls", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("stop-2", "tcp", "127.0.0.1", "0")

		err := srv.Stop(context.Background())
		if err != nil {
			t.Fatalf("first Stop returned %v", err)
		}

		err = srv.Stop(context.Background())
		if err != nil {
			t.Fatalf("second Stop returned %v", err)
		}
	})
}

func TestServer_Done(t *testing.T) {
	t.Parallel()

	t.Run("unblocks readers after Stop", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("done-1", "tcp", "127.0.0.1", "0")

		ready := make(chan struct{})
		done := make(chan struct{})

		go func() {
			close(ready)
			<-srv.Done()
			close(done)
		}()

		<-ready

		err := srv.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("expected reader to unblock after Stop")
		}
	})
}

func TestServer_Start(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrStart on invalid network", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("start-bad-net", "unknown-network", "127.0.0.1", "0")

		err := srv.Start(context.Background())
		if err == nil {
			t.Fatal("expected error from invalid network")
		}

		if !errors.Is(err, lifecycle.ErrStartFailed) {
			t.Fatalf("expected ErrStartFailed, got %v", err)
		}

		select {
		case <-srv.Done():
		default:
			t.Fatal("expected Done channel closed after Start failure")
		}
	})

	t.Run("blocking Start returns nil after Stop", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("start-block", "tcp", "127.0.0.1", "0")

		startErr := make(chan error, 1)

		go func() {
			startErr <- srv.Start(context.Background())
		}()

		// Give Start a moment to bind the listener and enter Serve.
		time.Sleep(50 * time.Millisecond)

		err := srv.Stop(context.Background())
		if err != nil {
			t.Fatalf("Stop returned %v", err)
		}

		select {
		case err := <-startErr:
			if err != nil {
				t.Fatalf("expected Start to return nil, got %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Start did not return after Stop")
		}

		select {
		case <-srv.Done():
		default:
			t.Fatal("expected Done channel closed after Start returned")
		}
	})
}

func TestRegisterService(t *testing.T) {
	t.Parallel()

	t.Run("registers service after construction", func(t *testing.T) {
		t.Parallel()

		srv := NewServer("register-1", "tcp", "127.0.0.1", "0")
		srv.RegisterService(newTestDesc(), &testServiceImpl{})
	})
}
