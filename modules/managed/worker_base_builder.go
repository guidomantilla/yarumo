package managed

import (
	"context"
	"time"

	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildBaseWorker creates a managed base worker component, starts it in a background goroutine, and returns a stop function.
func BuildBaseWorker(ctx context.Context, name string, _ any, errChan ErrChan) (Component[BaseWorker], StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	baseWorker := Component[BaseWorker]{name: name, internal: NewBaseWorker()}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := baseWorker.internal.Stop(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}
	}

	go func() {
		err := baseWorker.internal.Start(ctx)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
			select {
			case errChan <- err:
			default:
			}
			return
		}

		<-baseWorker.internal.Done()
	}()

	return baseWorker, stopFn, nil
}
