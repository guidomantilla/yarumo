package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	stdhttp "net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	t.Run("returns server with correct address", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("localhost", "8080", handler)

		expected := net.JoinHostPort("localhost", "8080")
		if srv.Address() != expected {
			t.Fatalf("expected address %q, got %q", expected, srv.Address())
		}
	})

	t.Run("applies default options", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("0.0.0.0", "9090", handler)

		s := srv.(*server)

		if s.inner.ReadHeaderTimeout != 5*time.Second {
			t.Fatalf("ReadHeaderTimeout = %v, want 5s", s.inner.ReadHeaderTimeout)
		}

		if s.inner.ReadTimeout != 15*time.Second {
			t.Fatalf("ReadTimeout = %v, want 15s", s.inner.ReadTimeout)
		}

		if s.inner.WriteTimeout != 15*time.Second {
			t.Fatalf("WriteTimeout = %v, want 15s", s.inner.WriteTimeout)
		}

		if s.inner.IdleTimeout != 60*time.Second {
			t.Fatalf("IdleTimeout = %v, want 60s", s.inner.IdleTimeout)
		}

		if s.inner.MaxHeaderBytes != 1<<20 {
			t.Fatalf("MaxHeaderBytes = %d, want %d", s.inner.MaxHeaderBytes, 1<<20)
		}

		if s.inner.TLSConfig != nil {
			t.Fatalf("TLSConfig = %v, want nil", s.inner.TLSConfig)
		}
	})

	t.Run("applies custom options", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS13}

		srv := NewServer("0.0.0.0", "9091", handler,
			WithServerReadHeaderTimeout(10*time.Second),
			WithServerReadTimeout(30*time.Second),
			WithServerWriteTimeout(45*time.Second),
			WithServerIdleTimeout(120*time.Second),
			WithServerMaxHeaderBytes(2<<20),
			WithServerTLSConfig(tlsCfg),
		)

		s := srv.(*server)

		if s.inner.ReadHeaderTimeout != 10*time.Second {
			t.Fatalf("ReadHeaderTimeout = %v, want 10s", s.inner.ReadHeaderTimeout)
		}

		if s.inner.ReadTimeout != 30*time.Second {
			t.Fatalf("ReadTimeout = %v, want 30s", s.inner.ReadTimeout)
		}

		if s.inner.WriteTimeout != 45*time.Second {
			t.Fatalf("WriteTimeout = %v, want 45s", s.inner.WriteTimeout)
		}

		if s.inner.IdleTimeout != 120*time.Second {
			t.Fatalf("IdleTimeout = %v, want 120s", s.inner.IdleTimeout)
		}

		if s.inner.MaxHeaderBytes != 2<<20 {
			t.Fatalf("MaxHeaderBytes = %d, want %d", s.inner.MaxHeaderBytes, 2<<20)
		}

		if s.inner.TLSConfig != tlsCfg {
			t.Fatalf("TLSConfig not set correctly")
		}
	})

	t.Run("sets handler", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("0.0.0.0", "9092", handler)

		s := srv.(*server)
		if s.inner.Handler != handler {
			t.Fatalf("handler not set correctly")
		}
	})
}

func TestServer_ListenAndServe(t *testing.T) {
	t.Parallel()

	t.Run("starts and can be closed", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("127.0.0.1", "0", handler)

		errCh := make(chan error, 1)

		go func() {
			errCh <- srv.ListenAndServe()
		}()

		// Give the server a moment to start, then close it.
		time.Sleep(50 * time.Millisecond)

		err := srv.Close()
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}

		listenErr := <-errCh
		if listenErr != nil && !errors.Is(listenErr, stdhttp.ErrServerClosed) {
			t.Fatalf("ListenAndServe returned unexpected error: %v", listenErr)
		}
	})
}

func TestServer_ListenAndServeTLS(t *testing.T) {
	t.Parallel()

	t.Run("invalid cert returns error", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("127.0.0.1", "0", handler)

		err := srv.ListenAndServeTLS("nonexistent.crt", "nonexistent.key")
		if err == nil {
			t.Fatal("expected error for invalid cert files")
		}
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("graceful shutdown", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("127.0.0.1", "0", handler)

		errCh := make(chan error, 1)

		go func() {
			errCh <- srv.ListenAndServe()
		}()

		time.Sleep(50 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			t.Fatalf("Shutdown returned error: %v", err)
		}

		listenErr := <-errCh
		if listenErr != nil && !errors.Is(listenErr, stdhttp.ErrServerClosed) {
			t.Fatalf("ListenAndServe returned unexpected error: %v", listenErr)
		}
	})
}

func TestServerAddress(t *testing.T) {
	t.Parallel()

	t.Run("returns host port combination", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("127.0.0.1", "8080", handler)

		expected := "127.0.0.1:8080"
		if srv.Address() != expected {
			t.Fatalf("expected %q, got %q", expected, srv.Address())
		}
	})

	t.Run("returns ipv6 address", func(t *testing.T) {
		t.Parallel()

		handler := stdhttp.NewServeMux()
		srv := NewServer("::1", "443", handler)

		expected := net.JoinHostPort("::1", "443")
		if srv.Address() != expected {
			t.Fatalf("expected %q, got %q", expected, srv.Address())
		}
	})
}
