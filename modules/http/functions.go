package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/guidomantilla/yarumo/common/lifecycle"
	clog "github.com/guidomantilla/yarumo/log"
)

// BuildServer creates a managed HTTP Server, starts it in a background
// goroutine, and returns a CloseFn for graceful shutdown.
//
// Startup errors are logged and forwarded to errChan (non-blocking). The
// returned Server implements lifecycle.Component for further inspection
// (Name, Done) but typically the caller just needs the CloseFn. The
// CloseFn must be invoked by the caller to release resources; it bounds
// the shutdown by the given timeout and blocks until the background
// goroutine has exited.
func BuildServer(ctx context.Context, name string, network string, host string, port string, handler nethttp.Handler, errChan lifecycle.ErrChan, options ...Option) (Server, lifecycle.CloseFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	component := NewServer(name, network, host, port, handler, options...)

	spawned := make(chan struct{})

	closeFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		err := lifecycle.Stop(ctx, component, timeout)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}

		<-spawned
	}

	go func() {
		defer close(spawned)

		err := lifecycle.Start(ctx, component, errChan)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
		}
	}()

	return component, closeFn, nil
}
