package managed

import (
	"context"
	"time"

	commongrpc "github.com/guidomantilla/yarumo/common/grpc"

	"github.com/rs/zerolog/log"
)

func BuildGrpcServer(ctx context.Context, name string, internal commongrpc.Server, errChan ErrChan) (Component[GrpcServer], StopFn, error) {
	log.Ctx(ctx).Info().Str("stage", "startup").Str("component", name).Msg("starting up")

	grpcServer := Component[GrpcServer]{name: name, internal: NewGrpcServer(internal, "tcp")}

	stopFn := func(ctx context.Context, timeout time.Duration) {
		log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopping")
		defer log.Ctx(ctx).Info().Str("stage", "shut down").Str("component", name).Msg("stopped")

		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := grpcServer.internal.Stop(timeoutCtx)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "shut down").Str("component", name).Msg("shutdown failed")
		}
	}

	go func() {
		err := grpcServer.internal.ListenAndServe()
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Str("stage", "startup").Str("component", name).Msg("failed to listen or serve")
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	return grpcServer, stopFn, nil
}
