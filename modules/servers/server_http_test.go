package servers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestBuildHttpServer(t *testing.T) {
	s := &http.Server{Addr: "127.0.0.1:0"} //nolint:gosec
	name, srv := BuildHttpServer(s)
	if name != "http-server" {
		t.Fatalf("unexpected name: %s", name)
	}
	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestNewHttpServer_AllowsNilForCoverage(t *testing.T) {
	// Exercise assert.NotNil branch (it only logs)
	_ = NewHttpServer(nil)
}

func TestHttpServer_RunAndStop_Success(t *testing.T) {
	s := &http.Server{Addr: "127.0.0.1:0"} //nolint:gosec
	hs := NewHttpServer(s).(*httpServer)

	ctx := context.Background()
	done := make(chan error, 1)
	go func() { done <- hs.Run(ctx) }()

	// Wait a bit for server to start listening
	time.Sleep(100 * time.Millisecond)

	// Stop will cause Shutdown -> ListenAndServe returns http.ErrServerClosed
	if err := hs.Stop(ctx); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not finish in time")
	}
}

func TestHttpServer_Run_ErrorOnStart(t *testing.T) {
	// Invalid port should make ListenAndServe fail immediately
	s := &http.Server{Addr: "127.0.0.1:-1"} //nolint:gosec
	hs := NewHttpServer(s)
	err := hs.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if se, ok := err.(*ServerError); !ok || se.Type != ServerStartType {
		t.Fatalf("expected ServerError type=start, got %#v", err)
	}
}

func TestHttpServer_Stop_WrapsError(t *testing.T) {
	// Start a server, trigger a long-running handler, then shutdown with short timeout
	// to force context.DeadlineExceeded from Shutdown.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()
	chosen := fmt.Sprintf("127.0.0.1:%d", addr.Port)

	entered := make(chan struct{})
	release := make(chan struct{})
	s := &http.Server{Addr: chosen, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:gosec
		// Signal we've entered the handler, then block until released
		close(entered)
		<-release
		w.WriteHeader(http.StatusOK)
	})}
	hs := NewHttpServer(s).(*httpServer)

	done := make(chan error, 1)
	go func() { done <- hs.Run(context.Background()) }()
	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Fire a request and wait until the handler is running
	clientDone := make(chan struct{})
	go func() {
		_, _ = http.Get("http://" + chosen)
		close(clientDone)
	}()
	<-entered

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	stopErr := hs.Stop(ctx)
	if stopErr == nil {
		t.Fatal("expected error from Stop with short timeout")
	}
	if se, ok := stopErr.(*ServerError); !ok || se.Type != ServerStopType {
		t.Fatalf("expected ServerError type=stop, got %#v", stopErr)
	}

	// Ensure the server is properly stopped for test cleanup
	// Release the handler in case Shutdown returned early but handler is still blocked
	close(release)
	_ = hs.Stop(context.Background())
	<-clientDone
	<-done
}
