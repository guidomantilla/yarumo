package managed

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func BuildHttpServer(ctx context.Context, name string, internal *http.Server, errChan ErrChan) (Component[HttpServer], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	httpServer := Component[HttpServer]{name: name, internal: NewHttpServer(internal)}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancelTimeoutFn := context.WithTimeout(ctx, timeout)
		defer cancelTimeoutFn()

		err := httpServer.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := httpServer.internal.ListenAndServe()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "runtime").Str("component", name).Msg("failed to listen or serve")
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	return httpServer, stopFn, nil
}
