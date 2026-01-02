package managed

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/common/diagnostics"
)

func BuildTraceFlightRecorderWorker(ctx context.Context, name string, fr diagnostics.TraceFlightRecorder, errChan ErrChan) (Component[TraceFlightRecorderWorker], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	traceWorker := Component[TraceFlightRecorderWorker]{name: name, internal: NewTraceFlightRecorderWorker(fr)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := traceWorker.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := traceWorker.internal.Start(ctx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "startup").Str("component", name).Msg("failed to start")
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
