package managed

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type httpAdapter struct {
	h *http.Server
}

func NewHttpServer(h *http.Server) HttpServer {
	return &httpAdapter{
		h: h,
	}
}

func (h *httpAdapter) ListenAndServe() error {
	err := h.h.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen or serve error: %w", err)
	}

	return nil
}

func (h *httpAdapter) ListenAndServeTLS(certFile string, keyFile string) error {
	err := h.h.ListenAndServeTLS(certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen or serve (tls) error: %w", err)
	}
	return nil
}

func (h *httpAdapter) Stop(ctx context.Context) error {
	errShutdown := h.h.Shutdown(ctx)
	if errShutdown != nil {
		errClose := h.h.Close()
		if errClose != nil {
			return fmt.Errorf("shutdown error: %w, close error: %w", errShutdown, errClose)
		}

		return fmt.Errorf("shutdown error: %w", errShutdown)
	}
	return nil
}
