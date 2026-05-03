package managed

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestBuildHttpServer(t *testing.T) {
	t.Run("build succeeds and stop completes", func(t *testing.T) {
		errCh := make(chan error, 1)

		srv := &mockHTTPServer{
			listenAndServeErr: http.ErrServerClosed,
		}

		component, stopFn, err := BuildHttpServer(context.Background(), "test-http", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if component.name != "test-http" {
			t.Fatalf("expected name test-http, got %s", component.name)
		}

		if stopFn == nil {
			t.Fatal("expected non-nil stopFn")
		}

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

		srv := &mockHTTPServer{
			listenAndServeErr: errors.New("listen failed"),
		}

		_, _, err := BuildHttpServer(context.Background(), "test-http-fail", srv, errCh)
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

		srv := &mockHTTPServer{
			listenAndServeErr: http.ErrServerClosed,
			shutdownErr:       errors.New("shutdown error"),
		}

		_, stopFn, err := BuildHttpServer(context.Background(), "test-http-timeout", srv, errCh)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		time.Sleep(50 * time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		stopFn(ctx, time.Nanosecond)
	})
}
