package managed

import (
	"context"
	"time"

	ccron "github.com/guidomantilla/yarumo/common/cron"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildCronWorker creates a managed cron worker component, starts it in a background goroutine, and returns a stop function.
func BuildCronWorker(ctx context.Context, name string, internal ccron.Scheduler, errChan ErrChan) (Component[CronWorker], StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	cronWorker := Component[CronWorker]{name: name, internal: NewCronWorker(internal)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := cronWorker.internal.Stop(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}
	}

	go func() {
		err := cronWorker.internal.Start(ctx)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
			select {
			case errChan <- err:
			default:
			}
			return
		}
		<-cronWorker.internal.Done()
	}()

	return cronWorker, stopFn, nil
}
