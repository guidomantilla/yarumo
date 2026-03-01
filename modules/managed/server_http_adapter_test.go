package managed

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type mockHTTPServer struct {
	addressVal           string
	listenAndServeErr    error
	listenAndServeTLSErr error
	shutdownErr          error
	closeErr             error
}

func (m *mockHTTPServer) Address() string { return m.addressVal }

func (m *mockHTTPServer) ListenAndServe() error { return m.listenAndServeErr }

func (m *mockHTTPServer) ListenAndServeTLS(_ string, _ string) error { return m.listenAndServeTLSErr }

func (m *mockHTTPServer) Shutdown(_ context.Context) error { return m.shutdownErr }

func (m *mockHTTPServer) Close() error { return m.closeErr }

func TestNewHttpServer(t *testing.T) {
	t.Parallel()

	srv := &mockHTTPServer{}
	adapter := NewHttpServer(srv)
	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
}

func Test_httpAdapter_ListenAndServe(t *testing.T) {
	t.Parallel()

	t.Run("returns nil on success", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServe(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("returns nil on ErrServerClosed", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{listenAndServeErr: http.ErrServerClosed}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServe(context.Background())
		if err != nil {
			t.Fatalf("expected nil error for ErrServerClosed, got %v", err)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{listenAndServeErr: errors.New("listen failed")}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServe(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_httpAdapter_ListenAndServeTLS(t *testing.T) {
	t.Parallel()

	t.Run("returns nil on success", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServeTLS(context.Background(), "cert.pem", "key.pem")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("returns nil on ErrServerClosed", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{listenAndServeTLSErr: http.ErrServerClosed}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServeTLS(context.Background(), "cert.pem", "key.pem")
		if err != nil {
			t.Fatalf("expected nil error for ErrServerClosed, got %v", err)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{listenAndServeTLSErr: errors.New("tls failed")}
		adapter := NewHttpServer(srv)

		err := adapter.ListenAndServeTLS(context.Background(), "cert.pem", "key.pem")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func Test_httpAdapter_Stop(t *testing.T) {
	t.Parallel()

	t.Run("shutdown succeeds", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{}
		adapter := NewHttpServer(srv)

		err := adapter.Stop(context.Background())
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("shutdown fails and close succeeds", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{shutdownErr: errors.New("shutdown error")}
		adapter := NewHttpServer(srv)

		err := adapter.Stop(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("shutdown fails and close fails", func(t *testing.T) {
		t.Parallel()

		srv := &mockHTTPServer{
			shutdownErr: errors.New("shutdown error"),
			closeErr:    errors.New("close error"),
		}
		adapter := NewHttpServer(srv)

		err := adapter.Stop(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
