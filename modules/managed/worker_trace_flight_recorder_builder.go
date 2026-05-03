package managed

import (
	"context"
	"time"

	cdiagnostics "github.com/guidomantilla/yarumo/common/diagnostics"
	clog "github.com/guidomantilla/yarumo/common/log"
)

// BuildTraceFlightRecorderWorker creates a managed trace flight recorder worker component, starts it in a background goroutine, and returns a stop function.
func BuildTraceFlightRecorderWorker(ctx context.Context, name string, fr cdiagnostics.TraceFlightRecorder, errChan ErrChan) (Component[TraceFlightRecorderWorker], StopFn, error) {
	clog.Info(ctx, "starting up", "stage", "startup", "component", name)

	traceWorker := Component[TraceFlightRecorderWorker]{name: name, internal: NewTraceFlightRecorderWorker(fr)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		clog.Info(ctx, "stopping", "stage", "shutdown", "component", name)
		defer clog.Info(ctx, "stopped", "stage", "shutdown", "component", name)

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := traceWorker.internal.Stop(timeoutCtx)
		if err != nil {
			clog.Error(ctx, "shutdown failed", "stage", "shutdown", "component", name, "error", err)
		}
	}

	go func() {
		err := traceWorker.internal.Start(ctx)
		if err != nil {
			clog.Error(ctx, "failed to start", "stage", "startup", "component", name, "error", err)
			select {
			case errChan <- err:
			default:
			}
			return
		}
		<-traceWorker.internal.Done()
	}()

	return traceWorker, stopFn, nil
}
