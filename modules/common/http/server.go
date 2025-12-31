package http

import (
	"context"
	"net"
	"net/http"
)

func NewServer(ctx context.Context, host string, port string, handler http.Handler, options ...Option) Server {
	opts := NewOptions(options...)
	return &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:              net.JoinHostPort(host, port),
		Handler:           handler,
		ReadHeaderTimeout: opts.serverReadHeaderTimeout,
		ReadTimeout:       opts.serverReadTimeout,
		WriteTimeout:      opts.serverWriteTimeout,
		IdleTimeout:       opts.serverIdleTimeout,
		MaxHeaderBytes:    opts.serverMaxHeaderBytes,
		TLSConfig:         opts.serverTLSConfig,
	}
}
