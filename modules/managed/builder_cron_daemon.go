package managed

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

func BuildCronServer(ctx context.Context, name string, internal *cron.Cron, errChan ErrChan) (Component[CronDaemon], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	cronServer := Component[CronDaemon]{name: name, internal: NewCronDaemon(internal)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := cronServer.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := cronServer.internal.Start()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "startup").Str("component", name).Msg("failed to start")
			select {
			case errChan <- err:
			default:
			}
			return
		}
		<-cronServer.internal.Done()
	}()

	return cronServer, stopFn, nil
}
