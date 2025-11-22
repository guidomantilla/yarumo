package servers

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
)

type fakeGrpc struct {
	serveErr error
	stopped  bool
}

func (f *fakeGrpc) Serve(l net.Listener) error { return f.serveErr }

func (f *fakeGrpc) GracefulStop() { f.stopped = true }

func TestBuildGrpcServer(t *testing.T) {
	f := &fakeGrpc{}
	name, srv := BuildGrpcServer("127.0.0.1:0", f)
	if name != "grpc-server" || srv == nil {
		t.Fatalf("unexpected: name=%s srv=%v", name, srv)
	}
}

func TestNewGrpcServer_AssertionsCoverage(t *testing.T) {
	// Exercise asserts for empty address and nil server (they only log)
	_ = NewGrpcServer("", nil)
}

func TestGrpcServer_Run_ServeClosedOK(t *testing.T) {
	f := &fakeGrpc{serveErr: http.ErrServerClosed}
	gs := NewGrpcServer("127.0.0.1:0", f).(*grpcServer)

	// Run should treat ErrServerClosed as success
	if err := gs.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGrpcServer_Run_ListenError(t *testing.T) {
	f := &fakeGrpc{}
	gs := NewGrpcServer("127.0.0.1:-1", f)
	if err := gs.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	} else if se, ok := err.(*Error); !ok || se.Type != ServerStartType {
		t.Fatalf("expected ServerError type=start, got %#v", err)
	}
}

func TestGrpcServer_Run_ServeError(t *testing.T) {
	f := &fakeGrpc{serveErr: errors.New("boom")}
	gs := NewGrpcServer("127.0.0.1:0", f)
	if err := gs.Run(context.Background()); err == nil {
		t.Fatal("expected error")
	} else if se, ok := err.(*Error); !ok || se.Type != ServerStartType {
		t.Fatalf("expected ServerError type=start, got %#v", err)
	}
}

func TestGrpcServer_Stop(t *testing.T) {
	f := &fakeGrpc{}
	gs := NewGrpcServer("127.0.0.1:0", f).(*grpcServer)
	if err := gs.Stop(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.stopped {
		t.Fatal("expected GracefulStop to be called")
	}
}
