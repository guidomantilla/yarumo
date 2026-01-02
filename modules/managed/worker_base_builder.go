package managed

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func BuildBaseWorker(ctx context.Context, name string, _ any, errChan ErrChan) (Component[BaseWorker], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	baseWorker := Component[BaseWorker]{name: name, internal: NewBaseWorker()}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := baseWorker.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := baseWorker.internal.Start(ctx)
		if err != nil {
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
