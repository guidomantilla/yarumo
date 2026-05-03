package http

import (
	"context"
	"net"
	"net/http"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

type server struct {
	inner *http.Server
}

// NewServer creates a new HTTP Server with secure timeout defaults.
func NewServer(host string, port string, handler http.Handler, options ...Option) Server {
	cassert.NotEmpty(host, "host is empty")
	cassert.NotEmpty(port, "port is empty")
	cassert.NotNil(handler, "handler is nil")

	opts := NewOptions(options...)

	return &server{
		inner: &http.Server{
			Addr:              net.JoinHostPort(host, port),
			Handler:           handler,
			ReadHeaderTimeout: opts.serverReadHeaderTimeout,
			ReadTimeout:       opts.serverReadTimeout,
			WriteTimeout:      opts.serverWriteTimeout,
			IdleTimeout:       opts.serverIdleTimeout,
			MaxHeaderBytes:    opts.serverMaxHeaderBytes,
			TLSConfig:         opts.serverTLSConfig,
		},
	}
}

// Address returns the network address the server is configured to listen on.
func (s *server) Address() string {
	cassert.NotNil(s, "server is nil")

	return s.inner.Addr
}

// ListenAndServe starts the server on the configured address.
func (s *server) ListenAndServe() error {
	cassert.NotNil(s, "server is nil")

	return s.inner.ListenAndServe()
}

// ListenAndServeTLS starts the server with TLS using the provided certificate and key files.
func (s *server) ListenAndServeTLS(certFile string, keyFile string) error {
	cassert.NotNil(s, "server is nil")

	return s.inner.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the server without interrupting active connections.
func (s *server) Shutdown(ctx context.Context) error {
	cassert.NotNil(s, "server is nil")

	return s.inner.Shutdown(ctx)
}

// Close immediately closes the server.
func (s *server) Close() error {
	cassert.NotNil(s, "server is nil")

	return s.inner.Close()
}
