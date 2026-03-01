package managed

import (
	"context"
	"errors"
	"net/http"

	chttp "github.com/guidomantilla/yarumo/common/http"
)

type httpAdapter struct {
	h chttp.Server
}

// NewHttpServer creates a new managed HTTP server wrapping the given server.
func NewHttpServer(h chttp.Server) HttpServer {
	return &httpAdapter{
		h: h,
	}
}

func (h *httpAdapter) ListenAndServe(_ context.Context) error {
	err := h.h.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return ErrServe(err)
	}

	return nil
}

func (h *httpAdapter) ListenAndServeTLS(_ context.Context, certFile string, keyFile string) error {
	err := h.h.ListenAndServeTLS(certFile, keyFile)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return ErrServe(err)
	}
	return nil
}

func (h *httpAdapter) Stop(ctx context.Context) error {
	errShutdown := h.h.Shutdown(ctx)
	if errShutdown != nil {
		errClose := h.h.Close()
		if errClose != nil {
			return ErrShutdown(errShutdown, errClose)
		}

		return ErrShutdown(errShutdown)
	}
	return nil
}
