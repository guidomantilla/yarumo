package http

import (
	"context"
	"errors"
	"net"
	nethttp "net/http"
	"sync"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/lifecycle"
)

// server implements Server. It wraps a *net/http.Server and exposes the
// configured listen address. Start is blocking (server-style lifecycle):
// it opens the listener and calls Serve (or ServeTLS when configured),
// returning only when the server has been shut down or the listener
// fails. Done closes when Start returns.
type server struct {
	*nethttp.Server

	name     string
	network  string
	addr     string
	listener net.Listener
	mutex    sync.Mutex

	tls struct {
		enabled  bool
		certFile string
		keyFile  string
	}

	done chan struct{}
	once sync.Once
}

// NewServer creates a new HTTP Server with the given name, network (e.g.,
// "tcp"), host, port, and handler, applying the provided options.
func NewServer(name string, network string, host string, port string, handler nethttp.Handler, options ...Option) Server {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotEmpty(network, "network is empty")
	cassert.NotEmpty(host, "host is empty")
	cassert.NotEmpty(port, "port is empty")
	cassert.NotNil(handler, "handler is nil")

	opts := NewOptions(options...)

	internal := &nethttp.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           handler,
		ReadTimeout:       opts.readTimeout,
		WriteTimeout:      opts.writeTimeout,
		IdleTimeout:       opts.idleTimeout,
		ReadHeaderTimeout: opts.readHeaderTimeout,
		MaxHeaderBytes:    opts.maxHeaderBytes,
		ErrorLog:          opts.errorLog,
		BaseContext:       opts.baseContext,
	}

	s := &server{
		Server:  internal,
		name:    name,
		network: network,
		addr:    internal.Addr,
		done:    make(chan struct{}),
	}

	s.tls.enabled = opts.tlsEnabled
	s.tls.certFile = opts.tlsCertFile
	s.tls.keyFile = opts.tlsKeyFile

	return s
}

// Name returns the server's identity used in logs.
func (s *server) Name() string {
	cassert.NotNil(s, "server is nil")

	return s.name
}

// Start opens the configured listener and serves HTTP requests on it. It
// blocks until the server is shut down (via Stop) or the listener fails.
// When configured with WithTLS, ServeTLS is used instead of Serve. Done
// is closed when Start returns, regardless of the exit path (success,
// listen error, or serve error).
func (s *server) Start(ctx context.Context) error {
	cassert.NotNil(s, "server is nil")

	defer s.once.Do(func() { close(s.done) })

	lc := net.ListenConfig{}

	listener, err := lc.Listen(ctx, s.network, s.addr)
	if err != nil {
		return lifecycle.ErrStart(err)
	}

	s.mutex.Lock()
	s.listener = listener
	s.mutex.Unlock()

	if s.tls.enabled {
		err = s.Server.ServeTLS(listener, s.tls.certFile, s.tls.keyFile)
	} else {
		err = s.Server.Serve(listener)
	}

	if err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		return lifecycle.ErrStart(err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server bounded by ctx's deadline.
// http.Server.Shutdown is ctx-aware: if ctx expires before active
// connections drain, this falls back to closing the listener and forcing
// Close, returning ErrShutdown wrapping ErrShutdownTimeout. Done is
// closed by the deferred once.Do here as a safety net for callers that
// invoke Stop without Start (the normal close path is Start's defer).
func (s *server) Stop(ctx context.Context) error {
	cassert.NotNil(s, "server is nil")

	defer s.once.Do(func() { close(s.done) })

	err := s.Server.Shutdown(ctx)
	if err == nil {
		return nil
	}

	s.mutex.Lock()
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.mutex.Unlock()

	_ = s.Server.Close()

	return lifecycle.ErrShutdown(lifecycle.ErrShutdownTimeout, err)
}

// Done returns the channel that is closed when Start has returned (server
// shut down) or when Stop has been invoked, whichever comes first.
func (s *server) Done() <-chan struct{} {
	cassert.NotNil(s, "server is nil")

	return s.done
}
