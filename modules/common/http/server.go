package http

import (
	"net"
	"net/http"
)

type server struct {
	*http.Server
}

func NewServer(host string, port string, handler http.Handler, options ...Option) Server {
	opts := NewOptions(options...)

	return &server{
		Server: &http.Server{
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

func (s *server) Address() string {
	return s.Addr
}
