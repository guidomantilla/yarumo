package managed

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func BuildBaseServer(ctx context.Context, name string, errChan ErrChan) (Component[BaseDaemon], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	base := Component[BaseDaemon]{name: name, internal: NewBaseDaemon()}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := base.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := base.internal.Start()
		if err != nil {
			select {
			case errChan <- err:
			default:
			}
			return
		}

		<-base.internal.Done()
	}()

	return base, stopFn, nil
}
