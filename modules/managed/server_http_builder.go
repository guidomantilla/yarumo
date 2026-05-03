package managed

import (
	"context"
	"time"

	chttp "github.com/guidomantilla/yarumo/common/http"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildHttpServer creates a managed HTTP server component, starts it in a background goroutine, and returns a stop function.
func BuildHttpServer(ctx context.Context, name string, internal chttp.Server, errChan ErrChan) (Component[HttpServer], StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	httpServer := Component[HttpServer]{name: name, internal: NewHttpServer(internal)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := httpServer.internal.Stop(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}
	}

	go func() {
		err := httpServer.internal.ListenAndServe(ctx)
		if err != nil {
			clog.Error(ctx, "failed to listen or serve", "stage", "startup", "component", name, "error", err)
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	return httpServer, stopFn, nil
}
