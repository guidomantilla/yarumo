package http

import (
	"context"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
)

func TestBuildServer(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil server and closeFn", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		srv, closeFn, err := BuildServer(context.Background(), "build-1", "tcp", "127.0.0.1", "0", noopHandler(), errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if srv == nil {
			t.Fatal("expected non-nil server")
		}

		if closeFn == nil {
			t.Fatal("expected non-nil closeFn")
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("server carries the given name", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		srv, closeFn, err := BuildServer(context.Background(), "build-named", "tcp", "127.0.0.1", "0", noopHandler(), errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		if srv.Name() != "build-named" {
			t.Fatalf("expected name %q, got %q", "build-named", srv.Name())
		}
	})

	t.Run("returned closeFn drains the background goroutine before returning", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		srv, closeFn, err := BuildServer(context.Background(), "build-drain", "tcp", "127.0.0.1", "0", noopHandler(), errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		// Give Start a moment to enter Serve.
		time.Sleep(50 * time.Millisecond)

		closeFn(context.Background(), time.Second)

		select {
		case <-srv.Done():
		default:
			t.Fatal("expected server Done closed after closeFn returned")
		}
	})

	t.Run("closeFn is safe to call from defer with the same ctx", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)
		ctx := context.Background()

		_, closeFn, err := BuildServer(ctx, "build-defer", "tcp", "127.0.0.1", "0", noopHandler(), errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(ctx, time.Second)
	})

	t.Run("matches the BuildServerFn signature", func(t *testing.T) {
		t.Parallel()

		var fn BuildServerFn = BuildServer

		errChan := make(chan error, 1)

		_, closeFn, err := fn(context.Background(), "build-fn", "tcp", "127.0.0.1", "0", noopHandler(), errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("errChan accepts startup errors without blocking", func(t *testing.T) {
		t.Parallel()

		// Unbuffered channel: a non-blocking send by lifecycle.Start
		// should fall through the default arm. The build itself must
		// still succeed.
		errChan := make(chan error)

		_, closeFn, err := BuildServer(context.Background(), "build-errchan", "tcp", "127.0.0.1", "0", noopHandler(), lifecycle.ErrChan(errChan))
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		closeFn(context.Background(), time.Second)
	})

	t.Run("options are propagated to the underlying *http.Server", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)

		srv, closeFn, err := BuildServer(context.Background(), "build-opts", "tcp", "127.0.0.1", "0", noopHandler(), errChan,
			WithReadTimeout(7*time.Second),
			WithMaxHeaderBytes(8192),
		)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), time.Second)

		impl, ok := srv.(*server)
		if !ok {
			t.Fatalf("expected *server, got %T", srv)
		}

		if impl.ReadTimeout != 7*time.Second {
			t.Fatalf("expected ReadTimeout=7s, got %v", impl.ReadTimeout)
		}

		if impl.MaxHeaderBytes != 8192 {
			t.Fatalf("expected MaxHeaderBytes=8192, got %d", impl.MaxHeaderBytes)
		}
	})

	t.Run("server actually serves an HTTP roundtrip", func(t *testing.T) {
		t.Parallel()

		handler := nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
			w.WriteHeader(nethttp.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		errChan := make(chan error, 1)

		srv, closeFn, err := BuildServer(context.Background(), "build-roundtrip", "tcp", "127.0.0.1", "0", handler, errChan)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		defer closeFn(context.Background(), 2*time.Second)

		// Wait for the listener to be bound; the impl exposes it via
		// the concrete struct.
		impl, ok := srv.(*server)
		if !ok {
			t.Fatalf("expected *server, got %T", srv)
		}

		deadline := time.Now().Add(2 * time.Second)
		var addr string
		for {
			impl.mutex.Lock()
			l := impl.listener
			impl.mutex.Unlock()

			if l != nil {
				addr = l.Addr().String()
				break
			}

			if time.Now().After(deadline) {
				t.Fatal("listener never bound")
			}

			time.Sleep(10 * time.Millisecond)
		}

		client := &nethttp.Client{Timeout: time.Second}

		resp, err := client.Get("http://" + addr + "/")
		if err != nil {
			t.Fatalf("GET failed: %v", err)
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
	})
}
